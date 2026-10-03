package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"nihongo-master-backend/internal/domain"
	"strings"
	"testing"
)

func TestGrammarAPIContractAndAuthentication(t *testing.T) {
	server := newServer()
	for _, path := range []string{"/v1/grammar/course", "/v1/grammar/lessons/grammar-n5-01"} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s must require login", path)
		}
	}
	token := login(t, server)
	request := httptest.NewRequest(http.MethodGet, "/v1/grammar/course", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	var course struct {
		Data domain.GrammarCourse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &course); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || len(course.Data.Lessons) != 40 {
		t.Fatalf("bad catalog: %s", response.Body.String())
	}
	for _, lesson := range course.Data.Lessons {
		r := httptest.NewRequest(http.MethodGet, "/v1/grammar/lessons/"+lesson.ID, nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		var payload struct {
			Data domain.GrammarLesson `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || payload.Data.ID != lesson.ID || len(payload.Data.Points) != 2 {
			t.Fatalf("invalid lesson %s", lesson.ID)
		}
		if strings.Contains(w.Body.String(), "correctOptionId") {
			t.Fatal("lesson response leaks quiz answers")
		}
	}
	r := httptest.NewRequest(http.MethodPost, "/v1/grammar/lessons/grammar-n5-01/attempts", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	var payload struct {
		Data domain.GrammarAttempt `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if w.Code != 201 || len(payload.Data.Questions) != 5 {
		t.Fatalf("bad quiz: %s", w.Body.String())
	}
	// First authored question asks for 学生です.
	q := payload.Data.Questions[0]
	var choice string
	for _, o := range q.Options {
		if o.Text == "です" {
			choice = o.ID
		}
	}
	r = httptest.NewRequest(http.MethodPost, "/v1/grammar/attempts/"+payload.Data.ID+"/answers", strings.NewReader(`{"questionId":"`+q.ID+`","optionId":"`+choice+`"}`))
	r.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"correct":true`) {
		t.Fatalf("bad feedback: %s", w.Body.String())
	}
}
