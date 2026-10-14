package store

import (
	"nihongo-master-backend/internal/testdb"
	"testing"
)

func TestDictionaryUsesExistingMariaDBConnection(t *testing.T) {
	s := NewMemoryStore()
	defer s.Close()
	if _, err := s.DictionaryCache(); err == nil {
		t.Fatal("dictionary cache initialized without database")
	}
	if err := s.EnableGamificationPersistence(testdb.DSN(t)); err != nil {
		t.Fatal(err)
	}
	if err := s.EnableBillingPersistence(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DictionaryCache(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.gamificationDB.QueryRow(`SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ('dictionary_responses','dictionary_entries','billing_purchases','gamification_users')`).Scan(&count); err != nil || count != 4 {
		t.Fatalf("database tables unavailable: %d %v", count, err)
	}
}
