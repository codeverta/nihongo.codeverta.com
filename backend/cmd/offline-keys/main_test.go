package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreatePrivateKeyPersistsAndReusesKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "offline-access-private-key")

	first, created, err := loadOrCreatePrivateKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if !created || len(first) != ed25519.SeedSize {
		t.Fatalf("first load: created=%t seed bytes=%d", created, len(first))
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(contents)))
	if err != nil || !bytes.Equal(decoded, first) {
		t.Fatal("saved key is not the generated base64 seed")
	}

	second, created, err := loadOrCreatePrivateKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if created || !bytes.Equal(second, first) {
		t.Fatal("existing key was rotated instead of reused")
	}
}

func TestLoadOrCreateEncodedSecretPersistsAndRejectsCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "offline-practice-signing-key")

	first, created, err := loadOrCreateEncodedSecret(path, 32)
	if err != nil {
		t.Fatal(err)
	}
	if !created || len(first) != 32 {
		t.Fatalf("first load: created=%t secret bytes=%d", created, len(first))
	}
	second, created, err := loadOrCreateEncodedSecret(path, 32)
	if err != nil {
		t.Fatal(err)
	}
	if created || !bytes.Equal(second, first) {
		t.Fatal("existing secret was rotated instead of reused")
	}

	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadOrCreateEncodedSecret(path, 32); err == nil {
		t.Fatal("corrupt existing secret was silently replaced")
	}
}
