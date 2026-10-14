// import-legacy transfers a read-only legacy export to an empty MariaDB database.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"nihongo-master-backend/internal/config"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	input := flag.String("input", "", "JSON export created by scripts/export_legacy_sqlite.py")
	flag.Parse()
	if *input == "" {
		return fmt.Errorf("--input is required")
	}
	if err := config.LoadEnvFile(".env", "backend/.env"); err != nil {
		return err
	}
	dsn, err := config.MariaDBDSN()
	if err != nil {
		return err
	}
	return importLegacy(dsn, *input)
}

func importLegacy(dsn, input string) error {
	raw, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	var data struct {
		Version int                                     `json:"version"`
		Tables  map[string][]map[string]json.RawMessage `json:"tables"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	if data.Version != 1 {
		return fmt.Errorf("unsupported export version")
	}
	s := store.NewMemoryStore()
	if err := s.EnableGamificationPersistence(dsn); err != nil {
		return err
	}
	defer s.Close()
	if err := s.EnableBillingPersistence(); err != nil {
		return err
	}
	if _, err := s.DictionaryCache(); err != nil {
		return err
	}
	db, err := database.Open(dsn)
	if err != nil {
		return err
	}
	pool, err := db.DB()
	if err != nil {
		return err
	}
	defer pool.Close()
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()
	tables := []struct {
		name    string
		columns []string
	}{
		{"gamification_users", []string{"user_id", "state"}},
		{"accounts", []string{"email", "state"}},
		{"billing_purchases", []string{"id", "state"}},
		{"billing_tokens", []string{"user_id", "token"}},
		{"deleted_demo", []string{"id"}},
		{"dictionary_responses", []string{"cache_key", "source", "body", "fetched_at"}},
		{"dictionary_entries", []string{"source", "entry_id", "kind", "body", "fetched_at"}},
	}
	for _, table := range tables {
		if _, ok := data.Tables[table.name]; !ok {
			return fmt.Errorf("missing table %s in export", table.name)
		}
		var count int64
		if err := tx.Table(table.name).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("target table %s is not empty; import stopped without overwriting data", table.name)
		}
	}
	for _, table := range tables {
		for _, row := range data.Tables[table.name] {
			values := make(map[string]any, len(table.columns))
			for _, column := range table.columns {
				value, ok := row[column]
				if !ok {
					return fmt.Errorf("missing %s.%s in export", table.name, column)
				}
				switch column {
				case "body", "state":
					var b []byte
					if err := json.Unmarshal(value, &b); err != nil {
						return err
					}
					if !json.Valid(b) {
						return fmt.Errorf("invalid JSON in %s.%s", table.name, column)
					}
					values[column] = b
				case "fetched_at":
					var n int64
					if err := json.Unmarshal(value, &n); err != nil {
						return err
					}
					values[column] = n
				default:
					var text string
					if err := json.Unmarshal(value, &text); err != nil {
						return err
					}
					values[column] = text
				}
			}
			if err := database.PrepareUUIDRow(table.name, values); err != nil {
				return err
			}
			if err := tx.Table(table.name).Create(values).Error; err != nil {
				return fmt.Errorf("import %s: %w", table.name, err)
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	for _, table := range tables {
		fmt.Printf("%s: %d records imported\n", table.name, len(data.Tables[table.name]))
	}
	return nil
}
