package store

import (
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/testdb"
	"testing"
)

func openPersistentStore(t *testing.T, dsn string) *MemoryStore {
	t.Helper()
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	if err := s.EnableBillingPersistence(); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSessionsSurviveServerRestart(t *testing.T) {
	dsn := testdb.DSN(t)
	s := openPersistentStore(t, dsn)
	profile, err := s.AuthenticateGoogle(domain.GoogleIdentity{Subject: "restart-user", Email: "restart@example.com", DisplayName: "Restart"})
	if err != nil {
		t.Fatal(err)
	}
	token, err := s.CreateSession(profile.ID)
	if err != nil {
		t.Fatal(err)
	}
	var plaintext int64
	if err = s.ormDB.Model(&database.Session{}).Where("token_hash = ?", token).Count(&plaintext).Error; err != nil || plaintext != 0 {
		t.Fatalf("raw session token stored in database: %d %v", plaintext, err)
	}
	s.Close()

	// A new process has an empty in-memory session map and must restore the login.
	s = openPersistentStore(t, dsn)
	if id, ok := s.ResolveSession(token); !ok || id != profile.ID {
		t.Fatalf("session lost after restart: %q %v", id, ok)
	}
	if _, ok := s.ResolveSession("unknown-token"); ok {
		t.Fatal("unknown token accepted")
	}
	s.RevokeSession(token)
	s.Close()

	s = openPersistentStore(t, dsn)
	defer s.Close()
	if _, ok := s.ResolveSession(token); ok {
		t.Fatal("revoked session survives restart")
	}
}
