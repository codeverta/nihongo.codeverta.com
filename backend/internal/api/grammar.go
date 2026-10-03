package api

import (
	"net/http"
	"strings"
)

func (s *Server) grammarCourse(w http.ResponseWriter, r *http.Request) {
	level := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("level")))
	if level == "" {
		level = "N5"
	}
	if level != "N5" && level != "N4" {
		writeError(w, http.StatusUnprocessableEntity, "invalid_level", "Grammar tersedia untuk N5 dan N4.")
		return
	}
	writeData(w, http.StatusOK, s.store.GrammarCourse(userID(r), level))
}
func (s *Server) grammarLesson(w http.ResponseWriter, r *http.Request) {
	lesson, err := s.store.GrammarLesson(userID(r), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, lesson)
}
func (s *Server) startGrammarAttempt(w http.ResponseWriter, r *http.Request) {
	attempt, err := s.store.StartGrammarAttempt(userID(r), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusCreated, attempt)
}
func (s *Server) answerGrammarQuestion(w http.ResponseWriter, r *http.Request) {
	var input struct {
		QuestionID string `json:"questionId"`
		OptionID   string `json:"optionId"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	feedback, err := s.store.AnswerGrammarQuestion(userID(r), r.PathValue("attemptID"), input.QuestionID, input.OptionID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, feedback)
}
func (s *Server) completeGrammarAttempt(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.CompleteGrammarAttempt(userID(r), r.PathValue("attemptID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}
