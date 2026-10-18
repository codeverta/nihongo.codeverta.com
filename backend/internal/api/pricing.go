package api

import (
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const defaultLifetimePriceIDR = 99000

type PricingConfig struct {
	LifetimePriceIDR int `json:"lifetimePriceIdr"`
}

type PricingStore struct {
	mu          sync.RWMutex
	value       PricingConfig
	path, token string
}

func (s *Server) ConfigurePricing(path, token string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("pricing config path must not be empty")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	pricing := &PricingStore{
		value: PricingConfig{LifetimePriceIDR: defaultLifetimePriceIDR},
		path:  absolute,
		token: token,
	}
	data, err := os.ReadFile(absolute)
	if err == nil {
		if err := json.Unmarshal(data, &pricing.value); err != nil {
			return fmt.Errorf("invalid pricing config: %w", err)
		}
		if err := validatePricing(pricing.value); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	s.pricing = pricing
	return nil
}

func validatePricing(pricing PricingConfig) error {
	if pricing.LifetimePriceIDR < 1000 || pricing.LifetimePriceIDR > 100_000_000 {
		return fmt.Errorf("harga Lifetime harus antara Rp1.000 dan Rp100.000.000")
	}
	return nil
}

func (s *Server) lifetimePriceIDR() int {
	pricing := s.pricing
	if pricing == nil {
		return defaultLifetimePriceIDR
	}
	pricing.mu.RLock()
	defer pricing.mu.RUnlock()
	return pricing.value.LifetimePriceIDR
}

func (s *Server) registerPricing(mux *http.ServeMux) {
	if s.pricing == nil {
		s.pricing = &PricingStore{value: PricingConfig{LifetimePriceIDR: defaultLifetimePriceIDR}}
	}
	mux.HandleFunc("GET /v1/pricing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		writeData(w, http.StatusOK, PricingConfig{LifetimePriceIDR: s.lifetimePriceIDR()})
	})
	mux.HandleFunc("GET /v1/admin/pricing", s.pricingAdmin(s.getPricing))
	mux.HandleFunc("PUT /v1/admin/pricing", s.pricingAdmin(s.updatePricing))
	mux.HandleFunc("GET /admin/pricing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(pricingAdminHTML))
	})
}

func (s *Server) pricingAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		pricing := s.pricing
		if pricing.token == "" || subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(pricing.token)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized", "kunci admin salah atau belum dikonfigurasi")
			return
		}
		next(w, r)
	}
}

func (s *Server) getPricing(w http.ResponseWriter, _ *http.Request) {
	writeData(w, http.StatusOK, PricingConfig{LifetimePriceIDR: s.lifetimePriceIDR()})
}

func (s *Server) updatePricing(w http.ResponseWriter, r *http.Request) {
	var input PricingConfig
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := validatePricing(input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_price", err.Error())
		return
	}

	pricing := s.pricing
	pricing.mu.Lock()
	defer pricing.mu.Unlock()
	if pricing.path != "" {
		if err := os.MkdirAll(filepath.Dir(pricing.path), 0755); err != nil {
			writeError(w, http.StatusInternalServerError, "save_failed", "gagal menyimpan harga")
			return
		}
		file, err := os.CreateTemp(filepath.Dir(pricing.path), ".pricing-*")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "save_failed", "gagal menyimpan harga")
			return
		}
		tempPath := file.Name()
		saved := false
		defer func() {
			if !saved {
				_ = os.Remove(tempPath)
			}
		}()
		if err := file.Chmod(0600); err != nil {
			_ = file.Close()
			writeError(w, http.StatusInternalServerError, "save_failed", "gagal menyimpan harga")
			return
		}
		data, err := json.MarshalIndent(input, "", "  ")
		if err == nil {
			_, err = file.Write(data)
		}
		if err == nil {
			err = file.Sync()
		}
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err == nil {
			err = os.Rename(tempPath, pricing.path)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "save_failed", "gagal menyimpan harga")
			return
		}
		saved = true
	}
	pricing.value = input
	writeData(w, http.StatusOK, input)
}

//go:embed pricing_admin.html
var pricingAdminHTML string
