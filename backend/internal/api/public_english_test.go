package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicEnglishPagesRequireReleaseGate(t *testing.T) {
	t.Setenv("PUBLIC_ENGLISH_CONTENT_ENABLED", "false")
	server := newServer()
	for _, path := range []string{"/en/", "/en/privacy-policy"} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("English page %s status = %d, want 404", path, response.Code)
		}
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), `href="/en/"`) {
		t.Fatalf("Indonesian homepage exposed a disabled English link: %d", response.Code)
	}

	t.Setenv("PUBLIC_ENGLISH_CONTENT_ENABLED", "true")
	server = newServer()
	response = httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/en/", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `<html lang="en">`) {
		t.Fatalf("English homepage was not served when enabled: %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/en/privacy-policy", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `<html lang="en">`) {
		t.Fatalf("English privacy page was not served when enabled: %d %s", response.Code, response.Body.String())
	}
}
