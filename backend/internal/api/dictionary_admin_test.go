package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/testdb"
	"strings"
	"testing"
)

type dictionaryTransport func(*http.Request) (*http.Response, error)

func (f dictionaryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDictionaryAdminRequiresTokenAndSupportsRefresh(t *testing.T) {
	st := store.NewMemoryStore()
	if err := st.EnableGamificationPersistence(testdb.DSN(t)); err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	cache, err := st.DictionaryCache()
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(st, nil)
	s.ConfigureDictionaryCache(cache)
	s.ConfigureDictionaryAdmin("dictionary-test-key")
	s.dictionary.HTTP.Transport = dictionaryTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"meta":{"status":200},"data":[]}`))}, nil
	})
	if _, err := s.dictionary.Search(context.Background(), "konnichiwa", 1, "N5", true); err != nil {
		t.Fatal(err)
	}
	h := s.Handler()
	for _, route := range []struct{ method, path string }{{"GET", "/v1/admin/dictionary"}, {"POST", "/v1/admin/dictionary/any/refresh"}} {
		if w := bannerRequest(h, route.method, route.path, "wrong", nil); w.Code != 401 {
			t.Fatal("admin route exposed", w.Code)
		}
	}
	w := bannerRequest(h, "GET", "/v1/admin/dictionary", "dictionary-test-key", nil)
	var payload struct {
		Data struct {
			Items []struct {
				ID     string
				Query  string
				Mode   string
				JLPT   string
				Common bool
			}
			Total int
		}
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil || w.Code != 200 || payload.Data.Total != 1 {
		t.Fatalf("list %d %s %v", w.Code, w.Body.String(), err)
	}
	record := payload.Data.Items[0]
	if record.Query != "こんにちは" || record.Mode != "japanese" || record.JLPT != "N5" || !record.Common {
		t.Fatalf("metadata %+v", record)
	}
	// Existing cache refresh must wait for the source throttle, not bypass it.
	if w = bannerRequest(h, "POST", "/v1/admin/dictionary/"+record.ID+"/refresh", "dictionary-test-key", nil); w.Code != 200 {
		t.Fatalf("refresh %d %s", w.Code, w.Body.String())
	}
	if w = bannerRequest(h, "POST", "/v1/admin/dictionary/missing/refresh", "dictionary-test-key", nil); w.Code != 404 {
		t.Fatal("missing entry", w.Code)
	}
	for _, path := range []string{"/v1/admin/dictionary?page=0", "/v1/admin/dictionary?page=oops", "/v1/dictionary/search?q=cat&mode=invalid"} {
		if w = bannerRequest(h, "GET", path, "dictionary-test-key", nil); w.Code != 400 {
			t.Fatal(path, w.Code)
		}
	}
	public := httptest.NewRecorder()
	h.ServeHTTP(public, httptest.NewRequest("GET", "/admin/dictionary", nil))
	if public.Code != 200 || !strings.Contains(public.Body.String(), "Cache Kamus") {
		t.Fatal("admin page missing")
	}
	s.ConfigureDictionaryAdmin("")
	if w = bannerRequest(h, "GET", "/v1/admin/dictionary", "dictionary-test-key", nil); w.Code != 401 {
		t.Fatal("empty token opened admin")
	}
}
