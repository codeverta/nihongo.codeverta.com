package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFinishedPracticeLessonEarnsXPOnce(t *testing.T) {
	e := newUnlockEnv(t, "")
	path := "/v1/me/practice-lessons/nursing-care/kaigo-01/complete"
	first := e.data(t, e.do(http.MethodPost, path, ""), http.StatusOK)
	xp, _ := first["xp"].(map[string]any)
	if xp["awarded"] != float64(5) || xp["streak"] != float64(1) || xp["completionId"] != "practice-lesson:nursing-care:kaigo-01" {
		t.Fatalf("first = %v", first)
	}
	again := e.data(t, e.do(http.MethodPost, path, ""), http.StatusOK)
	if again["xp"].(map[string]any)["awarded"] != float64(0) {
		t.Fatalf("a lesson pays once: %v", again)
	}
	if response := e.do(http.MethodPost, "/v1/me/practice-lessons/Nursing/x/complete", ""); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad id: %d", response.Code)
	}
	signedOut := &unlockEnv{handler: e.handler}
	if response := signedOut.do(http.MethodPost, path, ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("signed out: %d", response.Code)
	}
}

func TestRepeatXPCapIsValidatedAndKept(t *testing.T) {
	e := newUnlockEnv(t, `{"enabled":true,"xpRepeatDailyCap":40}`)
	save := func(body string) int {
		request := httptest.NewRequest(http.MethodPut, "/v1/admin/ads-config", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer ads-token")
		response := httptest.NewRecorder()
		e.handler.ServeHTTP(response, request)
		return response.Code
	}
	base := `"enabled":true,"banner":{"enabled":false,"unitId":""},"newsNative":{"enabled":false,"unitId":"","every":6},"jobsNative":{"enabled":false,"unitId":"","every":8},"aiChatRewarded":{"enabled":false,"unitId":""}`
	if code := save(`{` + base + `,"xpRepeatDailyCap":5}`); code != http.StatusUnprocessableEntity {
		t.Fatalf("cap of 5: %d", code)
	}
	if code := save(`{` + base + `}`); code != http.StatusOK {
		t.Fatalf("save without the cap: %d", code)
	}
	public := httptest.NewRecorder()
	e.handler.ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/v1/ads-config", nil))
	if !strings.Contains(public.Body.String(), `"xpRepeatDailyCap":40`) {
		t.Fatalf("an older dashboard must not reset the cap: %s", public.Body.String())
	}
}
