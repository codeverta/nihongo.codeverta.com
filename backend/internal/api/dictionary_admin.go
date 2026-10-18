package api

import (
	"crypto/subtle"
	_ "embed"
	"errors"
	"net/http"
	"nihongo-master-backend/internal/dictionary"
	"strconv"
	"unicode/utf8"
)

//go:embed dictionary_admin.html
var dictionaryAdminHTML string

func (s *Server) ConfigureDictionaryAdmin(token string) { s.dictionaryAdminToken = token }

func (s *Server) registerDictionaryAdmin(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/dictionary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(dictionaryAdminHTML))
	})
	mux.HandleFunc("GET /v1/admin/dictionary", s.dictionaryAdmin(s.listDictionaryCache))
	mux.HandleFunc("POST /v1/admin/dictionary/{id}/refresh", s.dictionaryAdmin(s.refreshDictionaryCache))
}

func (s *Server) dictionaryAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if s.dictionaryAdminToken == "" || subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(s.dictionaryAdminToken)) != 1 {
			writeError(w, 401, "unauthorized", "Kunci admin salah atau belum dikonfigurasi.")
			return
		}
		if s.dictionary.Persistence == nil {
			writeError(w, 503, "database_unavailable", "Database kamus belum tersedia.")
			return
		}
		next(w, r)
	}
}

func (s *Server) listDictionaryCache(w http.ResponseWriter, r *http.Request) {
	page := 1
	if value := r.URL.Query().Get("page"); value != "" {
		var err error
		page, err = strconv.Atoi(value)
		if err != nil || page < 1 || page > 100000 {
			writeError(w, 400, "invalid_page", "Halaman tidak valid.")
			return
		}
	}
	query := r.URL.Query().Get("q")
	if utf8.RuneCountInString(query) > 100 {
		writeError(w, 400, "invalid_query", "Pencarian maksimal 100 karakter.")
		return
	}
	records, total, err := s.dictionary.Persistence.List(r.Context(), page, query)
	if err != nil {
		writeError(w, 503, "database_unavailable", "Daftar cache belum dapat dibaca.")
		return
	}
	writeData(w, 200, map[string]any{"items": records, "total": total, "page": page, "pageSize": 50, "stats": s.dictionary.Stats()})
}

func (s *Server) refreshDictionaryCache(w http.ResponseWriter, r *http.Request) {
	err := s.dictionary.Refresh(r.Context(), r.PathValue("id"))
	if errors.Is(err, dictionary.ErrNotFound) {
		writeError(w, 404, "cache_not_found", "Pencarian tidak ditemukan.")
		return
	}
	if errors.Is(err, dictionary.ErrInvalid) {
		writeError(w, 400, "legacy_cache", "Cache lama belum memiliki metadata refresh. Cari ulang kata melalui aplikasi terlebih dahulu.")
		return
	}
	if err != nil {
		writeError(w, 503, "refresh_failed", "Refresh gagal. Hasil sebelumnya tetap disimpan.")
		return
	}
	writeData(w, 200, map[string]bool{"refreshed": true})
}
