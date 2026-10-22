package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLandingPage(t *testing.T) {
	server := newServer()
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	if response.Code != 200 || response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("landing HTML expected: %d %s", response.Code, response.Header())
	}
	body := response.Body.String()
	for _, required := range []string{
		"Nihongo Master",
		"https://play.google.com/store/apps/details?id=com.codeverta.nihongomaster",
		"/privacy-policy", "/delete-account", "/support",
		"contact@codeverta.com", "<nav", "<main",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("landing page missing %q", required)
		}
	}
	csp := response.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'none'") || strings.Contains(csp, "unsafe") {
		t.Fatalf("unexpected landing CSP %q", csp)
	}
	if response.Header().Get("Set-Cookie") != "" {
		t.Fatal("landing page must not set cookies")
	}

	response = httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest("POST", "/", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST / should be rejected, got %d", response.Code)
	}
}

func TestLandingStaticFiles(t *testing.T) {
	server := newServer()
	for path, contentType := range map[string]string{
		"/site/site.css":                                   "text/css; charset=utf-8",
		"/site/site.js":                                    "text/javascript; charset=utf-8",
		"/site/img/app-vocab.webp":                         "image/webp",
		"/site/img/icon-64.png":                            "image/png",
		"/site/fonts/noto-sans-jp-latin-wght-normal.woff2": "font/woff2",
	} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 200 || response.Header().Get("Content-Type") != contentType || response.Body.Len() == 0 {
			t.Errorf("%s: %d %q", path, response.Code, response.Header().Get("Content-Type"))
		}
	}
	for _, path := range []string{"/site/", "/site/img/", "/site/index.html", "/site/fonts/OFL.txt", "/site/../landing.go", "/site/missing.css"} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code == 200 {
			t.Errorf("%s should not be served", path)
		}
	}
}
