package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/domain"
)

func achievementsByID(items []domain.Achievement) map[string]domain.Achievement {
	byID := map[string]domain.Achievement{}
	for _, a := range items {
		byID[a.ID] = a
	}
	return byID
}

func TestAchievementCatalogIsAboutFiftyAndWellFormed(t *testing.T) {
	items := gamificationAchievements(newGamificationState(), leaderboardClock)
	if len(items) < 50 || len(items) > 55 {
		t.Fatalf("%d achievements, want about 50", len(items))
	}
	seen := map[string]bool{}
	for _, a := range items {
		if seen[a.ID] {
			t.Errorf("duplicate achievement id %q", a.ID)
		}
		seen[a.ID] = true
		if a.Name == "" || a.Description == "" || a.Icon == "" || a.Target <= 0 {
			t.Errorf("incomplete achievement %+v", a)
		}
		if a.Progress != 0 || a.UnlockedAt != nil {
			t.Errorf("new account already has progress on %q: %+v", a.ID, a)
		}
	}
}

func TestLessonLevelReadsJLPTLevelFromIDs(t *testing.T) {
	cases := map[string]string{
		"grammar-n4-01": "N4", "kanji-n3-12": "N3", "reading-n2-01": "N2",
		"n5-daily": "N5", "n1-analysis": "N1", "ssw-kaigo-01": "", "vocabulary-web": "",
		"grammar-n6-01": "", "Grammar-N5-02": "N5",
	}
	for id, want := range cases {
		if got := lessonLevel(id); got != want {
			t.Errorf("lessonLevel(%q) = %q, want %q", id, got, want)
		}
	}
}

func day(offset int) string {
	return time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, offset).Format("2006-01-02")
}

func TestAchievementProgressComesFromSavedState(t *testing.T) {
	state := newGamificationState()
	state.XP = 2600
	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("grammar-n4-%02d", i)
		state.Lessons[id] = domain.LessonProgress{LessonID: id, Completed: true, Completion: 90}
		state.Subjects[id] = "grammar"
	}
	for i := 0; i < 12; i++ {
		id := fmt.Sprintf("kanji-n5-%02d", i)
		state.Lessons[id] = domain.LessonProgress{LessonID: id, Completed: true, Completion: 100}
		state.Subjects[id] = "kanji"
	}
	state.Lessons["n1-analysis"] = domain.LessonProgress{LessonID: "n1-analysis", Completed: false, Completion: 100}
	state.Rewards["exam:a"], state.Rewards["exam:b"], state.Rewards["lesson:x"] = true, true, true
	state.Exams["a"], state.Exams["b"], state.Exams["c"] = true, true, false
	// Three days in a row, then a break of seven days, then one more day.
	state.Days[day(0)], state.Days[day(1)], state.Days[day(2)] = 1500, 3700, 600
	state.Days[day(10)] = 60
	state.Anki = map[string]*AnkiDeckState{
		"d1": {Cards: map[int]AnkiCardState{0: {Version: 3}, 1: {Version: 2}}},
		"d2": {Cards: map[int]AnkiCardState{0: {Version: 5}}},
	}
	got := achievementsByID(gamificationAchievements(state, time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)))
	want := map[string]int{
		"lessons-25": 25, "lessons-100": 42, "perfect-1": 1, "perfect-10": 10,
		"grammar-50": 30, "kanji-50": 12, "vocabulary-50": 0,
		"level-n4": 10, "level-n5": 10, "level-n1": 0,
		"streak-3": 3, "streak-14": 3,
		"xp-100": 100, "xp-2500": 2500, "xp-5000": 2600,
		"days-10": 4, "time-1h": 60, "goal-1": 1, "goal-7": 2, "marathon-1h": 60,
		"comeback": 1, "exams-3": 3, "exams-passed-5": 2, "anki-50": 10,
	}
	for id, progress := range want {
		a, ok := got[id]
		if !ok {
			t.Errorf("missing achievement %q", id)
			continue
		}
		if a.Progress != progress && !(progress > a.Target && a.Progress == a.Target) {
			t.Errorf("%s progress %d (target %d), want %d", id, a.Progress, a.Target, progress)
		}
	}
	// 1500+3700+600+60 seconds is 5860 s, i.e. 97 minutes; time-1h caps at its target.
	if got["time-1h"].Progress != 60 || got["time-10h"].Progress != 97 {
		t.Errorf("study time: %+v %+v", got["time-1h"], got["time-10h"])
	}
	unlockAchievements(state, time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC))
	for _, id := range []string{"lessons-25", "perfect-1", "perfect-10", "streak-3", "xp-2500", "comeback", "goal-1", "marathon-1h", "level-n4"} {
		if _, ok := state.Unlocked[id]; !ok {
			t.Errorf("%s should be unlocked", id)
		}
	}
	for _, id := range []string{"lessons-100", "streak-14", "xp-5000", "level-n1", "goal-7", "anki-50", "days-30"} {
		if _, ok := state.Unlocked[id]; ok {
			t.Errorf("%s should still be locked", id)
		}
	}
}

func TestComebackNeedsAWeekOffNotLess(t *testing.T) {
	check := func(gap int) bool {
		state := newGamificationState()
		state.Days[day(0)], state.Days[day(gap)] = 60, 60
		return achievementsByID(gamificationAchievements(state, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)))["comeback"].Progress == 1
	}
	// gap is the difference in days; a rest of 7 full days means a gap of 8.
	if check(7) || !check(8) || check(1) {
		t.Fatalf("comeback boundary wrong: gap7=%v gap8=%v gap1=%v", check(7), check(8), check(1))
	}
}

func TestWeeklyTop10AchievementOnlyForTheTopTen(t *testing.T) {
	s, now := newLeaderboardStore(t)
	ids := make([]string, 12)
	for i := range ids {
		ids[i] = addLeaderboardAccount(s, fmt.Sprintf("U%02d", i))
		for j := 0; j < 12-i; j++ {
			passLesson(t, s, ids[i], fmt.Sprintf("l%d", j), *now)
		}
	}
	*now = time.Date(2026, 11, 22, 17, 0, 0, 0, time.UTC)
	if err := s.CloseLeaderboardWeeks(context.Background()); err != nil {
		t.Fatal(err)
	}
	has := func(user, id string) bool {
		return achievementByID(t, s.Gamification(user), id).UnlockedAt != nil
	}
	for rank, user := range ids {
		wantTop10, wantPodium := rank < 10, rank < 3
		if has(user, achievementWeeklyTop10) != wantTop10 || has(user, achievementWeeklyPodium) != wantPodium {
			t.Errorf("rank %d: top10=%v podium=%v", rank+1, has(user, achievementWeeklyTop10), has(user, achievementWeeklyPodium))
		}
	}
}

func TestAnkiReviewAchievementsUnlockEvenAfterTheDailyXPCap(t *testing.T) {
	s, _ := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Anki50")
	deck, err := anki.Get("wakayama")
	if err != nil || deck.CardCount < 55 {
		t.Skip("wakayama deck needs 55 cards")
	}
	for i := 0; i < 55; i++ {
		if _, err := s.AnkiReview(user, deck.ID, AnkiReviewInput{Index: i, Rating: "good", RequestID: uuid.NewString()}); err != nil {
			t.Fatalf("review %d: %v", i, err)
		}
	}
	summary := s.Gamification(user)
	if summary.XP != 30 {
		t.Fatalf("anki xp %d, want the 30 XP daily cap", summary.XP)
	}
	if achievementByID(t, summary, "anki-50").UnlockedAt == nil {
		t.Fatal("anki-50 must unlock at card 50 even though the XP cap was reached")
	}
	if achievementByID(t, summary, "anki-500").UnlockedAt != nil {
		t.Fatal("anki-500 unlocked too early")
	}
}
