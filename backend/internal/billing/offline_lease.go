package billing

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"nihongo-master-backend/internal/domain"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const OfflineLeaseDuration = 30 * 24 * time.Hour

const (
	offlineLeaseIssuer   = "nihongo-master"
	offlineLeaseAudience = "nihongo-master-offline-premium"
	offlineLeaseKeyID    = "offline-access-v1"
)

var ErrOfflineLeaseUnavailable = errors.New("offline lease unavailable")

// OfflineLeaseSigner signs short-lived, account-bound proof for offline premium
// reads. Its private key is server configuration and is never sent to clients.
type OfflineLeaseSigner struct {
	privateKey ed25519.PrivateKey
}

func NewOfflineLeaseSigner(privateKey []byte) (*OfflineLeaseSigner, error) {
	switch len(privateKey) {
	case ed25519.SeedSize:
		privateKey = ed25519.NewKeyFromSeed(privateKey)
	case ed25519.PrivateKeySize:
		expected := ed25519.NewKeyFromSeed(privateKey[:ed25519.SeedSize])
		if !bytes.Equal(privateKey, expected) {
			return nil, fmt.Errorf("Ed25519 private key does not match its seed")
		}
		privateKey = append(ed25519.PrivateKey(nil), privateKey...)
	default:
		return nil, fmt.Errorf("offline access key must be a base64 Ed25519 seed or private key")
	}
	return &OfflineLeaseSigner{privateKey: ed25519.PrivateKey(privateKey)}, nil
}

// NewOfflineLeaseSignerFromEnv returns nil when offline access is not
// configured. Invalid non-empty configuration is an error, never a fallback.
func NewOfflineLeaseSignerFromEnv() (*OfflineLeaseSigner, error) {
	encoded := strings.TrimSpace(os.Getenv("OFFLINE_ACCESS_PRIVATE_KEY"))
	if encoded == "" {
		privateKeyPath := strings.TrimSpace(os.Getenv("OFFLINE_ACCESS_PRIVATE_KEY_FILE"))
		if privateKeyPath == "" {
			return nil, nil
		}
		keyBytes, err := readOfflineAccessPrivateKey(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read OFFLINE_ACCESS_PRIVATE_KEY_FILE: %w", err)
		}
		encoded = strings.TrimSpace(string(keyBytes))
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		key, err = base64.RawStdEncoding.DecodeString(encoded)
	}
	if err != nil {
		return nil, fmt.Errorf("decode OFFLINE_ACCESS_PRIVATE_KEY: %w", err)
	}
	return NewOfflineLeaseSigner(key)
}

func readOfflineAccessPrivateKey(path string) ([]byte, error) {
	if filepath.IsAbs(path) {
		return os.ReadFile(path)
	}
	key, err := os.ReadFile(path)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		return key, err
	}
	// The API runs from either the backend directory or the repository root.
	return os.ReadFile(filepath.Join("backend", path))
}

// PublicKeyBase64 exposes only the public verification key for release build
// configuration. The private signing key must remain on the API server.
func (s *OfflineLeaseSigner) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(s.privateKey.Public().(ed25519.PublicKey))
}

type offlineLeaseHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

type offlineLeaseClaims struct {
	Issuer     string `json:"iss"`
	Subject    string `json:"sub"`
	Audience   string `json:"aud"`
	Tier       string `json:"tier"`
	IsLifetime bool   `json:"isLifetime"`
	IssuedAt   int64  `json:"iat"`
	ExpiresAt  int64  `json:"exp"`
	PurchaseID string `json:"jti"`
}

// SignPurchase uses the time of the most recent successful store verification
// as the lease start, so reading /billing repeatedly cannot extend the lease.
func (s *OfflineLeaseSigner) SignPurchase(p domain.Purchase, now time.Time) (string, error) {
	if s == nil || p.UserID == "" || p.ID == "" || p.Platform == "" || p.ProductID == "" || p.State != "purchased" || p.CheckedAt.IsZero() {
		return "", ErrOfflineLeaseUnavailable
	}
	verifiedAt := p.CheckedAt.UTC()
	expiresAt := verifiedAt.Add(OfflineLeaseDuration)
	if verifiedAt.After(now.UTC().Add(5*time.Minute)) || !now.UTC().Before(expiresAt) {
		return "", ErrOfflineLeaseUnavailable
	}
	header, err := json.Marshal(offlineLeaseHeader{Algorithm: "EdDSA", Type: "JWT", KeyID: offlineLeaseKeyID})
	if err != nil {
		return "", err
	}
	claims, err := json.Marshal(offlineLeaseClaims{
		Issuer: offlineLeaseIssuer, Subject: p.UserID,
		Audience: offlineLeaseAudience, Tier: "premium", IsLifetime: true,
		IssuedAt: verifiedAt.Unix(), ExpiresAt: expiresAt.Unix(),
		PurchaseID: p.Platform + ":" + p.ID,
	})
	if err != nil {
		return "", err
	}
	encode := base64.RawURLEncoding.EncodeToString
	message := encode(header) + "." + encode(claims)
	signature := ed25519.Sign(s.privateKey, []byte(message))
	return message + "." + encode(signature), nil
}

// OfflineLease returns proof only for an active, correctly-owned purchase.
// CheckedAt is advanced only after the store verifier accepts a purchase.
func (s *Service) OfflineLease(user string, now time.Time) string {
	if s == nil || s.OfflineSigner == nil || user == "" {
		return ""
	}
	var latest *domain.Purchase
	for _, purchase := range s.Store.Purchases(user) {
		if purchase.UserID != user || purchase.ProductID != s.ProductID || purchase.State != "purchased" {
			continue
		}
		if latest == nil || purchase.CheckedAt.After(latest.CheckedAt) {
			copy := purchase
			latest = &copy
		}
	}
	if latest == nil {
		return ""
	}
	proof, err := s.OfflineSigner.SignPurchase(*latest, now)
	if err != nil {
		return ""
	}
	return proof
}
