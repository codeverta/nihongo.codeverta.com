package api

import (
	"crypto/subtle"
	_ "embed"
	"errors"
	"net/http"
	"strconv"

	"nihongo-master-backend/internal/manga"
)

//go:embed manga_admin.html
var mangaAdminHTML string

func (s *Server) ConfigureManga(repo *manga.Repository, token string) {
	s.manga = repo
	s.mangaAdminToken = token
}

// NewMangaHandler also supports an isolated domain service. It exposes no
// account/billing routes and does not acquire the account store's writer lock.
func NewMangaHandler(repo *manga.Repository, token string, access MangaAccessChecker) http.Handler {
	s := &Server{manga: repo, mangaAdminToken: token, mangaAccess: access}
	mux := http.NewServeMux()
	s.registerManga(mux)
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/admin/manga", http.StatusSeeOther) })
	return s.middleware(mux)
}

func (s *Server) registerManga(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/manga", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(mangaAdminHTML))
	})
	mux.HandleFunc("GET /v1/manga", s.requireMangaLifetime(s.mangaList))
	mux.HandleFunc("GET /v1/manga/{id}", s.requireMangaLifetime(s.mangaDetail))
	mux.HandleFunc("GET /v1/manga/chapters/{id}", s.requireMangaLifetime(s.mangaChapter))
	mux.HandleFunc("POST /v1/manga/chapters/{id}/prepare", s.requireMangaLifetime(s.prepareMangaChapter))
	mux.HandleFunc("GET /v1/admin/manga", s.mangaAdmin(func(w http.ResponseWriter, r *http.Request) {
		items, err := s.manga.AdminList(r.Context())
		if err != nil {
			mangaError(w, err)
			return
		}
		writeData(w, 200, map[string]any{"items": items, "cacheEnabled": s.manga.CacheEnabled()})
	}))
	mux.HandleFunc("POST /v1/admin/manga/series", s.mangaAdmin(func(w http.ResponseWriter, r *http.Request) {
		var in manga.Submission
		if decodeJSON(r, &in) != nil {
			mangaError(w, manga.ErrInvalid)
			return
		}
		if err := s.manga.Submit(r.Context(), in); err != nil {
			mangaError(w, err)
			return
		}
		writeData(w, 200, map[string]bool{"saved": true})
	}))
}
func mangaError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, manga.ErrInvalid):
		writeError(w, 400, "invalid_manga", "Permintaan manga tidak valid.")
	case errors.Is(err, manga.ErrNotFound):
		writeError(w, 404, "manga_not_found", "Manga atau chapter tidak tersedia.")
	case errors.Is(err, manga.ErrBusy):
		w.Header().Set("Retry-After", "300")
		writeError(w, 429, "manga_busy", "Antrean manga penuh. Coba lagi beberapa saat.")
	default:
		writeError(w, 503, "manga_unavailable", "Manga belum dapat dimuat. Konten tersimpan tetap tersedia.")
	}
}
func (s *Server) mangaAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if s.mangaAdminToken == "" || subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(s.mangaAdminToken)) != 1 {
			writeError(w, 401, "unauthorized", "Kunci admin salah atau belum dikonfigurasi.")
			return
		}
		if s.manga == nil {
			mangaError(w, errors.New("manga unavailable"))
			return
		}
		next(w, r)
	}
}
func (s *Server) mangaList(w http.ResponseWriter, r *http.Request) {
	page, size := 1, 20
	var err error
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil {
			mangaError(w, manga.ErrInvalid)
			return
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		size, err = strconv.Atoi(raw)
		if err != nil {
			mangaError(w, manga.ErrInvalid)
			return
		}
	}
	if page < 1 || page > 10000 || size < 1 || size > 40 {
		mangaError(w, manga.ErrInvalid)
		return
	}
	if s.manga == nil {
		writeData(w, 200, manga.Catalog{Items: []manga.Series{}, Page: page, PageSize: size, SourceName: "Rawkuma", SourceURL: "https://rawkuma.net/"})
		return
	}
	data, err := s.manga.List(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("genre"), page, size)
	if err != nil {
		mangaError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeData(w, 200, data)
}
func (s *Server) mangaDetail(w http.ResponseWriter, r *http.Request) {
	if s.manga == nil {
		mangaError(w, manga.ErrNotFound)
		return
	}
	data, err := s.manga.Detail(r.Context(), r.PathValue("id"))
	if err != nil {
		mangaError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeData(w, 200, data)
}
func (s *Server) mangaChapter(w http.ResponseWriter, r *http.Request) {
	if s.manga == nil {
		mangaError(w, manga.ErrNotFound)
		return
	}
	data, err := s.manga.Chapter(r.Context(), r.PathValue("id"))
	if err != nil {
		mangaError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, 200, data)
}
func (s *Server) prepareMangaChapter(w http.ResponseWriter, r *http.Request) {
	if s.manga == nil {
		mangaError(w, manga.ErrNotFound)
		return
	}
	data, err := s.manga.Prepare(r.Context(), r.PathValue("id"))
	if err != nil {
		mangaError(w, err)
		return
	}
	code := 200
	if data.Status == "queued" || data.Status == "processing" {
		code = 202
		w.Header().Set("Retry-After", "15")
	}
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, code, data)
}
