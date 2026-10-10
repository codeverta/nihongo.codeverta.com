package store

import (
	"path/filepath"
	"testing"
)

func TestDictionaryUsesExistingExclusiveDatabaseConnection(t *testing.T) {
	s := NewMemoryStore()
	defer s.Close()
	if _, err := s.DictionaryCache(); err == nil {
		t.Fatal("dictionary cache initialized without database")
	}
	if err := s.EnableGamificationPersistence(filepath.Join(t.TempDir(), "state.db")); err != nil {
		t.Fatal(err)
	}
	if err := s.EnableBillingPersistence(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DictionaryCache(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.gamificationDB.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name IN ('dictionary_responses','dictionary_entries','billing_purchases','gamification_users')`).Scan(&count); err != nil || count != 4 {
		t.Fatalf("database tables unavailable: %d %v", count, err)
	}
}
