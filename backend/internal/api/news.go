package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"nihongo-master-backend/internal/imageopt"
	"nihongo-master-backend/internal/news"
)

//go:embed news_admin.html
var newsAdminHTML string
var newsImageName = regexp.MustCompile(`^[a-f0-9]{64}\.(png|jpg)$`)

func (s *Server) ConfigureNews(repo *news.Repository, fetcher news.Fetcher, token, directory string, ctx context.Context) error {
	abs, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(abs, 0755); err != nil {
		return err
	}
	s.news = repo
	s.newsFetcher = fetcher
	s.newsAdminToken = token
	s.newsImageDirectory = abs
	s.newsContext = ctx
	return nil
}

// ConfigureNewsImageStore sends admin uploads to object storage. Without it,
// uploads stay on the local news image directory.
func (s *Server) ConfigureNewsImageStore(store news.ImageStore) { s.newsImageStore = store }
func (s *Server) registerNews(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/news", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(newsAdminHTML))
	})
	mux.HandleFunc("GET /v1/news", func(w http.ResponseWriter, r *http.Request) { s.listNews(w, r, false) })
	mux.HandleFunc("GET /v1/news/{id}", func(w http.ResponseWriter, r *http.Request) { s.newsDetail(w, r, false) })
	mux.HandleFunc("GET /news-images/{name}", s.newsImage)
	mux.HandleFunc("GET /v1/admin/news", s.newsAdmin(func(w http.ResponseWriter, r *http.Request) { s.listNews(w, r, true) }))
	mux.HandleFunc("GET /v1/admin/news/{id}", s.newsAdmin(func(w http.ResponseWriter, r *http.Request) { s.newsDetail(w, r, true) }))
	mux.HandleFunc("PUT /v1/admin/news/{id}", s.newsAdmin(s.editNews))
	mux.HandleFunc("DELETE /v1/admin/news/{id}", s.newsAdmin(func(w http.ResponseWriter, r *http.Request) {
		if err := s.news.Delete(r.Context(), r.PathValue("id")); err != nil {
			newsError(w, err)
			return
		}
		w.WriteHeader(204)
	}))
	mux.HandleFunc("PUT /v1/admin/news/sources/{id}", s.newsAdmin(s.newsSchedule))
	mux.HandleFunc("POST /v1/admin/news/sources/{id}/fetch", s.newsAdmin(s.fetchNews))
	mux.HandleFunc("POST /v1/admin/news/images", s.newsAdmin(s.uploadNewsImage))
}
func (s *Server) newsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !adminRequestAuthorized(r, s.newsAdminToken) {
			writeError(w, 401, "unauthorized", "Kunci admin salah atau belum dikonfigurasi.")
			return
		}
		if s.news == nil {
			writeError(w, 503, "news_unavailable", "Database News belum tersedia.")
			return
		}
		next(w, r)
	}
}
func newsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, news.ErrInvalid):
		writeError(w, 400, "invalid_news", "Pengaturan atau artikel tidak valid.")
	case errors.Is(err, news.ErrNotFound):
		writeError(w, 404, "news_not_found", "Artikel atau sumber tidak ditemukan.")
	case errors.Is(err, news.ErrBusy):
		writeError(w, 409, "news_busy", "Sumber sedang diproses atau baru diperiksa. Coba lagi sebentar.")
	case errors.Is(err, news.ErrConflict):
		writeError(w, 409, "news_changed", "Artikel berubah sejak editor dibuka. Muat ulang artikel sebelum memperbaiki furigana.")
	default:
		writeError(w, 503, "news_unavailable", "News belum dapat diproses. Artikel tersimpan tetap tersedia.")
	}
}
func (s *Server) listNews(w http.ResponseWriter, r *http.Request, admin bool) {
	page, size := 1, 20
	var err error
	if p := r.URL.Query().Get("page"); p != "" {
		page, err = strconv.Atoi(p)
		if err != nil {
			newsError(w, news.ErrInvalid)
			return
		}
	}
	if p := r.URL.Query().Get("limit"); p != "" {
		size, err = strconv.Atoi(p)
		if err != nil {
			newsError(w, news.ErrInvalid)
			return
		}
	}
	source := r.URL.Query().Get("source")
	if (source != "" && source != "global-voices" && source != "osaka") || page < 1 || page > 100000 || size < 1 || size > 50 {
		newsError(w, news.ErrInvalid)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if s.news == nil {
		writeData(w, 200, news.Page{Items: []news.Article{}, Sources: []news.Source{}, Page: page, PageSize: size})
		return
	}
	data, err := s.news.List(r.Context(), page, size, source, admin)
	if err != nil {
		newsError(w, err)
		return
	}
	writeData(w, 200, data)
}
func (s *Server) newsDetail(w http.ResponseWriter, r *http.Request, admin bool) {
	if s.news == nil {
		newsError(w, news.ErrNotFound)
		return
	}
	a, err := s.news.Detail(r.Context(), r.PathValue("id"), admin)
	if err != nil {
		newsError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, 200, a)
}
func (s *Server) newsSchedule(w http.ResponseWriter, r *http.Request) {
	var input news.Schedule
	if err := decodeJSON(r, &input); err != nil {
		newsError(w, news.ErrInvalid)
		return
	}
	if err := s.news.SetSchedule(r.Context(), r.PathValue("id"), input); err != nil {
		newsError(w, err)
		return
	}
	writeData(w, 200, map[string]bool{"saved": true})
}
func (s *Server) editNews(w http.ResponseWriter, r *http.Request) {
	var input news.ArticleEdit
	if err := decodeJSON(r, &input); err != nil {
		newsError(w, news.ErrInvalid)
		return
	}
	if input.ImageURL != "" && s.newsImageStore != nil && s.newsImageStore.Owns(input.ImageURL) {
		if input.ImageCredit == "" || input.ImageLicense == "" {
			newsError(w, news.ErrInvalid)
			return
		}
	} else if input.ImageURL != "" {
		const prefix = "/news-images/"
		if len(input.ImageURL) <= len(prefix) || input.ImageURL[:len(prefix)] != prefix || !newsImageName.MatchString(input.ImageURL[len(prefix):]) || input.ImageCredit == "" || input.ImageLicense == "" {
			newsError(w, news.ErrInvalid)
			return
		}
		if _, err := os.Stat(filepath.Join(s.newsImageDirectory, input.ImageURL[len(prefix):])); err != nil {
			newsError(w, news.ErrInvalid)
			return
		}
	}
	if err := s.news.Edit(r.Context(), r.PathValue("id"), input); err != nil {
		newsError(w, err)
		return
	}
	writeData(w, 200, map[string]bool{"saved": true})
}
func (s *Server) fetchNews(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id != "global-voices" && id != "osaka" {
		newsError(w, news.ErrNotFound)
		return
	}
	if s.newsFetcher == nil {
		newsError(w, news.ErrInvalid)
		return
	}
	sources, err := s.news.Sources(r.Context())
	if err != nil {
		newsError(w, err)
		return
	}
	for _, source := range sources {
		if source.ID == id && ((source.LeaseUntil != nil && source.LeaseUntil.After(time.Now())) || (source.LastStartedAt != nil && time.Since(*source.LastStartedAt) < time.Minute)) {
			newsError(w, news.ErrBusy)
			return
		}
	}
	ctx := s.newsContext
	if ctx == nil {
		ctx = context.Background()
	}
	go func() {
		work, cancel := context.WithTimeout(ctx, 9*time.Minute)
		defer cancel()
		if _, err := s.news.Sync(work, s.newsFetcher, id, true); err != nil {
			log.Printf("manual news fetch %s: %v", id, err)
		}
	}()
	writeData(w, 202, map[string]bool{"accepted": true})
}
func (s *Server) uploadNewsImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeError(w, 400, "invalid_image", "Unggah PNG/JPEG maksimal 5 MiB.")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		newsError(w, news.ErrInvalid)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		newsError(w, err)
		return
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") || config.Width < 1 || config.Height < 1 || config.Width > 4096 || config.Height > 4096 {
		newsError(w, news.ErrInvalid)
		return
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		newsError(w, news.ErrInvalid)
		return
	}
	optimized, err := imageopt.JPEG(decoded)
	if err != nil {
		newsError(w, err)
		return
	}
	if s.newsImageStore != nil {
		stored, err := s.newsImageStore.Put(r.Context(), optimized)
		if err != nil {
			newsError(w, err)
			return
		}
		writeData(w, 201, map[string]string{"url": stored})
		return
	}
	sum := sha256.Sum256(optimized)
	name := hex.EncodeToString(sum[:]) + ".jpg"
	temp, err := os.CreateTemp(s.newsImageDirectory, ".upload-*")
	if err != nil {
		newsError(w, err)
		return
	}
	defer os.Remove(temp.Name())
	_, err = temp.Write(optimized)
	closeErr := temp.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		target := filepath.Join(s.newsImageDirectory, name)
		if _, statErr := os.Stat(target); errors.Is(statErr, os.ErrNotExist) {
			err = os.Rename(temp.Name(), target)
			if err != nil {
				if _, statErr = os.Stat(target); statErr == nil {
					err = nil
				}
			}
		} else if statErr != nil {
			err = statErr
		}
	}
	if err != nil {
		newsError(w, err)
		return
	}
	writeData(w, 201, map[string]string{"url": "/news-images/" + name})
}
func (s *Server) newsImage(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if s.newsImageDirectory == "" || !newsImageName.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(s.newsImageDirectory, name)
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	if filepath.Ext(name) == ".png" || info.Size() > 200*1024 {
		optimizedName, changed, optimizeErr := optimizeStoredImage(s.newsImageDirectory, name)
		if optimizeErr == nil && changed {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			http.Redirect(w, r, "/news-images/"+optimizedName, http.StatusMovedPermanently)
			return
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if filepath.Ext(name) == ".png" {
		w.Header().Set("Content-Type", "image/png")
	} else {
		w.Header().Set("Content-Type", "image/jpeg")
	}
	http.ServeFile(w, r, path)
}
