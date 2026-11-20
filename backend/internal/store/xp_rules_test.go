package store

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/domain"
)

// Rules from docs/ads-xp-spec.md (phase B).

func TestCompletionsReturnWhatTheyEarned(t *testing.T) {
	s, now := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Rina")
	s.mu.Lock()
	first, err := s.recordStudyLocked(user, studyEvent{ID: "a", Kind: "lesson", Subject: "grammar", LessonID: "l1", Score: 90, Answered: 5, At: *now})
	again, err2 := s.recordStudyLocked(user, studyEvent{ID: "a", Kind: "lesson", Subject: "grammar", LessonID: "l1", Score: 90, Answered: 5, At: *now})
	s.mu.Unlock()
	if err != nil || err2 != nil {
		t.Fatal(err, err2)
	}
	if first.Awarded != 50 || first.Total != 50 || first.Streak != 1 || first.CompletionID != "a" {
		t.Fatalf("first = %+v", first)
	}
	if again.Awarded != 0 || again.Total != 50 {
		t.Fatalf("a repeated event must earn nothing: %+v", again)
	}
}

func TestSSWPracticeEarnsXPOnlyWhenComplete(t *testing.T) {
	s, now := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Sari")
	s.mu.Lock()
	partial, _ := s.recordStudyLocked(user, studyEvent{ID: "p1", Kind: "ssw-practice", Answered: 3, Complete: false, At: *now})
	full, _ := s.recordStudyLocked(user, studyEvent{ID: "p2", Kind: "ssw-practice", Answered: 10, Complete: true, At: *now})
	s.mu.Unlock()
	if partial.Awarded != 0 || partial.Streak != 1 {
		t.Fatalf("a partial practice counts for the streak only: %+v", partial)
	}
	if full.Awarded != xpSSWPractice {
		t.Fatalf("full practice = %+v", full)
	}
}

func TestPracticeLessonsPayOnceAndCountForTheStreak(t *testing.T) {
	s, _ := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Kaigo")
	first, err := s.CompletePracticeLesson(user, "nursing-care", "kaigo-01")
	if err != nil || first.Awarded != xpPracticeLesson || first.Streak != 1 {
		t.Fatalf("first = %+v, %v", first, err)
	}
	again, err := s.CompletePracticeLesson(user, "nursing-care", "kaigo-01")
	if err != nil || again.Awarded != 0 || again.Total != xpPracticeLesson {
		t.Fatalf("again = %+v, %v", again, err)
	}
	for _, bad := range [][2]string{{"Nursing", "x"}, {"nursing-care", "../x"}, {"", "x"}} {
		if _, err := s.CompletePracticeLesson(user, bad[0], bad[1]); !errors.Is(err, ErrInvalid) {
			t.Fatalf("%v: %v", bad, err)
		}
	}
	if _, err := s.CompletePracticeLesson(uuid.NewString(), "nursing-care", "kaigo-01"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown account: %v", err)
	}
}

func TestAnkiDayCountsForTheStreakAfterEnoughReviews(t *testing.T) {
	s, _ := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Anki")
	deck := firstFreeAnkiDeck(t, s, user)
	s.SetXPRules(0, 3)
	for i := 0; i < 2; i++ {
		if _, err := s.AnkiReview(user, deck, AnkiReviewInput{Index: i, Version: 0, Rating: "good", RequestID: uuid.NewString()}); err != nil {
			t.Fatal(err)
		}
	}
	if streak := s.Gamification(user).StreakDays; streak != 0 {
		t.Fatalf("streak after 2 of 3 reviews = %d", streak)
	}
	if _, err := s.AnkiReview(user, deck, AnkiReviewInput{Index: 2, Version: 0, Rating: "good", RequestID: uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	if streak := s.Gamification(user).StreakDays; streak != 1 {
		t.Fatalf("streak after 3 reviews = %d", streak)
	}
}

func TestAdminCanChangeTheRepeatCap(t *testing.T) {
	s, now := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Cap")
	s.SetXPRules(25, 0)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.recordStudyLocked(user, studyEvent{ID: "first", Kind: "lesson", LessonID: "l1", Score: 90, Answered: 5, At: *now}); err != nil {
		t.Fatal(err)
	}
	var last domain.XPGain
	for i := 0; i < 4; i++ {
		last, _ = s.recordStudyLocked(user, studyEvent{ID: string(rune('a' + i)), Kind: "lesson", LessonID: "l1", Score: 70, Answered: 5, At: *now})
	}
	if last.Total != 75 || last.Awarded != 0 {
		t.Fatalf("50 + a 25 XP cap = 75, got %+v", last)
	}
}

func TestOfflineCompletionIsBounded(t *testing.T) {
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	at := func(t time.Time) *time.Time { return &t }
	cases := []struct {
		name     string
		input    domain.OfflinePracticeSyncInput
		wantAt   time.Time
		wantSecs int
	}{
		{"no device data", domain.OfflinePracticeSyncInput{}, now, 0},
		{"yesterday", domain.OfflinePracticeSyncInput{CompletedAt: at(now.Add(-24 * time.Hour)), DurationSeconds: 300}, now.Add(-24 * time.Hour), 300},
		{"slightly ahead clock", domain.OfflinePracticeSyncInput{CompletedAt: at(now.Add(2 * time.Minute))}, now, 0},
		{"far future", domain.OfflinePracticeSyncInput{CompletedAt: at(now.Add(time.Hour))}, now, 0},
		{"too old", domain.OfflinePracticeSyncInput{CompletedAt: at(now.Add(-8 * 24 * time.Hour))}, now, 0},
		{"too long", domain.OfflinePracticeSyncInput{DurationSeconds: 99999}, now, 7200},
		{"negative", domain.OfflinePracticeSyncInput{DurationSeconds: -5}, now, 0},
	}
	for _, tc := range cases {
		gotAt, gotSecs := offlineCompletion(tc.input, now)
		if !gotAt.Equal(tc.wantAt) || gotSecs != tc.wantSecs {
			t.Errorf("%s: got %v, %d", tc.name, gotAt, gotSecs)
		}
	}
}
