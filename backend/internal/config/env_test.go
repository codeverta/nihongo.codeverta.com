package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	const loadedKey = "NIHONGO_TEST_ENV_LOADED"
	const existingKey = "NIHONGO_TEST_ENV_EXISTING"
	previous, hadPrevious := os.LookupEnv(loadedKey)
	if err := os.Unsetenv(loadedKey); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if hadPrevious {
			_ = os.Setenv(loadedKey, previous)
		} else {
			_ = os.Unsetenv(loadedKey)
		}
	})
	t.Setenv(existingKey, "from-process")

	path := filepath.Join(t.TempDir(), ".env")
	content := loadedKey + "='from-file'\n" + existingKey + "=from-file\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := LoadEnvFile(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(loadedKey); got != "from-file" {
		t.Fatalf("expected loaded value, got %q", got)
	}
	if got := os.Getenv(existingKey); got != "from-process" {
		t.Fatalf("process environment should win, got %q", got)
	}
}
