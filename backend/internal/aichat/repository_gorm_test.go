package aichat_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"nihongo-master-backend/internal/aichat"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/testdb"
)

const testUser = "11111111-1111-4111-8111-111111111111"

func newGormRepo(t *testing.T) *aichat.GormRepository {
	t.Helper()
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	repo, err := aichat.NewGormRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestGormQuotaIsAtomicAcrossConcurrentRequests(t *testing.T) {
	repo := newGormRepo(t)
	q := aichat.Quota{Repo: repo, Config: aichat.DefaultConfig}
	now := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	granted := 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := q.Reserve(context.Background(), testUser, false, true, now); err == nil {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if granted != 3 {
		t.Fatalf("granted %d, want 3", granted)
	}
}

func TestGormAdCreditReleaseAndCap(t *testing.T) {
	repo := newGormRepo(t)
	ctx := context.Background()
	cfg := aichat.DefaultConfig()
	cfg.FreePerDay, cfg.AdPerDay = 0, 1
	q := aichat.Quota{Repo: repo, Config: func() aichat.Config { return cfg }}
	now := time.Now()
	if _, err := q.Reserve(ctx, testUser, false, true, now); err != aichat.ErrAdRequired {
		t.Fatalf("err = %v", err)
	}
	if err := q.GrantAdCredit(ctx, testUser, false, now); err != nil {
		t.Fatal(err)
	}
	if err := q.GrantAdCredit(ctx, testUser, false, now); err != aichat.ErrAdCapReached {
		t.Fatalf("cap: %v", err)
	}
	r, err := q.Reserve(ctx, testUser, false, true, now)
	if err != nil || r.Slot != aichat.SlotAd {
		t.Fatalf("reserve ad: %v %v", r, err)
	}
	if err := q.Release(ctx, r); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Reserve(ctx, testUser, false, true, now); err != nil {
		t.Fatalf("released credit must be reusable: %v", err)
	}
}

func TestGormConversationHistoryAndCache(t *testing.T) {
	repo := newGormRepo(t)
	ctx := context.Background()
	if err := repo.AppendMessages(ctx,
		aichat.Message{UserID: testUser, LessonKey: "grammar:a", Role: aichat.RoleUser, Content: "q1"},
		aichat.Message{UserID: testUser, LessonKey: "grammar:a", Role: aichat.RoleAssistant, Content: "a1"},
	); err != nil {
		t.Fatal(err)
	}
	got, err := repo.RecentMessages(ctx, testUser, "grammar:a", 8)
	if err != nil || len(got) != 2 || got[0].Content != "q1" || got[1].Content != "a1" {
		t.Fatalf("history = %+v, %v", got, err)
	}
	if err := repo.HideConversation(ctx, testUser, "grammar:a"); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.RecentMessages(ctx, testUser, "grammar:a", 8); len(got) != 0 {
		t.Fatalf("hidden conversation visible: %+v", got)
	}
	if n, err := repo.PurgeBefore(ctx, time.Now().Add(time.Hour)); err != nil || n != 2 {
		t.Fatalf("purge = %d, %v", n, err)
	}
	if err := repo.StoreAnswer(ctx, "k", "first"); err != nil {
		t.Fatal(err)
	}
	if err := repo.StoreAnswer(ctx, "k", "second"); err != nil {
		t.Fatal(err)
	}
	if answer, ok, err := repo.CachedAnswer(ctx, "k"); err != nil || !ok || answer != "second" {
		t.Fatalf("cache = %q %v %v", answer, ok, err)
	}
	if n, err := repo.ClearAnswers(ctx); err != nil || n != 1 {
		t.Fatalf("clear = %d %v", n, err)
	}
}
