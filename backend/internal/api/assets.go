package api

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ConfigureAssets serves learning images independently of the mobile bundle.
func (s *Server) ConfigureAssets(directory string) error {
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("asset path is not a directory: %s", absolute)
	}
	s.assetDirectory = absolute
	return nil
}

func (s *Server) registerAssets(mux *http.ServeMux) {
	var files http.Handler
	if s.assetDirectory != "" {
		files = http.StripPrefix("/assets/", http.FileServer(http.Dir(s.assetDirectory)))
	}
	mux.Handle("GET /assets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/assets/")
		parts := strings.Split(key, "/")
		// Only public learning PNGs; never expose directory listings or config files.
		isOnboarding := key == "images/onboarding-japanese-journey.png"
		isLearning := len(parts) == 3 && parts[0] == "images" &&
			(parts[1] == "reading" || parts[1] == "vocabulary") &&
			path.Clean(key) == key && !strings.Contains(key, "\\") &&
			strings.HasSuffix(parts[2], ".png")
		if !isOnboarding && !isLearning {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Del("Content-Type")
		if files == nil {
			remote := s.store.AssetURL("assets/" + key)
			parsed, err := url.Parse(remote)
			if err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http") {
				http.Redirect(w, r, remote, http.StatusFound)
				return
			}
			http.NotFound(w, r)
			return
		}
		files.ServeHTTP(w, r)
	}))
}
