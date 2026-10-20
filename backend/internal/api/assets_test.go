package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"nihongo-master-backend/internal/store"
)

func TestLearningAssets(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "images", "reading")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	png := []byte("\x89PNG\r\n\x1a\nimage-content")
	if err := os.WriteFile(filepath.Join(dir, "story-01.png"), png, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "images", "onboarding-japanese-journey.png"), png, 0644); err != nil {
		t.Fatal(err)
	}
	s := NewServer(store.NewMemoryStore(), nil)
	if err := s.ConfigureAssets(root); err != nil {
		t.Fatal(err)
	}
	handler := s.Handler()
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/assets/images/reading/story-01.png", http.StatusOK},
		{"/assets/images/onboarding-japanese-journey.png", http.StatusOK},
		{"/assets/images/reading/missing.png", http.StatusNotFound},
		{"/assets/images/reading/", http.StatusNotFound},
		{"/assets/", http.StatusNotFound},
		{"/assets/.env", http.StatusNotFound},
		{"/assets/images/reading/content.json", http.StatusNotFound},
		{"/assets/images/reading/..%5c..%5c.env.png", http.StatusNotFound},
	} {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			if tc.status == http.StatusOK {
				if w.Body.String() != string(png) || w.Header().Get("Content-Type") != "image/png" || w.Header().Get("Cache-Control") == "" {
					t.Fatalf("incorrect image response: %v", w.Header())
				}
			}
		})
	}
}

func TestN5VocabularyArtworkIsServedByAssetRoute(t *testing.T) {
	root := filepath.Join("..", "..", "assets")
	server := NewServer(store.NewMemoryStore(), nil)
	if err := server.ConfigureAssets(root); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/images/vocabulary/n5-theme-01.png", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("N5 artwork status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("Content-Type") != "image/png" || response.Body.Len() < 100_000 {
		t.Fatalf("N5 artwork response did not contain a real PNG: type=%q bytes=%d", response.Header().Get("Content-Type"), response.Body.Len())
	}
}

func TestCDNAssetsWithoutLocalDirectory(t *testing.T) {
	memory := store.NewMemoryStore()
	s := NewServer(memory, nil)
	request := httptest.NewRequest(http.MethodGet, "/assets/images/onboarding-japanese-journey.png", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, request)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing local/CDN image status = %d", w.Code)
	}
	memory.SetAssetBaseURL("https://cdn.example.com")
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, request)
	if w.Code != http.StatusFound || w.Header().Get("Location") != "https://cdn.example.com/assets/images/onboarding-japanese-journey.png" {
		t.Fatalf("CDN redirect = %d %v", w.Code, w.Header())
	}
}
