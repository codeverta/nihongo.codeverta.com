package database

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type uuidTable struct {
	name   string
	keys   []string
	unique string
	user   bool
}

var uuidTables = []uuidTable{
	{"accounts", []string{"email"}, "accounts_email", false},
	{"gamification_users", []string{"user_id"}, "gamification_user_id", true},
	{"billing_purchases", []string{"purchase_key"}, "billing_purchase_key", true},
	{"billing_tokens", []string{"user_id"}, "billing_token_user_id", true},
	{"deleted_demo", []string{"user_id"}, "deleted_demo_user_id", true},
	{"dictionary_responses", []string{"cache_key"}, "dictionary_cache_key", false},
	{"dictionary_entries", []string{"source", "source_entry_id"}, "dictionary_entry_source", false},
}

func textValue(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	case nil:
		return ""
	default:
		return fmt.Sprint(value)
	}
}

// PrepareUUIDRow is shared by the MariaDB schema migration and legacy importer.
// External purchase tokens and dictionary source identifiers keep their values.
func PrepareUUIDRow(table string, row map[string]any) error {
	switch table {
	case "billing_purchases":
		if _, exists := row["purchase_key"]; !exists {
			row["purchase_key"] = row["id"]
			delete(row, "id")
		}
	case "deleted_demo":
		if _, exists := row["user_id"]; !exists {
			row["user_id"] = row["id"]
			delete(row, "id")
		}
	case "dictionary_entries":
		if old, exists := row["entry_id"]; exists {
			row["source_entry_id"] = old
			delete(row, "entry_id")
		}
	}
	if table == "accounts" || table == "billing_purchases" {
		var state map[string]json.RawMessage
		if err := json.Unmarshal([]byte(textValue(row["state"])), &state); err != nil {
			return fmt.Errorf("invalid %s state: %w", table, err)
		}
		field := "userId"
		if table == "accounts" {
			field = "ID"
		}
		var old string
		if err := json.Unmarshal(state[field], &old); err != nil {
			return fmt.Errorf("invalid %s identity: %w", table, err)
		}
		if table == "accounts" && old == "" {
			return fmt.Errorf("account ID is empty")
		}
		if old != "" {
			old = UserUUID(old)
			state[field], _ = json.Marshal(old)
		}
		if table == "accounts" {
			row["id"] = old
		} else {
			row["user_id"] = OptionalUserID(old)
		}
		raw, err := json.Marshal(state)
		if err != nil {
			return err
		}
		row["state"] = raw
	}
	if table == "gamification_users" || table == "billing_tokens" || table == "deleted_demo" {
		old := textValue(row["user_id"])
		if old == "" {
			return fmt.Errorf("%s user_id is empty", table)
		}
		row["user_id"] = UserUUID(old)
	}
	id := textValue(row["id"])
	if id == "" {
		id = uuid.NewString()
	}
	if err := uuid.Validate(id); err != nil {
		return fmt.Errorf("invalid UUID in %s: %w", table, err)
	}
	row["id"] = id
	return nil
}

func uuidPrimary(db *gorm.DB, table string) (bool, error) {
	var columns []string
	err := db.Raw(`SELECT COLUMN_NAME FROM information_schema.KEY_COLUMN_USAGE WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND CONSTRAINT_NAME='PRIMARY' ORDER BY ORDINAL_POSITION`, table).Scan(&columns).Error
	return len(columns) == 1 && columns[0] == "id", err
}

// EnsureUUIDSchema must run while the caller holds the database writer lock.
// MariaDB DDL commits implicitly: each step is restartable; data backfill is atomic.
func EnsureUUIDSchema(db *gorm.DB) error {
	existing := []uuidTable{}
	needed := false
	for _, spec := range uuidTables {
		if !db.Migrator().HasTable(spec.name) {
			continue
		}
		existing = append(existing, spec)
		primary, err := uuidPrimary(db, spec.name)
		if err != nil {
			return err
		}
		if !primary || !db.Migrator().HasColumn(spec.name, "id") {
			needed = true
		}
		if spec.name == "billing_purchases" && (!db.Migrator().HasColumn(spec.name, "purchase_key") || !db.Migrator().HasColumn(spec.name, "user_id")) {
			needed = true
		}
		if spec.name == "deleted_demo" && !db.Migrator().HasColumn(spec.name, "user_id") {
			needed = true
		}
		if spec.name == "dictionary_entries" && !db.Migrator().HasColumn(spec.name, "source_entry_id") {
			needed = true
		}
	}
	if !needed {
		return nil
	}
	for _, spec := range existing {
		var rename string
		switch spec.name {
		case "billing_purchases":
			if !db.Migrator().HasColumn(spec.name, "purchase_key") {
				rename = "CHANGE COLUMN id purchase_key VARBINARY(2048) NOT NULL"
			}
		case "deleted_demo":
			if !db.Migrator().HasColumn(spec.name, "user_id") {
				rename = "CHANGE COLUMN id user_id VARCHAR(191) NOT NULL"
			}
		case "dictionary_entries":
			if !db.Migrator().HasColumn(spec.name, "source_entry_id") {
				rename = "CHANGE COLUMN entry_id source_entry_id VARCHAR(512) NOT NULL"
			}
		}
		if rename != "" {
			if err := db.Exec("ALTER TABLE `" + spec.name + "` " + rename).Error; err != nil {
				return err
			}
		}
		if !db.Migrator().HasColumn(spec.name, "id") {
			if err := db.Exec("ALTER TABLE `" + spec.name + "` ADD COLUMN id CHAR(36) NULL").Error; err != nil {
				return err
			}
		}
		if spec.name == "billing_purchases" && !db.Migrator().HasColumn(spec.name, "user_id") {
			if err := db.Exec("ALTER TABLE billing_purchases ADD COLUMN user_id CHAR(36) NULL").Error; err != nil {
				return err
			}
		}
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, spec := range existing {
			var rows []map[string]any
			if err := tx.Table(spec.name).Find(&rows).Error; err != nil {
				return err
			}
			for _, row := range rows {
				query := tx.Table(spec.name)
				for _, key := range spec.keys {
					query = query.Where("`"+key+"` = ?", row[key])
				}
				if err := PrepareUUIDRow(spec.name, row); err != nil {
					return err
				}
				if err := query.Updates(row).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for _, spec := range existing {
		primary, err := uuidPrimary(db, spec.name)
		if err != nil {
			return err
		}
		if !primary {
			if err := db.Exec("ALTER TABLE `" + spec.name + "` DROP PRIMARY KEY, MODIFY COLUMN id CHAR(36) NOT NULL, ADD PRIMARY KEY(id)").Error; err != nil {
				return err
			}
		}
		if spec.user {
			nullable := "NOT NULL"
			if spec.name == "billing_purchases" {
				nullable = "NULL"
			}
			if err := db.Exec("ALTER TABLE `" + spec.name + "` MODIFY COLUMN user_id CHAR(36) " + nullable).Error; err != nil {
				return err
			}
		}
		if !db.Migrator().HasIndex(spec.name, spec.unique) {
			columns := "`" + strings.Join(spec.keys, "`,`") + "`"
			if err := db.Exec("ALTER TABLE `" + spec.name + "` ADD UNIQUE INDEX `" + spec.unique + "`(" + columns + ")").Error; err != nil {
				return err
			}
		}
	}
	return nil
}
