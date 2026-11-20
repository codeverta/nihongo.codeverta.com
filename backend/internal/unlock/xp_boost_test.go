package unlock_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"nihongo-master-backend/internal/unlock"
)

func scenarioXPBoostIntentIsConsumedOnce(t *testing.T, h *harness) {
	ctx := context.Background()
	key, ok := unlock.Key(unlock.KindXPBoost, "grammar:attempt-1")
	if !ok {
		t.Fatal("completion IDs with colons must be valid boost items")
	}
	// Double XP has its own limit: a full unlock day does not block it.
	h.limits.DailyCap = 0
	intent, err := h.service.CreateIntent(ctx, userA, key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.ConsumeVerified(ctx, userA, intent.Nonce); !errors.Is(err, unlock.ErrNotVerified) {
		t.Fatalf("consume before Google confirmed = %v", err)
	}
	if err := h.service.HandleCallback(ctx, h.callback(userA, intent.Nonce, "tx-boost", adUnit)); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.ConsumeVerified(ctx, userB, intent.Nonce); !errors.Is(err, unlock.ErrNotFound) {
		t.Fatalf("another account consumed it: %v", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	won, consumed := 0, 0
	run(10, func(int) {
		_, err := h.service.ConsumeVerified(ctx, userA, intent.Nonce)
		mu.Lock()
		defer mu.Unlock()
		switch {
		case err == nil:
			won++
		case errors.Is(err, unlock.ErrConsumed):
			consumed++
		default:
			t.Errorf("unexpected error: %v", err)
		}
	})
	wg.Wait()
	if won != 1 || consumed != 9 {
		t.Fatalf("won %d, already-consumed %d; want 1 and 9", won, consumed)
	}
	if h.service.Allowed(ctx, userA, key) {
		t.Fatal("a boost must never become an unlock")
	}
	if usage := h.usage(t, userA); usage.Used != 0 {
		t.Fatalf("a boost must not count toward the unlock limit: %+v", usage)
	}
}

func TestBoostCallbackAcceptsEitherConfiguredUnit(t *testing.T) {
	h := newHarness(t, unlock.NewMemoryRepository())
	ctx := context.Background()
	key, _ := unlock.Key(unlock.KindXPBoost, "simulation:simulation-abc")
	intent, err := h.service.CreateIntent(ctx, userA, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.service.HandleCallback(ctx, h.callback(userA, intent.Nonce, "tx-1", "7777777777")); err != nil {
		t.Fatalf("callback from the second configured unit: %v", err)
	}
	if err := h.service.HandleCallback(ctx, h.callback(userA, intent.Nonce, "tx-2", "1234567890")); !errors.Is(err, unlock.ErrInvalid) {
		t.Fatalf("callback from an unknown unit = %v", err)
	}
}
