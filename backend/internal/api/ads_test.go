package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"nihongo-master-backend/internal/store"
)

func newAdsTestServer(t *testing.T) (http.Handler, string) {
	t.Helper()
	s := NewServer(store.NewMemoryStore(), nil)
	path := filepath.Join(t.TempDir(), "ads.json")
	if err := s.ConfigureAds(path, "secret-token"); err != nil {
		t.Fatal(err)
	}
	return s.Handler(), path
}

func TestAdsConfigIsOffByDefaultAndPublic(t *testing.T) {
	handler, _ := newAdsTestServer(t)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/ads-config", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body struct{ Data AdsConfig }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Enabled {
		t.Fatal("ads must be off until the owner turns them on")
	}
}

func TestAdsConfigUpdateRequiresTokenAndValidates(t *testing.T) {
	handler, _ := newAdsTestServer(t)
	valid := `{"enabled":true,"banner":{"enabled":true,"unitId":"` + testBannerAdUnit + `"},"newsNative":{"enabled":true,"unitId":"` + testNativeAdUnit + `","every":6},"jobsNative":{"enabled":false,"unitId":"` + testNativeAdUnit + `","every":8}}`
	put := func(token, body string) int {
		req := httptest.NewRequest(http.MethodPut, "/v1/admin/ads-config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}
	if code := put("", valid); code != http.StatusUnauthorized {
		t.Fatalf("no token: %d", code)
	}
	if code := put("wrong", valid); code != http.StatusUnauthorized {
		t.Fatalf("wrong token: %d", code)
	}
	bad := strings.Replace(valid, testBannerAdUnit, "not-a-unit", 1)
	if code := put("secret-token", bad); code != http.StatusUnprocessableEntity {
		t.Fatalf("bad unit: %d", code)
	}
	tooOften := strings.Replace(valid, `"every":6`, `"every":1`, 1)
	if code := put("secret-token", tooOften); code != http.StatusUnprocessableEntity {
		t.Fatalf("bad spacing: %d", code)
	}
	if code := put("secret-token", valid); code != http.StatusOK {
		t.Fatalf("valid update: %d", code)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/ads-config", nil))
	if !strings.Contains(rec.Body.String(), `"enabled":true`) {
		t.Fatalf("update not visible: %s", rec.Body.String())
	}
}

func TestAdsConfigPersistsAcrossRestart(t *testing.T) {
	handler, path := newAdsTestServer(t)
	body := `{"enabled":true,"banner":{"enabled":true,"unitId":"` + testBannerAdUnit + `"},"newsNative":{"enabled":false,"unitId":"` + testNativeAdUnit + `","every":6},"jobsNative":{"enabled":false,"unitId":"` + testNativeAdUnit + `","every":8}}`
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/ads-config", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("save: %d %s", rec.Code, rec.Body.String())
	}
	s2 := NewServer(store.NewMemoryStore(), nil)
	if err := s2.ConfigureAds(path, "secret-token"); err != nil {
		t.Fatal(err)
	}
	if !s2.ads.value.Enabled || !s2.ads.value.Banner.Enabled {
		t.Fatal("saved config was not reloaded")
	}
}
