package api

import "net/http"

func (s *Server) diagnosticOverview(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.DiagnosticOverview(userID(r)))
}
func (s *Server) diagnosticSession(w http.ResponseWriter, r *http.Request) {
	a, err := s.store.DiagnosticSession(userID(r), r.PathValue("attemptID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, a)
}
func (s *Server) diagnosticAnswer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		QuestionID string `json:"questionId"`
		OptionID   string `json:"optionId"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := s.store.SaveDiagnosticAnswer(userID(r), r.PathValue("attemptID"), input.QuestionID, input.OptionID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"saved": true})
}
func (s *Server) diagnosticComplete(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.CompleteDiagnostic(userID(r), r.PathValue("attemptID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}
