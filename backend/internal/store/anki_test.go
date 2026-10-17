package store

import (
	"errors"
	"github.com/google/uuid"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/testdb"
	"sync"
	"testing"
	"time"
)

func TestAnkiSchedulingOwnershipConflictsAndConcurrentRetries(t *testing.T) {
	s := NewMemoryStore()
	now := time.Date(2026, 10, 17, 8, 0, 0, 0, time.UTC)
	s.gamificationNow = func() time.Time { return now }
	user := database.DemoUserID
	first, err := s.AnkiStudy(user, "wakayama", 4)
	if err != nil || len(first.Cards) != 4 {
		t.Fatal(first, err)
	}
	for i, grade := range []string{"again", "hard", "good", "easy"} {
		input := AnkiReviewInput{Index: i, Rating: grade, RequestID: uuid.NewString()}
		result, err := s.AnkiReview(user, "wakayama", input)
		if err != nil {
			t.Fatal(err)
		}
		interval := []time.Duration{time.Minute, 6 * time.Hour, 24 * time.Hour, 4 * 24 * time.Hour}[i]
		if result.Due.Sub(now) != interval || result.Version != 1 {
			t.Fatal("bad schedule", grade, result)
		}
		var group sync.WaitGroup
		for n := 0; n < 4; n++ {
			group.Add(1)
			go func() {
				defer group.Done()
				again, err := s.AnkiReview(user, "wakayama", input)
				if err != nil || again.Version != 1 || !again.Due.Equal(result.Due) {
					t.Error("retry changed schedule", err)
				}
			}()
		}
		group.Wait()
		input.RequestID = uuid.NewString()
		if _, err = s.AnkiReview(user, "wakayama", input); !errors.Is(err, ErrAnkiConflict) {
			t.Fatal("stale review accepted")
		}
	}
	view, err := s.AnkiDeck(user, "wakayama")
	if err != nil || view.Stats.Seen != 4 || view.Stats.Due != 0 {
		t.Fatal(view.Stats, err)
	}
	if s.Gamification(user).XP != 0 {
		t.Fatal("self-ratings must not award graded-quiz XP")
	}
	other, err := s.AuthenticateGoogle(domain.GoogleIdentity{Subject: "anki-other", Email: "anki-other@example.test", DisplayName: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	otherView, _ := s.AnkiDeck(other.ID, "wakayama")
	if otherView.Stats.Seen != 0 {
		t.Fatal("progress leaked between accounts")
	}
	now = now.Add(2 * time.Minute)
	due, err := s.AnkiStudy(user, "wakayama", 4)
	if err != nil || due.Cards[0].Index != 0 || due.Cards[0].Version != 1 {
		t.Fatal("due review must precede new cards", err)
	}
	if _, err = s.AnkiStudy(user, "n3", 20); !errors.Is(err, ErrPremium) {
		t.Fatal("premium deck exposed")
	}
	if _, err = s.AnkiReview(user, "wakayama", AnkiReviewInput{Index: 99, Rating: "easy", RequestID: uuid.NewString()}); !errors.Is(err, ErrAnkiConflict) {
		t.Fatal("out-of-queue new card accepted")
	}
	if _, err = s.AnkiReview(user, "wakayama", AnkiReviewInput{Index: 4, Rating: "made-up", RequestID: uuid.NewString()}); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid rating accepted")
	}
	if err = s.DeleteAccount(user); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.gamification[user]; ok {
		t.Fatal("deleted user's reviews retained")
	}
}

func TestAnkiMariaDBPersistenceAndFailedWrites(t *testing.T) {
	dsn := testdb.DSN(t)
	user := database.DemoUserID
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	input := AnkiReviewInput{Index: 0, Rating: "good", RequestID: uuid.NewString()}
	first, err := s.AnkiReview(user, "n5", input)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	restored := NewMemoryStore()
	if err = restored.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	view, err := restored.AnkiDeck(user, "n5")
	if err != nil || view.Stats.Seen != 1 {
		t.Fatal("review did not survive restart", err)
	}
	retry, err := restored.AnkiReview(user, "n5", input)
	if err != nil || retry.Version != first.Version || !retry.Due.Equal(first.Due) {
		t.Fatal("retry after restart was not idempotent", err)
	}
	if err = restored.persistenceConn.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = restored.AnkiReview(user, "n5", AnkiReviewInput{Index: 1, Rating: "easy", RequestID: uuid.NewString()}); err == nil {
		t.Fatal("database failure was hidden")
	}
	view, _ = restored.AnkiDeck(user, "n5")
	if view.Stats.Seen != 1 {
		t.Fatal("failed write changed live progress")
	}
}
