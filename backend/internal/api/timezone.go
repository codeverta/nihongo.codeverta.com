package api

import "net/http"

func (s *Server) updateStudyTimezone(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Timezone string `json:"timezone"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	profile, err := s.store.UpdateStudyTimezone(userID(r), input.Timezone)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, profile)
}
