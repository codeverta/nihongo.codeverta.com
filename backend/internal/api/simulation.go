package api

import (
	"net/http"
	"nihongo-master-backend/internal/store"
	"strings"
)

func (s *Server) simulationOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	level := strings.ToUpper(r.PathValue("level"))
	isNATPaper := strings.HasPrefix(level, "NAT-") && len(level) == 5 && level[4] >= '2' && level[4] <= '5'
	if level != "N5" && level != "N4" && level != "N3" && level != "N2" && level != "N1" && !store.IsJFTLevel(level) && level != "NAT" && !isNATPaper && !store.IsSSWLevel(level) {
		writeError(w, 400, "invalid_level", "Simulation supports N5, N4, N3, N2, N1, JFT Exams 1–5, NAT and SSW")
		return
	}
	result, err := s.store.SimulationOverviewChecked(userID(r), level)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, 200, result)
}

func (s *Server) simulationSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	result, err := s.store.SimulationSession(userID(r), r.PathValue("attemptID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, 200, result)
}

func (s *Server) simulationAnswer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		QuestionID string `json:"questionId"`
		OptionID   string `json:"optionId"`
		Marked     bool   `json:"marked"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", err.Error())
		return
	}
	if err := s.store.SaveSimulationAnswer(userID(r), r.PathValue("attemptID"), input.QuestionID, input.OptionID, input.Marked); err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, 200, map[string]bool{"saved": true})
}

func (s *Server) simulationSubmit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Section string `json:"section"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", err.Error())
		return
	}
	result, err := s.store.SubmitSimulationSection(userID(r), r.PathValue("attemptID"), input.Section)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, 200, result)
}
