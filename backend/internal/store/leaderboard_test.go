package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/testdb"
)

// leaderboardClock is Wednesday 2026-11-18 10:00 WIB.
var leaderboardClock = time.Date(2026, 11, 18, 3, 0, 0, 0, time.UTC)

func newLeaderboardStore(t *testing.T) (*MemoryStore, *time.Time) {
	t.Helper()
	s := NewMemoryStore()
	now := leaderboardClock
	s.gamificationNow = func() time.Time { return now }
	return s, &now
}

func addLeaderboardAccount(s *MemoryStore, name string) string {
	id := uuid.NewSHA1(uuid.NameSpaceOID, []byte("lb:"+name)).String()
	s.mu.Lock()
	a := account{ID: id, Email: name + "@example.com", DisplayName: name}
	s.accounts[a.Email] = a
	s.profiles[id] = profileForAccount(a)
	s.mu.Unlock()
	return id
}

// passLesson awards a first lesson pass (50 XP) at the given time.
func passLesson(t *testing.T, s *MemoryStore, user, lesson string, at time.Time) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.recordStudyLocked(user, studyEvent{ID: fmt.Sprintf("test:%s:%s:%d", user, lesson, at.UnixNano()), Kind: "lesson", Subject: "grammar", LessonID: lesson, Score: 100, Answered: 5, Seconds: 30, At: at}); err != nil {
		t.Fatal(err)
	}
}

func entryNames(page LeaderboardPage) []string {
	names := make([]string, len(page.Entries))
	for i, e := range page.Entries {
		names[i] = e.DisplayName
	}
	return names
}

func TestLeaderboardWeekStartIsMondayMidnightJakarta(t *testing.T) {
	cases := map[time.Time]string{
		time.Date(2026, 11, 15, 16, 59, 59, 0, time.UTC): "2026-11-09", // Sunday 23:59:59 WIB
		time.Date(2026, 11, 15, 17, 0, 0, 0, time.UTC):   "2026-11-16", // Monday 00:00 WIB
		time.Date(2026, 11, 22, 16, 0, 0, 0, time.UTC):   "2026-11-16", // Sunday 23:00 WIB
		time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC):    "2026-11-16",
	}
	for at, want := range cases {
		if got := weekKey(leaderboardWeekStart(at)); got != want {
			t.Errorf("week of %s = %s, want %s", at, got, want)
		}
	}
}

func TestLeaderboardRanksByXPThenEarliestAndTopFifty(t *testing.T) {
	s, now := newLeaderboardStore(t)
	ana := addLeaderboardAccount(s, "Ana")
	budi := addLeaderboardAccount(s, "Budi")
	citra := addLeaderboardAccount(s, "Citra")
	passLesson(t, s, budi, "l1", *now)
	passLesson(t, s, ana, "l1", now.Add(time.Minute)) // ties Budi, but later
	passLesson(t, s, citra, "l1", now.Add(2*time.Minute))
	passLesson(t, s, citra, "l2", now.Add(3*time.Minute))
	page, err := s.Leaderboard("week")
	if err != nil {
		t.Fatal(err)
	}
	if got := entryNames(page); fmt.Sprint(got) != "[Citra Budi Ana]" || page.Participants != 3 {
		t.Fatalf("order = %v, participants %d", got, page.Participants)
	}
	if page.Entries[0].Rank != 1 || page.Entries[2].Rank != 3 || page.Entries[0].XP != 100 {
		t.Fatalf("bad entries %+v", page.Entries)
	}
	if !page.WeekStart.Equal(time.Date(2026, 11, 15, 17, 0, 0, 0, time.UTC)) || !page.WeekEndsAt.Equal(time.Date(2026, 11, 22, 17, 0, 0, 0, time.UTC)) {
		t.Fatalf("week bounds %s %s", page.WeekStart, page.WeekEndsAt)
	}
	if _, err := s.Leaderboard("month"); err != ErrInvalid {
		t.Fatal("unknown period accepted")
	}

	for i := 0; i < 55; i++ {
		passLesson(t, s, addLeaderboardAccount(s, fmt.Sprintf("p%02d", i)), "l1", now.Add(time.Hour+time.Duration(i)*time.Second))
	}
	*now = now.Add(2 * time.Minute) // let the snapshot expire
	page, _ = s.Leaderboard("all")
	if len(page.Entries) != 50 || page.Participants != 58 {
		t.Fatalf("top list %d entries, %d participants", len(page.Entries), page.Participants)
	}
	me := s.MyLeaderboard(addLeaderboardAccount(s, "Zero"))
	if me.Week.Rank != 0 || me.Week.XP != 0 || me.Week.Participants != 58 {
		t.Fatalf("zero-XP user ranked: %+v", me.Week)
	}
	last := s.MyLeaderboard(fmt.Sprint(uuid.NewSHA1(uuid.NameSpaceOID, []byte("lb:p54"))))
	if last.Week.Rank != 58 || last.Week.GapToNext != 1 {
		t.Fatalf("own rank outside top 50: %+v", last.Week)
	}
}

func TestMyLeaderboardUsesLiveXPAndGapToRankAbove(t *testing.T) {
	s, now := newLeaderboardStore(t)
	ana := addLeaderboardAccount(s, "Ana")
	budi := addLeaderboardAccount(s, "Budi")
	passLesson(t, s, ana, "l1", *now)
	passLesson(t, s, ana, "l2", *now)
	passLesson(t, s, budi, "l1", *now)
	s.RefreshLeaderboard()
	me := s.MyLeaderboard(budi)
	if me.Week.Rank != 2 || me.Week.GapToNext != 51 || me.Week.Participants != 2 || me.All.Rank != 2 {
		t.Fatalf("before: %+v", me)
	}
	// Snapshot is still the old one; Budi's live XP must already count.
	passLesson(t, s, budi, "l2", now.Add(time.Second))
	passLesson(t, s, budi, "l3", now.Add(2*time.Second))
	me = s.MyLeaderboard(budi)
	if me.Week.Rank != 1 || me.Week.GapToNext != 0 || me.Week.XP != 150 || me.Week.Participants != 2 {
		t.Fatalf("live rank: %+v", me.Week)
	}
	if me.DisplayName != "Budi" || me.Tag == "" || me.LastWeek != nil {
		t.Fatalf("identity: %+v", me)
	}
}

func TestLeaderboardTieGoesToWhoeverReachedFirst(t *testing.T) {
	s, now := newLeaderboardStore(t)
	late := addLeaderboardAccount(s, "Late")
	early := addLeaderboardAccount(s, "Early")
	passLesson(t, s, early, "l1", *now)
	passLesson(t, s, late, "l1", now.Add(time.Second))
	page, _ := s.Leaderboard("week")
	if fmt.Sprint(entryNames(page)) != "[Early Late]" {
		t.Fatalf("tie order %v", entryNames(page))
	}
	if r := s.MyLeaderboard(late).Week; r.Rank != 2 || r.GapToNext != 1 {
		t.Fatalf("tie standing %+v", r)
	}
}

func TestLeaderboardHidesHiddenExcludedAndSpecialAccounts(t *testing.T) {
	s, now := newLeaderboardStore(t)
	if err := s.ConfigurePlayReviewAccount("reviewer@example.com", "review-password"); err != nil {
		t.Fatal(err)
	}
	if err := s.ConfigureAdminLoginAccount("admin@example.com", "admin-password-123"); err != nil {
		t.Fatal(err)
	}
	ana := addLeaderboardAccount(s, "Ana")
	hidden := addLeaderboardAccount(s, "Hidden")
	excluded := addLeaderboardAccount(s, "Excluded")
	passLesson(t, s, ana, "l1", *now)
	passLesson(t, s, hidden, "l1", now.Add(-time.Hour))
	passLesson(t, s, hidden, "l2", now.Add(-time.Hour))
	passLesson(t, s, excluded, "l1", now.Add(-time.Hour))
	passLesson(t, s, excluded, "l2", now.Add(-time.Hour))
	passLesson(t, s, database.DemoUserID, "l1", *now)
	passLesson(t, s, s.playReviewUserID, "l1", *now)
	passLesson(t, s, s.accounts["admin@example.com"].ID, "l1", *now)
	if err := s.SetLeaderboardHidden(hidden, true); err != nil {
		t.Fatal(err)
	}
	if err := s.SetLeaderboardExcluded(excluded, true); err != nil {
		t.Fatal(err)
	}
	for _, period := range []string{"week", "all"} {
		page, _ := s.Leaderboard(period)
		if fmt.Sprint(entryNames(page)) != "[Ana]" || page.Participants != 1 {
			t.Fatalf("%s list leaked accounts: %v", period, entryNames(page))
		}
	}
	// Hidden owner still sees where they would be if shown.
	me := s.MyLeaderboard(hidden)
	if !me.Hidden || me.Week.Rank != 1 || me.Week.Participants != 2 {
		t.Fatalf("hidden standing %+v", me)
	}
	// Admin exclusion is silent: the flag is not reported to the user.
	if s.MyLeaderboard(excluded).Hidden || !s.LeaderboardExcluded(excluded) {
		t.Fatal("exclusion must be silent and stored")
	}
	if s.MyLeaderboard(excluded).Week.XP != 100 || s.Gamification(excluded).XP != 100 {
		t.Fatal("exclusion changed XP")
	}
	if err := s.SetLeaderboardHidden(hidden, false); err != nil {
		t.Fatal(err)
	}
	page, _ := s.Leaderboard("week")
	if fmt.Sprint(entryNames(page)) != "[Hidden Ana]" {
		t.Fatalf("unhidden user missing: %v", entryNames(page))
	}
	if err := s.SetLeaderboardHidden("not-a-uuid", true); err != ErrInvalid {
		t.Fatal("invalid user accepted")
	}
	if err := s.SetLeaderboardExcluded(uuid.NewString(), true); err != ErrNotFound {
		t.Fatal("unknown user accepted")
	}
}

func TestLeaderboardNeverShowsEmailNames(t *testing.T) {
	s, now := newLeaderboardStore(t)
	id := addLeaderboardAccount(s, "x")
	s.mu.Lock()
	a := s.accounts["x@example.com"]
	a.DisplayName = "someone@gmail.com"
	s.accounts[a.Email] = a
	s.mu.Unlock()
	passLesson(t, s, id, "l1", *now)
	page, _ := s.Leaderboard("week")
	if page.Entries[0].DisplayName != "" || len(page.Entries[0].Tag) != 4 || page.Entries[0].Tag != leaderboardTag(id) {
		t.Fatalf("email exposed or bad tag: %+v", page.Entries[0])
	}
	if leaderboardTag(id) != leaderboardTag(id) || leaderboardDisplayName("  ") != "" || leaderboardDisplayName("Rina") != "Rina" {
		t.Fatal("name helpers")
	}
}

func TestXPSourcesAndDailyCaps(t *testing.T) {
	s, now := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Rina")
	record := func(id, kind, lesson, exam string, score, seconds int, at time.Time) {
		t.Helper()
		s.mu.Lock()
		defer s.mu.Unlock()
		if _, err := s.recordStudyLocked(user, studyEvent{ID: id, Kind: kind, Subject: "grammar", LessonID: lesson, ExamID: exam, Score: score, Answered: 5, Seconds: seconds, At: at}); err != nil {
			t.Fatal(err)
		}
	}
	record("a", "lesson", "l1", "", 90, 60, *now)
	if xp := s.Gamification(user).XP; xp != 50 {
		t.Fatalf("first pass xp %d", xp)
	}
	// Later completions scoring >= 50% pay 10 XP each.
	for i := 0; i < 5; i++ {
		record(fmt.Sprintf("re%d", i), "lesson", "l1", "", 85, 10, *now)
	}
	if xp := s.Gamification(user).XP; xp != 100 {
		t.Fatalf("repeat xp %d", xp)
	}
	// Below 50% pays nothing (it still counts for the streak).
	record("fail", "lesson", "l2", "", 40, 10, *now)
	if xp := s.Gamification(user).XP; xp != 100 {
		t.Fatalf("failed attempt paid xp %d", xp)
	}
	// Repeats stop at the 100 XP daily cap (50 already used).
	for i := 0; i < 8; i++ {
		record(fmt.Sprintf("more%d", i), "lesson", "l1", "", 60, 10, *now)
	}
	if xp := s.Gamification(user).XP; xp != 150 {
		t.Fatalf("repeat cap: xp %d", xp)
	}
	// Reaching the 20-minute daily goal pays 20 XP once.
	record("long", "lesson", "l3", "", 40, 1200, *now)
	record("long2", "lesson", "l4", "", 40, 600, *now)
	if xp := s.Gamification(user).XP; xp != 170 {
		t.Fatalf("daily goal: xp %d", xp)
	}
	// Exams: first pass 100, a later completion 20 (the cap resets next day).
	tomorrow := now.Add(24 * time.Hour)
	record("e1", "exam", "", "jlpt-n5-simulation-1", 90, 10, tomorrow)
	record("e2", "exam", "", "jlpt-n5-simulation-1", 90, 10, tomorrow)
	if xp := s.Gamification(user).XP; xp != 290 {
		t.Fatalf("exam xp %d", xp)
	}
	// Duplicate event IDs never pay twice.
	record("e2", "exam", "", "jlpt-n5-simulation-1", 90, 10, tomorrow)
	if xp := s.Gamification(user).XP; xp != 290 {
		t.Fatalf("duplicate event paid: %d", xp)
	}
	me := s.MyLeaderboard(user)
	if me.Week.XP != 290 || me.All.XP != 290 {
		t.Fatalf("leaderboard xp %+v", me)
	}
}

func TestAnkiReviewPaysOneXPPerCardUpToThirty(t *testing.T) {
	s, now := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Anki")
	deck := firstFreeAnkiDeck(t, s, user)
	for i := 0; i < 35; i++ {
		if _, err := s.AnkiReview(user, deck, AnkiReviewInput{Index: i, Version: 0, Rating: "good", RequestID: uuid.NewString()}); err != nil {
			t.Fatalf("review %d: %v", i, err)
		}
	}
	if xp := s.Gamification(user).XP; xp != 30 {
		t.Fatalf("anki xp %d, want 30", xp)
	}
	if s.Profile(user).XP != 30 || s.MyLeaderboard(user).Week.XP != 30 {
		t.Fatal("anki XP not reflected in profile/leaderboard")
	}
	*now = now.Add(24 * time.Hour)
	if _, err := s.AnkiReview(user, deck, AnkiReviewInput{Index: 35, Version: 0, Rating: "good", RequestID: uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	if xp := s.Gamification(user).XP; xp != 31 {
		t.Fatalf("next day anki xp %d", xp)
	}
}

func firstFreeAnkiDeck(t *testing.T, s *MemoryStore, user string) string {
	t.Helper()
	deck, err := anki.Get("wakayama")
	if err != nil || deck.CardCount < 40 {
		t.Skip("wakayama deck unavailable")
	}
	return deck.ID
}

func TestWeeklyCloseStoresResultsGrantsAchievementsAndResets(t *testing.T) {
	s, now := newLeaderboardStore(t)
	names := []string{"A", "B", "C", "D"}
	ids := map[string]string{}
	for i, n := range names {
		ids[n] = addLeaderboardAccount(s, n)
		for j := 0; j < len(names)-i; j++ {
			passLesson(t, s, ids[n], fmt.Sprintf("l%d", j), now.Add(time.Duration(i)*time.Minute))
		}
	}
	hidden := addLeaderboardAccount(s, "Hidden")
	for j := 0; j < 9; j++ {
		passLesson(t, s, hidden, fmt.Sprintf("l%d", j), *now)
	}
	if err := s.SetLeaderboardHidden(hidden, true); err != nil {
		t.Fatal(err)
	}
	// Still the same week: nothing to close.
	if err := s.CloseLeaderboardWeeks(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.MyLeaderboard(ids["A"]).LastWeek != nil {
		t.Fatal("closed an open week")
	}
	// Monday 00:00 WIB.
	*now = time.Date(2026, 11, 22, 17, 0, 0, 0, time.UTC)
	if err := s.CloseLeaderboardWeeks(context.Background()); err != nil {
		t.Fatal(err)
	}
	for i, n := range names {
		me := s.MyLeaderboard(ids[n])
		if me.LastWeek == nil || me.LastWeek.Rank != i+1 || me.LastWeek.XP != (len(names)-i)*50 || me.LastWeek.WeekStart != "2026-11-16" {
			t.Fatalf("%s last week %+v", n, me.LastWeek)
		}
		if me.Week.XP != 0 || me.Week.Rank != 0 || me.All.XP != (len(names)-i)*50 {
			t.Fatalf("%s new week not reset: %+v", n, me)
		}
	}
	if s.MyLeaderboard(hidden).LastWeek != nil {
		t.Fatal("hidden user got a final rank")
	}
	unlocked := func(user, id string) bool {
		return achievementByID(t, s.Gamification(user), id).UnlockedAt != nil
	}
	if !unlocked(ids["A"], achievementWeeklyChampion) || !unlocked(ids["A"], achievementWeeklyPodium) {
		t.Fatal("winner missing achievements")
	}
	if !unlocked(ids["C"], achievementWeeklyPodium) || unlocked(ids["C"], achievementWeeklyChampion) {
		t.Fatal("third place achievements wrong")
	}
	if unlocked(ids["D"], achievementWeeklyPodium) || unlocked(hidden, achievementWeeklyPodium) {
		t.Fatal("fourth place or hidden user got podium")
	}
	page, _ := s.Leaderboard("week")
	if len(page.Entries) != 0 {
		t.Fatalf("new week should start empty: %v", entryNames(page))
	}
	// Running again is a no-op and keeps the first unlock time.
	first := *achievementByID(t, s.Gamification(ids["A"]), achievementWeeklyChampion).UnlockedAt
	if err := s.CloseLeaderboardWeeks(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !achievementByID(t, s.Gamification(ids["A"]), achievementWeeklyChampion).UnlockedAt.Equal(first) {
		t.Fatal("re-run changed achievements")
	}
	// The card is only for the week that just ended.
	*now = now.Add(8 * 24 * time.Hour)
	if err := s.CloseLeaderboardWeeks(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.MyLeaderboard(ids["A"]).LastWeek != nil {
		t.Fatal("stale last-week result shown")
	}
}

func TestUpdateProfileNameSurvivesGoogleSignIn(t *testing.T) {
	s := NewMemoryStore()
	identity := domain.GoogleIdentity{Subject: "sub-1", Email: "rina@example.com", DisplayName: "Rina Google"}
	profile, err := s.AuthenticateGoogle(identity)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateProfile(profile.ID, "Rina Sakura", "JLPT N4"); err != nil {
		t.Fatal(err)
	}
	again, err := s.AuthenticateGoogle(identity)
	if err != nil {
		t.Fatal(err)
	}
	if again.DisplayName != "Rina Sakura" || s.Profile(profile.ID).DisplayName != "Rina Sakura" {
		t.Fatalf("display name reverted to %q", again.DisplayName)
	}
}

func TestLeaderboardPersistsAcrossRestart(t *testing.T) {
	path := testdb.DSN(t)
	s := NewMemoryStore()
	now := leaderboardClock
	s.gamificationNow = func() time.Time { return now }
	if err := s.EnableGamificationPersistence(path); err != nil {
		t.Fatal(err)
	}
	ana, budi := uuid.NewString(), uuid.NewString()
	passLesson(t, s, ana, "l1", now)
	passLesson(t, s, ana, "l2", now)
	passLesson(t, s, budi, "l1", now)
	if err := s.SetLeaderboardHidden(budi, true); err != nil {
		t.Fatal(err)
	}
	var events int64
	s.ormDB.Model(&database.XPEvent{}).Where("user_id = ?", ana).Count(&events)
	if events != 2 {
		t.Fatalf("xp events %d", events)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := NewMemoryStore()
	reopened.gamificationNow = func() time.Time { return now }
	if err := reopened.EnableGamificationPersistence(path); err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	page, _ := reopened.Leaderboard("week")
	if page.Participants != 1 || page.Entries[0].XP != 100 {
		t.Fatalf("restored week %+v", page)
	}
	if !reopened.MyLeaderboard(budi).Hidden {
		t.Fatal("hidden flag lost")
	}
	// First run creates the schedule, the next due run closes the week.
	if err := reopened.CloseLeaderboardWeeks(context.Background()); err != nil {
		t.Fatal(err)
	}
	now = time.Date(2026, 11, 22, 17, 0, 0, 0, time.UTC)
	if err := reopened.CloseLeaderboardWeeks(context.Background()); err != nil {
		t.Fatal(err)
	}
	if last := reopened.MyLeaderboard(ana).LastWeek; last == nil || last.Rank != 1 {
		t.Fatalf("persistent close: %+v", last)
	}
	var results int64
	reopened.ormDB.Model(&database.WeeklyResult{}).Count(&results)
	if results != 1 {
		t.Fatalf("weekly results %d", results)
	}
}
