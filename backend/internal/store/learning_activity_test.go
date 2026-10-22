package store

import (
	"errors"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/testdb"
	"testing"
	"time"
)

func TestLearningHistoryOrderingIsolationValidationAndScores(t *testing.T) {
	s := NewMemoryStore()
	user := database.DemoUserID
	now := time.Now().UTC().Add(-time.Minute)
	id := s.grammar[0].ID
	first, err := s.SaveLearningActivity(user, "grammar", id, 1, now)
	if err != nil || first.Position != 1 || first.Title != s.grammar[0].Title {
		t.Fatal(first, err)
	}
	if _, err = s.SaveLearningActivity(user, "grammar", id, 0, now.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if items := s.LearningHistory(user); len(items) != 1 || items[0].Position != 1 {
		t.Fatal("stale write changed position", items)
	}
	if _, err = s.SaveLearningActivity(user, "reading", s.reading[0].ID, 0, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if items := s.LearningHistory(user); len(items) != 2 || items[0].Kind != "reading" {
		t.Fatal(items)
	}
	if len(s.LearningHistory("other")) != 0 {
		t.Fatal("history leaked")
	}
	for _, tc := range []struct {
		kind, id string
		position int
		at       time.Time
		expected error
	}{
		{"grammar", id, -1, now, ErrInvalid},
		{"grammar", id, len(s.grammar[0].Points), now, ErrInvalid},
		{"grammar", id, 0, time.Now().Add(time.Hour), ErrInvalid},
		{"grammar", id, 0, time.Time{}, ErrInvalid},
		{"kamus", id, 0, now, ErrNotFound},
		{"grammar", "missing", 0, now, ErrNotFound},
	} {
		if _, err = s.SaveLearningActivity(user, tc.kind, tc.id, tc.position, tc.at); !errors.Is(err, tc.expected) {
			t.Fatal(tc, err)
		}
	}
	for _, lesson := range s.grammar {
		if PremiumLevel(lesson.Level) {
			if _, err = s.SaveLearningActivity(user, "grammar", lesson.ID, 0, now); !errors.Is(err, ErrPremium) {
				t.Fatal("premium bypass", err)
			}
			break
		}
	}
	if s.Gamification(user).XP != 0 || len(s.Progress(user)) != 0 {
		t.Fatal("navigation awarded progress or XP")
	}
	passStudyLesson(t, s, user, "grammar", id)
	if items := s.LearningHistory(user); !items[1].Completed {
		t.Fatal("completion was not derived from graded progress", items)
	}
	s.grammar = s.grammar[1:]
	if items := s.LearningHistory(user); items[1].Available {
		t.Fatal("removed material was not flagged", items)
	}
}

func TestLearningHistoryMariaDBRestartAndDeletion(t *testing.T) {
	dsn := testdb.DSN(t)
	user := database.DemoUserID
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	id := s.grammar[0].ID
	if _, err := s.SaveLearningActivity(user, "grammar", id, 1, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	restored := NewMemoryStore()
	if err := restored.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	if items := restored.LearningHistory(user); len(items) != 1 || items[0].Position != 1 {
		t.Fatal(items)
	}
	if err := restored.DeleteAccount(user); err != nil {
		t.Fatal(err)
	}
	if len(restored.LearningHistory(user)) != 0 {
		t.Fatal("deleted history retained")
	}
}
