package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nihongo-master-backend/internal/domain"
)

func TestKanjiN4APIContract(t *testing.T) {
	server := newServer()
	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/v1/kanji/course?level=N4", "/v1/kanji/lessons/kanji-n4-01"} {
		if request("GET", path, "", "").Code != http.StatusUnauthorized {
			t.Fatalf("%s must require login", path)
		}
	}
	token := login(t, server)
	for level, count := range map[string]int{"": 20, "N5": 20, "N4": 34} {
		w := request("GET", "/v1/kanji/course?level="+level, token, "")
		var payload struct {
			Data domain.KanjiCourse `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil || w.Code != 200 || len(payload.Data.Lessons) != count {
			t.Fatalf("invalid catalog: %s %v", w.Body.String(), err)
		}
	}
	for _, level := range []string{"N2", "N1", "n4"} {
		if request("GET", "/v1/kanji/course?level="+level, token, "").Code != 400 {
			t.Fatalf("unsupported level %s accepted", level)
		}
	}
	for _, id := range []string{"kanji-n4-01", "kanji-n4-34"} {
		w := request("GET", "/v1/kanji/lessons/"+id, token, "")
		var payload struct {
			Data domain.KanjiLesson `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil || w.Code != 200 || payload.Data.ID != id {
			t.Fatalf("invalid detail: %s", w.Body.String())
		}
		if strings.Contains(w.Body.String(), "Questions") || strings.Contains(w.Body.String(), "Correct") || strings.Contains(w.Body.String(), "correctOptionId") {
			t.Fatal("lesson exposes answer key")
		}
		w = request("POST", "/v1/kanji/lessons/"+id+"/attempts", token, "")
		var quiz struct {
			Data domain.KanjiAttempt `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &quiz); err != nil || w.Code != 201 || len(quiz.Data.Questions) != 2*len(payload.Data.Characters) {
			t.Fatalf("invalid quiz: %s", w.Body.String())
		}
		q := quiz.Data.Questions[0]
		optionID := ""
		for _, o := range q.Options {
			if o.Text == payload.Data.Characters[0].Meaning {
				optionID = o.ID
			}
		}
		body, _ := json.Marshal(map[string]string{"questionId": q.ID, "optionId": optionID})
		w = request("POST", "/v1/kanji/attempts/"+quiz.Data.ID+"/answers", token, string(body))
		var feedback struct {
			Data domain.KanjiAnswerFeedback `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &feedback); err != nil || w.Code != 200 || !feedback.Data.Correct || feedback.Data.Explanation == "" {
			t.Fatalf("answer feedback failed: %s", w.Body.String())
		}
	}
}
