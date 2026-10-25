package api

import (
	"crypto/hmac"
	"os"
	"path/filepath"
	"testing"
)

func TestOfflinePracticeSigningKeyFileIsStableAndFailsClosedWhenMissing(t *testing.T) {
	keyFile := filepath.Join(t.TempDir(), "offline-practice-key")
	if err := os.WriteFile(keyFile, []byte("stable-secret-from-key-helper\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fromFile := offlinePracticeSigningKey("inline-secret", keyFile, "database-secret")
	want := defaultOfflinePracticeSigningKey("stable-secret-from-key-helper", "")
	if !hmac.Equal(fromFile, want) {
		t.Fatal("key file did not take precedence or trim trailing whitespace")
	}
	missing := filepath.Join(t.TempDir(), "missing")
	if got := offlinePracticeSigningKey("inline-secret", missing, "database-secret"); len(got) != 0 {
		t.Fatal("configured but unreadable key file should disable package signing")
	}
}
