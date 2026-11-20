package store

import (
	"errors"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/testdb"
	"testing"
	"time"
)

func grantTestLifetime(t *testing.T, s *MemoryStore, user string) {
	t.Helper()
	token, err := s.BillingAccountToken(user)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.SaveVerifiedPurchase(user, domain.Purchase{Platform: "android", ID: "test-" + user, ProductID: "nihongo_master_lifetime", AccountToken: token, Proof: "test-" + user, State: "purchased", CheckedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
}
func TestBillingOwnershipFreshnessAndMultipleStores(t *testing.T) {
	s := NewMemoryStore()
	grantTestLifetime(t, s, "buyer")
	p, _ := s.Purchase("android", "test-buyer")
	s.BillingAccountToken("other")
	if err := s.SaveVerifiedPurchase("other", p); !errors.Is(err, ErrPurchaseOwned) {
		t.Fatal("purchase replay across accounts allowed")
	}
	if s.Profile("buyer").Entitlement.Tier != domain.TierPremium || s.Profile("other").Entitlement.Tier != domain.TierFree {
		t.Fatal("incorrect ownership")
	}
	p.CheckedAt = time.Now().Add(-16 * time.Minute)
	s.UpdateVerifiedPurchase(p)
	if s.Profile("buyer").Entitlement.Tier != domain.TierFree {
		t.Fatal("stale verification grants premium")
	}
	token, _ := s.BillingAccountToken("buyer")
	if err := s.SaveVerifiedPurchase("buyer", domain.Purchase{Platform: "ios", ID: "ios-purchase", ProductID: p.ProductID, AccountToken: token, State: "purchased", CheckedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	p.State = "revoked"
	p.CheckedAt = time.Now()
	s.UpdateVerifiedPurchase(p)
	if s.Profile("buyer").Entitlement.Tier != domain.TierPremium {
		t.Fatal("refund on one store revoked another active purchase")
	}
}
func TestPromoPurchaseBindsToFirstVerifier(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.AuthenticateGoogle(domain.GoogleIdentity{Subject: "a", Email: "abcdef@gmail.com", DisplayName: "A"}); err != nil {
		t.Fatal(err)
	}
	first, _ := s.AuthenticateGoogle(domain.GoogleIdentity{Subject: "a", Email: "abcdef@gmail.com", DisplayName: "A"})
	second, _ := s.AuthenticateGoogle(domain.GoogleIdentity{Subject: "b", Email: "zz@example.com", DisplayName: "B"})
	s.BillingAccountToken(first.ID)
	s.BillingAccountToken(second.ID)
	promo := domain.Purchase{Platform: "android", ID: "promo", ProductID: "nihongo_master_lifetime", Proof: "promo", State: "purchased", CheckedAt: time.Now()}
	if err := s.SaveVerifiedPurchase(first.ID, promo); err != nil {
		t.Fatal("promo purchase without an account token rejected:", err)
	}
	if s.Profile(first.ID).Entitlement.Tier != domain.TierPremium {
		t.Fatal("promo claim did not grant Lifetime")
	}
	err := s.SaveVerifiedPurchase(second.ID, promo)
	var owned *PurchaseOwnedError
	if !errors.As(err, &owned) || owned.Owner != "ab***@gmail.com" {
		t.Fatalf("claimed promo moved or owner hidden: %v", err)
	}
	if s.Profile(second.ID).Entitlement.Tier != domain.TierFree {
		t.Fatal("second account got the claimed promo")
	}
	// A purchase bought for the first account names it to the second.
	token, _ := s.BillingAccountToken(first.ID)
	bought := domain.Purchase{Platform: "android", ID: "bought", ProductID: "nihongo_master_lifetime", AccountToken: token, State: "purchased", CheckedAt: time.Now()}
	if err := s.SaveVerifiedPurchase(second.ID, bought); !errors.As(err, &owned) || owned.Owner != "ab***@gmail.com" {
		t.Fatalf("purchase for another live account accepted: %v", err)
	}
	// iOS purchases always need the account token.
	if err := s.SaveVerifiedPurchase(first.ID, domain.Purchase{Platform: "ios", ID: "ios", State: "purchased", CheckedAt: time.Now()}); !errors.Is(err, ErrPremium) {
		t.Fatal("iOS purchase without an account token accepted")
	}
}

func TestMaskEmail(t *testing.T) {
	for in, want := range map[string]string{"abcdef@gmail.com": "ab***@gmail.com", "a@x.io": "a***@x.io", "": "", "bad": ""} {
		if got := MaskEmail(in); got != want {
			t.Fatalf("MaskEmail(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBillingDeletionRestoreAndRestart(t *testing.T) {
	path := testdb.DSN(t)
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(path); err != nil {
		t.Fatal(err)
	}
	if err := s.EnableBillingPersistence(); err != nil {
		t.Fatal(err)
	}
	identity := domain.GoogleIdentity{Subject: "same-identity", Email: "buyer@example.com", DisplayName: "Buyer"}
	buyer, err := s.AuthenticateGoogle(identity)
	if err != nil {
		t.Fatal(err)
	}
	grantTestLifetime(t, s, buyer.ID)
	session, _ := s.CreateSession(buyer.ID)
	a, _ := s.StartGrammarAttempt(buyer.ID, "grammar-n3-01")
	if err = s.DeleteAccount(buyer.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.ResolveSession(session); ok {
		t.Fatal("session survives account deletion")
	}
	if _, err = s.AnswerGrammarQuestion(buyer.ID, a.ID, a.Questions[0].ID, a.Questions[0].Options[0].ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("deleted account quiz survives")
	}
	old, _ := s.Purchase("android", "test-"+buyer.ID)
	if old.UserID != "" || s.gamification[buyer.ID] != nil || s.progress[buyer.ID] != nil {
		t.Fatal("personal data survives deletion")
	}
	if err = s.SaveVerifiedPurchase(buyer.ID, old); err == nil {
		t.Fatal("in-flight verification resurrected deleted account")
	}
	s.Close()
	s = NewMemoryStore()
	defer s.Close()
	if err = s.EnableGamificationPersistence(path); err != nil {
		t.Fatal(err)
	}
	if err = s.EnableBillingPersistence(); err != nil {
		t.Fatal(err)
	}
	restored, err := s.AuthenticateGoogle(identity)
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID == buyer.ID {
		t.Fatal("deleted account ID resurrected")
	}
	s.BillingAccountToken(restored.ID)
	if err = s.SaveVerifiedPurchase(restored.ID, old); err != nil {
		t.Fatal(err)
	}
	if s.Profile(restored.ID).Entitlement.Tier != domain.TierPremium {
		t.Fatal("valid restore after deletion failed")
	}
}
func TestPaidExamPolicy(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.StartSimulation("free", "N3"); !errors.Is(err, ErrPremium) {
		t.Fatal("N3 simulation is free")
	}
	if _, err := s.StartSimulation("free", "N5"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartExam("free", "nat-5q-full-1"); !errors.Is(err, ErrPremium) {
		t.Fatal("NAT exam is free")
	}
	if len(s.Plans()) != 2 || s.Plans()[1].PriceIDR != 99000 || s.Plans()[1].BillingType != "lifetime" {
		t.Fatal("incorrect plan catalog")
	}
}
