package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"nihongo-master-backend/internal/store"
)

func TestAppUpdatePolicyAdminPersistenceAndRollback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "updates.json")
	newHandler := func() http.Handler {
		s := NewServer(store.NewMemoryStore(), nil)
		if err := s.ConfigureAppUpdates(path, "test-admin"); err != nil {
			t.Fatal(err)
		}
		return s.Handler()
	}
	handler := newHandler()
	policy := defaultAppUpdatePolicy()
	policy.Android.MinimumBuild = 3
	policy.Android.LatestBuild = 3
	policy.Android.MinimumAndroidAPI = 26
	policy.Android.LatestVersion = "1.0.2"
	put := func(handler http.Handler, token string, p AppUpdatePolicy) *httptest.ResponseRecorder {
		data, _ := json.Marshal(p)
		r := httptest.NewRequest("PUT", "/v1/admin/app-update-policy", bytes.NewReader(data))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	if w := put(handler, "wrong-token", policy); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized update accepted: %d", w.Code)
	}
	invalid := policy
	invalid.Android.MinimumBuild = 4
	if w := put(handler, "test-admin", invalid); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unpublished minimum accepted: %d", w.Code)
	}
	if w := put(handler, "test-admin", policy); w.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", w.Code, w.Body.String())
	}
	// Reload proves the production setting survives an API restart.
	handler = newHandler()
	check := func(expected int64) {
		t.Helper()
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", "/v1/app-update-policy", nil))
		var result struct {
			Data AppUpdatePolicy `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || result.Data.Android.MinimumBuild != expected || result.Data.IOS.Enabled {
			t.Fatalf("wrong public policy: %d %s", w.Code, w.Body.String())
		}
	}
	check(3)
	policy.Android.MinimumBuild = 0
	if w := put(handler, "test-admin", policy); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	check(0)
}

func TestAppUpdatePolicyRejectsUnsafeAndIncompleteStoreSettings(t *testing.T) {
	for _, raw := range []string{
		"https://example.com/app",
		"http://play.google.com/store/apps/details?id=com.codeverta.nihongomaster",
		"https://play.google.com/store/apps/details?id=com.example.other",
		"https://play.google.com:443/store/apps/details?id=com.codeverta.nihongomaster",
		"https://play.google.com/store/apps/details?id=com.codeverta.nihongomaster&id=other",
	} {
		p := defaultAppUpdatePolicy()
		p.Android.StoreURL = raw
		if validateAppUpdatePolicy(p) == nil {
			t.Errorf("accepted unsafe URL %q", raw)
		}
	}
	p := defaultAppUpdatePolicy()
	p.IOS.Enabled = true
	if validateAppUpdatePolicy(p) == nil {
		t.Fatal("enabled iOS without a published listing")
	}
	p.IOS.StoreURL = "https://apps.apple.com/app/id123456789"
	p.IOS.LatestBuild = 3
	p.IOS.LatestVersion = "1.0.2"
	p.IOS.MinimumIOSVersion = "16.4.1"
	if err := validateAppUpdatePolicy(p); err != nil {
		t.Fatal(err)
	}
	p.IOS.MinimumIOSVersion = "16.x"
	if validateAppUpdatePolicy(p) == nil {
		t.Fatal("accepted invalid OS version")
	}
}
