package api

import (
	"net/http"
	"time"
)

func (s *Server) learningHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, http.StatusOK, map[string]any{"items": s.store.LearningHistory(userID(r))})
}
func (s *Server) saveLearningActivity(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Position   *int      `json:"position"`
		AccessedAt time.Time `json:"accessedAt"`
	}
	if err := decodeJSON(r, &input); err != nil || input.Position == nil {
		writeError(w, http.StatusBadRequest, "invalid_activity", "Posisi dan waktu akses wajib diisi.")
		return
	}
	item, err := s.store.SaveLearningActivity(userID(r), r.PathValue("kind"), r.PathValue("lessonID"), *input.Position, input.AccessedAt)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, http.StatusOK, item)
}
