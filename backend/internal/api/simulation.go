package api

import "net/http"

func (s *Server) simulationOverview(w http.ResponseWriter, r *http.Request) {
	writeData(w, 200, s.store.SimulationOverview(userID(r)))
}

func (s *Server) simulationSession(w http.ResponseWriter, r *http.Request) {
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
