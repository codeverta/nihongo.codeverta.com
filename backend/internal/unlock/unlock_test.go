package unlock_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	"nihongo-master-backend/internal/unlock"
)

const (
	userA   = "11111111-1111-4111-8111-111111111111"
	userB   = "22222222-2222-4222-8222-222222222222"
	adUnit  = "1574208521"
	keyID   = "3335741209"
	lessonA = "grammar:n3-lesson-01"
	lessonB = "grammar:n3-lesson-02"
)

type staticKeys struct{ key *ecdsa.PublicKey }

func (s staticKeys) PublicKey(_ context.Context, id string) (*ecdsa.PublicKey, error) {
	if id != keyID {
		return nil, errors.New("unknown key")
	}
	return s.key, nil
}

type harness struct {
	t       *testing.T
	service *unlock.Service
	repo    unlock.Repository
	private *ecdsa.PrivateKey
	now     time.Time
	limits  unlock.Limits
}

func newHarness(t *testing.T, repo unlock.Repository) *harness {
	t.Helper()
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{
		t: t, repo: repo, private: private,
		now:    time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC),
		limits: unlock.Limits{DailyCap: 20, FallbackPerDay: 2},
	}
	h.service = &unlock.Service{
		Repo:   h.repo,
		Limits: func() unlock.Limits { return h.limits },
		Verify: &unlock.Verifier{Keys: staticKeys{&private.PublicKey}},
		Units: func() []string {
			return []string{"ca-app-pub-2242816010232507/" + adUnit, "ca-app-pub-2242816010232507/7777777777"}
		},
		Now: func() time.Time { return h.now },
	}
	return h
}

// callback builds a query string signed the way Google signs SSV requests.
func (h *harness) callback(user, nonce, transaction, unit string) string {
	h.t.Helper()
	values := url.Values{}
	values.Set("ad_network", "5450213213286189855")
	values.Set("ad_unit", unit)
	values.Set("custom_data", nonce)
	values.Set("reward_amount", "1")
	values.Set("reward_item", "coin")
	values.Set("timestamp", "1763543000000")
	values.Set("transaction_id", transaction)
	values.Set("user_id", user)
	signed := values.Encode()
	digest := sha256.Sum256([]byte(signed))
	signature, err := ecdsa.SignASN1(rand.Reader, h.private, digest[:])
	if err != nil {
		h.t.Fatal(err)
	}
	return signed + "&signature=" + base64.RawURLEncoding.EncodeToString(signature) + "&key_id=" + keyID
}

func (h *harness) verifiedIntent(user, key, transaction string) unlock.Intent {
	h.t.Helper()
	intent, err := h.service.CreateIntent(context.Background(), user, key)
	if err != nil {
		h.t.Fatal(err)
	}
	if err := h.service.HandleCallback(context.Background(), h.callback(user, intent.Nonce, transaction, adUnit)); err != nil {
		h.t.Fatal(err)
	}
	return intent
}

var scenarios = []struct {
	name string
	run  func(*testing.T, *harness)
}{
	{"VerifiedAdUnlocksTheItem", scenarioVerifiedAdUnlocksTheItem},
	{"ClaimWaitsForGoogleToConfirm", scenarioClaimWaitsForGoogleToConfirm},
	{"ClaimIsBoundToTheAccountThatStartedIt", scenarioClaimIsBoundToTheAccountThatStartedIt},
	{"CallbackRejectsForgeriesAndReplays", scenarioCallbackRejectsForgeriesAndReplays},
	{"ExpiredIntentCannotBeVerifiedOrClaimed", scenarioExpiredIntentCannotBeVerifiedOrClaimed},
	{"DailyCapStopsIntentsAndClaims", scenarioDailyCapStopsIntentsAndClaims},
	{"ClaimRespectsTheCapEvenAfterTheAdWasWatched", scenarioClaimRespectsTheCapEvenAfterTheAdWasWatched},
	{"FallbackIsBoundedAndCountsTowardTheDailyCap", scenarioFallbackIsBoundedAndCountsTowardTheDailyCap},
	{"ConcurrentClaimsNeverExceedTheCap", scenarioConcurrentClaimsNeverExceedTheCap},
	{"ManyConcurrentClaimsOnlyEverHitTheCap", scenarioManyConcurrentClaimsOnlyEverHitTheCap},
	{"ClaimsAndFallbacksRaceForTheSameLesson", scenarioClaimsAndFallbacksRaceForTheSameLesson},
	{"MixedClaimsAndFallbacksRespectBothLimits", scenarioMixedClaimsAndFallbacksRespectBothLimits},
	{"ManyAccountsAtOnce", scenarioManyAccountsAtOnce},
	{"CreditConvertsExactlyOnce", scenarioCreditConvertsExactlyOnce},
	{"StatsCountActivityByDayAndGroup", scenarioStatsCountActivityByDayAndGroup},
	{"XPBoostIntentIsConsumedOnce", scenarioXPBoostIntentIsConsumedOnce},
}

func TestClientCannotAskForAnAttemptUnlock(t *testing.T) {
	if _, ok := unlock.ParseKind(string(unlock.KindSimulationAttempt)); ok {
		t.Fatal("simulation attempts must only be opened by converting a credit")
	}
	for _, kind := range []unlock.Kind{unlock.KindAnki, unlock.KindSimulation, unlock.KindNews, unlock.KindJob} {
		if _, ok := unlock.ParseKind(string(kind)); !ok {
			t.Fatalf("kind %q rejected", kind)
		}
	}
}

func TestMemoryRepository(t *testing.T) {
	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) { tc.run(t, newHarness(t, unlock.NewMemoryRepository())) })
	}
}

func TestKeyValidation(t *testing.T) {
	if key, ok := unlock.Key(unlock.KindGrammar, "n3-lesson-01"); !ok || key != lessonA {
		t.Fatalf("Key = %q, %v", key, ok)
	}
	for _, bad := range []struct {
		kind unlock.Kind
		id   string
	}{{"podcast", "x"}, {unlock.KindGrammar, ""}, {unlock.KindGrammar, "a/b"}, {unlock.KindGrammar, "a b"}} {
		if _, ok := unlock.Key(bad.kind, bad.id); ok {
			t.Fatalf("Key(%q, %q) accepted", bad.kind, bad.id)
		}
	}
}

func scenarioVerifiedAdUnlocksTheItem(t *testing.T, h *harness) {
	ctx := context.Background()
	intent := h.verifiedIntent(userA, lessonA, "tx-1")
	granted, err := h.service.Claim(ctx, userA, intent.Nonce)
	if err != nil || granted.ContentKey != lessonA || granted.Source != unlock.SourceRewarded {
		t.Fatalf("Claim = %+v, %v", granted, err)
	}
	if !h.service.Allowed(ctx, userA, lessonA) {
		t.Fatal("item not unlocked")
	}
	if h.service.Allowed(ctx, userB, lessonA) || h.service.Allowed(ctx, userA, lessonB) {
		t.Fatal("unlock leaked to another account or item")
	}
	again, err := h.service.Claim(ctx, userA, intent.Nonce)
	if err != nil || again.ContentKey != lessonA {
		t.Fatalf("repeat Claim = %+v, %v", again, err)
	}
	if remaining, _, _ := h.service.Remaining(ctx, userA); remaining != 19 {
		t.Fatalf("remaining = %d, want 19 (a repeat claim must not cost twice)", remaining)
	}
}

func scenarioClaimWaitsForGoogleToConfirm(t *testing.T, h *harness) {
	ctx := context.Background()
	intent, err := h.service.CreateIntent(ctx, userA, lessonA)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.Claim(ctx, userA, intent.Nonce); !errors.Is(err, unlock.ErrNotVerified) {
		t.Fatalf("Claim before callback = %v, want ErrNotVerified", err)
	}
	if h.service.Allowed(ctx, userA, lessonA) {
		t.Fatal("unlocked without a verified ad")
	}
}

func scenarioClaimIsBoundToTheAccountThatStartedIt(t *testing.T, h *harness) {
	intent := h.verifiedIntent(userA, lessonA, "tx-1")
	if _, err := h.service.Claim(context.Background(), userB, intent.Nonce); !errors.Is(err, unlock.ErrNotFound) {
		t.Fatalf("other account Claim = %v, want ErrNotFound", err)
	}
}

func scenarioCallbackRejectsForgeriesAndReplays(t *testing.T, h *harness) {
	ctx := context.Background()
	intent, err := h.service.CreateIntent(ctx, userA, lessonA)
	if err != nil {
		t.Fatal(err)
	}
	good := h.callback(userA, intent.Nonce, "tx-1", adUnit)

	tampered := good[:len("ad_network=")] + "9" + good[len("ad_network=")+1:]
	if err := h.service.HandleCallback(ctx, tampered); !errors.Is(err, unlock.ErrBadSignature) {
		t.Fatalf("tampered query = %v, want ErrBadSignature", err)
	}
	if err := h.service.HandleCallback(ctx, "user_id="+userA+"&custom_data="+intent.Nonce); !errors.Is(err, unlock.ErrBadSignature) {
		t.Fatalf("unsigned query = %v, want ErrBadSignature", err)
	}
	other, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	foreign := &unlock.Service{Repo: h.repo, Limits: h.service.Limits, Verify: &unlock.Verifier{Keys: staticKeys{&other.PublicKey}}, Now: h.service.Now}
	if err := foreign.HandleCallback(ctx, good); !errors.Is(err, unlock.ErrBadSignature) {
		t.Fatalf("callback signed by another key = %v, want ErrBadSignature", err)
	}
	if err := h.service.HandleCallback(ctx, h.callback(userA, intent.Nonce, "tx-1", "9999999999")); !errors.Is(err, unlock.ErrInvalid) {
		t.Fatalf("callback for another ad unit = %v, want ErrInvalid", err)
	}
	if err := h.service.HandleCallback(ctx, h.callback(userB, intent.Nonce, "tx-1", adUnit)); !errors.Is(err, unlock.ErrInvalid) {
		t.Fatalf("callback for another account = %v, want ErrInvalid", err)
	}
	if _, err := h.service.Claim(ctx, userA, intent.Nonce); !errors.Is(err, unlock.ErrNotVerified) {
		t.Fatalf("rejected callbacks must not verify the intent: %v", err)
	}

	if err := h.service.HandleCallback(ctx, good); err != nil {
		t.Fatalf("good callback: %v", err)
	}
	if err := h.service.HandleCallback(ctx, good); err != nil {
		t.Fatalf("Google retrying the same callback must stay a success: %v", err)
	}

	second, err := h.service.CreateIntent(ctx, userA, lessonB)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.service.HandleCallback(ctx, h.callback(userA, second.Nonce, "tx-1", adUnit)); !errors.Is(err, unlock.ErrReplay) {
		t.Fatalf("reused transaction = %v, want ErrReplay", err)
	}
}

func scenarioExpiredIntentCannotBeVerifiedOrClaimed(t *testing.T, h *harness) {
	ctx := context.Background()
	intent, err := h.service.CreateIntent(ctx, userA, lessonA)
	if err != nil {
		t.Fatal(err)
	}
	h.now = h.now.Add(unlock.IntentTTL + time.Second)
	if err := h.service.HandleCallback(ctx, h.callback(userA, intent.Nonce, "tx-1", adUnit)); !errors.Is(err, unlock.ErrExpired) {
		t.Fatalf("late callback = %v, want ErrExpired", err)
	}
	if _, err := h.service.Claim(ctx, userA, intent.Nonce); !errors.Is(err, unlock.ErrNotVerified) {
		t.Fatalf("Claim = %v, want ErrNotVerified", err)
	}
}

func scenarioDailyCapStopsIntentsAndClaims(t *testing.T, h *harness) {
	ctx := context.Background()
	h.limits.DailyCap = 2
	for i, key := range []string{"grammar:a", "grammar:b"} {
		intent := h.verifiedIntent(userA, key, "tx-"+key)
		if _, err := h.service.Claim(ctx, userA, intent.Nonce); err != nil {
			t.Fatalf("claim %d: %v", i, err)
		}
	}
	if _, err := h.service.CreateIntent(ctx, userA, "grammar:c"); !errors.Is(err, unlock.ErrCapReached) {
		t.Fatalf("intent over the cap = %v, want ErrCapReached", err)
	}
	if _, err := h.service.CreateIntent(ctx, userA, "grammar:a"); !errors.Is(err, unlock.ErrAlreadyUnlocked) {
		t.Fatalf("intent for an open item = %v, want ErrAlreadyUnlocked", err)
	}
	if _, err := h.service.CreateIntent(ctx, userB, "grammar:c"); err != nil {
		t.Fatalf("another account is unaffected: %v", err)
	}
	h.now = h.now.Add(24 * time.Hour)
	if _, err := h.service.CreateIntent(ctx, userA, "grammar:c"); err != nil {
		t.Fatalf("the cap must reset the next day: %v", err)
	}
}

func scenarioClaimRespectsTheCapEvenAfterTheAdWasWatched(t *testing.T, h *harness) {
	ctx := context.Background()
	h.limits.DailyCap = 1
	// Two intents are created before either is claimed, so both pass the soft check.
	first, err := h.service.CreateIntent(ctx, userA, "grammar:a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := h.service.CreateIntent(ctx, userA, "grammar:b")
	if err != nil {
		t.Fatal(err)
	}
	for i, intent := range []unlock.Intent{first, second} {
		if err := h.service.HandleCallback(ctx, h.callback(userA, intent.Nonce, "tx-"+intent.ContentKey, adUnit)); err != nil {
			t.Fatalf("callback %d: %v", i, err)
		}
	}
	if _, err := h.service.Claim(ctx, userA, first.Nonce); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.Claim(ctx, userA, second.Nonce); !errors.Is(err, unlock.ErrCapReached) {
		t.Fatalf("claim over the cap = %v, want ErrCapReached", err)
	}
	if h.service.Allowed(ctx, userA, "grammar:b") {
		t.Fatal("item unlocked beyond the cap")
	}
}

func scenarioFallbackIsBoundedAndCountsTowardTheDailyCap(t *testing.T, h *harness) {
	ctx := context.Background()
	for _, key := range []string{"grammar:a", "grammar:b"} {
		granted, err := h.service.Fallback(ctx, userA, key)
		if err != nil || granted.Source != unlock.SourceFallback {
			t.Fatalf("Fallback(%s) = %+v, %v", key, granted, err)
		}
	}
	if _, err := h.service.Fallback(ctx, userA, "grammar:c"); !errors.Is(err, unlock.ErrFallbackExhausted) {
		t.Fatalf("third fallback = %v, want ErrFallbackExhausted", err)
	}
	if again, err := h.service.Fallback(ctx, userA, "grammar:a"); err != nil || again.ContentKey != "grammar:a" {
		t.Fatalf("repeat fallback on an open item = %+v, %v", again, err)
	}
	if remaining, fallbackLeft, _ := h.service.Remaining(ctx, userA); remaining != 18 || fallbackLeft != 0 {
		t.Fatalf("Remaining = %d, %d", remaining, fallbackLeft)
	}
	h.limits = unlock.Limits{DailyCap: 2, FallbackPerDay: 5}
	if _, err := h.service.Fallback(ctx, userB, "grammar:a"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.Fallback(ctx, userB, "grammar:b"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.Fallback(ctx, userB, "grammar:c"); !errors.Is(err, unlock.ErrCapReached) {
		t.Fatalf("fallback over the daily cap = %v, want ErrCapReached", err)
	}
}

func scenarioConcurrentClaimsNeverExceedTheCap(t *testing.T, h *harness) {
	ctx := context.Background()
	h.limits.DailyCap = 3
	var intents []unlock.Intent
	for i := 0; i < 10; i++ {
		intent, err := h.service.CreateIntent(ctx, userA, "grammar:l"+string(rune('a'+i)))
		if err != nil {
			t.Fatal(err)
		}
		if err := h.service.HandleCallback(ctx, h.callback(userA, intent.Nonce, "tx-"+intent.Nonce, adUnit)); err != nil {
			t.Fatal(err)
		}
		intents = append(intents, intent)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	granted := 0
	for _, intent := range intents {
		wg.Add(1)
		go func(nonce string) {
			defer wg.Done()
			if _, err := h.service.Claim(ctx, userA, nonce); err == nil {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		}(intent.Nonce)
	}
	wg.Wait()
	if granted != 3 {
		t.Fatalf("granted %d, want 3", granted)
	}
}

func TestUnconfiguredVerifierRejectsCallbacks(t *testing.T) {
	service := &unlock.Service{Repo: unlock.NewMemoryRepository(), Limits: func() unlock.Limits { return unlock.Limits{DailyCap: 1} }}
	if err := service.HandleCallback(context.Background(), "a=b&signature=x&key_id=1"); !errors.Is(err, unlock.ErrUnavailable) {
		t.Fatalf("HandleCallback = %v, want ErrUnavailable", err)
	}
}
