package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// MariaDBDSN builds the driver configuration without logging credentials.
func MariaDBDSN() (string, error) {
	value := func(key, fallback string) string {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
		return fallback
	}
	port := value("DB_PORT", "3306")
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("DB_PORT must be between 1 and 65535")
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("DB_PASSWORD is required; configure MariaDB in backend/.env")
	}
	cfg := mysql.NewConfig()
	cfg.User = value("DB_USER", "nihongo_master")
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(value("DB_HOST", "127.0.0.1"), port)
	cfg.DBName = value("DB_NAME", "nihongo_master")
	cfg.ParseTime = true
	cfg.Collation = "utf8mb4_bin"
	cfg.Params = map[string]string{"charset": "utf8mb4"}
	cfg.Timeout = 5 * time.Second
	cfg.ReadTimeout = 30 * time.Second
	cfg.WriteTimeout = 30 * time.Second
	cfg.TLSConfig = value("DB_TLS", "false")
	if cfg.TLSConfig != "false" && cfg.TLSConfig != "true" {
		return "", fmt.Errorf("DB_TLS must be true or false")
	}
	return cfg.FormatDSN(), nil
}
