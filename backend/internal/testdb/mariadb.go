// Package testdb provisions isolated MariaDB schemas for integration tests.
package testdb

import (
	"database/sql"
	"nihongo-master-backend/internal/config"
	"os"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

func DSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("MARIADB_TEST_DSN")
	if dsn == "" {
		if err := config.LoadEnvFile(".env.test", "../../.env.test"); err != nil {
			t.Fatal(err)
		}
		dsn = os.Getenv("MARIADB_TEST_DSN")
	}
	if dsn == "" {
		t.Skip("set MARIADB_TEST_DSN to run MariaDB integration tests")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid MARIADB_TEST_DSN")
	}
	cfg.DBName = ""
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	name := "nihongo_test_" + uuid.NewString()[:8]
	if _, err := db.Exec("CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec("DROP DATABASE `" + name + "`"); err != nil {
			t.Error(err)
		}
		db.Close()
	})
	cfg.DBName = name
	cfg.Params = map[string]string{"charset": "utf8mb4"}
	cfg.Collation = "utf8mb4_bin"
	return cfg.FormatDSN()
}
