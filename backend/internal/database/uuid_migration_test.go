package database_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/testdb"
)

func TestUUIDMigrationPreservesRelationsAndExternalIDs(t *testing.T) {
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	defer pool.Close()
	statements := []string{
		`CREATE TABLE accounts(email VARCHAR(320) PRIMARY KEY,state LONGBLOB NOT NULL)`,
		`CREATE TABLE gamification_users(user_id VARCHAR(191) PRIMARY KEY,state LONGBLOB NOT NULL)`,
		`CREATE TABLE billing_purchases(id VARBINARY(2048) PRIMARY KEY,state LONGBLOB NOT NULL)`,
		`CREATE TABLE billing_tokens(user_id VARCHAR(191) PRIMARY KEY,token CHAR(36) UNIQUE NOT NULL)`,
		`CREATE TABLE deleted_demo(id VARCHAR(191) PRIMARY KEY)`,
		`CREATE TABLE dictionary_responses(cache_key CHAR(64) PRIMARY KEY,source VARCHAR(64),body LONGBLOB,fetched_at BIGINT)`,
		`CREATE TABLE dictionary_entries(source VARCHAR(64),entry_id VARCHAR(512),kind VARCHAR(32),body LONGBLOB,fetched_at BIGINT,PRIMARY KEY(source,entry_id))`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	oldUser := "google-legacy-identity"
	account, _ := json.Marshal(map[string]any{"ID": oldUser, "Email": "legacy@example.com", "DisplayName": "Legacy"})
	purchase, _ := json.Marshal(map[string]any{"id": "external-store-token", "userId": oldUser, "proof": "private-proof", "platform": "android"})
	token := uuid.NewString()
	inputs := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO accounts VALUES(?,?)`, []any{"legacy@example.com", account}},
		{`INSERT INTO gamification_users VALUES(?,?)`, []any{oldUser, []byte(`{"XP":50}`)}},
		{`INSERT INTO billing_purchases VALUES(?,?)`, []any{"android:external-store-token", purchase}},
		{`INSERT INTO billing_tokens VALUES(?,?)`, []any{oldUser, token}},
		{`INSERT INTO deleted_demo VALUES(?)`, []any{"demo-user"}},
		{`INSERT INTO dictionary_responses VALUES(?,?,?,?)`, []any{"cache-hash", "jisho.org", []byte(`{"word":"猫"}`), int64(123)}},
		{`INSERT INTO dictionary_entries VALUES(?,?,?,?,?)`, []any{"jisho.org", "猫", "word", []byte(`{"word":"猫"}`), int64(123)}},
	}
	for _, input := range inputs {
		if err := db.Exec(input.query, input.args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := database.EnsureUUIDSchema(db); err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &database.Account{}, &database.GamificationUser{}, &database.BillingPurchase{}, &database.BillingToken{}, &database.DeletedDemo{}, &database.DictionaryResponse{}, &database.DictionaryEntry{}); err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, table := range []string{"accounts", "gamification_users", "billing_purchases", "billing_tokens", "deleted_demo", "dictionary_responses", "dictionary_entries"} {
		var row struct{ ID string }
		if err := db.Table(table).Take(&row).Error; err != nil {
			t.Fatal(err)
		}
		if err := uuid.Validate(row.ID); err != nil {
			t.Fatalf("%s has non-UUID ID", table)
		}
		ids[table] = row.ID
	}
	wantUser := database.UserUUID(oldUser)
	var a database.Account
	db.First(&a)
	if a.ID != wantUser {
		t.Fatal("account ID does not match migrated user relations")
	}
	var state map[string]any
	json.Unmarshal(a.State, &state)
	if state["ID"] != wantUser {
		t.Fatal("account payload kept old user ID")
	}
	var g database.GamificationUser
	db.First(&g)
	var b database.BillingToken
	db.First(&b)
	if g.UserID != wantUser || b.UserID != wantUser || b.Token != token {
		t.Fatal("progress/billing ownership changed")
	}
	var p database.BillingPurchase
	db.First(&p)
	json.Unmarshal(p.State, &state)
	if p.UserID == nil || *p.UserID != wantUser || p.PurchaseKey != "android:external-store-token" || state["userId"] != wantUser || state["proof"] != "private-proof" {
		t.Fatal("purchase identity/proof changed")
	}
	var d database.DeletedDemo
	db.First(&d)
	if d.UserID != database.DemoUserID {
		t.Fatal("deleted demo account mapping changed")
	}
	var entry database.DictionaryEntry
	db.First(&entry)
	if entry.EntryID != "猫" || string(entry.Body) != `{"word":"猫"}` {
		t.Fatal("source archive changed")
	}
	if err := database.EnsureUUIDSchema(db); err != nil {
		t.Fatal(err)
	}
	for table, id := range ids {
		var row struct{ ID string }
		db.Table(table).Take(&row)
		if row.ID != id {
			t.Fatal("repeated migration changed row ID")
		}
	}
}
