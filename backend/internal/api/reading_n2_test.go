package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nihongo-master-backend/internal/domain"
)

func TestReadingN2API(t *testing.T) {
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
	if request("GET", "/v1/reading/course?level=N2", "", "").Code != http.StatusUnauthorized {
		t.Fatal("N2 course must require authentication")
	}
	token := login(t, server)
	w := request("GET", "/v1/reading/course?level=N2", token, "")
	var course struct {
		Data domain.ReadingCourse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &course); err != nil || w.Code != 200 || course.Data.Level != "N2" || len(course.Data.Lessons) != 12 {
		t.Fatalf("N2 course failed: %s %v", w.Body.String(), err)
	}
	for _, summary := range course.Data.Lessons {
		w = request("GET", "/v1/reading/lessons/"+summary.ID, token, "")
		var detail struct {
			Data domain.ReadingLesson `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil || w.Code != 200 || len(detail.Data.Pages) != 3 {
			t.Fatalf("lesson unavailable: %s", w.Body.String())
		}
	}
	if request("GET", "/v1/reading/course?level=N1", token, "").Code != http.StatusBadRequest {
		t.Fatal("unsupported level accepted")
	}
	w = request("POST", "/v1/reading/lessons/reading-n2-01/attempts", token, "")
	var quiz struct {
		Data domain.ReadingAttempt `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &quiz); err != nil || w.Code != 201 || len(quiz.Data.Questions) != 5 {
		t.Fatalf("N2 quiz failed: %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "correctOptionId") {
		t.Fatal("attempt leaked answer key")
	}
	answers := []string{
		"すべての人にとっての使いやすさ。",
		"操作を間違えるのが怖かったから。",
		"利用をやめて声を届けない人がいるから。",
		"他の手段と支援を残して進めるべきだ。",
		"成果は効率と実際の利用可能性の両方から評価する。",
	}
	for i, question := range quiz.Data.Questions {
		optionID := ""
		for _, option := range question.Options {
			if option.Text == answers[i] {
				optionID = option.ID
			}
		}
		if optionID == "" {
			t.Fatalf("missing expected answer %d", i)
		}
		body, err := json.Marshal(map[string]string{"questionId": question.ID, "optionId": optionID})
		if err != nil {
			t.Fatal(err)
		}
		w = request("POST", "/v1/reading/attempts/"+quiz.Data.ID+"/answers", token, string(body))
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"correct":true`) {
			t.Fatalf("N2 grading failed: %s", w.Body.String())
		}
	}
	w = request("POST", "/v1/reading/attempts/"+quiz.Data.ID+"/complete", token, "")
	var result struct {
		Data domain.ReadingAttemptResult `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || w.Code != 200 || result.Data.Score != 100 || result.Data.Progress != 8 {
		t.Fatalf("N2 completion failed: %s", w.Body.String())
	}
}
