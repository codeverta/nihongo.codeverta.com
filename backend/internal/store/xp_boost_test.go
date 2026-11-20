package store

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// Double XP rules (docs/ads-xp-spec.md, phase C).

func finishLesson(t *testing.T, s *MemoryStore, user, id, lesson string, score int, at time.Time) string {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	gain, err := s.recordStudyLocked(user, studyEvent{ID: id, Kind: "lesson", Subject: "grammar", LessonID: lesson, Score: score, Answered: 5, At: at})
	if err != nil {
		t.Fatal(err)
	}
	return gain.CompletionID
}

func TestDoubleXPDoublesOneCompletionOnce(t *testing.T) {
	s, now := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Rina")
	id := finishLesson(t, s, user, "grammar:a1", "l1", 90, *now)
	if kind, err := s.XPBoostable(user, id, 0); err != nil || kind != "lessons" {
		t.Fatalf("boostable = %q, %v", kind, err)
	}
	gain, err := s.BoostXP(user, id, 0)
	if err != nil || gain.Awarded != 50 || gain.Total != 100 || gain.Kind != "lessons" {
		t.Fatalf("boost = %+v, %v", gain, err)
	}
	if _, err := s.BoostXP(user, id, 0); !errors.Is(err, ErrBoostUsed) {
		t.Fatalf("second boost = %v", err)
	}
	if s.Gamification(user).XP != 100 || s.MyLeaderboard(user).Week.XP != 100 {
		t.Fatal("the bonus must reach the profile and the weekly leaderboard")
	}
}

func TestDoubleXPNeedsXPAndAFreshCompletion(t *testing.T) {
	s, now := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Sari")
	nothing := finishLesson(t, s, user, "grammar:low", "l1", 30, *now)
	if _, err := s.XPBoostable(user, nothing, 0); !errors.Is(err, ErrNotFound) {
		t.Fatalf("a completion that earned nothing = %v", err)
	}
	if _, err := s.BoostXP(user, "grammar:unknown", 0); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown completion = %v", err)
	}
	id := finishLesson(t, s, user, "grammar:ok", "l2", 90, *now)
	*now = now.Add(31 * time.Minute)
	if _, err := s.BoostXP(user, id, 0); !errors.Is(err, ErrBoostExpired) {
		t.Fatalf("after 31 minutes = %v", err)
	}
	if _, err := s.BoostXP(addLeaderboardAccount(s, "Other"), id, 0); !errors.Is(err, ErrNotFound) {
		t.Fatalf("another account boosted it: %v", err)
	}
}

func TestDoubleXPHasADailyLimit(t *testing.T) {
	s, now := newLeaderboardStore(t)
	user := addLeaderboardAccount(s, "Cap")
	for i := 0; i < 3; i++ {
		id := finishLesson(t, s, user, fmt.Sprintf("grammar:%d", i), fmt.Sprintf("l%d", i), 90, *now)
		_, err := s.BoostXP(user, id, 2)
		if i < 2 && err != nil {
			t.Fatalf("boost %d: %v", i, err)
		}
		if i == 2 && !errors.Is(err, ErrBoostLimit) {
			t.Fatalf("third boost = %v", err)
		}
	}
	// The limit resets the next day.
	*now = now.Add(24 * time.Hour)
	id := finishLesson(t, s, user, "grammar:next", "l9", 90, *now)
	if _, err := s.BoostXP(user, id, 2); err != nil {
		t.Fatalf("next day: %v", err)
	}
}
