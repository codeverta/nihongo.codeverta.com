package api_test

import (
	"net/http/httptest"
	"testing"
)

func TestDictionaryValidatesPublicRequestsWithoutLogin(t *testing.T) {
	handler := newServer()
	for _, path := range []string{"/v1/dictionary/search?q=", "/v1/dictionary/search?q=cat&page=oops", "/v1/dictionary/search?q=cat&jlpt=N8", "/v1/dictionary/search?q=cat&common=yes", "/v1/dictionary/kanji/abc"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 400 {
			t.Fatalf("invalid dictionary request was not rejected: %s %d", path, w.Code)
		}
	}
}
