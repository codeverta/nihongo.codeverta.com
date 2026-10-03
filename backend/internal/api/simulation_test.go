package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"nihongo-master-backend/internal/domain"
	"strings"
	"testing"
)

func TestSimulationAPIAuthAndContract(t *testing.T) {
	server := newServer()
	endpoints := []struct{ method, path string }{
		{"GET", "/v1/simulation/n5"}, {"GET", "/v1/simulation/attempts/test"},
		{"PUT", "/v1/simulation/attempts/test/answers"}, {"POST", "/v1/simulation/attempts/test/submit-section"},
	}
	for _, e := range endpoints {
		w := httptest.NewRecorder()
		server.ServeHTTP(w, httptest.NewRequest(e.method, e.path, nil))
		if w.Code != 401 {
			t.Fatalf("%s %s must require auth", e.method, e.path)
		}
	}
	token := login(t, server)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	started := request("POST", "/v1/exam-attempts", `{"examId":"jlpt-n5-simulation-1"}`)
	var attempt struct {
		Data domain.ExamAttempt `json:"data"`
	}
	if err := json.Unmarshal(started.Body.Bytes(), &attempt); err != nil {
		t.Fatal(err)
	}
	if started.Code != 201 || attempt.Data.ID == "" {
		t.Fatalf("start: %s", started.Body.String())
	}
	path := "/v1/simulation/attempts/" + attempt.Data.ID
	loaded := request("GET", path, "")
	var session struct {
		Data domain.SimulationSession `json:"data"`
	}
	json.Unmarshal(loaded.Body.Bytes(), &session)
	if loaded.Code != 200 || len(session.Data.Questions) != 21 || strings.Contains(loaded.Body.String(), "correctOptionId") {
		t.Fatalf("bad session: %s", loaded.Body.String())
	}
	q := session.Data.Questions[0]
	data, _ := json.Marshal(map[string]any{"questionId": q.ID, "optionId": q.Options[0].ID, "marked": true})
	if w := request("PUT", path+"/answers", string(data)); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := request("PUT", path+"/answers", `{"questionId":"n5-sim-01-q22","optionId":"invalid"}`); w.Code != 422 {
		t.Fatal("future question accepted")
	}
	if w := request("POST", path+"/submit-section", `{"section":"vocabulary"}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := request("PUT", path+"/answers", string(data)); w.Code != 422 {
		t.Fatal("closed section accepted")
	}
	if w := request("POST", path+"/submit-section", `{"section":"grammar"}`); w.Code != 200 || !strings.Contains(w.Body.String(), "correctOptionId") {
		t.Fatal("missing review after final submission")
	}
	if w := request("GET", "/v1/simulation/attempts/not-owned", ""); w.Code != http.StatusNotFound {
		t.Fatal("unknown attempt must be hidden")
	}
}
