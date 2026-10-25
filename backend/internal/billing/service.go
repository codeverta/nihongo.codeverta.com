package billing

import (
	"context"
	"errors"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
	"sync"
	"time"
)

// Serialize verified state changes, including notifications and reconciliation,
// to prevent late network responses overwriting a newer revocation.
type Service struct {
	mu        sync.Mutex
	Store     *store.MemoryStore
	Verifier  Verifier
	ProductID string
	// OfflineSigner remains nil unless a stable signing key is configured.
	OfflineSigner *OfflineLeaseSigner
}

func (s *Service) Verify(ctx context.Context, user string, proof domain.PurchaseProof) (domain.Purchase, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, err := s.Verifier.Verify(ctx, proof)
	if err != nil {
		return p, err
	}
	if p.State == "pending" {
		return p, nil
	}
	if err = s.Store.SaveVerifiedPurchase(user, p); err != nil {
		return p, err
	}
	if err = s.Verifier.Acknowledge(ctx, p); err != nil {
		return p, err
	}
	return p, nil
}
func (s *Service) refreshLocked(ctx context.Context, p domain.Purchase) error {
	current, err := s.Verifier.Verify(ctx, domain.PurchaseProof{Platform: p.Platform, ProductID: p.ProductID, VerificationData: p.Proof})
	if errors.Is(err, ErrInvalid) {
		p.State = "revoked"
		p.CheckedAt = time.Now().UTC()
		return s.Store.UpdateVerifiedPurchase(p)
	}
	if err != nil {
		return err
	}
	if current.ID != p.ID {
		return ErrInvalid
	}
	if err = s.Store.UpdateVerifiedPurchase(current); err != nil {
		return err
	}
	return s.Verifier.Acknowledge(ctx, current)
}
func (s *Service) RefreshUser(ctx context.Context, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.Store.Purchases(user) {
		if time.Since(p.CheckedAt) > 5*time.Minute {
			if err := s.refreshLocked(ctx, p); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *Service) Reconcile(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var first error
	for _, p := range s.Store.Purchases("") {
		if time.Since(p.CheckedAt) > 5*time.Minute {
			if err := s.refreshLocked(ctx, p); err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}
func (s *Service) RefreshPurchase(ctx context.Context, platform, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.Store.Purchase(platform, id)
	if !ok {
		return nil
	}
	return s.refreshLocked(ctx, p)
}
func (s *Service) AppleNotification(ctx context.Context, signed string) error {
	verifier, ok := s.Verifier.(*StoreVerifier)
	if !ok || verifier.appleKey == nil {
		return ErrUnavailable
	}
	var notification struct {
		Type string `json:"notificationType"`
		Data struct {
			BundleID    string `json:"bundleId"`
			Environment string `json:"environment"`
			Transaction string `json:"signedTransactionInfo"`
		} `json:"data"`
	}
	if err := verifier.decodeAppleJWS(signed, &notification); err != nil {
		return err
	}
	if notification.Type == "TEST" {
		return nil
	}
	if notification.Data.BundleID != verifier.BundleID || notification.Data.Environment != verifier.AppleEnvironment {
		return ErrInvalid
	}
	if notification.Data.Transaction == "" {
		return nil
	}
	var transaction appleTransaction
	if err := verifier.decodeAppleJWS(notification.Data.Transaction, &transaction); err != nil {
		return err
	}
	if transaction.BundleID != verifier.BundleID || transaction.ProductID != s.ProductID || transaction.Environment != verifier.AppleEnvironment {
		return ErrInvalid
	}
	return s.RefreshPurchase(ctx, "ios", transaction.OriginalID)
}
