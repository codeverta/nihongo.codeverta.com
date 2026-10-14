package store

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/testdb"
)

func TestUUIDAccountOwnershipAndStablePurchasePrimaryKey(t *testing.T) {
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(testdb.DSN(t)); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.EnableBillingPersistence(); err != nil {
		t.Fatal(err)
	}
	identity := domain.GoogleIdentity{Subject: "subject", Email: "uuid@example.com", DisplayName: "UUID account"}
	profile, err := s.AuthenticateGoogle(identity)
	if err != nil {
		t.Fatal(err)
	}
	if err := uuid.Validate(profile.ID); err != nil {
		t.Fatal("account user ID is not UUID")
	}
	again, err := s.AuthenticateGoogle(identity)
	if err != nil || again.ID != profile.ID {
		t.Fatal("login changed user UUID")
	}
	token, err := s.BillingAccountToken(profile.ID)
	if err != nil {
		t.Fatal(err)
	}
	purchase := domain.Purchase{Platform: "android", ID: "external-purchase-id", Proof: "external-proof", AccountToken: token, State: "purchased", CheckedAt: time.Now()}
	if err := s.SaveVerifiedPurchase(profile.ID, purchase); err != nil {
		t.Fatal(err)
	}
	var row database.BillingPurchase
	if err := s.ormDB.Where("purchase_key = ?", "android:external-purchase-id").Take(&row).Error; err != nil {
		t.Fatal(err)
	}
	if err := uuid.Validate(row.ID); err != nil {
		t.Fatal("purchase primary key is not UUID")
	}
	if row.UserID == nil || *row.UserID != profile.ID {
		t.Fatal("purchase owner is not account UUID")
	}
	firstID := row.ID
	if err := s.SaveVerifiedPurchase(profile.ID, purchase); err != nil {
		t.Fatal(err)
	}
	if err := s.ormDB.Where("purchase_key = ?", "android:external-purchase-id").Take(&row).Error; err != nil || row.ID != firstID {
		t.Fatal("purchase upsert changed row UUID")
	}
	var account database.Account
	var billingToken database.BillingToken
	if err := s.ormDB.Where("email = ?", identity.Email).Take(&account).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.ormDB.Where("user_id = ?", profile.ID).Take(&billingToken).Error; err != nil {
		t.Fatal(err)
	}
	if account.ID != profile.ID || uuid.Validate(billingToken.ID) != nil || billingToken.UserID != account.ID {
		t.Fatal("UUID relations disagree")
	}
}
