package store

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"nihongo-master-backend/internal/database"
)

// XP rules. First passes are in recordStudyLocked; these are the extra sources.
const (
	xpLessonPass          = 50  // first time a lesson scores >= 80%
	xpExamPass            = 100 // first time an exam scores >= 80%
	xpLessonComplete      = 10  // any later lesson scoring >= xpRepeatMinScore
	xpExamComplete        = 20  // any later exam scoring >= xpRepeatMinScore
	xpSSWPractice         = 10  // a fully answered SSW practice session
	xpPracticeLesson      = 5   // first completion of a Kaigo/SSW curriculum lesson
	xpRepeatMinScore      = 50  // below this a completion counts for the streak only
	xpDailyGoal           = 20
	xpAnkiCard            = 1
	xpAnkiMaxPerDay       = 30
	dailyGoalMinutes      = 20

	// Admin-tunable defaults (see SetXPRules).
	defaultXPRepeatDailyCap = 100 // XP per study day from repeats and SSW practice
	defaultAnkiStreakCards  = 10  // Anki reviews that make a day count for the streak
	leaderboardTopSize    = 50
	leaderboardSnapshotTT = time.Minute
	xpEventRetention      = 12 * 7 * 24 * time.Hour
	leaderboardJobID      = "weekly-close"

	achievementWeeklyPodium   = "weekly-podium"
	achievementWeeklyChampion = "weekly-champion"
)

// xpAward is one XP grant, written to the event log and leaderboard tables.
type xpAward struct {
	Amount int
	Source string
}

// xpDay is the per-day ledger for capped XP sources.
type xpDay struct {
	Anki   int  `json:",omitempty"`
	Repass int  `json:",omitempty"` // superseded by Repeat; kept to read old states
	Goal   bool `json:",omitempty"`
	// Repeat is XP earned today from capped sources: repeat completions and
	// SSW practice.
	Repeat int `json:",omitempty"`
	// Reviews counts today's Anki reviews, for the streak threshold.
	Reviews int `json:",omitempty"`
	// Boosts counts today's double XP claims.
	Boosts int `json:",omitempty"`
}

func (state *gamificationState) xpDayLocked(day string) *xpDay {
	if state.XPDays == nil {
		state.XPDays = map[string]*xpDay{}
	}
	entry := state.XPDays[day]
	if entry == nil {
		entry = &xpDay{}
		state.XPDays[day] = entry
	}
	return entry
}

// pruneXPDays keeps only the ledger days that can still be capped.
func (state *gamificationState) pruneXPDays(now time.Time) {
	cutoff := now.In(studyTimezone(state)).AddDate(0, 0, -14).Format("2006-01-02")
	for day := range state.XPDays {
		if day < cutoff {
			delete(state.XPDays, day)
		}
	}
}

func totalXP(awards []xpAward) int {
	total := 0
	for _, a := range awards {
		total += a.Amount
	}
	return total
}

// leaderboardWeekStart is Monday 00:00 Asia/Jakarta for everyone.
func leaderboardWeekStart(at time.Time) time.Time {
	local := at.In(studyLocation)
	offset := (int(local.Weekday()) + 6) % 7
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, studyLocation)
	return day.AddDate(0, 0, -offset)
}

func weekKey(start time.Time) string { return start.In(studyLocation).Format("2006-01-02") }

func parseWeekKey(key string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", key, studyLocation)
}

type lbScore struct {
	XP        int
	ReachedAt time.Time
}

type lbFlags struct {
	Hidden   bool
	Excluded bool
}

// WeeklyResult is a user's final position in a closed week.
type WeeklyResult struct {
	WeekStart string `json:"weekStart"`
	Rank      int    `json:"rank"`
	XP        int    `json:"xp"`
}

// leaderboardState is guarded by MemoryStore.mu, except snapshot fields which
// use their own mutex so readers never wait on writers.
type leaderboardState struct {
	allTime        map[string]lbScore
	weekly         map[string]map[string]lbScore
	flags          map[string]lbFlags
	lastWeekStart  string
	lastWeek       map[string]WeeklyResult
	lastClosedWeek string // used when no database is configured

	snapMu   sync.Mutex
	snapshot *lbSnapshot
}

func newLeaderboardState() *leaderboardState {
	return &leaderboardState{allTime: map[string]lbScore{}, weekly: map[string]map[string]lbScore{}, flags: map[string]lbFlags{}, lastWeek: map[string]WeeklyResult{}}
}

type lbRow struct {
	UserID    string
	XP        int
	ReachedAt time.Time
	Name      string
	Tag       string
	PhotoURL  string
}

type lbBoard struct {
	Rows  []lbRow
	Index map[string]int
}

type lbSnapshot struct {
	BuiltAt   time.Time
	WeekStart string
	Week      lbBoard
	All       lbBoard
}

// lbBetter reports whether a ranks above b: more XP, then reached it first.
func lbBetter(aXP int, aAt time.Time, aID string, bXP int, bAt time.Time, bID string) bool {
	if aXP != bXP {
		return aXP > bXP
	}
	if !aAt.Equal(bAt) {
		return aAt.Before(bAt)
	}
	return aID < bID
}

// leaderboardTag is a stable 4-digit number used as "Pelajar #1234".
func leaderboardTag(userID string) string {
	sum := sha256.Sum256([]byte("nihongo-master:leaderboard:" + userID))
	return fmt.Sprintf("%04d", binary.BigEndian.Uint32(sum[:4])%10000)
}

// leaderboardDisplayName never exposes an email: names containing "@" or
// placeholders fall back to the tag on the client.
func leaderboardDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || strings.Contains(name, "@") || name == "Nihongo Learner" {
		return ""
	}
	if runes := []rune(name); len(runes) > 40 {
		name = string(runes[:40])
	}
	return name
}

// specialLeaderboardUserLocked reports reviewer, admin and demo accounts.
func (s *MemoryStore) specialLeaderboardUserLocked(userID string) bool {
	if userID == database.DemoUserID || (s.playReviewUserID != "" && userID == s.playReviewUserID) {
		return true
	}
	if s.adminLoginEmail != "" {
		if a, ok := s.accounts[s.adminLoginEmail]; ok && a.ID == userID {
			return true
		}
	}
	return false
}

func (s *MemoryStore) leaderboardEligibleLocked(userID string) bool {
	flags := s.lb.flags[userID]
	return !flags.Hidden && !flags.Excluded && !s.specialLeaderboardUserLocked(userID)
}

// recordXPLocked updates the in-memory leaderboard after a committed award.
func (s *MemoryStore) recordXPLocked(userID string, total int, awarded int, at time.Time) {
	if awarded <= 0 {
		return
	}
	at = at.UTC()
	s.lb.allTime[userID] = lbScore{XP: total, ReachedAt: at}
	key := weekKey(leaderboardWeekStart(at))
	week := s.lb.weekly[key]
	if week == nil {
		week = map[string]lbScore{}
		s.lb.weekly[key] = week
	}
	current := week[userID]
	week[userID] = lbScore{XP: current.XP + awarded, ReachedAt: at}
}

// persistXPLocked writes the next state together with the award log, the
// weekly score and the indexed all-time columns in one transaction.
func (s *MemoryStore) persistXPLocked(userID string, next *gamificationState, awards []xpAward, at time.Time) error {
	if s.gamificationDB == nil {
		return nil
	}
	raw, err := json.Marshal(next)
	if err != nil {
		return err
	}
	awarded := totalXP(awards)
	if awarded <= 0 {
		return s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state"})}).Create(&database.GamificationUser{UserID: userID, State: raw}).Error
	}
	at = at.UTC()
	key := weekKey(leaderboardWeekStart(at))
	weekXP := s.lb.weekly[key][userID].XP + awarded
	return s.ormWriter.Transaction(func(tx *gorm.DB) error {
		reached := at
		if err := tx.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state", "xp_total", "reached_at"})}).Create(&database.GamificationUser{UserID: userID, State: raw, XPTotal: next.XP, ReachedAt: &reached}).Error; err != nil {
			return err
		}
		events := make([]database.XPEvent, 0, len(awards))
		for _, a := range awards {
			if a.Amount > 0 {
				events = append(events, database.XPEvent{UserID: userID, Amount: a.Amount, Source: a.Source, AwardedAt: at})
			}
		}
		if err := tx.Create(&events).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"xp", "reached_at"})}).Create(&database.WeeklyScore{WeekStart: key, UserID: userID, XP: weekXP, ReachedAt: at}).Error
	})
}

// loadLeaderboard restores leaderboard state at startup. records are
// the already-loaded gamification rows; rows whose indexed XP column is out
// of date (written by an older server) are backfilled once.
func (s *MemoryStore) loadLeaderboard(orm, writer *gorm.DB, records []database.GamificationUser, states map[string]*gamificationState) (*leaderboardState, error) {
	lb := newLeaderboardState()
	if err := database.Migrate(orm, &database.XPEvent{}, &database.WeeklyScore{}, &database.WeeklyResult{}, &database.LeaderboardFlag{}, &database.LeaderboardJob{}); err != nil {
		return nil, fmt.Errorf("initialize leaderboard schema: %w", err)
	}
	now := s.gamificationNow().UTC()
	backfill := map[string]int{}
	for _, record := range records {
		state := states[record.UserID]
		if state == nil || state.XP <= 0 {
			continue
		}
		reached := now
		if record.ReachedAt != nil {
			reached = record.ReachedAt.UTC()
		}
		if record.XPTotal != state.XP || record.ReachedAt == nil {
			backfill[record.UserID] = state.XP
		}
		lb.allTime[record.UserID] = lbScore{XP: state.XP, ReachedAt: reached}
	}
	if len(backfill) > 0 {
		if err := writer.Transaction(func(tx *gorm.DB) error {
			for id, xp := range backfill {
				reached := lb.allTime[id].ReachedAt
				if err := tx.Model(&database.GamificationUser{}).Where("user_id = ?", id).Updates(map[string]any{"xp_total": xp, "reached_at": reached}).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("backfill leaderboard XP: %w", err)
		}
	}
	current := leaderboardWeekStart(now)
	var scores []database.WeeklyScore
	if err := orm.Where("week_start >= ?", weekKey(current.AddDate(0, 0, -7))).Find(&scores).Error; err != nil {
		return nil, err
	}
	for _, score := range scores {
		week := lb.weekly[score.WeekStart]
		if week == nil {
			week = map[string]lbScore{}
			lb.weekly[score.WeekStart] = week
		}
		week[score.UserID] = lbScore{XP: score.XP, ReachedAt: score.ReachedAt.UTC()}
	}
	var flags []database.LeaderboardFlag
	if err := orm.Find(&flags).Error; err != nil {
		return nil, err
	}
	for _, f := range flags {
		lb.flags[f.UserID] = lbFlags{Hidden: f.Hidden, Excluded: f.Excluded}
	}
	var latest struct{ WeekStart string }
	if err := orm.Model(&database.WeeklyResult{}).Select("MAX(week_start) AS week_start").Scan(&latest).Error; err != nil {
		return nil, err
	}
	if latest.WeekStart != "" {
		var results []database.WeeklyResult
		if err := orm.Where("week_start = ?", latest.WeekStart).Find(&results).Error; err != nil {
			return nil, err
		}
		lb.lastWeekStart = latest.WeekStart
		for _, r := range results {
			lb.lastWeek[r.UserID] = WeeklyResult{WeekStart: r.WeekStart, Rank: r.Rank, XP: r.XP}
		}
	}
	return lb, nil
}

// leaderboardSnapshot returns a sorted snapshot at most a minute old and
// always for the current week.
func (s *MemoryStore) leaderboardSnapshot() *lbSnapshot {
	now := s.gamificationNow()
	week := weekKey(leaderboardWeekStart(now))
	s.lb.snapMu.Lock()
	defer s.lb.snapMu.Unlock()
	if snap := s.lb.snapshot; snap != nil && snap.WeekStart == week && now.Sub(snap.BuiltAt) < leaderboardSnapshotTT && !now.Before(snap.BuiltAt) {
		return snap
	}
	snap := s.buildLeaderboardSnapshot(now)
	s.lb.snapshot = snap
	return snap
}

// RefreshLeaderboard rebuilds the snapshot now (used by the ticker and after
// settings changes).
func (s *MemoryStore) RefreshLeaderboard() {
	now := s.gamificationNow()
	snap := s.buildLeaderboardSnapshot(now)
	s.lb.snapMu.Lock()
	s.lb.snapshot = snap
	s.lb.snapMu.Unlock()
}

func (s *MemoryStore) buildLeaderboardSnapshot(now time.Time) *lbSnapshot {
	week := weekKey(leaderboardWeekStart(now))
	s.mu.RLock()
	defer s.mu.RUnlock()
	type identity struct{ name, photo string }
	people := make(map[string]identity, len(s.accounts))
	for _, a := range s.accounts {
		people[a.ID] = identity{a.profileDisplayName(), a.profilePhotoURL()}
	}
	build := func(scores map[string]lbScore) lbBoard {
		rows := make([]lbRow, 0, len(scores))
		for id, score := range scores {
			if score.XP <= 0 || !s.leaderboardEligibleLocked(id) {
				continue
			}
			who, ok := people[id]
			if !ok {
				if profile, exists := s.profiles[id]; exists {
					who = identity{profile.DisplayName, profile.PhotoURL}
				}
			}
			rows = append(rows, lbRow{UserID: id, XP: score.XP, ReachedAt: score.ReachedAt, Name: leaderboardDisplayName(who.name), Tag: leaderboardTag(id), PhotoURL: who.photo})
		}
		sort.Slice(rows, func(i, j int) bool {
			return lbBetter(rows[i].XP, rows[i].ReachedAt, rows[i].UserID, rows[j].XP, rows[j].ReachedAt, rows[j].UserID)
		})
		index := make(map[string]int, len(rows))
		for i, r := range rows {
			index[r.UserID] = i
		}
		return lbBoard{Rows: rows, Index: index}
	}
	return &lbSnapshot{BuiltAt: now, WeekStart: week, Week: build(s.lb.weekly[week]), All: build(s.lb.allTime)}
}

// LeaderboardEntry is one public row. It never carries an email or user ID.
type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	DisplayName string `json:"displayName"`
	Tag         string `json:"tag"`
	PhotoURL    string `json:"photoUrl,omitempty"`
	XP          int    `json:"xp"`
}

// LeaderboardPage is the public top list, identical for every viewer.
type LeaderboardPage struct {
	Period       string             `json:"period"`
	WeekStart    time.Time          `json:"weekStart"`
	WeekEndsAt   time.Time          `json:"weekEndsAt"`
	GeneratedAt  time.Time          `json:"generatedAt"`
	Participants int                `json:"participants"`
	Entries      []LeaderboardEntry `json:"entries"`
}

func (s *MemoryStore) Leaderboard(period string) (LeaderboardPage, error) {
	if period == "" {
		period = "week"
	}
	if period != "week" && period != "all" {
		return LeaderboardPage{}, ErrInvalid
	}
	snap := s.leaderboardSnapshot()
	board := snap.Week
	if period == "all" {
		board = snap.All
	}
	start, _ := parseWeekKey(snap.WeekStart)
	page := LeaderboardPage{Period: period, WeekStart: start.UTC(), WeekEndsAt: start.AddDate(0, 0, 7).UTC(), GeneratedAt: snap.BuiltAt.UTC(), Participants: len(board.Rows), Entries: []LeaderboardEntry{}}
	for i, row := range board.Rows {
		if i >= leaderboardTopSize {
			break
		}
		page.Entries = append(page.Entries, LeaderboardEntry{Rank: i + 1, DisplayName: row.Name, Tag: row.Tag, PhotoURL: row.PhotoURL, XP: row.XP})
	}
	return page, nil
}

// LeaderboardStanding is the viewer's own position in one period.
type LeaderboardStanding struct {
	XP           int `json:"xp"`
	Rank         int `json:"rank"`
	Participants int `json:"participants"`
	// GapToNext is the XP needed to pass the person one rank above.
	GapToNext int `json:"gapToNext"`
}

// MyLeaderboard is the response of GET /v1/me/leaderboard.
type MyLeaderboard struct {
	Hidden      bool                `json:"hidden"`
	DisplayName string              `json:"displayName"`
	Tag         string              `json:"tag"`
	PhotoURL    string              `json:"photoUrl,omitempty"`
	WeekStart   time.Time           `json:"weekStart"`
	WeekEndsAt  time.Time           `json:"weekEndsAt"`
	Week        LeaderboardStanding `json:"week"`
	All         LeaderboardStanding `json:"all"`
	LastWeek    *WeeklyResult       `json:"lastWeek"`
}

// standing ranks a live score against the snapshot without rebuilding it. A
// hidden, excluded or special user gets the rank they would have if shown.
func standing(board lbBoard, userID string, score lbScore) LeaderboardStanding {
	result := LeaderboardStanding{XP: score.XP, Participants: len(board.Rows)}
	self, inBoard := board.Index[userID]
	if score.XP <= 0 {
		if inBoard {
			result.Participants--
		}
		return result
	}
	if !inBoard {
		result.Participants++
	}
	better := sort.Search(len(board.Rows), func(i int) bool {
		r := board.Rows[i]
		return !lbBetter(r.XP, r.ReachedAt, r.UserID, score.XP, score.ReachedAt, userID)
	})
	above := better - 1
	if inBoard && self < better {
		better--
		if above == self {
			above--
		}
	}
	result.Rank = better + 1
	if above >= 0 {
		result.GapToNext = board.Rows[above].XP - score.XP + 1
		if result.GapToNext < 1 {
			result.GapToNext = 1
		}
	}
	return result
}

func (s *MemoryStore) MyLeaderboard(userID string) MyLeaderboard {
	snap := s.leaderboardSnapshot()
	start, _ := parseWeekKey(snap.WeekStart)
	s.mu.RLock()
	weekScore := s.lb.weekly[snap.WeekStart][userID]
	allScore := s.lb.allTime[userID]
	flags := s.lb.flags[userID]
	result := MyLeaderboard{Hidden: flags.Hidden, Tag: leaderboardTag(userID), WeekStart: start.UTC(), WeekEndsAt: start.AddDate(0, 0, 7).UTC()}
	for _, a := range s.accounts {
		if a.ID == userID {
			result.DisplayName, result.PhotoURL = leaderboardDisplayName(a.profileDisplayName()), a.profilePhotoURL()
			break
		}
	}
	if last, ok := s.lb.lastWeek[userID]; ok && s.lb.lastWeekStart == weekKey(start.AddDate(0, 0, -7)) {
		lastWeek := last
		result.LastWeek = &lastWeek
	}
	s.mu.RUnlock()
	result.Week = standing(snap.Week, userID, weekScore)
	result.All = standing(snap.All, userID, allScore)
	return result
}

// SetLeaderboardHidden stores the user's "hide me from the leaderboard" switch.
func (s *MemoryStore) SetLeaderboardHidden(userID string, hidden bool) error {
	if err := s.updateLeaderboardFlags(userID, func(f *lbFlags) { f.Hidden = hidden }); err != nil {
		return err
	}
	s.RefreshLeaderboard()
	return nil
}

// SetLeaderboardExcluded silently removes or restores an account (admin only).
// XP is untouched.
func (s *MemoryStore) SetLeaderboardExcluded(userID string, excluded bool) error {
	if err := s.updateLeaderboardFlags(userID, func(f *lbFlags) { f.Excluded = excluded }); err != nil {
		return err
	}
	s.RefreshLeaderboard()
	return nil
}

func (s *MemoryStore) LeaderboardExcluded(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lb.flags[userID].Excluded
}

func (s *MemoryStore) updateLeaderboardFlags(userID string, change func(*lbFlags)) error {
	if _, err := uuid.Parse(userID); err != nil {
		return ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	known := false
	for _, a := range s.accounts {
		if a.ID == userID {
			known = true
			break
		}
	}
	if _, ok := s.profiles[userID]; ok {
		known = true
	}
	if !known {
		return ErrNotFound
	}
	flags := s.lb.flags[userID]
	change(&flags)
	if s.gamificationDB != nil {
		if err := s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"hidden", "excluded"})}).Create(&database.LeaderboardFlag{UserID: userID, Hidden: flags.Hidden, Excluded: flags.Excluded}).Error; err != nil {
			return fmt.Errorf("persist leaderboard setting: %w", err)
		}
	}
	if flags == (lbFlags{}) {
		delete(s.lb.flags, userID)
	} else {
		s.lb.flags[userID] = flags
	}
	return nil
}

// forgetLeaderboardLocked drops a deleted account from every ranking.
func (s *MemoryStore) forgetLeaderboardLocked(userID string) {
	delete(s.lb.allTime, userID)
	for _, week := range s.lb.weekly {
		delete(week, userID)
	}
	delete(s.lb.flags, userID)
	delete(s.lb.lastWeek, userID)
}

func deleteLeaderboardRows(tx *gorm.DB, userID string) error {
	for _, model := range []any{&database.XPEvent{}, &database.WeeklyScore{}, &database.WeeklyResult{}, &database.LeaderboardFlag{}} {
		if err := tx.Where("user_id = ?", userID).Delete(model).Error; err != nil {
			return err
		}
	}
	return nil
}

// RunLeaderboard rebuilds the snapshot every minute and runs the weekly close
// when it is due. It returns when ctx is cancelled.
func (s *MemoryStore) RunLeaderboard(ctx context.Context) {
	ticker := time.NewTicker(leaderboardSnapshotTT)
	defer ticker.Stop()
	for {
		if err := s.CloseLeaderboardWeeks(ctx); err != nil && ctx.Err() == nil {
			log.Printf("leaderboard weekly close: %v", err)
		}
		s.RefreshLeaderboard()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

var errLeaderboardBusy = errors.New("leaderboard close is running elsewhere")

// CloseLeaderboardWeeks finalizes every finished week that has not been closed:
// it stores final ranks, grants the weekly achievements and prunes old XP
// events. With a database it holds a lease so it runs once across instances.
func (s *MemoryStore) CloseLeaderboardWeeks(ctx context.Context) error {
	now := s.gamificationNow().UTC()
	current := leaderboardWeekStart(now)
	nextRun := current.AddDate(0, 0, 7).UTC()
	s.mu.RLock()
	persistent := s.gamificationDB != nil
	s.mu.RUnlock()
	if !persistent {
		s.mu.RLock()
		last := s.lb.lastClosedWeek
		s.mu.RUnlock()
		if last == "" {
			last = s.firstUnclosedWeekGuess(current)
		}
		for week := s.nextWeekToClose(last); week.Before(current); week = week.AddDate(0, 0, 7) {
			if err := s.closeWeek(ctx, weekKey(week), now); err != nil {
				return err
			}
			s.mu.Lock()
			s.lb.lastClosedWeek = weekKey(week)
			s.mu.Unlock()
		}
		return nil
	}
	owner := uuid.NewString()
	var job database.LeaderboardJob
	notDue := false
	err := s.ormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&job, "id = ?", leaderboardJobID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// First run: nothing before this week was recorded, so start here.
			job = database.LeaderboardJob{ID: leaderboardJobID, NextRunAt: nextRun, LastClosedWeek: weekKey(current.AddDate(0, 0, -7))}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&job).Error; err != nil {
				return err
			}
			// Commit the new schedule row; returning an error here would roll it back.
			notDue = true
			return nil
		}
		if err != nil {
			return err
		}
		if job.NextRunAt.After(now) || (job.LeaseUntil != nil && job.LeaseUntil.After(now)) {
			return errLeaderboardBusy
		}
		until := now.Add(10 * time.Minute)
		return tx.Model(&job).Updates(map[string]any{"lease_owner": owner, "lease_until": until}).Error
	})
	if errors.Is(err, errLeaderboardBusy) || (err == nil && notDue) {
		return nil
	}
	if err != nil {
		return err
	}
	closed := job.LastClosedWeek
	var runErr error
	for week := s.nextWeekToClose(closed); week.Before(current); week = week.AddDate(0, 0, 7) {
		if runErr = s.closeWeek(ctx, weekKey(week), now); runErr != nil {
			break
		}
		closed = weekKey(week)
	}
	if runErr == nil {
		runErr = s.pruneXPEvents(ctx, now)
	}
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	updates := map[string]any{"lease_owner": "", "lease_until": nil, "last_closed_week": closed, "last_error": ""}
	if runErr == nil {
		updates["next_run_at"] = nextRun
	} else {
		message := runErr.Error()
		if len(message) > 2000 {
			message = message[:2000]
		}
		updates["last_error"] = message
	}
	finishErr := s.ormDB.WithContext(finishCtx).Model(&database.LeaderboardJob{}).Where("id = ? AND lease_owner = ?", leaderboardJobID, owner).Updates(updates).Error
	return errors.Join(runErr, finishErr)
}

// firstUnclosedWeekGuess is used without a database: the oldest week still in
// memory, or last week.
func (s *MemoryStore) firstUnclosedWeekGuess(current time.Time) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	oldest := weekKey(current)
	for key := range s.lb.weekly {
		if key < oldest {
			oldest = key
		}
	}
	start, _ := parseWeekKey(oldest)
	return weekKey(start.AddDate(0, 0, -7))
}

func (s *MemoryStore) nextWeekToClose(lastClosed string) time.Time {
	start, err := parseWeekKey(lastClosed)
	if err != nil {
		return leaderboardWeekStart(s.gamificationNow())
	}
	return start.AddDate(0, 0, 7)
}

type finalRow struct {
	UserID    string
	XP        int
	ReachedAt time.Time
}

func (s *MemoryStore) closeWeek(ctx context.Context, week string, now time.Time) error {
	s.mu.RLock()
	scores, inMemory := s.lb.weekly[week]
	rows := make([]finalRow, 0, len(scores))
	for id, score := range scores {
		rows = append(rows, finalRow{id, score.XP, score.ReachedAt})
	}
	persistent := s.gamificationDB != nil
	s.mu.RUnlock()
	if !inMemory && persistent {
		var stored []database.WeeklyScore
		if err := s.ormDB.WithContext(ctx).Where("week_start = ? AND xp > 0", week).Find(&stored).Error; err != nil {
			return err
		}
		for _, score := range stored {
			rows = append(rows, finalRow{score.UserID, score.XP, score.ReachedAt.UTC()})
		}
	}
	s.mu.RLock()
	ranked := rows[:0]
	for _, r := range rows {
		if r.XP > 0 && s.leaderboardEligibleLocked(r.UserID) {
			ranked = append(ranked, r)
		}
	}
	s.mu.RUnlock()
	sort.Slice(ranked, func(i, j int) bool {
		return lbBetter(ranked[i].XP, ranked[i].ReachedAt, ranked[i].UserID, ranked[j].XP, ranked[j].ReachedAt, ranked[j].UserID)
	})
	results := make([]database.WeeklyResult, len(ranked))
	for i, r := range ranked {
		results[i] = database.WeeklyResult{WeekStart: week, UserID: r.UserID, Rank: i + 1, XP: r.XP}
	}
	if persistent && len(results) > 0 {
		if err := s.ormDB.WithContext(ctx).Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"rank_position", "xp"})}).CreateInBatches(results, 1000).Error; err != nil {
			return fmt.Errorf("store weekly results: %w", err)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range results {
		if i >= 10 {
			break
		}
		if err := s.grantWeeklyAchievementsLocked(r.UserID, r.Rank, now); err != nil {
			return err
		}
	}
	s.lb.lastWeekStart = week
	s.lb.lastWeek = make(map[string]WeeklyResult, len(results))
	for _, r := range results {
		s.lb.lastWeek[r.UserID] = WeeklyResult{WeekStart: week, Rank: r.Rank, XP: r.XP}
	}
	// Keep only weeks that are still open (or not yet closed) in memory.
	for key := range s.lb.weekly {
		if key <= week {
			delete(s.lb.weekly, key)
		}
	}
	return nil
}

// grantWeeklyAchievementsLocked unlocks "Podium Mingguan" for ranks 1-3 and
// "Juara Mingguan" for rank 1, once per account lifetime.
func (s *MemoryStore) grantWeeklyAchievementsLocked(userID string, rank int, now time.Time) error {
	current := s.gamification[userID]
	if current == nil {
		current = newGamificationState()
	}
	ids := []string{achievementWeeklyTop10}
	if rank <= 3 {
		ids = append(ids, achievementWeeklyPodium)
	}
	if rank == 1 {
		ids = append(ids, achievementWeeklyChampion)
	}
	missing := false
	for _, id := range ids {
		if _, ok := current.Unlocked[id]; !ok {
			missing = true
		}
	}
	if !missing {
		return nil
	}
	raw, err := json.Marshal(current)
	if err != nil {
		return err
	}
	next := newGamificationState()
	if err := json.Unmarshal(raw, next); err != nil {
		return err
	}
	for _, id := range ids {
		if _, ok := next.Unlocked[id]; !ok {
			next.Unlocked[id] = now.UTC()
		}
	}
	if err := s.persistXPLocked(userID, next, nil, now); err != nil {
		return fmt.Errorf("grant weekly achievement: %w", err)
	}
	s.gamification[userID] = next
	return nil
}

// pruneXPEvents deletes raw XP events older than 12 weeks in small batches.
func (s *MemoryStore) pruneXPEvents(ctx context.Context, now time.Time) error {
	cutoff := now.Add(-xpEventRetention)
	for {
		result := s.ormDB.WithContext(ctx).Exec("DELETE FROM xp_events WHERE awarded_at < ? LIMIT 5000", cutoff)
		if result.Error != nil {
			return fmt.Errorf("prune XP events: %w", result.Error)
		}
		if result.RowsAffected < 5000 {
			return nil
		}
	}
}
