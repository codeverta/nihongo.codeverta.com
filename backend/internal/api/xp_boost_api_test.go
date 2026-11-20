package api_test

import (
	"net/http"
	"net/url"
	"testing"
)

// Double XP over HTTP (docs/ads-xp-spec.md, phase C). A finished Kaigo
// lesson is used as the completion because it earns XP without quiz data.

func (e *unlockEnv) finishPracticeLesson(t *testing.T, lesson string) string {
	t.Helper()
	data := e.data(t, e.do(http.MethodPost, "/v1/me/practice-lessons/nursing-care/"+lesson+"/complete", ""), http.StatusOK)
	id, _ := data["xp"].(map[string]any)["completionId"].(string)
	if id == "" {
		t.Fatalf("no completion id: %v", data)
	}
	return id
}

func boostItem(completionID string) string {
	return `{"kind":"xp-boost","itemId":"` + completionID + `"}`
}

func TestWatchingAnAdDoublesTheXPOfOneActivity(t *testing.T) {
	e := newUnlockEnv(t, "")
	completion := e.finishPracticeLesson(t, "kaigo-01")

	claimed := e.watchAdFor(t, "xp-boost", completion, "tx-boost")
	xp, _ := claimed["xp"].(map[string]any)
	if claimed["unlocked"] != true || xp["awarded"] != float64(5) || xp["total"] != float64(10) {
		t.Fatalf("claim = %v", claimed)
	}
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/intents", boostItem(completion)), http.StatusConflict); body["code"] != "boost_used" {
		t.Fatalf("second intent = %v", body)
	}
	// Double XP is not an unlock and does not use the unlock limit.
	list := e.data(t, e.do(http.MethodGet, "/v1/unlocks", ""), http.StatusOK)
	if items, _ := list["items"].([]any); len(items) != 0 || list["remaining"] != float64(20) {
		t.Fatalf("unlock list = %v", list)
	}
}

func TestDoubleXPHasNoFallbackAndNeedsTheSwitch(t *testing.T) {
	e := newUnlockEnv(t, "")
	completion := e.finishPracticeLesson(t, "kaigo-02")
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/fallback", boostItem(completion)), http.StatusForbidden); body["code"] != "boost_no_fallback" {
		t.Fatalf("fallback = %v", body)
	}
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/intents", boostItem("practice-lesson:nursing-care:never")), http.StatusUnprocessableEntity); body["code"] != "not_locked" {
		t.Fatalf("unknown completion = %v", body)
	}

	off := newUnlockEnv(t, `{"enabled":true,"xpBoostKinds":{"lessons":true,"exams":true,"practice":false}}`)
	offCompletion := off.finishPracticeLesson(t, "kaigo-01")
	if body := off.errorBody(t, off.do(http.MethodPost, "/v1/unlocks/intents", boostItem(offCompletion)), http.StatusForbidden); body["code"] != "unlock_unavailable" {
		t.Fatalf("switched off = %v", body)
	}
}

func TestLifetimeClaimsDoubleXPWithoutAnAd(t *testing.T) {
	e := newUnlockEnv(t, "")
	completion := e.finishPracticeLesson(t, "kaigo-03")
	path := "/v1/xp-boosts/" + url.PathEscape(completion) + "/free"
	if body := e.errorBody(t, e.do(http.MethodPost, path, ""), http.StatusPaymentRequired); body["code"] != "premium_required" {
		t.Fatalf("free claim without Lifetime = %v", body)
	}
	e.makeLifetime(t)
	data := e.data(t, e.do(http.MethodPost, path, ""), http.StatusOK)
	if data["xp"].(map[string]any)["awarded"] != float64(5) {
		t.Fatalf("free claim = %v", data)
	}
	if body := e.errorBody(t, e.do(http.MethodPost, path, ""), http.StatusConflict); body["code"] != "boost_used" {
		t.Fatalf("second free claim = %v", body)
	}
}

func TestDoubleXPDailyLimit(t *testing.T) {
	e := newUnlockEnv(t, `{"enabled":true,"xpBoostDailyCap":1}`)
	e.makeLifetime(t)
	first := e.finishPracticeLesson(t, "kaigo-01")
	second := e.finishPracticeLesson(t, "kaigo-02")
	e.data(t, e.do(http.MethodPost, "/v1/xp-boosts/"+url.PathEscape(first)+"/free", ""), http.StatusOK)
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/xp-boosts/"+url.PathEscape(second)+"/free", ""), http.StatusTooManyRequests); body["code"] != "boost_limit" {
		t.Fatalf("over the limit = %v", body)
	}
}
