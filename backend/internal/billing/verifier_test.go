package billing

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
	"strings"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestGoogleProductV2Validation(t *testing.T) {
	for _, tc := range []struct {
		name, product, state, consumed, expected string
		refundable                               int
	}{
		{"paid", "lifetime", "PURCHASED", "CONSUMPTION_STATE_YET_TO_BE_CONSUMED", "purchased", 1},
		{"pending", "lifetime", "PENDING", "CONSUMPTION_STATE_YET_TO_BE_CONSUMED", "pending", 1},
		{"cancelled", "lifetime", "CANCELLED", "CONSUMPTION_STATE_YET_TO_BE_CONSUMED", "revoked", 0},
		{"full-refund", "lifetime", "PURCHASED", "CONSUMPTION_STATE_YET_TO_BE_CONSUMED", "revoked", 0},
		{"consumed", "lifetime", "PURCHASED", "CONSUMPTION_STATE_CONSUMED", "revoked", 1},
		{"wrong-product", "other", "PURCHASED", "CONSUMPTION_STATE_YET_TO_BE_CONSUMED", "invalid", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host != "androidpublisher.googleapis.com" || !strings.Contains(r.URL.Path, "/applications/com.codeverta.nihongomaster/") {
					t.Fatal("wrong verification endpoint")
				}
				raw, _ := json.Marshal(map[string]any{"productLineItem": []any{map[string]any{"productId": tc.product, "productOfferDetails": map[string]any{"refundableQuantity": tc.refundable, "consumptionState": tc.consumed}}}, "purchaseStateContext": map[string]string{"purchaseState": tc.state}, "obfuscatedExternalAccountId": "bound-account", "acknowledgementState": "ACKNOWLEDGEMENT_STATE_ACKNOWLEDGED"})
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
			})}
			v := &StoreVerifier{ProductID: "lifetime", PackageID: "com.codeverta.nihongomaster", googleClient: client}
			p, err := v.Verify(context.Background(), domain.PurchaseProof{Platform: "android", ProductID: "lifetime", VerificationData: "token"})
			if tc.expected == "invalid" {
				if !errors.Is(err, ErrInvalid) {
					t.Fatal("wrong product accepted")
				}
				return
			}
			if err != nil || p.State != tc.expected || p.AccountToken != "bound-account" {
				t.Fatalf("wrong verified state %+v %v", p, err)
			}
			if err = v.Acknowledge(context.Background(), p); err != nil || calls != 1 {
				t.Fatal("acknowledged purchase acknowledged twice")
			}
		})
	}
}

type fakeVerifier struct {
	p           domain.Purchase
	err, ackErr error
	acks        int
}

func (v *fakeVerifier) Platforms() []string { return []string{"android"} }
func (v *fakeVerifier) Verify(context.Context, domain.PurchaseProof) (domain.Purchase, error) {
	p := v.p
	p.CheckedAt = time.Now()
	return p, v.err
}
func (v *fakeVerifier) Acknowledge(context.Context, domain.Purchase) error { v.acks++; return v.ackErr }
func TestVerificationRetryReconciliationAndOwnership(t *testing.T) {
	memory := store.NewMemoryStore()
	token, _ := memory.BillingAccountToken("buyer")
	verifier := &fakeVerifier{p: domain.Purchase{Platform: "android", ID: "token", ProductID: "lifetime", AccountToken: token, Proof: "token", State: "purchased"}, ackErr: ErrUnavailable}
	service := &Service{Store: memory, Verifier: verifier, ProductID: "lifetime"}
	proof := domain.PurchaseProof{Platform: "android", ProductID: "lifetime", VerificationData: "token"}
	if _, err := service.Verify(context.Background(), "buyer", proof); !errors.Is(err, ErrUnavailable) {
		t.Fatal("ack failure hidden")
	}
	if len(memory.Purchases("buyer")) != 1 {
		t.Fatal("verified ownership lost during ack failure")
	}
	verifier.ackErr = nil
	if _, err := service.Verify(context.Background(), "buyer", proof); err != nil || len(memory.Purchases("buyer")) != 1 {
		t.Fatal("retry duplicated ownership")
	}
	memory.BillingAccountToken("other")
	if _, err := service.Verify(context.Background(), "other", proof); !errors.Is(err, store.ErrPremium) {
		t.Fatal("replay to other account accepted")
	}
	verifier.p.State = "revoked"
	if err := service.RefreshPurchase(context.Background(), "android", "token"); err != nil {
		t.Fatal(err)
	}
	if memory.Profile("buyer").Entitlement.Tier != domain.TierFree {
		t.Fatal("refund did not revoke access")
	}
	verifier.err = ErrUnavailable
	verifier.p.State = "purchased"
	if err := service.RefreshPurchase(context.Background(), "android", "token"); err == nil {
		t.Fatal("outage hidden")
	}
	if memory.Profile("buyer").Entitlement.Tier != domain.TierFree {
		t.Fatal("outage resurrected refunded access")
	}
}
