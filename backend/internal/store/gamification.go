package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"log"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"regexp"
	"sort"
	"strings"
	"time"
)

const defaultStudyTimezone = "Asia/Jakarta"

var studyLocation = time.FixedZone(defaultStudyTimezone, 7*60*60)

// Only server-graded attempts can create these events. This is not a public input.
type studyEvent struct {
	ID, Kind, Subject, LessonID, ExamID string
	Score, Answered, Seconds            int
	// Complete marks an SSW practice where every question was answered.
	Complete                 bool
	At                       time.Time
	Progress                 *domain.LessonProgress
	SSWAttempt               *SSWAttemptRecord
	OfflinePracticeAttemptID string
	OfflinePracticeRecord    *offlinePracticeAttemptRecord
}

// completionRecord is a finished activity that earned XP, kept briefly so
// its XP can be doubled (docs/ads-xp-spec.md, phase C).
type completionRecord struct {
	Kind       string
	Awarded    int
	RecordedAt time.Time // server time, which the claim window is measured from
	Boosted    bool
}

type gamificationState struct {
	Completions             map[string]completionRecord             `json:",omitempty"`
	LearningActivity        map[string]LearningActivity             `json:",omitempty"`
	OfflinePracticeAttempts map[string]offlinePracticeAttemptRecord `json:",omitempty"`
	SSWAttempts             map[string]SSWAttemptRecord
	Anki                    map[string]*AnkiDeckState
	XP                      int
	Events                  map[string]bool
	Rewards                 map[string]bool
	Lessons                 map[string]domain.LessonProgress
	Subjects                map[string]string
	Days                    map[string]int
	Unlocked                map[string]time.Time
	Exams                   map[string]bool
	PassedExam              bool
	Timezone                string
	// XPDays caps the daily XP sources (Anki, re-pass, daily goal).
	XPDays map[string]*xpDay `json:",omitempty"`
}
type GamificationSummary struct {
	XP               int                  `json:"xp"`
	Level            int                  `json:"level"`
	LevelXP          int                  `json:"levelXp"`
	NextLevelXP      int                  `json:"nextLevelXp"`
	StreakDays       int                  `json:"streakDays"`
	BestStreak       int                  `json:"bestStreak"`
	LessonsCompleted int                  `json:"lessonsCompleted"`
	ExamsCompleted   int                  `json:"examsCompleted"`
	ActiveDays       int                  `json:"activeDays"`
	MinutesToday     int                  `json:"minutesToday"`
	Timezone         string               `json:"timezone"`
	Achievements     []domain.Achievement `json:"achievements"`
}

func newGamificationState() *gamificationState {
	return &gamificationState{Events: map[string]bool{}, Rewards: map[string]bool{}, Lessons: map[string]domain.LessonProgress{}, Subjects: map[string]string{}, Days: map[string]int{}, Unlocked: map[string]time.Time{}, Exams: map[string]bool{}, OfflinePracticeAttempts: map[string]offlinePracticeAttemptRecord{}, Timezone: defaultStudyTimezone}
}

func isValidStudyTimezone(timezone string) bool {
	if timezone == "" || strings.TrimSpace(timezone) != timezone {
		return false
	}
	_, err := time.LoadLocation(timezone)
	return err == nil
}

func studyTimezone(state *gamificationState) *time.Location {
	if state != nil && isValidStudyTimezone(state.Timezone) {
		if location, err := time.LoadLocation(state.Timezone); err == nil {
			return location
		}
	}
	return studyLocation
}

// EnableGamificationPersistence connects to MariaDB before serving traffic.
// Cached account/progress state has one writer per database. The dedicated
// connection owns an advisory lock and performs every account/progress write;
// a lost connection fails writes instead of silently losing the lock.
func (s *MemoryStore) EnableGamificationPersistence(dsn string) error {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil || cfg.DBName == "" {
		return fmt.Errorf("a valid MariaDB DSN with a database name is required")
	}
	orm, err := database.Open(dsn)
	if err != nil {
		return fmt.Errorf("open MariaDB: %w", err)
	}
	db, err := orm.DB()
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(6)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(3 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	failed := true
	var writer *sql.Conn
	lockName := fmt.Sprintf("nihongo-master:%x", sha256.Sum256([]byte(cfg.DBName)))[:63]
	defer func() {
		if failed {
			if writer != nil {
				writer.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", lockName)
				writer.Close()
			}
			db.Close()
		}
	}()
	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to MariaDB: %w", err)
	}
	writer, err = db.Conn(ctx)
	if err != nil {
		return err
	}
	var acquired int
	if err = writer.QueryRowContext(ctx, "SELECT GET_LOCK(?,0)", lockName).Scan(&acquired); err != nil {
		return err
	}
	if acquired != 1 {
		return fmt.Errorf("another API process already owns this MariaDB database")
	}
	writerORM, err := database.Writer(writer)
	if err != nil {
		return err
	}
	schemaCtx, cancelSchema := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelSchema()
	if err := database.EnsureUUIDSchema(orm.WithContext(schemaCtx)); err != nil {
		return fmt.Errorf("migrate UUID schema: %w", err)
	}
	if err := database.Migrate(orm.WithContext(schemaCtx), &database.GamificationUser{}, &database.UserAnkiDeck{}, &database.UserAnkiDeckMedia{}); err != nil {
		return fmt.Errorf("initialize GORM gamification schema: %w", err)
	}
	var records []database.GamificationUser
	if err := orm.Find(&records).Error; err != nil {
		return err
	}
	loaded := map[string]*gamificationState{}
	for _, record := range records {
		state := newGamificationState()
		if err := json.Unmarshal(record.State, state); err != nil {
			return fmt.Errorf("invalid gamification record: %w", err)
		}
		if !isValidStudyTimezone(state.Timezone) {
			state.Timezone = defaultStudyTimezone
		}
		if state.XP < 0 || state.Events == nil || state.Rewards == nil || state.Lessons == nil || state.Subjects == nil || state.Days == nil || state.Unlocked == nil || state.Exams == nil {
			return fmt.Errorf("invalid gamification record for user %q", record.UserID)
		}
		for attemptID, r := range state.SSWAttempts {
			if r.Session.Attempt.UserID != record.UserID || r.Session.Attempt.ID != attemptID || len(r.Session.Sections) == 0 || r.Correct == nil || r.Marked == nil {
				return fmt.Errorf("invalid SSW attempt")
			}
		}
		loaded[record.UserID] = state
	}
	var persistedDecks []database.UserAnkiDeck
	if err := orm.Find(&persistedDecks).Error; err != nil {
		return err
	}
	var persistedMedia []userAnkiMediaMetadata
	if err := orm.Model(&database.UserAnkiDeckMedia{}).
		Select("user_id, deck_id, name, mime, OCTET_LENGTH(data) AS size").
		Scan(&persistedMedia).Error; err != nil {
		return err
	}
	mediaByDeck := make(map[string][]userAnkiMediaMetadata, len(persistedDecks))
	for _, item := range persistedMedia {
		mediaByDeck[item.DeckID] = append(mediaByDeck[item.DeckID], item)
	}
	loadedUserDecks := make(map[string]*userAnkiDeckRecord, len(persistedDecks))
	loadedUserDeckRequests := make(map[string]string, len(persistedDecks))
	for _, persisted := range persistedDecks {
		record, err := persistedUserAnkiDeck(persisted, mediaByDeck[persisted.ID])
		if err != nil {
			return fmt.Errorf("invalid imported Anki deck %q: %w", persisted.ID, err)
		}
		request, err := uuid.Parse(record.RequestID)
		if err != nil || request == uuid.Nil {
			return fmt.Errorf("invalid imported Anki deck request ID")
		}
		record.RequestID = request.String()
		requestKey := ankiUserDeckRequestKey(record.UserID, record.RequestID)
		if _, duplicateID := loadedUserDecks[record.ID]; duplicateID {
			return fmt.Errorf("duplicate imported Anki deck ID")
		}
		if _, duplicateRequest := loadedUserDeckRequests[requestKey]; duplicateRequest {
			return fmt.Errorf("duplicate imported Anki deck request ID")
		}
		loadedUserDecks[record.ID] = record
		loadedUserDeckRequests[requestKey] = record.ID
	}

	lb, err := s.loadLeaderboard(orm, writerORM, records, loaded)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gamificationDB != nil {
		return fmt.Errorf("gamification persistence already enabled")
	}
	s.lb = lb
	s.gamificationDB = db
	s.ormDB = orm
	s.ormWriter = writerORM
	s.persistenceConn = writer
	s.persistenceLockName = lockName
	s.gamification = loaded
	s.userAnkiDecks = loadedUserDecks
	s.userAnkiDeckRequests = loadedUserDeckRequests
	for id, state := range loaded {
		for attemptID, r := range state.SSWAttempts {
			s.simulationAttempts[attemptID] = &simulationAttemptState{r.Session, r.Questions, r.Correct, r.Explanations, r.Marked}
			s.simulationOrder = append(s.simulationOrder, attemptID)
		}
		s.progress[id] = map[string]domain.LessonProgress{}
		for key, progress := range state.Lessons {
			s.progress[id][key] = progress
		}
		profile := s.ensureProfileLocked(id)
		profile.XP = state.XP
		profile.Level = 1 + state.XP/500
		s.profiles[id] = profile
	}
	sort.SliceStable(s.simulationOrder, func(i, j int) bool {
		return s.simulationAttempts[s.simulationOrder[i]].session.Attempt.StartedAt.Before(s.simulationAttempts[s.simulationOrder[j]].session.Attempt.StartedAt)
	})
	failed = false
	s.persistenceStop = make(chan struct{})
	go keepMariaDBWriterAlive(writer, s.persistenceStop)
	return nil
}

func keepMariaDBWriterAlive(conn *sql.Conn, stop <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := conn.PingContext(ctx)
			cancel()
			if err != nil {
				select {
				case <-stop:
					return
				default:
				}
				log.Printf("MariaDB writer connection lost; restart the API: %v", err)
				return
			}
		}
	}
}

func (s *MemoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gamificationDB == nil || s.persistenceClosed {
		return nil
	}
	s.persistenceClosed = true
	close(s.persistenceStop)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, lockErr := s.persistenceConn.ExecContext(ctx, "SELECT RELEASE_LOCK(?)", s.persistenceLockName)
	return errors.Join(lockErr, s.persistenceConn.Close(), s.gamificationDB.Close())
}

// Commit the full next state before changing the live cache or attempt result.
// A failed disk write returns an error; the completion can be retried safely.
// recordStudyLocked records one finished activity and returns what it
// earned. Repeating an event ID earns nothing and changes nothing.
func (s *MemoryStore) recordStudyLocked(userID string, event studyEvent) (domain.XPGain, error) {
	current := s.gamification[userID]
	if current == nil {
		current = newGamificationState()
		if profile := s.profiles[userID]; profile.Timezone != "" {
			current.Timezone = profile.Timezone
		}
	}
	if current.Events[event.ID] {
		streak, _ := studyStreaks(current, event.At)
		return domain.XPGain{Total: current.XP, Streak: streak, CompletionID: event.ID}, nil
	}
	raw, err := json.Marshal(current)
	if err != nil {
		return domain.XPGain{}, err
	}
	next := newGamificationState()
	if err = json.Unmarshal(raw, next); err != nil {
		return domain.XPGain{}, err
	}
	if event.SSWAttempt != nil {
		if next.SSWAttempts == nil {
			next.SSWAttempts = map[string]SSWAttemptRecord{}
		}
		raw, err := json.Marshal(event.SSWAttempt)
		if err != nil {
			return domain.XPGain{}, err
		}
		var record SSWAttemptRecord
		if err := json.Unmarshal(raw, &record); err != nil {
			return domain.XPGain{}, err
		}
		next.SSWAttempts[record.Session.Attempt.ID] = record
	}
	if event.OfflinePracticeRecord != nil {
		if event.OfflinePracticeAttemptID == "" {
			return domain.XPGain{}, fmt.Errorf("offline practice attempt ID is required")
		}
		if next.OfflinePracticeAttempts == nil {
			next.OfflinePracticeAttempts = map[string]offlinePracticeAttemptRecord{}
		}
		// Stored below, once the XP it earned is known.
	}
	next.Events[event.ID] = true
	var awards []xpAward
	if event.Progress != nil {
		next.Lessons[event.LessonID] = *event.Progress
		next.Subjects[event.LessonID] = event.Subject
	}
	if event.Answered > 0 {
		day := event.At.In(studyTimezone(next)).Format("2006-01-02")
		seconds := event.Seconds
		if seconds < 0 {
			seconds = 0
		}
		if seconds > 7200 {
			seconds = 7200
		}
		next.Days[day] += seconds
		// Prevent disconnected/overlapping attempts from reporting more than one day.
		if next.Days[day] > 86400 {
			next.Days[day] = 86400
		}
		ledger := next.xpDayLocked(day)
		// Capped sources share one daily budget so practice is rewarded but
		// cannot be farmed for the weekly leaderboard.
		repeat := func(amount int, source string) {
			room := s.xpRepeatDailyCap() - ledger.Repeat
			if room <= 0 {
				return
			}
			if amount > room {
				amount = room
			}
			ledger.Repeat += amount
			awards = append(awards, xpAward{amount, source})
		}
		switch event.Kind {
		case "lesson":
			key := "lesson:" + event.LessonID
			if event.Score >= 80 && !next.Rewards[key] {
				next.Rewards[key] = true
				awards = append(awards, xpAward{xpLessonPass, "lesson"})
			} else if event.Score >= xpRepeatMinScore {
				repeat(xpLessonComplete, "lesson_repeat")
			}
		case "exam":
			next.Exams[event.ExamID] = true
			key := "exam:" + event.ExamID
			if event.Score >= 80 {
				next.PassedExam = true
			}
			if event.Score >= 80 && !next.Rewards[key] {
				next.Rewards[key] = true
				awards = append(awards, xpAward{xpExamPass, "exam"})
			} else if event.Score >= xpRepeatMinScore {
				repeat(xpExamComplete, "exam_repeat")
			}
		case "ssw-practice":
			if event.Complete {
				repeat(xpSSWPractice, "ssw_practice")
			}
		case "practice-lesson":
			key := "practice-lesson:" + event.LessonID
			if !next.Rewards[key] {
				next.Rewards[key] = true
				awards = append(awards, xpAward{xpPracticeLesson, "practice_lesson"})
			}
		}
		if !ledger.Goal && next.Days[day] >= dailyGoalMinutes*60 {
			ledger.Goal = true
			awards = append(awards, xpAward{xpDailyGoal, "daily_goal"})
		}
		next.pruneXPDays(event.At)
		next.XP += totalXP(awards)
	}
	if event.Kind != "ssw-practice" {
		unlockAchievements(next, event.At)
	}
	gain := domain.XPGain{Awarded: totalXP(awards), Total: next.XP, CompletionID: event.ID, Kind: completionKind(event.Kind)}
	gain.Streak, _ = studyStreaks(next, event.At)
	if gain.Awarded > 0 && gain.Kind != "" {
		recorded := s.gamificationNow().UTC()
		if next.Completions == nil {
			next.Completions = map[string]completionRecord{}
		}
		next.Completions[event.ID] = completionRecord{Kind: gain.Kind, Awarded: gain.Awarded, RecordedAt: recorded}
		pruneCompletions(next, recorded)
	}
	if event.OfflinePracticeRecord != nil {
		// The stored result carries the XP too, so a repeated sync, even
		// after a restart, returns exactly the first answer.
		record := *event.OfflinePracticeRecord
		stored := gain
		record.Result.XP = &stored
		next.OfflinePracticeAttempts[event.OfflinePracticeAttemptID] = record
	}
	if err = s.persistXPLocked(userID, next, awards, event.At); err != nil {
		return domain.XPGain{}, fmt.Errorf("persist study reward: %w", err)
	}
	s.recordXPLocked(userID, next.XP, totalXP(awards), event.At)
	s.gamification[userID] = next
	if event.Progress != nil {
		if s.progress[userID] == nil {
			s.progress[userID] = map[string]domain.LessonProgress{}
		}
		s.progress[userID][event.LessonID] = *event.Progress
	}
	profile := s.ensureProfileLocked(userID)
	profile.XP = next.XP
	profile.Level = 1 + next.XP/500
	profile.StreakDays, _ = studyStreaks(next, event.At)
	s.profiles[userID] = profile
	return gain, nil
}

// SetXPRules sets the admin-tunable XP limits. Zero keeps the default.
func (s *MemoryStore) SetXPRules(repeatDailyCap, ankiStreakCards int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.xpRepeatCap = repeatDailyCap
	s.ankiStreakCards = ankiStreakCards
}

func (s *MemoryStore) xpRepeatDailyCap() int {
	if s.xpRepeatCap > 0 {
		return s.xpRepeatCap
	}
	return defaultXPRepeatDailyCap
}

func (s *MemoryStore) ankiStreakCardsLocked() int {
	if s.ankiStreakCards > 0 {
		return s.ankiStreakCards
	}
	return defaultAnkiStreakCards
}

// practiceLessonPattern bounds the IDs a client may report for curriculum
// lessons, which are app content the server does not hold.
var practiceLessonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// CompletePracticeLesson records a finished Kaigo or SSW curriculum lesson:
// it counts for the streak and earns XP the first time. The server cannot
// verify reading, so the reward is small and given once per lesson.
func (s *MemoryStore) CompletePracticeLesson(userID, sectorID, lessonID string) (domain.XPGain, error) {
	if !practiceLessonPattern.MatchString(sectorID) || !practiceLessonPattern.MatchString(lessonID) {
		return domain.XPGain{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.profiles[userID]; !exists {
		return domain.XPGain{}, ErrNotFound
	}
	key := sectorID + ":" + lessonID
	return s.recordStudyLocked(userID, studyEvent{
		ID: "practice-lesson:" + key, Kind: "practice-lesson", LessonID: key,
		Answered: 1, At: s.gamificationNow().UTC(),
	})
}

func unlockAchievements(state *gamificationState, at time.Time) {
	for _, a := range gamificationAchievements(state, at) {
		if a.Progress >= a.Target {
			if _, ok := state.Unlocked[a.ID]; !ok {
				state.Unlocked[a.ID] = at.UTC()
			}
		}
	}
}

func studyStreaks(state *gamificationState, now time.Time) (int, int) {
	dates := make([]string, 0, len(state.Days))
	for day := range state.Days {
		dates = append(dates, day)
	}
	sort.Strings(dates)
	best, run := 0, 0
	previous := time.Time{}
	location := studyTimezone(state)
	for _, day := range dates {
		date, err := time.ParseInLocation("2006-01-02", day, location)
		if err != nil {
			continue
		}
		if date.Equal(previous.AddDate(0, 0, 1)) {
			run++
		} else {
			run = 1
		}
		if run > best {
			best = run
		}
		previous = date
	}
	today := now.In(location)
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location)
	if _, ok := state.Days[date.Format("2006-01-02")]; !ok {
		date = date.AddDate(0, 0, -1)
	}
	current := 0
	for {
		if _, ok := state.Days[date.Format("2006-01-02")]; !ok {
			break
		}
		current++
		date = date.AddDate(0, 0, -1)
	}
	return current, best
}
func gamificationAchievements(state *gamificationState, now time.Time) []domain.Achievement {
	counts := map[string]int{}
	total := 0
	for id, p := range state.Lessons {
		if p.Completed {
			total++
			counts[state.Subjects[id]]++
		}
	}
	subjects := 0
	for _, v := range counts {
		if v > 0 {
			subjects++
		}
	}
	_, best := studyStreaks(state, now)
	passed := 0
	if state.PassedExam {
		passed = 1
	}
	items := []domain.Achievement{
		{ID: "first-lesson", Name: "First Step", Description: "Lulus satu kuis lesson dengan nilai minimal 80%", Icon: "footprints", Progress: total, Target: 1},
		{ID: "lessons-10", Name: "Pelajar Tekun", Description: "Lulus 10 lesson berbeda", Icon: "book", Progress: total, Target: 10},
		{ID: "lessons-50", Name: "50 Lesson", Description: "Lulus 50 lesson berbeda", Icon: "book", Progress: total, Target: 50},
		{ID: "vocabulary-10", Name: "Word Collector", Description: "Lulus 10 lesson kosakata", Icon: "words", Progress: counts["vocabulary"], Target: 10},
		{ID: "grammar-10", Name: "Grammar Builder", Description: "Lulus 10 lesson grammar", Icon: "grammar", Progress: counts["grammar"], Target: 10},
		{ID: "kanji-10", Name: "Kanji Explorer", Description: "Lulus 10 lesson kanji", Icon: "kanji", Progress: counts["kanji"], Target: 10},
		{ID: "reading-5", Name: "Bookworm", Description: "Lulus 5 lesson reading", Icon: "book", Progress: counts["reading"], Target: 5},
		{ID: "all-subjects", Name: "Serba Bisa", Description: "Lulus lesson di keempat bidang", Icon: "star", Progress: subjects, Target: 4},
		{ID: "streak-7", Name: "7-Day Streak", Description: "Selesaikan kuis atau ujian selama 7 hari berturut-turut", Icon: "flame", Progress: best, Target: 7},
		{ID: "streak-30", Name: "30-Day Streak", Description: "Selesaikan kuis atau ujian selama 30 hari berturut-turut", Icon: "flame", Progress: best, Target: 30},
		{ID: "xp-500", Name: "Level Up", Description: "Kumpulkan 500 XP", Icon: "bolt", Progress: state.XP, Target: 500},
		{ID: "exam-ready", Name: "Exam Ready", Description: "Dapatkan minimal 80% pada diagnostic atau simulasi", Icon: "trophy", Progress: passed, Target: 1},
		{ID: achievementWeeklyPodium, Name: "Podium Mingguan", Description: "Finis di peringkat 1–3 leaderboard mingguan", Icon: "trophy", Progress: unlockedCount(state, achievementWeeklyPodium), Target: 1},
		{ID: achievementWeeklyChampion, Name: "Juara Mingguan", Description: "Finis di peringkat 1 leaderboard mingguan", Icon: "trophy", Progress: unlockedCount(state, achievementWeeklyChampion), Target: 1},
	}
	items = append(items, extraAchievements(state, total, counts, best)...)
	for i := range items {
		if items[i].Progress > items[i].Target {
			items[i].Progress = items[i].Target
		}
		if at, ok := state.Unlocked[items[i].ID]; ok {
			stamp := at
			items[i].UnlockedAt = &stamp
		}
	}
	return items
}
func unlockedCount(state *gamificationState, id string) int {
	if _, ok := state.Unlocked[id]; ok {
		return 1
	}
	return 0
}

func (s *MemoryStore) Gamification(userID string) GamificationSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state := s.gamification[userID]
	if state == nil {
		state = newGamificationState()
		if profile := s.profiles[userID]; profile.Timezone != "" {
			state.Timezone = profile.Timezone
		}
	}
	now := s.gamificationNow()
	current, best := studyStreaks(state, now)
	location := studyTimezone(state)
	result := GamificationSummary{XP: state.XP, Level: 1 + state.XP/500, LevelXP: state.XP % 500, NextLevelXP: 500, StreakDays: current, BestStreak: best, ExamsCompleted: len(state.Exams), ActiveDays: len(state.Days), Timezone: location.String(), MinutesToday: state.Days[now.In(location).Format("2006-01-02")] / 60, Achievements: gamificationAchievements(state, now)}
	for _, p := range state.Lessons {
		if p.Completed {
			result.LessonsCompleted++
		}
	}
	return result
}

// Double XP (docs/ads-xp-spec.md, phase C).
const (
	xpBoostWindow          = 30 * time.Minute
	defaultXPBoostDailyCap = 10
	completionRetention    = 48 * time.Hour
)

var (
	ErrBoostExpired = errors.New("double XP can only be claimed within 30 minutes")
	ErrBoostUsed    = errors.New("double XP was already claimed for this activity")
	ErrBoostLimit   = errors.New("daily double XP limit reached")
)

// completionKind maps an event kind to its double XP group.
func completionKind(eventKind string) string {
	switch eventKind {
	case "lesson":
		return "lessons"
	case "exam":
		return "exams"
	case "ssw-practice", "practice-lesson":
		return "practice"
	}
	return ""
}

func pruneCompletions(state *gamificationState, now time.Time) {
	for id, record := range state.Completions {
		if now.Sub(record.RecordedAt) > completionRetention {
			delete(state.Completions, id)
		}
	}
}

func boostCheck(state *gamificationState, completionID string, dailyCap int, now time.Time) (completionRecord, error) {
	if state == nil {
		return completionRecord{}, ErrNotFound
	}
	record, ok := state.Completions[completionID]
	if !ok || record.Awarded <= 0 {
		return completionRecord{}, ErrNotFound
	}
	if record.Boosted {
		return record, ErrBoostUsed
	}
	if now.Sub(record.RecordedAt) > xpBoostWindow {
		return record, ErrBoostExpired
	}
	if dailyCap <= 0 {
		dailyCap = defaultXPBoostDailyCap
	}
	day := now.In(studyTimezone(state)).Format("2006-01-02")
	if ledger := state.XPDays[day]; ledger != nil && ledger.Boosts >= dailyCap {
		return record, ErrBoostLimit
	}
	return record, nil
}

// XPBoostable reports whether the XP of completionID can still be doubled,
// and its group (lessons, exams, practice).
func (s *MemoryStore) XPBoostable(userID, completionID string, dailyCap int) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, err := boostCheck(s.gamification[userID], completionID, dailyCap, s.gamificationNow().UTC())
	return record.Kind, err
}

// BoostXP doubles the XP one finished activity earned: once, within the claim
// window, and within the daily limit. The bonus goes through the XP ledger, so
// it counts toward the weekly leaderboard like any other XP.
func (s *MemoryStore) BoostXP(userID, completionID string, dailyCap int) (domain.XPGain, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.gamificationNow().UTC()
	current := s.gamification[userID]
	record, err := boostCheck(current, completionID, dailyCap, now)
	if err != nil {
		return domain.XPGain{}, err
	}
	raw, err := json.Marshal(current)
	if err != nil {
		return domain.XPGain{}, err
	}
	next := newGamificationState()
	if err = json.Unmarshal(raw, next); err != nil {
		return domain.XPGain{}, err
	}
	record.Boosted = true
	next.Completions[completionID] = record
	next.xpDayLocked(now.In(studyTimezone(next)).Format("2006-01-02")).Boosts++
	awards := []xpAward{{record.Awarded, "double_xp"}}
	next.XP += record.Awarded
	next.pruneXPDays(now)
	if err = s.persistXPLocked(userID, next, awards, now); err != nil {
		return domain.XPGain{}, fmt.Errorf("persist double XP: %w", err)
	}
	s.recordXPLocked(userID, next.XP, record.Awarded, now)
	s.gamification[userID] = next
	profile := s.ensureProfileLocked(userID)
	profile.XP = next.XP
	profile.Level = 1 + next.XP/500
	streak, _ := studyStreaks(next, now)
	profile.StreakDays = streak
	s.profiles[userID] = profile
	return domain.XPGain{Awarded: record.Awarded, Total: next.XP, Streak: streak, CompletionID: completionID, Kind: record.Kind}, nil
}
