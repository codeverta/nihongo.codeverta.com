package api

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
)

//go:embed legal/*.html
var legalFiles embed.FS

func (s *Server) registerLegal(mux *http.ServeMux) {
	pages := []struct{ path, name, title string }{
		{"/legal", "index", "Informasi dan kebijakan"},
		{"/privacy-policy", "privacy", "Kebijakan Privasi / Privacy Policy"},
		{"/delete-account", "deletion", "Hapus akun dan data"},
		{"/refund-policy", "refund", "Kebijakan pembelian dan refund"},
		{"/terms-of-service", "terms", "Ketentuan Penggunaan / Terms of Service"},
		{"/support", "support", "Bantuan dan kontak"},
		{"/data-safety", "safety", "Ringkasan keamanan data"},
	}
	for _, page := range pages {
		view := template.Must(template.ParseFS(legalFiles, "legal/layout.html", "legal/"+page.name+".html"))
		var rendered bytes.Buffer
		if err := view.ExecuteTemplate(&rendered, "layout.html", struct{ Title, Email string }{page.title, "contact@codeverta.com"}); err != nil {
			panic(err)
		}
		body := append([]byte(nil), rendered.Bytes()...)
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")
			w.Header().Set("Cache-Control", "public, max-age=300, no-transform")
			w.Write(body)
		}
		mux.HandleFunc("GET "+page.path, handler)
	}
	for alias, target := range map[string]string{"/privacy": "/privacy-policy", "/terms": "/terms-of-service", "/refunds": "/refund-policy", "/account-deletion": "/delete-account"} {
		destination := target
		mux.HandleFunc("GET "+alias, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, destination, http.StatusPermanentRedirect)
		})
	}
}
