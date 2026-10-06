package api

import "net/http"

func (s *Server) readingCourse(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level == "" {
		level = "N5"
	}
	if level != "N5" && level != "N4" && level != "N3" && level != "N2" {
		writeError(w, http.StatusBadRequest, "invalid_level", "Reading tersedia untuk N5, N4, N3, dan N2.")
		return
	}
	writeData(w, http.StatusOK, s.store.ReadingCourse(userID(r), level))
}
func (s *Server) readingLesson(w http.ResponseWriter, r *http.Request) {
	lesson, err := s.store.ReadingLesson(userID(r), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, lesson)
}
func (s *Server) startReadingAttempt(w http.ResponseWriter, r *http.Request) {
	attempt, err := s.store.StartReadingAttempt(userID(r), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusCreated, attempt)
}
func (s *Server) answerReadingQuestion(w http.ResponseWriter, r *http.Request) {
	var input struct {
		QuestionID string `json:"questionId"`
		OptionID   string `json:"optionId"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	feedback, err := s.store.AnswerReadingQuestion(userID(r), r.PathValue("attemptID"), input.QuestionID, input.OptionID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, feedback)
}
func (s *Server) completeReadingAttempt(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.CompleteReadingAttempt(userID(r), r.PathValue("attemptID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}
