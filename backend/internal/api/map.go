package api

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"io"
	"log"
	"net/http"
	"strings"

	"nihongo-master-backend/internal/imageopt"
	"nihongo-master-backend/internal/places"
)

const maxMapImageBytes = 8 << 20

//go:embed map_admin.html
var mapAdminHTML string

// ConfigureMap enables the Japan map. local is used when photos are stored on
// disk instead of COS; it may be nil.
func (s *Server) ConfigureMap(service *places.Service, token string, local *places.LocalImages) {
	s.mapSpots, s.mapAdminToken, s.mapLocalImages = service, strings.TrimSpace(token), local
}

func (s *Server) registerMap(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/map/spots", s.mapSpotList)
	mux.HandleFunc("GET /v1/jobs/prefecture-counts", s.jobPrefectureCounts)
	mux.HandleFunc("GET /map-images/{name}", s.mapImage)
	mux.HandleFunc("GET /admin/map", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(mapAdminHTML))
	})
	mux.HandleFunc("GET /v1/admin/map/spots", s.mapAdmin(s.mapAdminList))
	mux.HandleFunc("POST /v1/admin/map/spots", s.mapAdmin(s.mapAdminImport))
	mux.HandleFunc("PUT /v1/admin/map/spots/{id}", s.mapAdmin(s.mapAdminUpdate))
	mux.HandleFunc("DELETE /v1/admin/map/spots/{id}", s.mapAdmin(s.mapAdminDelete))
	mux.HandleFunc("POST /v1/admin/map/spots/{id}/refresh", s.mapAdmin(s.mapAdminRefresh))
	mux.HandleFunc("POST /v1/admin/map/spots/{id}/image", s.mapAdmin(s.mapAdminImage))
	mux.HandleFunc("DELETE /v1/admin/map/spots/{id}/image", s.mapAdmin(s.mapAdminClearImage))
}

func (s *Server) mapSpotList(w http.ResponseWriter, r *http.Request) {
	if s.mapSpots == nil {
		writeError(w, 503, "map_unavailable", "Peta belum tersedia. Coba lagi nanti.")
		return
	}
	spots, err := s.mapSpots.Store.Published(r.Context())
	if err != nil {
		log.Printf("list map spots: %v", err)
		writeError(w, 503, "map_unavailable", "Peta belum tersedia. Coba lagi nanti.")
		return
	}
	public := make([]places.PublicSpot, 0, len(spots))
	for i := range spots {
		public = append(public, spots[i].Public())
	}
	body, err := json.Marshal(map[string]any{"data": map[string]any{"spots": public, "categories": places.Categories}})
	if err != nil {
		writeError(w, 500, "internal_error", "Gagal menyiapkan data peta.")
		return
	}
	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`
	h := w.Header()
	h.Set("ETag", etag)
	h.Set("Cache-Control", "public, max-age=300")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h.Set("Content-Type", "application/json")
	_, _ = w.Write(append(body, '\n'))
}

func (s *Server) jobPrefectureCounts(w http.ResponseWriter, r *http.Request) {
	if s.jobs == nil {
		writeError(w, 503, "jobs_unavailable", "Katalog lowongan belum terhubung. Coba lagi nanti.")
		return
	}
	counts, err := s.jobs.PrefectureCounts(r.Context())
	if err != nil {
		writeJobError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeData(w, 200, map[string]any{"counts": counts})
}

func (s *Server) mapImage(w http.ResponseWriter, r *http.Request) {
	if s.mapLocalImages == nil {
		http.NotFound(w, r)
		return
	}
	file, err := s.mapLocalImages.Open(r.PathValue("name"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = io.Copy(w, file)
}

func (s *Server) mapAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if s.mapAdminToken == "" || subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(s.mapAdminToken)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized", "kunci admin salah atau belum dikonfigurasi")
			return
		}
		if s.mapSpots == nil {
			writeError(w, 503, "map_unavailable", "fitur peta belum dikonfigurasi")
			return
		}
		next(w, r)
	}
}

type mapAdminSpot struct {
	places.Spot
	Phrases           []places.Phrase `json:"phrases"`
	MissingIndonesian bool            `json:"missingIndonesian"`
	Publishable       bool            `json:"publishable"`
}

func adminSpot(spot places.Spot) mapAdminSpot {
	return mapAdminSpot{spot, spot.Phrases(), spot.MissingIndonesian(), spot.Publishable()}
}

func (s *Server) mapAdminList(w http.ResponseWriter, r *http.Request) {
	spots, err := s.mapSpots.Store.All(r.Context())
	if err != nil {
		writeMapError(w, err)
		return
	}
	out := make([]mapAdminSpot, 0, len(spots))
	for _, spot := range spots {
		out = append(out, adminSpot(spot))
	}
	writeData(w, 200, map[string]any{"spots": out, "categories": places.Categories})
}

func (s *Server) mapAdminImport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Input    string `json:"input"`
		Category string `json:"category"`
	}
	if decodeJSON(r, &body) != nil {
		writeError(w, 400, "invalid_request", "kirim input dan kategori")
		return
	}
	spot, err := s.mapSpots.Import(r.Context(), body.Input, body.Category)
	if errors.Is(err, places.ErrExists) {
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": map[string]string{"code": "exists", "message": "spot ini sudah ada"},
			"data":  adminSpot(spot),
		})
		return
	}
	if err != nil {
		writeMapError(w, err)
		return
	}
	writeData(w, http.StatusCreated, adminSpot(spot))
}

func (s *Server) mapAdminUpdate(w http.ResponseWriter, r *http.Request) {
	var edit places.Edit
	if decodeJSON(r, &edit) != nil {
		writeError(w, 400, "invalid_request", "data spot tidak valid")
		return
	}
	spot, err := s.mapSpots.Update(r.Context(), r.PathValue("id"), edit)
	if err != nil {
		writeMapError(w, err)
		return
	}
	writeData(w, 200, adminSpot(spot))
}

func (s *Server) mapAdminDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.mapSpots.Store.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeMapError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) mapAdminRefresh(w http.ResponseWriter, r *http.Request) {
	spot, err := s.mapSpots.Refresh(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMapError(w, err)
		return
	}
	writeData(w, 200, adminSpot(spot))
}

func (s *Server) mapAdminImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMapImageBytes+(1<<20))
	if err := r.ParseMultipartForm(maxMapImageBytes); err != nil {
		writeError(w, 413, "invalid_upload", "unggah satu PNG atau JPEG, maksimal 8 MB")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		writeError(w, 422, "missing_image", "pilih gambar PNG atau JPEG")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxMapImageBytes+1))
	if err != nil || len(data) > maxMapImageBytes {
		writeError(w, 413, "image_too_large", "ukuran gambar maksimal 8 MB")
		return
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") || int64(config.Width)*int64(config.Height) > 40000000 {
		writeError(w, 422, "invalid_image", "gunakan PNG/JPEG valid")
		return
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		writeError(w, 422, "invalid_image", "berkas gambar rusak atau tidak lengkap")
		return
	}
	optimized, err := imageopt.JPEG(decoded)
	if err != nil {
		writeError(w, 422, "invalid_image", "gambar tidak dapat dikompresi")
		return
	}
	spot, err := s.mapSpots.SetImage(r.Context(), r.PathValue("id"), optimized, r.FormValue("author"), r.FormValue("license"))
	if err != nil {
		writeMapError(w, err)
		return
	}
	writeData(w, 200, adminSpot(spot))
}

func (s *Server) mapAdminClearImage(w http.ResponseWriter, r *http.Request) {
	spot, err := s.mapSpots.ClearImage(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMapError(w, err)
		return
	}
	writeData(w, 200, adminSpot(spot))
}

func writeMapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, places.ErrInvalid):
		writeError(w, 400, "invalid_spot", "data tidak valid: periksa input, kategori, koordinat (harus di Jepang), prefektur, dan frasa (teks Jepang + arti Indonesia wajib). Spot hanya bisa diterbitkan jika punya nama, koordinat, prefektur, dan kategori.")
	case errors.Is(err, places.ErrNotFound):
		writeError(w, 404, "not_found", "spot tidak ditemukan (judul Wikipedia atau ID Wikidata tidak ada)")
	case errors.Is(err, places.ErrSource):
		writeError(w, 502, "source_unavailable", "Wikimedia sedang tidak bisa diakses dari server. Coba lagi nanti.")
	default:
		log.Printf("map admin: %v", err)
		writeError(w, 500, "internal_error", "terjadi kesalahan di server")
	}
}
