package api

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/imageopt"
)

const maxBannerImageBytes = 5 << 20

var (
	bannerIDPattern    = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	bannerImagePattern = regexp.MustCompile(`^[a-f0-9]{64}\.(png|jpg)$`)
)

var defaultBannerImages = map[string]string{
	"learn": "/assets/images/exams/n5.webp",
	"ssw":   "/assets/images/ssw/ssw-careers.webp",
	"jobs":  "/assets/images/jobs/job-category-atlas.webp",
}

type homeBannerConfig struct {
	Banners []domain.HomeBanner `json:"banners"`
}

type HomeBannerStore struct {
	mu                          sync.RWMutex
	value                       homeBannerConfig
	path, imageDirectory, token string
}

func defaultHomeBanners() homeBannerConfig {
	return homeBannerConfig{Banners: []domain.HomeBanner{
		{ID: "learn", Title: "Belajar Jepang, selangkah setiap hari", Description: "Kosakata, kanji, grammar, dan reading untuk target JLPT-mu.", ImageURL: defaultBannerImages["learn"], Action: "learn", ButtonLabel: "Mulai belajar", Order: 0, Enabled: true},
		{ID: "ssw", Title: "Siapkan langkahmu untuk Tokutei Ginou", Description: "Kenali bidang SSW, pelajari materi, dan latih kesiapan ujianmu.", ImageURL: defaultBannerImages["ssw"], Action: "ssw", ButtonLabel: "Jelajahi SSW", Order: 1, Enabled: true},
		{ID: "jobs", Title: "Temukan peluang kerja di Jepang", Description: "Jelajahi lowongan, lokasi, gaji, dan kualifikasi yang sesuai denganmu.", ImageURL: defaultBannerImages["jobs"], Action: "jobs", ButtonLabel: "Jelajahi lowongan", Order: 2, Enabled: true},
	}}
}

// ConfigureHomeBanners keeps settings and uploaded images under the durable data directory.
func (s *Server) ConfigureHomeBanners(path, token string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("home banner config path must not be empty")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	b := &HomeBannerStore{value: defaultHomeBanners(), path: absolute, imageDirectory: filepath.Join(filepath.Dir(absolute), "home-banner-images"), token: token}
	data, err := os.ReadFile(absolute)
	if err == nil {
		if err = json.Unmarshal(data, &b.value); err != nil {
			return fmt.Errorf("invalid home banner config: %w", err)
		}
		migrated := migrateDefaultBannerImages(&b.value)
		if err = b.validate(b.value); err != nil {
			return err
		}
		for i := range b.value.Banners {
			imageURL := b.value.Banners[i].ImageURL
			if !strings.HasPrefix(imageURL, "/home-banner-images/") {
				continue
			}
			name := strings.TrimPrefix(imageURL, "/home-banner-images/")
			optimizedName, changed, optimizeErr := optimizeStoredImage(b.imageDirectory, name)
			if optimizeErr != nil {
				return fmt.Errorf("failed to optimize saved home banner: %w", optimizeErr)
			}
			if changed {
				b.value.Banners[i].ImageURL = "/home-banner-images/" + optimizedName
				migrated = true
			}
		}
		if migrated {
			data, err = json.MarshalIndent(b.value, "", "  ")
			if err == nil {
				err = atomicBannerWrite(absolute, data, 0600)
			}
			if err != nil {
				return fmt.Errorf("failed to save optimized home banner config: %w", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	s.homeBanners = b
	return nil
}

func migrateDefaultBannerImages(config *homeBannerConfig) bool {
	changed := false
	for i := range config.Banners {
		for _, imageURL := range defaultBannerImages {
			if config.Banners[i].ImageURL == strings.TrimSuffix(imageURL, ".webp")+".png" {
				config.Banners[i].ImageURL = imageURL
				changed = true
				break
			}
		}
	}
	return changed
}

func (b *HomeBannerStore) publicBanners() []domain.HomeBanner {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]domain.HomeBanner, 0, len(b.value.Banners))
	for _, banner := range b.value.Banners {
		if banner.Enabled {
			result = append(result, banner)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func (b *HomeBannerStore) validate(config homeBannerConfig) error {
	if len(config.Banners) > 12 {
		return fmt.Errorf("maksimal 12 banner")
	}
	seen := map[string]bool{}
	for _, banner := range config.Banners {
		if !bannerIDPattern.MatchString(banner.ID) || seen[banner.ID] {
			return fmt.Errorf("ID banner harus unik: huruf kecil, angka, dan tanda hubung, maksimal 64 karakter")
		}
		seen[banner.ID] = true
		if strings.TrimSpace(banner.Title) == "" || len([]rune(banner.Title)) > 80 || len([]rune(banner.Description)) > 180 || strings.TrimSpace(banner.ButtonLabel) == "" || len([]rune(banner.ButtonLabel)) > 40 {
			return fmt.Errorf("judul wajib (maks 80), deskripsi maks 180, teks tombol wajib (maks 40 karakter)")
		}
		if _, ok := defaultBannerImages[banner.Action]; !ok {
			return fmt.Errorf("tujuan banner harus learn, ssw, atau jobs")
		}
		if banner.Order < 0 || banner.Order > 999 {
			return fmt.Errorf("urutan harus antara 0 dan 999")
		}
		raw := banner.ImageURL
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "/home-banner-images/") {
			name := strings.TrimPrefix(raw, "/home-banner-images/")
			if !bannerImagePattern.MatchString(name) {
				return fmt.Errorf("alamat gambar unggahan tidak valid")
			}
			info, err := os.Stat(filepath.Join(b.imageDirectory, name))
			if err != nil || !info.Mode().IsRegular() {
				return fmt.Errorf("gambar unggahan tidak ditemukan; unggah ulang sebelum menyimpan")
			}
			continue
		}
		local := false
		for _, value := range defaultBannerImages {
			if raw == value {
				local = true
				break
			}
		}
		if local {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || len(raw) > 2048 || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return fmt.Errorf("gambar harus URL HTTPS atau gambar yang diunggah dari halaman ini")
		}
	}
	return nil
}

func (s *Server) registerHomeBanners(mux *http.ServeMux) {
	if s.homeBanners == nil {
		s.homeBanners = &HomeBannerStore{value: defaultHomeBanners()}
	}
	mux.HandleFunc("GET /v1/admin/home-banners", s.homeBannerAdmin(s.getHomeBannerConfig))
	mux.HandleFunc("PUT /v1/admin/home-banners", s.homeBannerAdmin(s.updateHomeBanners))
	mux.HandleFunc("POST /v1/admin/home-banners/images", s.homeBannerAdmin(s.uploadHomeBannerImage))
	mux.HandleFunc("GET /home-banner-images/{filename}", s.serveHomeBannerImage)
	mux.HandleFunc("GET /admin/home-banners", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(homeBannersAdminHTML))
	})
}

func (s *Server) homeBannerAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		b := s.homeBanners
		if b.token == "" || subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(b.token)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized", "kunci admin salah atau belum dikonfigurasi")
			return
		}
		next(w, r)
	}
}

func (s *Server) getHomeBannerConfig(w http.ResponseWriter, r *http.Request) {
	b := s.homeBanners
	b.mu.RLock()
	defer b.mu.RUnlock()
	writeData(w, http.StatusOK, b.value)
}

func (s *Server) updateHomeBanners(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var input homeBannerConfig
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if input.Banners == nil {
		writeError(w, 422, "invalid_banners", "banners wajib berupa daftar; gunakan [] untuk mengosongkan carousel")
		return
	}
	b := s.homeBanners
	if err := b.validate(input); err != nil {
		writeError(w, 422, "invalid_banners", err.Error())
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.path == "" {
		writeError(w, 503, "not_configured", "penyimpanan banner belum dikonfigurasi")
		return
	}
	data, err := json.MarshalIndent(input, "", "  ")
	if err == nil {
		err = atomicBannerWrite(b.path, data, 0600)
	}
	if err != nil {
		writeError(w, 500, "save_failed", "gagal menyimpan banner")
		return
	}
	b.value = input
	writeData(w, http.StatusOK, b.value)
}

func atomicBannerWrite(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".banner-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(mode); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), path)
}

func (s *Server) uploadHomeBannerImage(w http.ResponseWriter, r *http.Request) {
	b := s.homeBanners
	if b.imageDirectory == "" {
		writeError(w, 503, "not_configured", "penyimpanan gambar belum dikonfigurasi")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBannerImageBytes+(1<<20))
	if err := r.ParseMultipartForm(maxBannerImageBytes); err != nil {
		writeError(w, 413, "invalid_upload", "unggah satu PNG atau JPEG, maksimal 5 MB")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	f, _, err := r.FormFile("image")
	if err != nil {
		writeError(w, 422, "missing_image", "pilih gambar PNG atau JPEG")
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxBannerImageBytes+1))
	if err != nil || len(data) > maxBannerImageBytes {
		writeError(w, 413, "image_too_large", "ukuran gambar maksimal 5 MB")
		return
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") || config.Width < 1 || config.Height < 1 || int64(config.Width)*int64(config.Height) > 12000000 {
		writeError(w, 422, "invalid_image", "gunakan PNG/JPEG valid dengan maksimal 12 juta piksel")
		return
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		writeError(w, 422, "invalid_image", "berkas gambar rusak atau tidak lengkap")
		return
	}
	optimized, err := imageopt.JPEG(decoded)
	if err != nil {
		writeError(w, 422, "invalid_image", "gambar tidak dapat dikompresi di bawah 200 KB")
		return
	}
	hash := sha256.Sum256(optimized)
	name := hex.EncodeToString(hash[:]) + ".jpg"
	if err := atomicBannerWrite(filepath.Join(b.imageDirectory, name), optimized, 0644); err != nil {
		writeError(w, 500, "upload_failed", "gagal menyimpan gambar")
		return
	}
	writeData(w, http.StatusCreated, map[string]string{"imageUrl": "/home-banner-images/" + name})
}

func (s *Server) serveHomeBannerImage(w http.ResponseWriter, r *http.Request) {
	b := s.homeBanners
	name := r.PathValue("filename")
	if b.imageDirectory == "" || !bannerImagePattern.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(filepath.Join(b.imageDirectory, name))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	if filepath.Ext(name) == ".png" || info.Size() > 200*1024 {
		optimizedName, changed, optimizeErr := optimizeStoredImage(b.imageDirectory, name)
		if optimizeErr == nil && changed {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			http.Redirect(w, r, "/home-banner-images/"+optimizedName, http.StatusMovedPermanently)
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
	http.ServeContent(w, r, name, info.ModTime(), f)
}

//go:embed home_banners_admin.html
var homeBannersAdminHTML string
