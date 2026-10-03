package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestBrandingAuthorizationValidationAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "branding.json")
	server := &Server{}
	if err := server.ConfigureBranding(path, "test-admin"); err != nil {
		t.Fatal(err)
	}
	handler := server.Handler()
	value := Branding{AppName: "New name", Tagline: "Learn Japanese", IconURL: "https://cdn.example.com/icon.png", BackgroundColor: "#FAFAFA"}
	request := func(token string, input Branding) int {
		data, _ := json.Marshal(input)
		r := httptest.NewRequest(http.MethodPut, "/v1/admin/branding", bytes.NewReader(data))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}
	if got := request("wrong", value); got != 401 {
		t.Fatalf("unauthorized: %d", got)
	}
	invalid := value
	invalid.IconURL = "javascript:alert(1)"
	if got := request("test-admin", invalid); got != 422 {
		t.Fatalf("invalid URL: %d", got)
	}
	if got := request("test-admin", value); got != 200 {
		t.Fatalf("save: %d", got)
	}
	value.AppName = "Second update"
	if got := request("test-admin", value); got != 200 {
		t.Fatalf("replace: %d", got)
	}
	restored := &Server{}
	if err := restored.ConfigureBranding(path, "test-admin"); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	restored.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/v1/branding", nil))
	var response struct {
		Data Branding `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data != value {
		t.Fatalf("not persisted: %+v", response.Data)
	}
}

func TestBrandingAdminDisabledByDefault(t *testing.T) {
	server := &Server{}
	handler := server.Handler()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("PUT", "/v1/admin/branding", nil))
	if w.Code != 401 {
		t.Fatalf("unconfigured admin: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/v1/branding", nil))
	if w.Code != 200 {
		t.Fatalf("public default: %d", w.Code)
	}
}
