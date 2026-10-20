package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

func TestMangaLifetimeGuardAndRevocation(t *testing.T) {
	memory := store.NewMemoryStore()
	buyer, err := memory.AuthenticateGoogle(domain.GoogleIdentity{Subject: "manga-guard", Email: "guard@example.com", DisplayName: "Buyer"})
	if err != nil {
		t.Fatal(err)
	}
	token, _ := memory.CreateSession(buyer.ID)
	server := api.NewServer(memory, fakeGoogleVerifier{})
	handler := server.Handler()
	call := func(method, path, session string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		if session != "" {
			r.Header.Set("Authorization", "Bearer "+session)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/v1/manga", "/v1/manga/series", "/v1/manga/chapters/chapter", "/v1/manga/chapters/chapter/prepare"} {
		method := "GET"
		if strings.HasSuffix(path, "prepare") {
			method = "POST"
		}
		for _, session := range []string{"", "invalid"} {
			if w := call(method, path, session); w.Code != 401 {
				t.Fatalf("guest bypass: %s %d", path, w.Code)
			}
		}
		if w := call(method, path, token); w.Code != 402 {
			t.Fatalf("free bypass: %s %d", path, w.Code)
		}
	}
	account, _ := memory.BillingAccountToken(buyer.ID)
	purchase := domain.Purchase{ID: "guard-purchase", Platform: "android", AccountToken: account, State: "purchased", CheckedAt: time.Now()}
	if err := memory.SaveVerifiedPurchase(buyer.ID, purchase); err != nil {
		t.Fatal(err)
	}
	if w := call("GET", "/v1/manga", token); w.Code != 200 || w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("lifetime blocked or public cache", w.Code, w.Body.String())
	}
	purchase.State = "refunded"
	if err := memory.SaveVerifiedPurchase(buyer.ID, purchase); err != nil {
		t.Fatal(err)
	}
	if w := call("GET", "/v1/manga", token); w.Code != 402 {
		t.Fatal("refunded purchase retains access", w.Code)
	}
	if err := memory.DeleteAccount(buyer.ID); err != nil {
		t.Fatal(err)
	}
	if w := call("GET", "/v1/manga", token); w.Code != 401 {
		t.Fatal("deleted account retains access", w.Code)
	}
}

func TestIsolatedMangaValidatesBillingAndFailsClosed(t *testing.T) {
	state := `{"data":{"entitlement":{"tier":"premium","isLifetime":true}}}`
	status := http.StatusOK
	billing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/billing" || r.Header.Get("Authorization") != "Bearer session" {
			t.Error("missing account API validation")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(state))
	}))
	defer billing.Close()
	handler := api.NewMangaHandler(nil, "", api.NewMangaBillingChecker(billing.URL+"/v1"))
	call := func() int {
		r := httptest.NewRequest("GET", "/v1/manga", nil)
		r.Header.Set("Authorization", "Bearer session")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}
	if got := call(); got != 200 {
		t.Fatal("lifetime denied", got)
	}
	state = `{"data":{"entitlement":{"tier":"premium","isLifetime":false}}}`
	if got := call(); got != 402 {
		t.Fatal("non-lifetime premium allowed", got)
	}
	status = 401
	if got := call(); got != 401 {
		t.Fatal("invalid session allowed", got)
	}
	status = 503
	if got := call(); got != 503 {
		t.Fatal("billing outage granted access", got)
	}
	noCheck := api.NewMangaHandler(nil, "", nil)
	r := httptest.NewRequest("GET", "/v1/manga", nil)
	r.Header.Set("Authorization", "Bearer session")
	w := httptest.NewRecorder()
	noCheck.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatal("unconfigured auth allowed access", w.Code)
	}
}
