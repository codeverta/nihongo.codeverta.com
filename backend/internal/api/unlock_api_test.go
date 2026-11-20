package api_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/unlock"
)

const (
	unlockUnit   = "ca-app-pub-2242816010232507/1574208521"
	unlockKeyID  = "3335741209"
	unlockAdUnit = "1574208521"
)

type unlockStaticKeys struct{ key *ecdsa.PublicKey }

func (k unlockStaticKeys) PublicKey(_ context.Context, id string) (*ecdsa.PublicKey, error) {
	if id != unlockKeyID {
		return nil, errors.New("unknown key")
	}
	return k.key, nil
}

type unlockEnv struct {
	handler http.Handler
	memory  *store.MemoryStore
	private *ecdsa.PrivateKey
	token   string
	userID  string
	locked  string // an N3 grammar lesson
	other   string // a second N3 grammar lesson
	free    string // an N5 grammar lesson
}

// newUnlockEnv builds a server whose unlock ad is configured. adsConfig is the
// ads.json content; pass "" for the default (master switch on, unlock ad on).
func newUnlockEnv(t *testing.T, adsConfig string) *unlockEnv {
	t.Helper()
	return newUnlockEnvWith(t, adsConfig, nil)
}

// newUnlockEnvWith is newUnlockEnv with extra server wiring (news, jobs).
func newUnlockEnvWith(t *testing.T, adsConfig string, configure func(*api.Server)) *unlockEnv {
	t.Helper()
	memory := store.NewMemoryStore()
	t.Cleanup(func() { _ = memory.Close() })
	server := api.NewServer(memory, nil)
	if configure != nil {
		configure(server)
	}
	if adsConfig == "" {
		adsConfig = `{"enabled":true,"unlockRewarded":{"enabled":true,"unitId":"` + unlockUnit + `"},"unlockDailyCap":20,"unlockFallbackPerDay":2}`
	}
	adsPath := filepath.Join(t.TempDir(), "ads.json")
	if err := os.WriteFile(adsPath, []byte(adsConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := server.ConfigureAds(adsPath, "ads-token"); err != nil {
		t.Fatal(err)
	}
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	server.ConfigureUnlocks(unlock.NewMemoryRepository(), unlockStaticKeys{&private.PublicKey})
	server.ConfigureUnlockTiming(60*time.Millisecond, 10*time.Millisecond)
	handler := server.Handler()
	session := aiLogin(t, handler)
	lessons := memory.GrammarCourse("", "N3").Lessons
	return &unlockEnv{
		handler: handler, memory: memory, private: private, token: session.Token, userID: session.User.ID,
		locked: lessons[0].ID, other: lessons[1].ID, free: memory.GrammarCourse("", "N5").Lessons[0].ID,
	}
}

func (e *unlockEnv) do(method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if e.token != "" {
		request.Header.Set("Authorization", "Bearer "+e.token)
	}
	response := httptest.NewRecorder()
	e.handler.ServeHTTP(response, request)
	return response
}

func (e *unlockEnv) data(t *testing.T, response *httptest.ResponseRecorder, want int) map[string]any {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status %d, want %d: %s", response.Code, want, response.Body.String())
	}
	var payload struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Data
}

func (e *unlockEnv) errorBody(t *testing.T, response *httptest.ResponseRecorder, want int) map[string]any {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status %d, want %d: %s", response.Code, want, response.Body.String())
	}
	var payload struct {
		Error map[string]any `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Error
}

func (e *unlockEnv) lessonPath(id string) string { return "/v1/grammar/lessons/" + id }

func (e *unlockEnv) item(id string) string { return `{"kind":"grammar","itemId":"` + id + `"}` }

// callback sends a callback signed the way Google signs SSV requests.
func (e *unlockEnv) callback(user, nonce, transaction string) *httptest.ResponseRecorder {
	values := url.Values{}
	values.Set("ad_network", "5450213213286189855")
	values.Set("ad_unit", unlockAdUnit)
	values.Set("custom_data", nonce)
	values.Set("reward_amount", "1")
	values.Set("reward_item", "coin")
	values.Set("timestamp", "1763543000000")
	values.Set("transaction_id", transaction)
	values.Set("user_id", user)
	signed := values.Encode()
	digest := sha256.Sum256([]byte(signed))
	signature, err := ecdsa.SignASN1(rand.Reader, e.private, digest[:])
	if err != nil {
		panic(err)
	}
	query := signed + "&signature=" + base64.RawURLEncoding.EncodeToString(signature) + "&key_id=" + unlockKeyID
	response := httptest.NewRecorder()
	e.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/ads/ssv?"+query, nil))
	return response
}

// watchAd runs the whole flow for one lesson and returns the claim response.
func (e *unlockEnv) watchAd(t *testing.T, lesson, transaction string) map[string]any {
	t.Helper()
	intent := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(lesson)), http.StatusOK)
	nonce, _ := intent["nonce"].(string)
	if nonce == "" || intent["userId"] != e.userID {
		t.Fatalf("intent = %v", intent)
	}
	if response := e.callback(e.userID, nonce, transaction); response.Code != http.StatusOK {
		t.Fatalf("callback %d: %s", response.Code, response.Body.String())
	}
	return e.data(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+nonce+`"}`), http.StatusOK)
}

func (e *unlockEnv) makeLifetime(t *testing.T) {
	t.Helper()
	token, err := e.memory.BillingAccountToken(e.userID)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.memory.SaveVerifiedPurchase(e.userID, domain.Purchase{
		Platform: "android", ID: "unlock-test-" + e.userID, ProductID: "nihongo_master_lifetime",
		AccountToken: token, Proof: "unlock-test", State: "purchased", CheckedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLockedLessonOffersAnAdOnlyWhenOneCanBeShown(t *testing.T) {
	e := newUnlockEnv(t, "")
	body := e.errorBody(t, e.do(http.MethodGet, e.lessonPath(e.locked), ""), http.StatusPaymentRequired)
	if body["code"] != "premium_required" || body["unlockable"] != true {
		t.Fatalf("error = %v", body)
	}
	ref, _ := body["unlock"].(map[string]any)
	if ref["kind"] != "grammar" || ref["itemId"] != e.locked {
		t.Fatalf("unlock ref = %v", ref)
	}
	if response := e.do(http.MethodGet, e.lessonPath(e.free), ""); response.Code != http.StatusOK {
		t.Fatalf("free lesson: %d", response.Code)
	}

	off := newUnlockEnv(t, `{"enabled":false,"unlockRewarded":{"enabled":true,"unitId":"`+unlockUnit+`"}}`)
	body = off.errorBody(t, off.do(http.MethodGet, off.lessonPath(off.locked), ""), http.StatusPaymentRequired)
	if _, offered := body["unlockable"]; offered {
		t.Fatalf("an ad must not be offered while ads are switched off: %v", body)
	}
}

func TestWatchingAnAdOpensOnlyThatLesson(t *testing.T) {
	e := newUnlockEnv(t, "")
	intent := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.locked)), http.StatusOK)
	nonce := intent["nonce"].(string)

	if response := e.do(http.MethodGet, e.lessonPath(e.locked), ""); response.Code != http.StatusPaymentRequired {
		t.Fatalf("lesson opened before any ad: %d", response.Code)
	}
	pending := e.data(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+nonce+`"}`), http.StatusAccepted)
	if pending["pending"] != true {
		t.Fatalf("claim before the callback = %v", pending)
	}
	if response := e.do(http.MethodGet, e.lessonPath(e.locked), ""); response.Code != http.StatusPaymentRequired {
		t.Fatalf("lesson opened by an unverified claim: %d", response.Code)
	}

	if response := e.callback(e.userID, nonce, "tx-1"); response.Code != http.StatusOK {
		t.Fatalf("callback: %d", response.Code)
	}
	claimed := e.data(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+nonce+`"}`), http.StatusOK)
	if claimed["unlocked"] != true || claimed["contentKey"] != "grammar:"+e.locked || claimed["remaining"] != float64(19) {
		t.Fatalf("claim = %v", claimed)
	}

	if response := e.do(http.MethodGet, e.lessonPath(e.locked), ""); response.Code != http.StatusOK {
		t.Fatalf("unlocked lesson: %d %s", response.Code, response.Body.String())
	}
	attempt := e.data(t, e.do(http.MethodPost, e.lessonPath(e.locked)+"/attempts", "{}"), http.StatusCreated)
	attemptID, _ := attempt["id"].(string)
	if attemptID == "" {
		t.Fatalf("attempt = %v", attempt)
	}
	// The quiz of an unlocked lesson follows the lesson, so its attempt routes open too.
	if response := e.do(http.MethodPost, "/v1/grammar/attempts/"+attemptID+"/answers", `{}`); response.Code == http.StatusPaymentRequired {
		t.Fatalf("attempt route of an unlocked lesson was refused: %s", response.Body.String())
	}

	if response := e.do(http.MethodGet, e.lessonPath(e.other), ""); response.Code != http.StatusPaymentRequired {
		t.Fatalf("another locked lesson opened: %d", response.Code)
	}
	if response := e.do(http.MethodGet, "/v1/offline/practice/grammar/"+e.locked, ""); response.Code != http.StatusPaymentRequired {
		t.Fatalf("offline package of an ad-unlocked lesson must stay Lifetime-only: %d", response.Code)
	}

	list := e.data(t, e.do(http.MethodGet, "/v1/unlocks", ""), http.StatusOK)
	items, _ := list["items"].([]any)
	if len(items) != 1 || items[0] != "grammar:"+e.locked || list["remaining"] != float64(19) || list["available"] != true {
		t.Fatalf("list = %v", list)
	}
	gates, _ := list["gates"].(map[string]any)
	offered, _ := list["offered"].(map[string]any)
	if gates["lessons"] != true || gates["news"] != false || offered["lessons"] != true || offered["news"] != true {
		t.Fatalf("gates = %v, offered = %v", gates, offered)
	}
}

func TestCatalogStaysOpenSoFreeAccountsCanSeeWhatToUnlock(t *testing.T) {
	e := newUnlockEnv(t, "")
	data := e.data(t, e.do(http.MethodGet, "/v1/grammar/course?level=N3", ""), http.StatusOK)
	if lessons, _ := data["lessons"].([]any); len(lessons) == 0 {
		t.Fatalf("catalog = %v", data)
	}
}

func TestIntentRefusesItemsThatNeedNoUnlock(t *testing.T) {
	e := newUnlockEnv(t, "")
	for name, body := range map[string]string{
		"free lesson":     e.item(e.free),
		"unknown lesson":  e.item("does-not-exist"),
		"bad kind":        `{"kind":"news","itemId":"x"}`,
		"bad item id":     `{"kind":"grammar","itemId":"../x"}`,
		"unknown field":   `{"kind":"grammar","itemId":"x","extra":1}`,
		"missing details": `{}`,
	} {
		response := e.do(http.MethodPost, "/v1/unlocks/intents", body)
		if response.Code != http.StatusUnprocessableEntity && response.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 422/400: %s", name, response.Code, response.Body.String())
		}
	}
	signedOut := &unlockEnv{handler: e.handler}
	if response := signedOut.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.locked)); response.Code != http.StatusUnauthorized {
		t.Fatalf("signed-out intent: %d", response.Code)
	}
}

func TestIntentIsUnavailableWhileAdsAreOff(t *testing.T) {
	e := newUnlockEnv(t, `{"enabled":false,"unlockRewarded":{"enabled":true,"unitId":"`+unlockUnit+`"}}`)
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.locked)), http.StatusForbidden); body["code"] != "unlock_unavailable" {
		t.Fatalf("error = %v", body)
	}
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/fallback", e.item(e.locked)), http.StatusForbidden); body["code"] != "unlock_unavailable" {
		t.Fatalf("fallback error = %v", body)
	}
}

func TestLifetimeAccountsAreToldTheyNeedNoUnlock(t *testing.T) {
	e := newUnlockEnv(t, "")
	e.makeLifetime(t)
	data := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.locked)), http.StatusOK)
	if data["alreadyUnlocked"] != true || data["reason"] != "lifetime" {
		t.Fatalf("data = %v", data)
	}
	if response := e.do(http.MethodGet, e.lessonPath(e.locked), ""); response.Code != http.StatusOK {
		t.Fatalf("lifetime lesson: %d", response.Code)
	}
}

func TestSecondIntentForAnOpenLessonSkipsTheAd(t *testing.T) {
	e := newUnlockEnv(t, "")
	e.watchAd(t, e.locked, "tx-1")
	data := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.locked)), http.StatusOK)
	if data["alreadyUnlocked"] != true {
		t.Fatalf("data = %v", data)
	}
}

func TestCallbackNeedsGoogleSignatureAndCannotBeReplayed(t *testing.T) {
	e := newUnlockEnv(t, "")
	intent := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.locked)), http.StatusOK)
	nonce := intent["nonce"].(string)

	forged := httptest.NewRecorder()
	e.handler.ServeHTTP(forged, httptest.NewRequest(http.MethodGet, "/v1/ads/ssv?custom_data="+nonce+"&user_id="+e.userID+"&transaction_id=x", nil))
	if forged.Code != http.StatusBadRequest {
		t.Fatalf("unsigned callback: %d, want 400", forged.Code)
	}
	if body := e.data(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+nonce+`"}`), http.StatusAccepted); body["pending"] != true {
		t.Fatal("a forged callback verified the intent")
	}

	// A genuine callback for an unknown intent is acknowledged but does nothing.
	if response := e.callback(e.userID, strings.Repeat("0", 32), "tx-unknown"); response.Code != http.StatusOK {
		t.Fatalf("callback for an unknown intent: %d", response.Code)
	}
	// Someone else's callback cannot verify this intent.
	if response := e.callback("33333333-3333-4333-8333-333333333333", nonce, "tx-other"); response.Code != http.StatusOK {
		t.Fatalf("callback for another account: %d", response.Code)
	}
	if body := e.data(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+nonce+`"}`), http.StatusAccepted); body["pending"] != true {
		t.Fatal("another account's callback verified the intent")
	}

	if response := e.callback(e.userID, nonce, "tx-1"); response.Code != http.StatusOK {
		t.Fatalf("callback: %d", response.Code)
	}
	second := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.other)), http.StatusOK)["nonce"].(string)
	if response := e.callback(e.userID, second, "tx-1"); response.Code != http.StatusOK {
		t.Fatalf("replayed transaction must still be acknowledged: %d", response.Code)
	}
	if body := e.data(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+second+`"}`), http.StatusAccepted); body["pending"] != true {
		t.Fatal("a replayed transaction verified a second intent")
	}
}

func TestClaimRejectsBadCodesAndSignedOutRequests(t *testing.T) {
	e := newUnlockEnv(t, "")
	nonce := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.locked)), http.StatusOK)["nonce"].(string)
	e.callback(e.userID, nonce, "tx-1")
	signedOut := &unlockEnv{handler: e.handler}
	if response := signedOut.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+nonce+`"}`); response.Code != http.StatusUnauthorized {
		t.Fatalf("claim without a session: %d", response.Code)
	}
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"nothex"}`), http.StatusUnprocessableEntity); body["code"] != "invalid_nonce" {
		t.Fatalf("error = %v", body)
	}
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+strings.Repeat("a", 32)+`"}`), http.StatusNotFound); body["code"] != "intent_not_found" {
		t.Fatalf("error = %v", body)
	}
}

func TestDailyLimitStopsFurtherUnlocks(t *testing.T) {
	e := newUnlockEnv(t, `{"enabled":true,"unlockRewarded":{"enabled":true,"unitId":"`+unlockUnit+`"},"unlockDailyCap":1,"unlockFallbackPerDay":2}`)
	e.watchAd(t, e.locked, "tx-1")
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/intents", e.item(e.other)), http.StatusTooManyRequests); body["code"] != "unlock_daily_limit" {
		t.Fatalf("error = %v", body)
	}
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/fallback", e.item(e.other)), http.StatusTooManyRequests); body["code"] != "unlock_daily_limit" {
		t.Fatalf("fallback over the daily limit: %v", body)
	}
}

func TestFallbackOpensALessonWhenNoAdCanLoadButOnlyTwiceADay(t *testing.T) {
	e := newUnlockEnv(t, "")
	lessons := e.memory.GrammarCourse("", "N3").Lessons
	for i := 0; i < 2; i++ {
		data := e.data(t, e.do(http.MethodPost, "/v1/unlocks/fallback", e.item(lessons[i].ID)), http.StatusOK)
		if data["unlocked"] != true || data["source"] != "fallback" {
			t.Fatalf("fallback %d = %v", i, data)
		}
		if response := e.do(http.MethodGet, e.lessonPath(lessons[i].ID), ""); response.Code != http.StatusOK {
			t.Fatalf("lesson %d after fallback: %d", i, response.Code)
		}
	}
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/fallback", e.item(lessons[2].ID)), http.StatusTooManyRequests); body["code"] != "unlock_fallback_exhausted" {
		t.Fatalf("error = %v", body)
	}
	if response := e.do(http.MethodGet, e.lessonPath(lessons[2].ID), ""); response.Code != http.StatusPaymentRequired {
		t.Fatalf("third lesson opened without an ad: %d", response.Code)
	}
}

func TestUnlockedLessonKeepsItsAITutor(t *testing.T) {
	e := newUnlockEnv(t, "")
	if before := e.do(http.MethodGet, "/v1/ai/conversations/grammar/"+e.locked, ""); before.Code != http.StatusPaymentRequired {
		t.Fatalf("AI tutor of a locked lesson: %d", before.Code)
	}
	e.watchAd(t, e.locked, "tx-1")
	if after := e.do(http.MethodGet, "/v1/ai/conversations/grammar/"+e.locked, ""); after.Code != http.StatusOK {
		t.Fatalf("AI tutor of an unlocked lesson: %d %s", after.Code, after.Body.String())
	}
}

func TestAdminSeesUnlockStatistics(t *testing.T) {
	e := newUnlockEnv(t, "")
	stats := func(query, token string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/v1/admin/unlock-stats"+query, nil)
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		e.handler.ServeHTTP(response, request)
		return response
	}
	if response := stats("", ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("stats without the admin key: %d", response.Code)
	}
	if response := stats("?days=0", "ads-token"); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("days=0: %d", response.Code)
	}

	e.watchAd(t, e.locked, "tx-1")
	e.fallback(t, "grammar", e.other)
	forged := httptest.NewRecorder()
	e.handler.ServeHTTP(forged, httptest.NewRequest(http.MethodGet, "/v1/ads/ssv?custom_data=x&signature=AAAA&key_id=1", nil))

	data := e.data(t, stats("?days=7", "ads-token"), http.StatusOK)
	var lessons map[string]any
	for _, row := range data["totals"].([]any) {
		if row.(map[string]any)["group"] == "lessons" {
			lessons = row.(map[string]any)
		}
	}
	if lessons["intents"] != float64(1) || lessons["verified"] != float64(1) || lessons["rewarded"] != float64(1) || lessons["fallback"] != float64(1) {
		t.Fatalf("lesson totals = %v", lessons)
	}
	if days := data["days"].([]any); len(days) != 7*6 {
		t.Fatalf("days rows = %d, want 7 days x 6 groups (incl. double XP)", len(days))
	}
	callbacks := data["callbacks"].(map[string]any)
	if callbacks["ok"] != float64(1) || callbacks["badSignature"] != float64(1) {
		t.Fatalf("callbacks = %v", callbacks)
	}
	if data["accounts"] != float64(1) || data["accountsNowLifetime"] != float64(0) {
		t.Fatalf("accounts = %v / %v", data["accounts"], data["accountsNowLifetime"])
	}
	e.makeLifetime(t)
	if data := e.data(t, stats("", "ads-token"), http.StatusOK); data["accountsNowLifetime"] != float64(1) {
		t.Fatalf("an account that unlocked and then bought Lifetime = %v", data["accountsNowLifetime"])
	}
}

func TestSavingAdsFromAnOlderDashboardKeepsTheUnlockSettings(t *testing.T) {
	e := newUnlockEnv(t, `{"enabled":true,"unlockRewarded":{"enabled":true,"unitId":"`+unlockUnit+`"},"unlockDailyCap":7,"unlockFallbackPerDay":1}`)
	save := func(body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPut, "/v1/admin/ads-config", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer ads-token")
		response := httptest.NewRecorder()
		e.handler.ServeHTTP(response, request)
		return response
	}
	// A dashboard built before the unlock settings existed sends none of them.
	legacy := `{"enabled":true,"banner":{"enabled":true,"unitId":"ca-app-pub-3940256099942544/6300978111"},"newsNative":{"enabled":true,"unitId":"ca-app-pub-3940256099942544/2247696110","every":6},"jobsNative":{"enabled":true,"unitId":"ca-app-pub-3940256099942544/2247696110","every":8},"aiChatRewarded":{"enabled":true,"unitId":"ca-app-pub-3940256099942544/5224354917"}}`
	if response := save(legacy); response.Code != http.StatusOK {
		t.Fatalf("save: %d %s", response.Code, response.Body.String())
	}
	public := httptest.NewRecorder()
	e.handler.ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/v1/ads-config", nil))
	var payload struct {
		Data struct {
			UnlockRewarded struct {
				Enabled bool   `json:"enabled"`
				UnitID  string `json:"unitId"`
			} `json:"unlockRewarded"`
			UnlockDailyCap       int `json:"unlockDailyCap"`
			UnlockFallbackPerDay int `json:"unlockFallbackPerDay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(public.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	got := payload.Data
	if !got.UnlockRewarded.Enabled || got.UnlockRewarded.UnitID != unlockUnit || got.UnlockDailyCap != 7 || got.UnlockFallbackPerDay != 1 {
		t.Fatalf("unlock settings were reset by a save that did not mention them: %+v", got)
	}

	// Sending them explicitly is validated.
	explicit := `{"enabled":true,"banner":{"enabled":false,"unitId":""},"newsNative":{"enabled":false,"unitId":"","every":6},"jobsNative":{"enabled":false,"unitId":"","every":8},"aiChatRewarded":{"enabled":false,"unitId":""},"unlockRewarded":{"enabled":true,"unitId":"` + unlockUnit + `"},"unlockDailyCap":500,"unlockFallbackPerDay":1}`
	if response := save(explicit); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("daily cap of 500: %d, want 422", response.Code)
	}
}
