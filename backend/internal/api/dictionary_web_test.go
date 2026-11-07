package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/testdb"
)

const webDictionaryFixture = `{"meta":{"status":200},"data":[{"slug":"猫","is_common":true,"tags":[],"jlpt":["jlpt-n5"],"japanese":[{"word":"猫","reading":"ねこ"},{"word":"ネコ"}],"senses":[{"english_definitions":["cat","<script>alert(1)</script>"],"parts_of_speech":["Noun"],"tags":["Usually written using kana alone"],"info":[],"see_also":["犬 1"],"antonyms":[]}]}]}`

func newWebDictionaryServer(t *testing.T, status int, body string) (*Server, http.Handler, *atomic.Int32) {
	t.Helper()
	s := NewServer(store.NewMemoryStore(), nil)
	var calls atomic.Int32
	s.dictionary.HTTP.Transport = dictionaryTransport(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	return s, s.Handler(), &calls
}

func webGet(h http.Handler, path, visitor string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", path, nil)
	r.RemoteAddr = "127.0.0.1:41000"
	if visitor != "" {
		r.Header.Set("CF-Connecting-IP", visitor)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestDictionaryWebIndexIsIndexable(t *testing.T) {
	_, h, calls := newWebDictionaryServer(t, 200, webDictionaryFixture)
	w := webGet(h, "/kamus", "")
	body := w.Body.String()
	if w.Code != 200 || calls.Load() != 0 {
		t.Fatalf("index: %d calls=%d", w.Code, calls.Load())
	}
	if !strings.Contains(body, `<link rel="canonical" href="https://nihongo.codeverta.com/kamus">`) || strings.Contains(body, "noindex") {
		t.Fatal("dictionary index must be canonical and indexable")
	}
	if !strings.Contains(body, `action="/kamus"`) || !strings.Contains(body, "jisho.org") {
		t.Fatal("index needs the search form and attribution")
	}
	if csp := w.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "form-action 'self'") || strings.Contains(csp, "unsafe") {
		t.Fatalf("unexpected CSP %q", csp)
	}
	if w = webGet(h, "/kamus/", ""); w.Code != http.StatusPermanentRedirect || w.Header().Get("Location") != "/kamus" {
		t.Fatalf("trailing slash should redirect: %d %s", w.Code, w.Header().Get("Location"))
	}
}

func TestDictionaryWebSearchRendersEscapedResults(t *testing.T) {
	_, h, _ := newWebDictionaryServer(t, 200, webDictionaryFixture)
	w := webGet(h, "/kamus?q=neko&jlpt=n5&common=1", "203.0.113.5")
	body := w.Body.String()
	if w.Code != 200 {
		t.Fatalf("search status %d", w.Code)
	}
	for _, want := range []string{
		`<meta name="robots" content="noindex,follow">`,
		`href="/kamus/kata/%e7%8c%ab"`,
		`ねこ`, `JLPT N5`, `&lt;script&gt;alert(1)&lt;/script&gt;`,
		`<option value="N5" selected>`, `name="common" value="1" checked`,
	} {
		if !strings.Contains(strings.ToLower(body), strings.ToLower(want)) {
			t.Errorf("search page missing %q", want)
		}
	}
	if strings.Contains(body, "<script>alert") {
		t.Fatal("upstream text must be escaped")
	}
	if strings.Contains(body, "Muat lebih banyak") {
		t.Fatal("a single result must not offer another page")
	}
}

func TestDictionaryWebLimitsOnlyUpstreamLookups(t *testing.T) {
	s, h, calls := newWebDictionaryServer(t, 200, webDictionaryFixture)
	s.webDictionaryLimiter.perIP = 2
	for _, q := range []string{"cat", "dog"} {
		if w := webGet(h, "/kamus?q="+q, "203.0.113.5"); w.Code != 200 {
			t.Fatalf("%s: %d", q, w.Code)
		}
	}
	w := webGet(h, "/kamus?q=fish", "203.0.113.5")
	if w.Code != http.StatusTooManyRequests || calls.Load() != 2 {
		t.Fatalf("third new word should be limited: %d calls=%d", w.Code, calls.Load())
	}
	if !strings.Contains(w.Body.String(), "jisho.org/search/fish") || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("limited page should link to Jisho and not be cached")
	}
	if w = webGet(h, "/kamus?q=cat", "203.0.113.5"); w.Code != 200 || calls.Load() != 2 {
		t.Fatalf("cached word must stay available: %d calls=%d", w.Code, calls.Load())
	}
	if w = webGet(h, "/kamus?q=fish", "203.0.113.6"); w.Code != 200 || calls.Load() != 3 {
		t.Fatalf("another visitor has its own quota: %d calls=%d", w.Code, calls.Load())
	}
	// The mobile API is not subject to the website quota.
	r := httptest.NewRequest("GET", "/v1/dictionary/search?q=bird", nil)
	r.RemoteAddr = "127.0.0.1:41000"
	r.Header.Set("CF-Connecting-IP", "203.0.113.5")
	api := httptest.NewRecorder()
	h.ServeHTTP(api, r)
	if api.Code != 200 {
		t.Fatalf("app API must ignore the web quota: %d", api.Code)
	}
}

func TestWebDictionaryGlobalLimit(t *testing.T) {
	s, h, _ := newWebDictionaryServer(t, 200, webDictionaryFixture)
	s.webDictionaryLimiter.total = 1
	if w := webGet(h, "/kamus?q=cat", "203.0.113.5"); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := webGet(h, "/kamus?q=dog", "203.0.113.9"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("global cap must apply across visitors: %d", w.Code)
	}
}

func TestWebDictionaryClientTrustsCloudflareOnlyFromLoopback(t *testing.T) {
	for _, tc := range []struct{ remote, header, want string }{
		{"127.0.0.1:1", "203.0.113.5", "203.0.113.5"},
		{"[::1]:1", "203.0.113.5", "203.0.113.5"},
		{"198.51.100.7:1", "203.0.113.5", "198.51.100.7"},
		{"127.0.0.1:1", "not-an-ip", "127.0.0.1"},
		{"127.0.0.1:1", "2001:db8:1:2:3:4:5:6", "2001:db8:1:2::/64"},
	} {
		r := httptest.NewRequest("GET", "/kamus", nil)
		r.RemoteAddr = tc.remote
		r.Header.Set("CF-Connecting-IP", tc.header)
		if got := webDictionaryClient(r); got != tc.want {
			t.Errorf("%s + %s = %s, want %s", tc.remote, tc.header, got, tc.want)
		}
	}
}

func TestDictionaryWebDetailPagesNeverFetchMissingWords(t *testing.T) {
	_, h, calls := newWebDictionaryServer(t, 200, webDictionaryFixture)
	w := webGet(h, "/kamus/kata/"+url.PathEscape("猫"), "")
	if w.Code != http.StatusNotFound || calls.Load() != 0 {
		t.Fatalf("missing word page: %d calls=%d", w.Code, calls.Load())
	}
	if !strings.Contains(w.Body.String(), "noindex") || !strings.Contains(strings.ToLower(w.Body.String()), "/kamus?q=%e7%8c%ab") {
		t.Fatal("missing word page should be noindex and offer a search")
	}
	if w = webGet(h, "/kamus/kanji/a", ""); w.Code != http.StatusNotFound || calls.Load() != 0 {
		t.Fatalf("invalid kanji: %d calls=%d", w.Code, calls.Load())
	}
}

func TestDictionaryWebUpstreamFailure(t *testing.T) {
	_, h, _ := newWebDictionaryServer(t, 500, "oops")
	w := webGet(h, "/kamus?q=cat", "203.0.113.5")
	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "jisho.org/search/cat") {
		t.Fatalf("upstream failure should be 503 with a Jisho link: %d", w.Code)
	}
}

func TestRobotsAndSitemap(t *testing.T) {
	_, h, _ := newWebDictionaryServer(t, 200, webDictionaryFixture)
	robots := webGet(h, "/robots.txt", "").Body.String()
	for _, want := range []string{"Disallow: /kamus?", "Disallow: /v1/", "Sitemap: https://nihongo.codeverta.com/sitemap.xml"} {
		if !strings.Contains(robots, want) {
			t.Errorf("robots.txt missing %q", want)
		}
	}
	w := webGet(h, "/sitemap.xml", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "<loc>https://nihongo.codeverta.com/kamus</loc>") {
		t.Fatalf("sitemap: %d %s", w.Code, w.Body.String())
	}
}

func TestLandingLinksToWebDictionary(t *testing.T) {
	_, h, _ := newWebDictionaryServer(t, 200, webDictionaryFixture)
	body := webGet(h, "/", "").Body.String()
	if !strings.Contains(body, `<a href="/kamus">Kamus</a>`) || !strings.Contains(body, `<form class="ex-search" action="/kamus"`) {
		t.Fatal("landing page should link to the web dictionary")
	}
}

func TestDictionaryWebWordPagesFromDatabase(t *testing.T) {
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
	var calls atomic.Int32
	s.dictionary.HTTP.Transport = dictionaryTransport(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(webDictionaryFixture))}, nil
	})
	if _, err := s.dictionary.Search(context.Background(), "neko", 1, "", false); err != nil {
		t.Fatal(err)
	}
	h := s.Handler()
	w := webGet(h, "/kamus/kata/"+url.PathEscape("猫"), "")
	body := w.Body.String()
	if w.Code != 200 || calls.Load() != 1 {
		t.Fatalf("archived word page: %d calls=%d", w.Code, calls.Load())
	}
	for _, want := range []string{"Arti 猫 (ねこ)", `rel="canonical"`, "ネコ", `href="/kamus/kanji/%e7%8c%ab"`, "Usually written using kana alone"} {
		if !strings.Contains(strings.ToLower(body), strings.ToLower(want)) {
			t.Errorf("word page missing %q", want)
		}
	}
	if strings.Contains(body, "noindex") {
		t.Fatal("archived word pages should be indexable")
	}
	sitemap := webGet(h, "/sitemap.xml", "").Body.String()
	if !strings.Contains(sitemap, "https://nihongo.codeverta.com/kamus/kata/%E7%8C%AB") {
		t.Fatalf("sitemap should list archived words: %s", sitemap)
	}
}
