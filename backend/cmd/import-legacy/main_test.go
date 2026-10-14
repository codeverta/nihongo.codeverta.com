package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/testdb"
)

func TestLegacyImportPreservesBlobsAndRejectsOverwrite(t *testing.T) {
	dsn := testdb.DSN(t)
	state := []byte(`{"XP":50,"Events":{},"Rewards":{},"Lessons":{},"Subjects":{},"Days":{},"Unlocked":{},"Exams":{}}`)
	body := []byte(`{"kanji":"猫","stroke_count":11}`)
	tables := map[string]any{
		"gamification_users":   []map[string]any{{"user_id": "legacy-user", "state": state}},
		"accounts":             []map[string]any{},
		"billing_purchases":    []map[string]any{},
		"billing_tokens":       []map[string]any{},
		"deleted_demo":         []map[string]any{},
		"dictionary_responses": []map[string]any{},
		"dictionary_entries":   []map[string]any{{"source": "kanjiapi.dev", "entry_id": "猫", "kind": "kanji", "body": body, "fetched_at": int64(1700000000000)}},
	}
	raw, err := json.Marshal(map[string]any{"version": 1, "tables": tables})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "export.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := importLegacy(dsn, path); err != nil {
		t.Fatal(err)
	}
	if err := importLegacy(dsn, path); err == nil {
		t.Fatal("nonempty target overwritten")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var actual []byte
	if err := db.QueryRow(`SELECT body FROM dictionary_entries WHERE source_entry_id='猫'`).Scan(&actual); err != nil || !bytes.Equal(actual, body) {
		t.Fatalf("legacy UTF-8 blob changed: %v", err)
	}
	s := store.NewMemoryStore()
	if err := s.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.Gamification(database.UserUUID("legacy-user")).XP != 50 {
		t.Fatal("legacy XP not restored")
	}
}
