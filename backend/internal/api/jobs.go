package api

import (
	"errors"
	"net/http"
	"nihongo-master-backend/internal/jobs"
)

func (s *Server) ConfigureJobs(repository *jobs.Repository) { s.jobs = repository }
func (s *Server) registerJobs(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/jobs/meta", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		writeData(w, 200, jobs.Metadata())
	})
	mux.HandleFunc("GET /v1/jobs/sources", s.jobSources)
	mux.HandleFunc("GET /v1/jobs", s.jobList)
	mux.HandleFunc("GET /v1/jobs/{jobID}", s.jobDetail)
}
func (s *Server) jobList(w http.ResponseWriter, r *http.Request) {
	if s.jobs == nil {
		writeError(w, 503, "jobs_unavailable", "Katalog lowongan belum terhubung. Coba lagi nanti.")
		return
	}
	filter, err := jobs.ParseFilter(r.URL.Query())
	if err != nil {
		writeJobError(w, err)
		return
	}
	page, err := s.jobs.List(r.Context(), filter)
	if err != nil {
		writeJobError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeData(w, 200, page)
}
func (s *Server) jobDetail(w http.ResponseWriter, r *http.Request) {
	if s.jobs == nil {
		writeError(w, 503, "jobs_unavailable", "Katalog lowongan belum terhubung. Coba lagi nanti.")
		return
	}
	job, err := s.jobs.Get(r.Context(), r.PathValue("jobID"))
	if err != nil {
		writeJobError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeData(w, 200, job)
}
func (s *Server) jobSources(w http.ResponseWriter, r *http.Request) {
	if s.jobs == nil {
		writeError(w, 503, "jobs_unavailable", "Direktori belum terhubung. Coba lagi nanti.")
		return
	}
	sources, err := s.jobs.Sources(r.Context())
	if err != nil {
		writeJobError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeData(w, 200, map[string]any{"items": sources})
}
func writeJobError(w http.ResponseWriter, err error) {
	if errors.Is(err, jobs.ErrInvalid) {
		writeError(w, 400, "invalid_filter", "Filter lowongan tidak valid. Gunakan nominal JPY dan periode gaji yang sama.")
		return
	}
	if errors.Is(err, jobs.ErrNotFound) {
		writeError(w, 404, "job_not_found", "Lowongan tidak ditemukan.")
		return
	}
	writeError(w, 503, "jobs_unavailable", "Data lowongan belum dapat dimuat. Coba lagi.")
}
