package billing

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nihongo-master-backend/internal/domain"
)

func TestOfflineLeaseIsSignedAccountBoundAndFixedToVerifiedAt(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	signer, err := NewOfflineLeaseSigner(seed)
	if err != nil {
		t.Fatal(err)
	}
	verifiedAt := time.Date(2026, 10, 23, 1, 2, 3, 0, time.UTC)
	now := verifiedAt.Add(time.Hour)
	purchase := domain.Purchase{
		Platform: "android", ID: "store-123", ProductID: "nihongo_master_lifetime",
		UserID: "account-1", State: "purchased", CheckedAt: verifiedAt,
	}

	proof, err := signer.SignPurchase(purchase, now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(proof, ".")
	if len(parts) != 3 {
		t.Fatalf("lease parts = %d, want 3", len(parts))
	}
	message := parts[0] + "." + parts[1]
	publicKey := signer.privateKey.Public().(ed25519.PublicKey)
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !ed25519.Verify(publicKey, []byte(message), signature) {
		t.Fatal("lease signature did not verify")
	}
	var header offlineLeaseHeader
	var claims offlineLeaseClaims
	headerData, _ := base64.RawURLEncoding.DecodeString(parts[0])
	claimsData, _ := base64.RawURLEncoding.DecodeString(parts[1])
	if err := json.Unmarshal(headerData, &header); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(claimsData, &claims); err != nil {
		t.Fatal(err)
	}
	if header.Algorithm != "EdDSA" || header.KeyID != offlineLeaseKeyID || claims.Subject != "account-1" || claims.Tier != "premium" || !claims.IsLifetime {
		t.Fatalf("unexpected signed lease: header=%+v claims=%+v", header, claims)
	}
	if claims.IssuedAt != verifiedAt.Unix() || claims.ExpiresAt != verifiedAt.Add(OfflineLeaseDuration).Unix() {
		t.Fatalf("lease was not anchored to verification time: %+v", claims)
	}
	if _, err := signer.SignPurchase(purchase, verifiedAt.Add(OfflineLeaseDuration)); !errors.Is(err, ErrOfflineLeaseUnavailable) {
		t.Fatalf("expired lease error = %v", err)
	}
	if _, err := signer.SignPurchase(domain.Purchase{Platform: "android", ID: "store-123", ProductID: purchase.ProductID, UserID: "account-1", State: "revoked", CheckedAt: verifiedAt}, now); !errors.Is(err, ErrOfflineLeaseUnavailable) {
		t.Fatalf("revoked purchase error = %v", err)
	}
}

func TestOfflineLeaseSignerLoadsPersistentKeyFile(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(255 - i)
	}
	path := filepath.Join(t.TempDir(), "offline-access-private-key")
	if err := os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(seed)), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OFFLINE_ACCESS_PRIVATE_KEY", "")
	t.Setenv("OFFLINE_ACCESS_PRIVATE_KEY_FILE", path)

	signer, err := NewOfflineLeaseSignerFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	want, err := NewOfflineLeaseSigner(seed)
	if err != nil {
		t.Fatal(err)
	}
	if signer.PublicKeyBase64() != want.PublicKeyBase64() {
		t.Fatal("file configured signer does not match stored seed")
	}
}
