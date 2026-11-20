package api_test

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLeaderboardAPI(t *testing.T) {
	server := newServer()
	get := func(path, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, nil)
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	// Guests can view the public list; it is cacheable and has no emails.
	w := get("/v1/leaderboard?period=week", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"entries":[]`) || !strings.Contains(w.Body.String(), `"weekEndsAt"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "public, max-age=60" {
		t.Fatal("public list must be cacheable", w.Header())
	}
	if w = get("/v1/leaderboard?period=year", ""); w.Code != 422 {
		t.Fatal("bad period accepted", w.Code)
	}
	if w = get("/v1/me/leaderboard", ""); w.Code != 401 {
		t.Fatal("own standing without sign-in", w.Code)
	}
	token := login(t, server)
	w = get("/v1/me/leaderboard", token)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"hidden":false`) || !strings.Contains(w.Body.String(), `"lastWeek":null`) {
		t.Fatal(w.Body.String())
	}
	put := func(body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("PUT", "/v1/me/leaderboard/settings", strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	if w = put(`{}`); w.Code != 400 {
		t.Fatal("missing hidden accepted", w.Code)
	}
	if w = put(`{"hidden":true}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w = get("/v1/me/leaderboard", token); !strings.Contains(w.Body.String(), `"hidden":true`) {
		t.Fatal("setting not applied", w.Body.String())
	}
	// Admin exclusion needs an admin dashboard session.
	r := httptest.NewRequest("PUT", "/v1/admin/users/a5090412-9ab4-5852-8229-c7d19eb5cb1c/leaderboard", strings.NewReader(`{"excluded":true}`))
	w = httptest.NewRecorder()
	server.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("admin exclusion without session", w.Code)
	}
	// The per-user rate limit eventually answers 429.
	limited := false
	for i := 0; i < 40 && !limited; i++ {
		limited = get("/v1/me/leaderboard", token).Code == 429
	}
	if !limited {
		t.Fatal("leaderboard is not rate limited")
	}
}
