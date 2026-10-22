package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLearningHistoryRequiresAuthAndRejectsSpoofedMetadata(t *testing.T) {
	server := newServer()
	request := func(method, path, body, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	path := "/v1/me/learning-history"
	if w := request("GET", path, "", ""); w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
	token := login(t, server)
	if w := request("GET", path, "", token); w.Code != 200 || !strings.Contains(w.Body.String(), `"items":[]`) {
		t.Fatal(w.Code, w.Body.String())
	}
	body := `{"position":0,"accessedAt":"` + time.Now().UTC().Format(time.RFC3339Nano) + `"}`
	if w := request("PUT", path+"/grammar/grammar-n5-01", body, token); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request("GET", path, "", token); w.Code != 200 || !strings.Contains(w.Body.String(), `"kind":"grammar"`) || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request("PUT", path+"/grammar/grammar-n5-01", `{"position":0,"title":"spoof"}`, token); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request("PUT", path+"/grammar/grammar-n5-01", `{"accessedAt":"`+time.Now().UTC().Format(time.RFC3339Nano)+`"}`, token); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
}
