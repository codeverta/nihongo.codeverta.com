package unlock_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"nihongo-master-backend/internal/unlock"
)

// keyServer serves verifier keys in the format Google publishes them.
type keyServer struct {
	mu    sync.Mutex
	keys  map[string]*ecdsa.PublicKey
	calls int
	fail  bool
}

func (k *keyServer) set(id string) *ecdsa.PrivateKey {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	k.mu.Lock()
	defer k.mu.Unlock()
	k.keys = map[string]*ecdsa.PublicKey{id: &private.PublicKey}
	return private
}

func (k *keyServer) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.calls++
	if k.fail {
		http.Error(w, "down", http.StatusServiceUnavailable)
		return
	}
	type entry struct {
		KeyID json.Number `json:"keyId"`
		PEM   string      `json:"pem"`
	}
	var entries []entry
	for id, key := range k.keys {
		der, _ := x509.MarshalPKIXPublicKey(key)
		entries = append(entries, entry{json.Number(id), string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"keys": entries})
}

func TestGoogleKeysAreCachedAndRefreshedWithoutBeingHammered(t *testing.T) {
	server := &keyServer{}
	first := server.set("111")
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()

	now := time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)
	keys := &unlock.GoogleKeys{URL: httpServer.URL, Client: httpServer.Client(), Now: func() time.Time { return now }}
	ctx := context.Background()

	got, err := keys.PublicKey(ctx, "111")
	if err != nil || !got.Equal(&first.PublicKey) {
		t.Fatalf("PublicKey(111) = %v, %v", got, err)
	}
	if _, err := keys.PublicKey(ctx, "111"); err != nil || server.calls != 1 {
		t.Fatalf("a cached key must not be fetched again: calls=%d err=%v", server.calls, err)
	}

	// Google rotates keys: an unknown ID triggers one refresh, but not within a minute of the last attempt.
	second := server.set("222")
	if _, err := keys.PublicKey(ctx, "222"); err == nil {
		t.Fatal("unknown key accepted within the refresh interval")
	}
	if server.calls != 1 {
		t.Fatalf("refresh was not throttled: calls=%d", server.calls)
	}
	now = now.Add(2 * time.Minute)
	got, err = keys.PublicKey(ctx, "222")
	if err != nil || !got.Equal(&second.PublicKey) || server.calls != 2 {
		t.Fatalf("rotated key: %v, calls=%d, err=%v", got, server.calls, err)
	}
	if _, err := keys.PublicKey(ctx, "999"); err == nil {
		t.Fatal("a key Google never published was accepted")
	}
}

func TestGoogleKeysRefuseBrokenResponses(t *testing.T) {
	server := &keyServer{fail: true}
	server.set("111")
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	keys := &unlock.GoogleKeys{URL: httpServer.URL, Client: httpServer.Client()}
	if _, err := keys.PublicKey(context.Background(), "111"); err == nil {
		t.Fatal("a failing key server must not yield a key")
	}
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"keys":[]}`)) }))
	defer empty.Close()
	keys = &unlock.GoogleKeys{URL: empty.URL, Client: empty.Client()}
	if _, err := keys.PublicKey(context.Background(), "111"); err == nil {
		t.Fatal("an empty key list must not yield a key")
	}
}
