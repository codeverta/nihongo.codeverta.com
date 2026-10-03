package api

import "net/http"

func (s *Server) kanjiCourse(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.KanjiCourse(userID(r)))
}
func (s *Server) kanjiLesson(w http.ResponseWriter, r *http.Request) {
	lesson, err := s.store.KanjiLesson(userID(r), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, lesson)
}
func (s *Server) startKanjiAttempt(w http.ResponseWriter, r *http.Request) {
	attempt, err := s.store.StartKanjiAttempt(userID(r), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusCreated, attempt)
}
func (s *Server) answerKanjiQuestion(w http.ResponseWriter, r *http.Request) {
	var input struct {
		QuestionID string `json:"questionId"`
		OptionID   string `json:"optionId"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	feedback, err := s.store.AnswerKanjiQuestion(userID(r), r.PathValue("attemptID"), input.QuestionID, input.OptionID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, feedback)
}
func (s *Server) completeKanjiAttempt(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.CompleteKanjiAttempt(userID(r), r.PathValue("attemptID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}
