package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// LoadEnvFile loads the first available env file without overriding variables
// that are already present in the process environment.
func LoadEnvFile(paths ...string) error {
	for _, path := range paths {
		file, err := os.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("open env file %q: %w", path, err)
		}
		err = load(file)
		closeErr := file.Close()
		if err != nil {
			return fmt.Errorf("parse env file %q: %w", path, err)
		}
		if closeErr != nil {
			return fmt.Errorf("close env file %q: %w", path, closeErr)
		}
		return nil
	}
	return nil
}

func load(file *os.File) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			return fmt.Errorf("invalid line for key %q", key)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %q: %w", key, err)
		}
	}
	return scanner.Err()
}
