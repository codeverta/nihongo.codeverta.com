// Package aichat implements the per-lesson AI tutor: quota accounting,
// conversation storage, prompt assembly and the DeepSeek streaming client.
package aichat

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config holds the admin-tunable limits. Changing it never requires a deploy.
type Config struct {
	// Enabled is the kill switch. When false every chat request is refused.
	Enabled bool `json:"enabled"`
	// FreePerDay is the number of messages a free account may send per day
	// before each further message needs a rewarded-ad credit.
	FreePerDay int `json:"freePerDay"`
	// AdPerDay caps how many ad credits a free account may earn per day.
	AdPerDay int `json:"adPerDay"`
	// PremiumPerDay is the daily limit for Lifetime accounts (no ads).
	PremiumPerDay int `json:"premiumPerDay"`
	// PerMinute is the hidden burst limit that applies to every account.
	PerMinute int `json:"perMinute"`
	// AlertPerDay records an admin diagnostic once an account has sent this
	// many messages in a day. Zero disables the alert.
	AlertPerDay int `json:"alertPerDay"`
}

func DefaultConfig() Config {
	return Config{Enabled: true, FreePerDay: 3, AdPerDay: 20, PremiumPerDay: 100, PerMinute: 6, AlertPerDay: 80}
}

func (c Config) Validate() error {
	switch {
	case c.FreePerDay < 0 || c.FreePerDay > 100:
		return errors.New("kuota gratis harus 0 sampai 100 pesan per hari")
	case c.AdPerDay < 0 || c.AdPerDay > 200:
		return errors.New("batas pesan dari iklan harus 0 sampai 200 per hari")
	case c.PremiumPerDay < 1 || c.PremiumPerDay > 1000:
		return errors.New("batas Lifetime harus 1 sampai 1000 pesan per hari")
	case c.PerMinute < 1 || c.PerMinute > 60:
		return errors.New("batas per menit harus 1 sampai 60")
	case c.AlertPerDay < 0 || c.AlertPerDay > 10000:
		return errors.New("ambang peringatan harus 0 sampai 10000")
	}
	return nil
}

// ConfigStore persists Config as a JSON file, following the ads.json pattern.
type ConfigStore struct {
	mu    sync.RWMutex
	value Config
	path  string
}

// LoadConfigStore reads path when it exists and falls back to defaults.
// An empty path keeps the configuration in memory only (tests).
func LoadConfigStore(path string) (*ConfigStore, error) {
	store := &ConfigStore{value: DefaultConfig(), path: path}
	if path == "" {
		return store, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &store.value); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := store.value.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return store, nil
}

func (s *ConfigStore) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

func (s *ConfigStore) Set(value Config) error {
	if err := value.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path != "" {
		if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
			return err
		}
		data, _ := json.MarshalIndent(value, "", "  ")
		if err := os.WriteFile(s.path+".tmp", data, 0o600); err != nil {
			return err
		}
		if err := os.Rename(s.path+".tmp", s.path); err != nil {
			return err
		}
	}
	s.value = value
	return nil
}
