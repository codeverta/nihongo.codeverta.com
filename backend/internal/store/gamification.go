package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
	"nihongo-master-backend/internal/domain"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var studyLocation = time.FixedZone("Asia/Jakarta", 7*60*60)

// Only server-graded attempts can create these events. This is not a public input.
type studyEvent struct {
	ID, Kind, Subject, LessonID, ExamID string
	Score, Answered, Seconds            int
	At                                  time.Time
	Progress                            *domain.LessonProgress
}
type gamificationState struct {
	XP         int
	Events     map[string]bool
	Rewards    map[string]bool
	Lessons    map[string]domain.LessonProgress
	Subjects   map[string]string
	Days       map[string]int
	Unlocked   map[string]time.Time
	Exams      map[string]bool
	PassedExam bool
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
	return &gamificationState{Events: map[string]bool{}, Rewards: map[string]bool{}, Lessons: map[string]domain.LessonProgress{}, Subjects: map[string]string{}, Days: map[string]int{}, Unlocked: map[string]time.Time{}, Exams: map[string]bool{}}
}

// EnableGamificationPersistence must run before serving traffic. A deployment uses
// one API process per database; SQLite is a local durable store, not shared NFS.
func (s *MemoryStore) EnableGamificationPersistence(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("gamification database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	failed := true
	defer func() {
		if failed {
			db.Close()
		}
	}()
	for _, statement := range []string{"PRAGMA busy_timeout=5000", "PRAGMA locking_mode=EXCLUSIVE", "PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", `CREATE TABLE IF NOT EXISTS gamification_users(user_id TEXT PRIMARY KEY, state BLOB NOT NULL)`} {
		if _, err = db.Exec(statement); err != nil {
			return fmt.Errorf("initialize gamification database: %w", err)
		}
	}
	rows, err := db.Query("SELECT user_id,state FROM gamification_users")
	if err != nil {
		return err
	}
	defer rows.Close()
	loaded := map[string]*gamificationState{}
	for rows.Next() {
		var id string
		var raw []byte
		if err = rows.Scan(&id, &raw); err != nil {
			return err
		}
		state := newGamificationState()
		if err = json.Unmarshal(raw, state); err != nil {
			return fmt.Errorf("invalid gamification record: %w", err)
		}
		if state.XP < 0 || state.Events == nil || state.Rewards == nil || state.Lessons == nil || state.Subjects == nil || state.Days == nil || state.Unlocked == nil || state.Exams == nil {
			return fmt.Errorf("invalid gamification record for user %q", id)
		}
		loaded[id] = state
	}
	if err = rows.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gamificationDB != nil {
		return fmt.Errorf("gamification persistence already enabled")
	}
	s.gamificationDB = db
	s.gamification = loaded
	for id, state := range loaded {
		s.progress[id] = map[string]domain.LessonProgress{}
		for key, progress := range state.Lessons {
			s.progress[id][key] = progress
		}
		profile := s.ensureProfileLocked(id)
		profile.XP = state.XP
		profile.Level = 1 + state.XP/500
		s.profiles[id] = profile
	}
	failed = false
	return nil
}

func (s *MemoryStore) Close() error {
	if s.gamificationDB != nil {
		return s.gamificationDB.Close()
	}
	return nil
}

// Commit the full next state before changing the live cache or attempt result.
// A failed disk write returns an error; the completion can be retried safely.
func (s *MemoryStore) recordStudyLocked(userID string, event studyEvent) error {
	current := s.gamification[userID]
	if current == nil {
		current = newGamificationState()
	}
	if current.Events[event.ID] {
		return nil
	}
	raw, err := json.Marshal(current)
	if err != nil {
		return err
	}
	next := newGamificationState()
	if err = json.Unmarshal(raw, next); err != nil {
		return err
	}
	next.Events[event.ID] = true
	if event.Progress != nil {
		next.Lessons[event.LessonID] = *event.Progress
		next.Subjects[event.LessonID] = event.Subject
	}
	if event.Answered > 0 {
		day := event.At.In(studyLocation).Format("2006-01-02")
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
		reward := ""
		xp := 0
		if event.Kind == "lesson" && event.Score >= 80 {
			reward = "lesson:" + event.LessonID
			xp = 50
		}
		if event.Kind == "exam" {
			next.Exams[event.ExamID] = true
			if event.Score >= 80 {
				next.PassedExam = true
				reward = "exam:" + event.ExamID
				xp = 100
			}
		}
		if reward != "" && !next.Rewards[reward] {
			next.XP += xp
			next.Rewards[reward] = true
		}
	}
	for _, a := range gamificationAchievements(next, event.At) {
		if a.Progress >= a.Target {
			if _, ok := next.Unlocked[a.ID]; !ok {
				next.Unlocked[a.ID] = event.At.UTC()
			}
		}
	}
	if s.gamificationDB != nil {
		raw, err = json.Marshal(next)
		if err != nil {
			return err
		}
		if _, err = s.gamificationDB.Exec(`INSERT INTO gamification_users(user_id,state) VALUES(?,?) ON CONFLICT(user_id) DO UPDATE SET state=excluded.state`, userID, raw); err != nil {
			return fmt.Errorf("persist study reward: %w", err)
		}
	}
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
	return nil
}

func studyStreaks(state *gamificationState, now time.Time) (int, int) {
	dates := make([]string, 0, len(state.Days))
	for day := range state.Days {
		dates = append(dates, day)
	}
	sort.Strings(dates)
	best, run := 0, 0
	previous := time.Time{}
	for _, day := range dates {
		date, err := time.ParseInLocation("2006-01-02", day, studyLocation)
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
	today := now.In(studyLocation)
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, studyLocation)
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
	}
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
func (s *MemoryStore) Gamification(userID string) GamificationSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state := s.gamification[userID]
	if state == nil {
		state = newGamificationState()
	}
	now := s.gamificationNow()
	current, best := studyStreaks(state, now)
	result := GamificationSummary{XP: state.XP, Level: 1 + state.XP/500, LevelXP: state.XP % 500, NextLevelXP: 500, StreakDays: current, BestStreak: best, ExamsCompleted: len(state.Exams), ActiveDays: len(state.Days), Timezone: "Asia/Jakarta", MinutesToday: state.Days[now.In(studyLocation).Format("2006-01-02")] / 60, Achievements: gamificationAchievements(state, now)}
	for _, p := range state.Lessons {
		if p.Completed {
			result.LessonsCompleted++
		}
	}
	return result
}
