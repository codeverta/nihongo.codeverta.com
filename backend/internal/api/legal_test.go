package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicLegalPages(t *testing.T) {
	server := newServer()
	for _, path := range []string{"/", "/legal", "/privacy-policy", "/terms-of-service", "/refund-policy", "/delete-account", "/support", "/data-safety"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			server.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
			if response.Code != 200 || response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
				t.Fatalf("public HTML expected: %d %s", response.Code, response.Header())
			}
			for _, required := range []string{"Nihongo Master", "contact@codeverta.com", "<nav", "<main", "14 Oktober 2026"} {
				if !strings.Contains(response.Body.String(), required) {
					t.Errorf("missing %q", required)
				}
			}
			if !strings.Contains(response.Header().Get("Cache-Control"), "no-transform") {
				t.Fatal("public contact must not be obfuscated by the CDN")
			}
			if response.Header().Get("Set-Cookie") != "" || response.Header().Get("Content-Security-Policy") == "" {
				t.Fatal("unexpected cookie or missing CSP")
			}
		})
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest("GET", "/delete-account", nil))
	if !strings.Contains(response.Body.String(), "mailto:contact@codeverta.com?subject=") {
		t.Fatal("deletion request pathway missing")
	}
	for _, path := range []string{"/privacy-policy/unknown", "/support/unknown"} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("unexpected fallback for %s", path)
		}
	}
	response = httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest("POST", "/delete-account", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatal("public page must not accept account deletion writes")
	}
}

func TestPublicLegalAliases(t *testing.T) {
	server := newServer()
	for path, destination := range map[string]string{"/privacy": "/privacy-policy", "/terms": "/terms-of-service", "/refunds": "/refund-policy", "/account-deletion": "/delete-account"} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 308 || response.Header().Get("Location") != destination {
			t.Fatalf("invalid redirect for %s", path)
		}
	}
}
