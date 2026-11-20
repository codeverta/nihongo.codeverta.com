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
	// PremiumPerDay is the number of messages a Lifetime account may send per
	// day before it can watch an (optional) rewarded ad for more.
	PremiumPerDay int `json:"premiumPerDay"`
	// LifetimeAdBonus is how many messages one rewarded ad adds for a
	// Lifetime account (a free account gets one message per ad).
	LifetimeAdBonus int `json:"lifetimeAdBonus"`
	// LifetimeAdPerDay caps the rewarded ads a Lifetime account may watch per
	// day. Zero means Lifetime accounts stop at PremiumPerDay.
	LifetimeAdPerDay int `json:"lifetimeAdPerDay"`
	// PerMinute is the hidden burst limit that applies to every account.
	PerMinute int `json:"perMinute"`
	// AlertPerDay records an admin diagnostic once an account has sent this
	// many messages in a day. Zero disables the alert.
	AlertPerDay int `json:"alertPerDay"`
}

func DefaultConfig() Config {
	return Config{Enabled: true, FreePerDay: 3, AdPerDay: 20, PremiumPerDay: 15, LifetimeAdBonus: 5, LifetimeAdPerDay: 4, PerMinute: 6, AlertPerDay: 80}
}

func (c Config) Validate() error {
	switch {
	case c.FreePerDay < 0 || c.FreePerDay > 100:
		return errors.New("kuota gratis harus 0 sampai 100 pesan per hari")
	case c.AdPerDay < 0 || c.AdPerDay > 200:
		return errors.New("batas pesan dari iklan harus 0 sampai 200 per hari")
	case c.PremiumPerDay < 1 || c.PremiumPerDay > 1000:
		return errors.New("batas Lifetime harus 1 sampai 1000 pesan per hari")
	case c.LifetimeAdBonus < 1 || c.LifetimeAdBonus > 50:
		return errors.New("pesan per iklan Lifetime harus 1 sampai 50")
	case c.LifetimeAdPerDay < 0 || c.LifetimeAdPerDay > 20:
		return errors.New("maks iklan Lifetime harus 0 sampai 20 per hari")
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
	if migrateLifetimeLimit(data, &store.value) {
		if err := store.Set(store.value); err != nil {
			return nil, err
		}
	}
	return store, nil
}

// migrateLifetimeLimit lowers a saved Lifetime limit of 100 (the old default)
// to the new 15 once, when the file predates the Lifetime ad settings. Any
// other saved value was an admin choice and is kept. Saving the file with the
// new fields makes this run only once.
func migrateLifetimeLimit(data []byte, value *Config) bool {
	var keys map[string]json.RawMessage
	if json.Unmarshal(data, &keys) != nil {
		return false
	}
	if _, ok := keys["lifetimeAdBonus"]; ok {
		return false
	}
	if value.PremiumPerDay == 100 {
		value.PremiumPerDay = DefaultConfig().PremiumPerDay
	}
	return true
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
