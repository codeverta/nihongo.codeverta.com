package api

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// The public marketing page for nihongo.codeverta.com. Everything it needs
// (HTML, CSS, JS, images, fonts) is embedded in the binary and served from
// the same origin, so the CSP can stay at 'self'.
//
//go:embed site/index.html site/static
var landingFiles embed.FS

const landingCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'"

var landingStaticTypes = map[string]string{
	".css":   "text/css; charset=utf-8",
	".js":    "text/javascript; charset=utf-8",
	".webp":  "image/webp",
	".png":   "image/png",
	".woff2": "font/woff2",
}

func (s *Server) registerLanding(mux *http.ServeMux) {
	page, err := landingFiles.ReadFile("site/index.html")
	if err != nil {
		panic(err)
	}
	static, err := fs.Sub(landingFiles, "site/static")
	if err != nil {
		panic(err)
	}

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", landingCSP)
		h.Set("Cache-Control", "public, max-age=300, no-transform")
		w.Write(page)
	})

	mux.HandleFunc("GET /site/{file...}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		contentType, ok := landingStaticTypes[path.Ext(name)]
		if !ok || name == "" || path.Clean(name) != name || strings.HasPrefix(name, ".") || strings.Contains(name, "\\") {
			http.NotFound(w, r)
			return
		}
		body, err := fs.ReadFile(static, name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		h := w.Header()
		h.Set("Content-Type", contentType)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "public, max-age=86400")
		if contentType == "font/woff2" {
			h.Set("Cache-Control", "public, max-age=31536000, immutable")
			h.Set("Access-Control-Allow-Origin", "*")
		}
		w.Write(body)
	})
}
