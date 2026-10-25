package api_test

import (
	"context"
	"crypto/ed25519"
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

type billingVerifier struct {
	token, state string
	err          error
	checkedAt    time.Time
}

func (v *billingVerifier) Platforms() []string                                { return []string{"android", "ios"} }
func (v *billingVerifier) Acknowledge(context.Context, domain.Purchase) error { return nil }
func (v *billingVerifier) Verify(_ context.Context, p domain.PurchaseProof) (domain.Purchase, error) {
	if v.err != nil {
		return domain.Purchase{}, v.err
	}
	if p.VerificationData != "real-proof" || p.ProductID != "nihongo_master_lifetime" {
		return domain.Purchase{}, billing.ErrInvalid
	}
	checkedAt := v.checkedAt
	if checkedAt.IsZero() {
		checkedAt = time.Now()
	}
	return domain.Purchase{Platform: p.Platform, ID: "transaction", ProductID: p.ProductID, Proof: "real-proof", AccountToken: v.token, State: v.state, CheckedAt: checkedAt}, nil
}
func TestBillingAPIUnlockRevokeAndDelete(t *testing.T) {
	memory := store.NewMemoryStore()
	verifier := &billingVerifier{state: "purchased"}
	server := api.NewServer(memory, fakeGoogleVerifier{})
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	signer, err := billing.NewOfflineLeaseSigner(seed)
	if err != nil {
		t.Fatal(err)
	}
	server.ConfigureBilling(&billing.Service{Store: memory, Verifier: verifier, ProductID: "nihongo_master_lifetime", OfflineSigner: signer}, "", "")
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
	verifier.checkedAt = time.Now().UTC().Add(-5 * time.Minute)
	if w = call("POST", "/v1/billing/verify", body); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w = call("GET", "/v1/grammar/lessons/grammar-n2-01", ""); w.Code != 200 {
		t.Fatal("paid user cannot read N2")
	}
	w = call("GET", "/v1/billing", "")
	var activeBilling struct {
		Data struct {
			OfflineLease string `json:"offlineLease"`
		} `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &activeBilling) != nil || activeBilling.Data.OfflineLease == "" {
		t.Fatal("verified lifetime purchase did not receive a signed offline lease")
	}
	if same := call("GET", "/v1/billing", ""); same.Code == 200 {
		var snapshot struct {
			Data struct {
				OfflineLease string `json:"offlineLease"`
			} `json:"data"`
		}
		if json.Unmarshal(same.Body.Bytes(), &snapshot) != nil || snapshot.Data.OfflineLease != activeBilling.Data.OfflineLease {
			t.Fatal("billing polling extended the offline lease without store revalidation")
		}
	} else {
		t.Fatalf("second billing status = %d", same.Code)
	}
	verifiedPurchase, _ := memory.Purchase("android", "transaction")
	verifiedPurchase.CheckedAt = time.Now().UTC().Add(-10 * time.Minute)
	if err := memory.UpdateVerifiedPurchase(verifiedPurchase); err != nil {
		t.Fatal(err)
	}
	verifier.err = billing.ErrUnavailable
	w = call("GET", "/v1/billing", "")
	if w.Code != 503 {
		t.Fatalf("billing outage status = %d, want 503", w.Code)
	}
	stillPurchased, _ := memory.Purchase("android", "transaction")
	if stillPurchased.State != "purchased" {
		t.Fatal("temporary store outage revoked the server purchase")
	}
	verifier.err = nil
	verifier.checkedAt = time.Now().UTC()
	if w = call("GET", "/v1/billing", ""); w.Code != 200 {
		t.Fatal("billing did not recover after temporary store outage")
	}
	var renewedBilling struct {
		Data struct {
			OfflineLease string `json:"offlineLease"`
		} `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &renewedBilling) != nil || renewedBilling.Data.OfflineLease == "" || renewedBilling.Data.OfflineLease == activeBilling.Data.OfflineLease {
		t.Fatal("authoritative store revalidation did not issue a renewed lease")
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
	w = call("GET", "/v1/billing", "")
	var revokedBilling struct {
		Data struct {
			Entitlement  domain.Entitlement `json:"entitlement"`
			OfflineLease string             `json:"offlineLease"`
		} `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &revokedBilling) != nil || revokedBilling.Data.Entitlement.Tier != domain.TierFree || revokedBilling.Data.OfflineLease != "" {
		t.Fatal("confirmed refund did not clear the offline lease")
	}
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
