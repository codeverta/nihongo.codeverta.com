package api_test

import (
	"encoding/json"
	"net/http/httptest"
	"nihongo-master-backend/internal/domain"
	"strings"
	"testing"
)

func TestGamificationAPIAuthenticatesAndTracksDiagnostic(t *testing.T) {
	server := newServer()
	for _, path := range []string{"/v1/me/gamification", "/v1/me/achievements"} {
		w := httptest.NewRecorder()
		server.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 401 {
			t.Fatal("unauthenticated gamification access")
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
	w := request("GET", "/v1/me/gamification", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"xp":0`) || !strings.Contains(w.Body.String(), `"timezone":"Asia/Jakarta"`) || strings.Contains(w.Body.String(), "unlockedAt") {
		t.Fatal("new account must not show dummy rewards", w.Body.String())
	}
	// Manual legacy completion must never grant XP or badges.
	request("PUT", "/v1/me/progress/fake-lesson", `{"completion":100,"lastPosition":9999}`)
	if w = request("GET", "/v1/me/gamification", ""); !strings.Contains(w.Body.String(), `"lessonsCompleted":0`) {
		t.Fatal("forged progress granted achievement")
	}
	w = request("POST", "/v1/exam-attempts", `{"examId":"jlpt-n5-diagnostic"}`)
	var attempt struct {
		Data domain.ExamAttempt `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &attempt); err != nil {
		t.Fatal(err)
	}
	path := "/v1/exam-attempts/" + attempt.Data.ID
	w = request("GET", path, "")
	var session struct {
		Data domain.DiagnosticSession `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	q := session.Data.Questions[0]
	raw, _ := json.Marshal(map[string]string{"questionId": q.ID, "optionId": q.Options[0].ID})
	if w = request("PUT", path+"/answers", string(raw)); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w = request("POST", path+"/complete", ""); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w = request("GET", "/v1/me/gamification", "")
	if !strings.Contains(w.Body.String(), `"xp":0`) || !strings.Contains(w.Body.String(), `"streakDays":1`) || !strings.Contains(w.Body.String(), `"examsCompleted":1`) {
		t.Fatal("failed diagnostic should count activity but not XP", w.Body.String())
	}
	if w = request("GET", "/v1/home", ""); !strings.Contains(w.Body.String(), `"streakDays":1`) {
		t.Fatal("dashboard uses stale streak")
	}
}
