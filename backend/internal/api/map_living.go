package api

import (
	"log"
	"net/http"
	"time"

	"nihongo-master-backend/internal/prefinfo"
)

// ConfigurePrefectureLiving enables the per-prefecture wage and living-cost
// data on the Japan map. A nil store serves the bundled defaults.
func (s *Server) ConfigurePrefectureLiving(store *prefinfo.Store) {
	s.prefLiving = store
}

func (s *Server) registerPrefectureLiving(mux *http.ServeMux) {
	if s.prefLiving == nil {
		store, err := prefinfo.Load("")
		if err != nil {
			log.Printf("prefecture living data unavailable: %v", err)
		}
		s.prefLiving = store
	}
	mux.HandleFunc("GET /v1/map/prefecture-living", s.prefectureLiving)
	mux.HandleFunc("GET /v1/map/prefecture-profiles", s.prefectureProfiles)
	mux.HandleFunc("GET /v1/admin/prefecture-living", s.prefectureLivingAdmin(s.prefectureLiving))
	mux.HandleFunc("PUT /v1/admin/prefecture-living", s.prefectureLivingAdmin(s.updatePrefectureLiving))
}

func (s *Server) prefectureLivingAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !adminRequestAuthorized(r, s.mapAdminToken) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "kunci admin salah atau belum dikonfigurasi")
			return
		}
		next(w, r)
	}
}

func (s *Server) prefectureLiving(w http.ResponseWriter, r *http.Request) {
	if s.prefLiving == nil {
		writeError(w, http.StatusServiceUnavailable, "prefecture_living_unavailable", "Data upah dan biaya hidup belum tersedia.")
		return
	}
	if w.Header().Get("Cache-Control") == "" {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
	writeData(w, http.StatusOK, s.prefLiving.Get())
}

func (s *Server) updatePrefectureLiving(w http.ResponseWriter, r *http.Request) {
	if s.prefLiving == nil {
		writeError(w, http.StatusServiceUnavailable, "prefecture_living_unavailable", "Data upah dan biaya hidup belum tersedia.")
		return
	}
	var input prefinfo.Data
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "format data tidak valid")
		return
	}
	saved, err := s.prefLiving.Set(input, time.Now())
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_prefecture_living", err.Error())
		return
	}
	writeData(w, http.StatusOK, saved)
}

func (s *Server) prefectureProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := prefinfo.LoadProfiles()
	if err != nil {
		log.Printf("prefecture profiles: %v", err)
		writeError(w, http.StatusServiceUnavailable, "prefecture_profiles_unavailable", "Profil prefektur belum tersedia.")
		return
	}
	h := w.Header()
	h.Set("ETag", profiles.ETag)
	h.Set("Cache-Control", "public, max-age=86400")
	if match := r.Header.Get("If-None-Match"); match != "" && match == profiles.ETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h.Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"data":`))
	_, _ = w.Write(profiles.Raw)
	_, _ = w.Write([]byte(`}`))
}
