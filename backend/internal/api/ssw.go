package api

import "net/http"

func (s *Server) sswPractice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	result, err := s.store.SSWPractice(userID(r), r.PathValue("sector"), r.URL.Query().Get("module"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, 200, map[string]any{"questions": result})
}
func (s *Server) sswPracticeCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	var input struct {
		QuestionID string `json:"questionId"`
		Answer     *int   `json:"answer"`
	}
	if err := decodeJSON(r, &input); err != nil || input.Answer == nil {
		writeError(w, 400, "invalid_input", "questionId and answer are required")
		return
	}
	result, err := s.store.SSWPracticeCheck(userID(r), r.PathValue("sector"), input.QuestionID, *input.Answer)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, 200, result)
}
