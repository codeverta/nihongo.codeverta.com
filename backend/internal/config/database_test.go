package config

import (
	"github.com/go-sql-driver/mysql"
	"strings"
	"testing"
)

func TestMariaDBConfigDefaultsAndPasswordEscaping(t *testing.T) {
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_TLS"} {
		t.Setenv(key, "")
	}
	t.Setenv("DB_PASSWORD", "colon:@slash/?#日本語")
	dsn, err := MariaDBDSN()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "127.0.0.1:3306" || cfg.DBName != "nihongo_master" || cfg.User != "nihongo_master" || cfg.Passwd != "colon:@slash/?#日本語" || !strings.Contains(dsn, "charset=utf8mb4") {
		t.Fatal("invalid database configuration")
	}
}

func TestMariaDBConfigRejectsMissingPasswordAndInvalidPort(t *testing.T) {
	t.Setenv("DB_PASSWORD", "")
	if _, err := MariaDBDSN(); err == nil {
		t.Fatal("missing password accepted")
	}
	t.Setenv("DB_PASSWORD", "secret")
	for _, port := range []string{"0", "65536", "wrong"} {
		t.Setenv("DB_PORT", port)
		if _, err := MariaDBDSN(); err == nil {
			t.Fatal("invalid port accepted")
		}
	}
}
