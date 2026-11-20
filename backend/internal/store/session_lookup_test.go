package store

import (
	"database/sql"
	"errors"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// A database that cannot be reached must not look like a revoked login.
func TestResolveSessionReportsLookupFailureInsteadOfInvalid(t *testing.T) {
	sqlDB, err := sql.Open("mysql", "nobody:x@tcp(127.0.0.1:1)/none?timeout=200ms")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	orm, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	s := NewMemoryStore()
	s.ormDB, s.billingEnabled = orm, true

	if _, ok, err := s.ResolveSessionChecked("some-token"); ok || !errors.Is(err, ErrSessionLookup) {
		t.Fatalf("unreachable database: ok=%v err=%v, want lookup error", ok, err)
	}
	// Tokens already in memory keep working while the database is down.
	s.sessions["cached"] = "user-1"
	if id, ok, err := s.ResolveSessionChecked("cached"); !ok || err != nil || id != "user-1" {
		t.Fatalf("cached token: %q %v %v", id, ok, err)
	}
	// Without persistence an unknown token is plainly invalid, not an error.
	plain := NewMemoryStore()
	if _, ok, err := plain.ResolveSessionChecked("nope"); ok || err != nil {
		t.Fatalf("unknown token: ok=%v err=%v", ok, err)
	}
}
