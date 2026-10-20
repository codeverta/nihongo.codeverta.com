package api_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/billing"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
	"strings"
	"testing"
	"time"
)

type billingVerifier struct{ token, state string }

func (v *billingVerifier) Platforms() []string                                { return []string{"android", "ios"} }
func (v *billingVerifier) Acknowledge(context.Context, domain.Purchase) error { return nil }
func (v *billingVerifier) Verify(_ context.Context, p domain.PurchaseProof) (domain.Purchase, error) {
	if p.VerificationData != "real-proof" || p.ProductID != "nihongo_master_lifetime" {
		return domain.Purchase{}, billing.ErrInvalid
	}
	return domain.Purchase{Platform: p.Platform, ID: "transaction", ProductID: p.ProductID, Proof: "real-proof", AccountToken: v.token, State: v.state, CheckedAt: time.Now()}, nil
}
func TestBillingAPIUnlockRevokeAndDelete(t *testing.T) {
	memory := store.NewMemoryStore()
	verifier := &billingVerifier{state: "purchased"}
	server := api.NewServer(memory, fakeGoogleVerifier{})
	server.ConfigureBilling(&billing.Service{Store: memory, Verifier: verifier, ProductID: "nihongo_master_lifetime"}, "", "")
	handler := server.Handler()
	token := login(t, handler)
	for _, endpoint := range []struct{ method, path string }{{"GET", "/v1/billing"}, {"POST", "/v1/billing/verify"}} {
		r := httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader("{}"))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal("shared demo account can purchase or restore")
		}
	}
	buyer, err := memory.AuthenticateGoogle(domain.GoogleIdentity{Subject: "buyer", Email: "buyer@example.com", DisplayName: "Buyer"})
	if err != nil {
		t.Fatal(err)
	}
	token, err = memory.CreateSession(buyer.ID)
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/v1/grammar/course?level=N3", "/v1/grammar/lessons/grammar-n2-01", "/v1/kanji/lessons/kanji-n3-01", "/v1/reading/lessons/reading-n2-01", "/v1/vocabulary/lessons/n3-daily-life", "/v1/simulation/n4", "/v1/simulation/n3", "/v1/simulation/jft", "/v1/simulation/nat"} {
		if w := call("GET", path, ""); w.Code != 402 {
			t.Fatalf("free user accessed %s: %d %s", path, w.Code, w.Body.String())
		}
	}
	for _, path := range []string{"/v1/grammar/course?level=N4", "/v1/simulation/n5"} {
		if w := call("GET", path, ""); w.Code != 200 {
			t.Fatal(w.Body.String())
		}
	}
	var status struct {
		Data struct {
			AccountToken string `json:"accountToken"`
		} `json:"data"`
	}
	w := call("GET", "/v1/billing", "")
	if json.Unmarshal(w.Body.Bytes(), &status) != nil || status.Data.AccountToken == "" {
		t.Fatal("missing billing account")
	}
	verifier.token = status.Data.AccountToken
	if w = call("POST", "/v1/billing/verify", `{"platform":"android","productId":"nihongo_master_lifetime","verificationData":"fake"}`); w.Code != 422 {
		t.Fatal("forged purchase accepted")
	}
	body := `{"platform":"android","productId":"nihongo_master_lifetime","verificationData":"real-proof"}`
	verifier.state = "pending"
	if w = call("POST", "/v1/billing/verify", body); w.Code != 202 || memory.Profile(buyer.ID).Entitlement.Tier != domain.TierFree {
		t.Fatal("pending payment unlocks premium")
	}
	verifier.state = "purchased"
	if w = call("POST", "/v1/billing/verify", body); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w = call("GET", "/v1/grammar/lessons/grammar-n2-01", ""); w.Code != 200 {
		t.Fatal("paid user cannot read N2")
	}
	w = call("POST", "/v1/grammar/lessons/grammar-n2-01/attempts", "")
	var attempt struct {
		Data domain.GrammarAttempt `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &attempt)
	w = call("POST", "/v1/vocabulary/lessons/n3-daily-life/attempts", "")
	var vocabAttempt struct {
		Data domain.VocabularyAttempt `json:"data"`
	}
	if w.Code != 200 && w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &vocabAttempt); err != nil || vocabAttempt.Data.ID == "" {
		t.Fatal("missing vocabulary attempt")
	}
	p, _ := memory.Purchase("android", "transaction")
	p.State = "revoked"
	memory.UpdateVerifiedPurchase(p)
	if w = call("POST", "/v1/grammar/attempts/"+attempt.Data.ID+"/complete", ""); w.Code != 402 {
		t.Fatal("revoked user completes premium quiz")
	}
	if w = call("POST", "/v1/vocabulary/attempts/"+vocabAttempt.Data.ID+"/complete", ""); w.Code != 402 {
		t.Fatal("revoked user completes premium vocabulary quiz")
	}
	if w = call("DELETE", "/v1/me/account", `{"confirmation":"wrong"}`); w.Code != 422 {
		t.Fatal("deletion without confirmation accepted")
	}
	if w = call("DELETE", "/v1/me/account", `{"confirmation":"DELETE"}`); w.Code != 204 {
		t.Fatal(w.Body.String())
	}
	if w = call("GET", "/v1/auth/me", ""); w.Code != 401 {
		t.Fatal("deleted account keeps session")
	}
}
