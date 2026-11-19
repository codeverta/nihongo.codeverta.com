// Package prefinfo holds per-prefecture living information for Japan
// (minimum wage, rent, price level) shown on the app's Japan map. Values come
// from official Japanese statistics and are edited by admins when they
// change (minimum wages are revised every October).
package prefinfo

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// Meta describes the sources and the assumptions used for estimates.
type Meta struct {
	WageFiscalYear     string  `json:"wageFiscalYear"`
	WageSource         string  `json:"wageSource"`
	RentMeasure        string  `json:"rentMeasure"`
	RentSurveyYear     string  `json:"rentSurveyYear"`
	RentSource         string  `json:"rentSource"`
	PriceIndexYear     string  `json:"priceIndexYear"`
	PriceIndexSource   string  `json:"priceIndexSource"`
	IDRPerJPY          float64 `json:"idrPerJpy"`
	ExchangeRateDate   string  `json:"exchangeRateDate"`
	ExchangeRateSource string  `json:"exchangeRateSource"`
	// HoursPerDay and DaysPerMonth turn the hourly minimum wage into an
	// indicative full-time monthly wage.
	HoursPerDay  float64 `json:"hoursPerDay"`
	DaysPerMonth int     `json:"daysPerMonth"`
	UpdatedAt    string  `json:"updatedAt"`
}

// Prefecture is one prefecture's figures. Nullable values are unknown.
type Prefecture struct {
	Code              string   `json:"code"`
	NameJa            string   `json:"nameJa"`
	MinimumWage       int      `json:"minimumWage"`
	WageEffectiveDate string   `json:"wageEffectiveDate"`
	RentPer33m2       *int     `json:"rentPer33m2"`
	Rent1KEstimate    *int     `json:"rent1kEstimate"`
	PriceIndex        *float64 `json:"priceIndex"`
}

type Data struct {
	Meta        Meta         `json:"meta"`
	Prefectures []Prefecture `json:"prefectures"`
}

//go:embed default.json
var defaultJSON []byte

// Default returns the bundled figures researched from official sources.
func Default() (Data, error) {
	var d Data
	if err := json.Unmarshal(defaultJSON, &d); err != nil {
		return Data{}, fmt.Errorf("parse bundled prefecture data: %w", err)
	}
	return d, d.Validate()
}

var (
	codePattern = regexp.MustCompile(`^(0[1-9]|[1-3][0-9]|4[0-7])$`)
	datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// Validate checks that all 47 prefectures are present exactly once and that
// the numbers are plausible, so a typo in the admin cannot break the map.
func (d Data) Validate() error {
	if len(d.Prefectures) != 47 {
		return fmt.Errorf("harus ada 47 prefektur, ditemukan %d", len(d.Prefectures))
	}
	seen := map[string]bool{}
	for _, p := range d.Prefectures {
		if !codePattern.MatchString(p.Code) || seen[p.Code] {
			return fmt.Errorf("kode prefektur %q tidak valid atau ganda", p.Code)
		}
		seen[p.Code] = true
		if p.MinimumWage < 500 || p.MinimumWage > 5000 {
			return fmt.Errorf("%s: upah minimum %d tidak masuk akal (500–5000 yen)", p.Code, p.MinimumWage)
		}
		if p.WageEffectiveDate != "" && !datePattern.MatchString(p.WageEffectiveDate) {
			return fmt.Errorf("%s: tanggal berlaku harus YYYY-MM-DD", p.Code)
		}
		if p.Rent1KEstimate != nil && (*p.Rent1KEstimate < 5000 || *p.Rent1KEstimate > 500000) {
			return fmt.Errorf("%s: perkiraan sewa 1K tidak masuk akal", p.Code)
		}
		if p.RentPer33m2 != nil && (*p.RentPer33m2 < 500 || *p.RentPer33m2 > 100000) {
			return fmt.Errorf("%s: sewa per 3,3 m² tidak masuk akal", p.Code)
		}
		if p.PriceIndex != nil && (*p.PriceIndex < 50 || *p.PriceIndex > 150) {
			return fmt.Errorf("%s: indeks harga harus 50–150", p.Code)
		}
	}
	m := d.Meta
	switch {
	case m.IDRPerJPY < 0 || m.IDRPerJPY > 10000:
		return errors.New("kurs IDR per 1 JPY tidak masuk akal")
	case m.HoursPerDay < 1 || m.HoursPerDay > 12:
		return errors.New("jam kerja per hari harus 1–12")
	case m.DaysPerMonth < 1 || m.DaysPerMonth > 31:
		return errors.New("hari kerja per bulan harus 1–31")
	}
	return nil
}

// Store keeps the data in memory and persists admin edits to a JSON file,
// falling back to the bundled defaults when no file exists yet.
type Store struct {
	mu    sync.RWMutex
	value Data
	path  string
}

func Load(path string) (*Store, error) {
	value, err := Default()
	if err != nil {
		return nil, err
	}
	store := &Store{value: value, path: path}
	if path == "" {
		return store, nil
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	var saved Data
	if err := json.Unmarshal(raw, &saved); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := saved.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	store.value = saved
	return store, nil
}

func (s *Store) Get() Data {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

func (s *Store) Set(value Data, now time.Time) (Data, error) {
	if err := value.Validate(); err != nil {
		return Data{}, err
	}
	value.Meta.UpdatedAt = now.UTC().Format(time.RFC3339)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path != "" {
		if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
			return Data{}, err
		}
		raw, _ := json.MarshalIndent(value, "", "  ")
		if err := os.WriteFile(s.path+".tmp", raw, 0o600); err != nil {
			return Data{}, err
		}
		if err := os.Rename(s.path+".tmp", s.path); err != nil {
			return Data{}, err
		}
	}
	s.value = value
	return value, nil
}
