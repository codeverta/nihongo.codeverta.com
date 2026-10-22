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
		{"GET", "/v1/simulation/n5"}, {"GET", "/v1/simulation/n4"}, {"GET", "/v1/simulation/n3"}, {"GET", "/v1/simulation/attempts/test"},
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

func TestSimulationN4API(t *testing.T) {
	server := newServer(true)
	token := login(t, server)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	overview := request("GET", "/v1/simulation/n4", "")
	if overview.Code != 200 || !strings.Contains(overview.Body.String(), `"questionCount":57`) || !strings.Contains(overview.Body.String(), `"durationMinutes":80`) {
		t.Fatal(overview.Body.String())
	}
	if w := request("GET", "/v1/simulation/unsupported", ""); w.Code != 400 {
		t.Fatal("unsupported level accepted")
	}
	w := request("POST", "/v1/exam-attempts", `{"examId":"jlpt-n4-simulation-1"}`)
	var attempt struct {
		Data domain.ExamAttempt `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &attempt); err != nil {
		t.Fatal(err)
	}
	if w.Code != 201 || attempt.Data.ExamID != "jlpt-n4-simulation-1" {
		t.Fatal(w.Body.String())
	}
	path := "/v1/simulation/attempts/" + attempt.Data.ID
	w = request("GET", path, "")
	var session struct {
		Data domain.SimulationSession `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if len(session.Data.Questions) != 28 || session.Data.Sections[0].DurationSeconds != 1500 || session.Data.Sections[1].DurationSeconds != 3300 {
		t.Fatal("wrong N4 session")
	}
	w = request("POST", path+"/submit-section", `{"section":"vocabulary"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if len(session.Data.Questions) != 29 || session.Data.Questions[21].Passage == "" {
		t.Fatal("missing N4 reading")
	}
	w = request("POST", path+"/submit-section", `{"section":"grammar"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Data.Result == nil || session.Data.Result.Total != 57 || len(session.Data.Result.Reviews) != 57 {
		t.Fatal("missing review")
	}
	if w = request("GET", "/v1/simulation/n5", ""); strings.Contains(w.Body.String(), attempt.Data.ID) {
		t.Fatal("N4 result leaked into N5 history")
	}
}

func TestSimulationN3API(t *testing.T) {
	server := newServer(true)
	token := login(t, server)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	overview := request("GET", "/v1/simulation/n3", "")
	if overview.Code != 200 || !strings.Contains(overview.Body.String(), `"questionCount":74`) || !strings.Contains(overview.Body.String(), `"durationMinutes":100`) {
		t.Fatal(overview.Body.String())
	}
	if w := request("GET", "/v1/simulation/unsupported", ""); w.Code != 400 {
		t.Fatal("unsupported level accepted")
	}
	w := request("POST", "/v1/exam-attempts", `{"examId":"jlpt-n3-simulation-1"}`)
	var attempt struct {
		Data domain.ExamAttempt `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &attempt); err != nil {
		t.Fatal(err)
	}
	if w.Code != 201 || attempt.Data.ExamID != "jlpt-n3-simulation-1" {
		t.Fatal(w.Body.String())
	}
	path := "/v1/simulation/attempts/" + attempt.Data.ID
	w = request("GET", path, "")
	var session struct {
		Data domain.SimulationSession `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if len(session.Data.Questions) != 35 || session.Data.Sections[0].DurationSeconds != 1800 || session.Data.Sections[1].DurationSeconds != 4200 {
		t.Fatal("wrong N3 session")
	}
	w = request("POST", path+"/submit-section", `{"section":"vocabulary"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if len(session.Data.Questions) != 39 || session.Data.Questions[23].Passage == "" {
		t.Fatal("missing N3 reading")
	}
	w = request("POST", path+"/submit-section", `{"section":"grammar"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Data.Result == nil || session.Data.Result.Total != 74 || len(session.Data.Result.Reviews) != 74 {
		t.Fatal("missing review")
	}
	if w = request("GET", "/v1/simulation/n5", ""); strings.Contains(w.Body.String(), attempt.Data.ID) {
		t.Fatal("N3 result leaked into N5 history")
	}
}

func TestJFTAPI(t *testing.T) {
	server := newServer(true)
	token := login(t, server)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	overview := request("GET", "/v1/simulation/jft", "")
	if overview.Code != 200 || !strings.Contains(overview.Body.String(), `"questionCount":36`) || !strings.Contains(overview.Body.String(), `"durationMinutes":45`) {
		t.Fatal(overview.Body.String())
	}
	if w := request("GET", "/v1/simulation/unsupported", ""); w.Code != 400 {
		t.Fatal("unsupported level accepted")
	}
	w := request("POST", "/v1/exam-attempts", `{"examId":"jft-a2-full-1"}`)
	var attempt struct {
		Data domain.ExamAttempt `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &attempt); err != nil {
		t.Fatal(err)
	}
	if w.Code != 201 || attempt.Data.ExamID != "jft-a2-full-1" {
		t.Fatal(w.Body.String())
	}
	path := "/v1/simulation/attempts/" + attempt.Data.ID
	w = request("GET", path, "")
	var session struct {
		Data domain.SimulationSession `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if len(session.Data.Questions) != 12 || session.Data.Sections[0].DurationSeconds != 2700 || session.Data.Sections[1].DurationSeconds != 0 {
		t.Fatal("wrong N3 session")
	}
	w = request("POST", path+"/submit-section", `{"section":"vocabulary"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Data.ActiveSection != "conversation" || len(session.Data.Questions) != 12 {
		t.Fatal("missing conversation section")
	}
	w = request("POST", path+"/submit-section", `{"section":"conversation"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Data.ActiveSection != "reading" || len(session.Data.Questions) != 12 || session.Data.Questions[0].Passage == "" {
		t.Fatal("missing reading")
	}
	w = request("POST", path+"/submit-section", `{"section":"reading"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Data.Result == nil || session.Data.Result.Total != 36 || len(session.Data.Result.Reviews) != 36 {
		t.Fatal("missing review")
	}
	if w = request("GET", "/v1/simulation/n5", ""); strings.Contains(w.Body.String(), attempt.Data.ID) {
		t.Fatal("JFT result leaked into N5 history")
	}
}

func TestSimulationNATAPI(t *testing.T) {
	server := newServer(true)
	token := login(t, server)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	overview := request("GET", "/v1/simulation/nat", "")
	if overview.Code != 200 || !strings.Contains(overview.Body.String(), `"questionCount":43`) || !strings.Contains(overview.Body.String(), `"durationMinutes":60`) {
		t.Fatal(overview.Body.String())
	}
	if w := request("GET", "/v1/simulation/unsupported", ""); w.Code != 400 {
		t.Fatal("unsupported level accepted")
	}
	w := request("POST", "/v1/exam-attempts", `{"examId":"nat-5q-full-1"}`)
	var attempt struct {
		Data domain.ExamAttempt `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &attempt); err != nil {
		t.Fatal(err)
	}
	if w.Code != 201 || attempt.Data.ExamID != "nat-5q-full-1" {
		t.Fatal(w.Body.String())
	}
	path := "/v1/simulation/attempts/" + attempt.Data.ID
	w = request("GET", path, "")
	var session struct {
		Data domain.SimulationSession `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if len(session.Data.Questions) != 21 || session.Data.Sections[0].DurationSeconds != 1200 || session.Data.Sections[1].DurationSeconds != 2400 {
		t.Fatal("wrong NAT session")
	}
	w = request("POST", path+"/submit-section", `{"section":"vocabulary"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if len(session.Data.Questions) != 22 || session.Data.Questions[17].Passage == "" {
		t.Fatal("missing NAT reading")
	}
	w = request("POST", path+"/submit-section", `{"section":"grammar"}`)
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Data.Result == nil || session.Data.Result.Total != 43 || len(session.Data.Result.Reviews) != 43 {
		t.Fatal("missing review")
	}
	if w = request("GET", "/v1/simulation/n5", ""); strings.Contains(w.Body.String(), attempt.Data.ID) {
		t.Fatal("NAT result leaked into N5 history")
	}
}
