package unlock_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"nihongo-master-backend/internal/unlock"
)

// These scenarios hammer the repository from many goroutines at once. On
// MariaDB they run with the retry switched off (see TestGormRepository), so a
// single deadlock or duplicate-key error fails them.

// results collects the outcome of concurrent calls.
type results struct {
	mu      sync.Mutex
	granted []unlock.Unlock
	errs    []error
}

func (r *results) add(u unlock.Unlock, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err != nil {
		r.errs = append(r.errs, err)
		return
	}
	r.granted = append(r.granted, u)
}

// onlyExpected fails the test for any error that is not an expected limit.
func (r *results) onlyExpected(t *testing.T, allowed ...error) {
	t.Helper()
	for _, err := range r.errs {
		ok := false
		for _, want := range allowed {
			if errors.Is(err, want) {
				ok = true
			}
		}
		if !ok {
			t.Fatalf("unexpected error under concurrency: %v", err)
		}
	}
}

func run(n int, call func(i int)) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start // release every goroutine at the same moment
			call(i)
		}(i)
	}
	close(start)
	wg.Wait()
}

func (h *harness) verifiedIntents(t *testing.T, user string, keys []string) []unlock.Intent {
	t.Helper()
	intents := make([]unlock.Intent, 0, len(keys))
	for _, key := range keys {
		intent, err := h.service.CreateIntent(context.Background(), user, key)
		if err != nil {
			t.Fatal(err)
		}
		if err := h.service.HandleCallback(context.Background(), h.callback(user, intent.Nonce, "tx-"+intent.Nonce, adUnit)); err != nil {
			t.Fatal(err)
		}
		intents = append(intents, intent)
	}
	return intents
}

func keys(prefix string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("grammar:%s-%02d", prefix, i)
	}
	return out
}

func (h *harness) usage(t *testing.T, user string) unlock.Usage {
	t.Helper()
	usage, err := h.repo.Usage(context.Background(), user, unlock.Day(h.now))
	if err != nil {
		t.Fatal(err)
	}
	return usage
}

func scenarioManyConcurrentClaimsOnlyEverHitTheCap(t *testing.T, h *harness) {
	h.limits.DailyCap = 7
	intents := h.verifiedIntents(t, userA, keys("many", 40))
	var r results
	run(len(intents), func(i int) { r.add(h.service.Claim(context.Background(), userA, intents[i].Nonce)) })
	r.onlyExpected(t, unlock.ErrCapReached)
	if len(r.granted) != 7 {
		t.Fatalf("granted %d, want exactly 7", len(r.granted))
	}
	if used := h.usage(t, userA).Used; used != 7 {
		t.Fatalf("counter = %d, want 7", used)
	}
	listed, _ := h.repo.List(context.Background(), userA)
	if len(listed) != 7 {
		t.Fatalf("stored unlocks = %d, want 7", len(listed))
	}
}

func scenarioClaimsAndFallbacksRaceForTheSameLesson(t *testing.T, h *harness) {
	const key = "grammar:same-lesson"
	// Several intents for one lesson exist before any of them is claimed.
	intents := h.verifiedIntents(t, userA, []string{key, key, key, key, key, key})
	var r results
	run(12, func(i int) {
		if i < 6 {
			r.add(h.service.Claim(context.Background(), userA, intents[i].Nonce))
			return
		}
		r.add(h.service.Fallback(context.Background(), userA, key))
	})
	r.onlyExpected(t) // every call must succeed: the lesson is simply open
	if len(r.granted) != 12 {
		t.Fatalf("granted %d calls, want all 12", len(r.granted))
	}
	for _, u := range r.granted {
		if u.ContentKey != key {
			t.Fatalf("granted %q", u.ContentKey)
		}
	}
	if usage := h.usage(t, userA); usage.Used != 1 {
		t.Fatalf("one lesson was counted %d times", usage.Used)
	}
	listed, _ := h.repo.List(context.Background(), userA)
	if len(listed) != 1 {
		t.Fatalf("stored unlocks = %d, want 1", len(listed))
	}
	for _, intent := range intents {
		if got, _, _ := h.repo.Intent(context.Background(), intent.Nonce); got.ConsumedAt == nil {
			// A claimed intent stays claimable only by returning the same unlock.
			if _, err := h.service.Claim(context.Background(), userA, intent.Nonce); err != nil {
				t.Fatalf("intent %s: %v", intent.Nonce, err)
			}
		}
	}
}

func scenarioMixedClaimsAndFallbacksRespectBothLimits(t *testing.T, h *harness) {
	h.limits = unlock.Limits{DailyCap: 5, FallbackPerDay: 2}
	intents := h.verifiedIntents(t, userA, keys("claim", 8))
	fallbackKeys := keys("fallback", 6)
	var r results
	run(len(intents)+len(fallbackKeys), func(i int) {
		if i < len(intents) {
			r.add(h.service.Claim(context.Background(), userA, intents[i].Nonce))
			return
		}
		r.add(h.service.Fallback(context.Background(), userA, fallbackKeys[i-len(intents)]))
	})
	r.onlyExpected(t, unlock.ErrCapReached, unlock.ErrFallbackExhausted)
	fallbacks := 0
	for _, u := range r.granted {
		if u.Source == unlock.SourceFallback {
			fallbacks++
		}
	}
	if len(r.granted) != 5 || fallbacks > 2 {
		t.Fatalf("granted %d (%d fallback), want exactly 5 with at most 2 fallback", len(r.granted), fallbacks)
	}
	if usage := h.usage(t, userA); usage.Used != 5 || usage.FallbackUsed != fallbacks {
		t.Fatalf("counters = %+v, want used 5 and fallback %d", usage, fallbacks)
	}
}

func scenarioCreditConvertsExactlyOnce(t *testing.T, h *harness) {
	ctx := context.Background()
	const credit = "simulation:jlpt-n3-simulation-1"
	if _, err := h.service.Fallback(ctx, userA, credit); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	converted := 0
	run(10, func(i int) {
		ok, err := h.service.Convert(ctx, userA, credit, fmt.Sprintf("simulation-attempt:attempt-%02d", i))
		if err != nil {
			t.Errorf("convert %d: %v", i, err)
		}
		if ok {
			mu.Lock()
			converted++
			mu.Unlock()
		}
	})
	if converted != 1 {
		t.Fatalf("one credit was spent %d times", converted)
	}
	if h.service.Allowed(ctx, userA, credit) {
		t.Fatal("the credit is still held after being spent")
	}
	if ok, _ := h.service.Convert(ctx, userA, credit, "simulation-attempt:late"); ok {
		t.Fatal("a spent credit converted again")
	}
	if ok, _ := h.service.Convert(ctx, userB, credit, "simulation-attempt:other"); ok {
		t.Fatal("another account spent this credit")
	}
	// A new credit cannot overwrite an attempt that is already open.
	listed, _ := h.repo.List(ctx, userA)
	if len(listed) != 1 {
		t.Fatalf("unlocks = %v", listed)
	}
	if _, err := h.service.Fallback(ctx, userA, credit); err != nil {
		t.Fatal(err)
	}
	if ok, _ := h.service.Convert(ctx, userA, credit, listed[0].ContentKey); ok {
		t.Fatal("converted onto an attempt that is already open")
	}
	if !h.service.Allowed(ctx, userA, credit) {
		t.Fatal("a refused conversion must keep the credit")
	}
}

func scenarioManyAccountsAtOnce(t *testing.T, h *harness) {
	h.limits.DailyCap = 4
	const accounts, perAccount = 8, 6
	type job struct {
		user  string
		nonce string
	}
	var jobs []job
	for a := 0; a < accounts; a++ {
		user := fmt.Sprintf("%08d-0000-4000-8000-%012d", a+10, a+10)
		for _, intent := range h.verifiedIntents(t, user, keys(fmt.Sprintf("acct%d", a), perAccount)) {
			jobs = append(jobs, job{user, intent.Nonce})
		}
	}
	var r results
	run(len(jobs), func(i int) { r.add(h.service.Claim(context.Background(), jobs[i].user, jobs[i].nonce)) })
	r.onlyExpected(t, unlock.ErrCapReached)
	perUser := map[string]int{}
	for _, u := range r.granted {
		perUser[u.UserID]++
	}
	if len(perUser) != accounts {
		t.Fatalf("%d accounts got unlocks, want %d", len(perUser), accounts)
	}
	for user, n := range perUser {
		if n != 4 {
			t.Fatalf("account %s got %d unlocks, want exactly 4", user, n)
		}
	}
}
