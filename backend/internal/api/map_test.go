package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nihongo-master-backend/internal/places"
	"nihongo-master-backend/internal/store"
)

func newMapServer(t *testing.T) (*Server, http.Handler, *places.Service) {
	t.Helper()
	fetcher := places.NewFetcher("")
	fetcher.HTTP.Transport = dictionaryTransport(func(r *http.Request) (*http.Response, error) {
		body := `{"entities":{"Q7":{"id":"Q7","labels":{"en":{"value":"Shibuya Crossing"},"id":{"value":"Persimpangan Shibuya"},"ja":{"value":"渋谷スクランブル交差点"}},"sitelinks":{},"claims":{"P625":[{"rank":"normal","mainsnak":{"datavalue":{"value":{"latitude":35.6595,"longitude":139.7005}}}}]}}}}`
		if r.URL.Host != "www.wikidata.org" {
			return &http.Response{StatusCode: 404, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	fetcher.Gap = 0
	service := &places.Service{Store: places.NewMemoryStore(), Fetcher: fetcher}
	s := NewServer(store.NewMemoryStore(), nil)
	s.ConfigureMap(service, "map-key", nil)
	return s, s.Handler(), service
}

func mapRequest(h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	r := httptest.NewRequest(method, path, reader)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestMapAdminFlowAndPublicList(t *testing.T) {
	_, h, service := newMapServer(t)
	if w := mapRequest(h, "GET", "/v1/admin/map/spots", "wrong", nil); w.Code != 401 {
		t.Fatalf("admin without key: %d", w.Code)
	}
	w := mapRequest(h, "POST", "/v1/admin/map/spots", "map-key", map[string]string{"input": "Q7", "category": "city"})
	if w.Code != 201 {
		t.Fatalf("import: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID          string `json:"id"`
			Prefecture  string `json:"prefecture"`
			Published   bool   `json:"published"`
			Publishable bool   `json:"publishable"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if created.Data.Prefecture != "13" || created.Data.Published || !created.Data.Publishable {
		t.Fatalf("imported spot: %+v", created.Data)
	}
	if w = mapRequest(h, "POST", "/v1/admin/map/spots", "map-key", map[string]string{"input": "Q7", "category": "city"}); w.Code != http.StatusConflict {
		t.Fatalf("duplicate import: %d", w.Code)
	}

	list := mapRequest(h, "GET", "/v1/map/spots", "", nil)
	if list.Code != 200 || strings.Contains(list.Body.String(), "Shibuya") {
		t.Fatalf("drafts must stay hidden: %d %s", list.Code, list.Body.String())
	}
	edit := places.Edit{Category: "city", NameEN: "Shibuya Crossing", NameID: "Persimpangan Shibuya", NameJA: "渋谷スクランブル交差点",
		ManualSummaryID: "Persimpangan pejalan kaki tersibuk.", Phrases: []places.Phrase{{Japanese: "信号", Reading: "しんごう", MeaningID: "lampu lalu lintas", MeaningEN: "traffic light"}}, Published: true}
	if w = mapRequest(h, "PUT", "/v1/admin/map/spots/"+created.Data.ID, "map-key", edit); w.Code != 200 {
		t.Fatalf("publish: %d %s", w.Code, w.Body.String())
	}
	list = mapRequest(h, "GET", "/v1/map/spots", "", nil)
	var public struct {
		Data struct {
			Spots      []places.PublicSpot `json:"spots"`
			Categories []string            `json:"categories"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &public); err != nil || len(public.Data.Spots) != 1 || len(public.Data.Categories) != 6 {
		t.Fatalf("public list: %v %s", err, list.Body.String())
	}
	spot := public.Data.Spots[0]
	if spot.Name["id"] != "Persimpangan Shibuya" || spot.Summary["id"].Text == "" || spot.Lat == 0 || len(spot.Phrases) != 1 {
		t.Fatalf("public spot: %+v", spot)
	}
	etag := list.Header().Get("ETag")
	r := httptest.NewRequest("GET", "/v1/map/spots", nil)
	r.Header.Set("If-None-Match", etag)
	cached := httptest.NewRecorder()
	h.ServeHTTP(cached, r)
	if etag == "" || cached.Code != http.StatusNotModified {
		t.Fatalf("ETag revalidation: %q %d", etag, cached.Code)
	}

	if w = mapRequest(h, "DELETE", "/v1/admin/map/spots/"+created.Data.ID, "map-key", nil); w.Code != 204 {
		t.Fatalf("delete: %d", w.Code)
	}
	if spots, _ := service.Store.All(context.Background()); len(spots) != 0 {
		t.Fatal("spot should be deleted")
	}
}

func TestMapEndpointsWithoutConfiguration(t *testing.T) {
	s := NewServer(store.NewMemoryStore(), nil)
	h := s.Handler()
	if w := mapRequest(h, "GET", "/v1/map/spots", "", nil); w.Code != 503 {
		t.Fatalf("unconfigured map: %d", w.Code)
	}
	if w := mapRequest(h, "GET", "/v1/jobs/prefecture-counts", "", nil); w.Code != 503 {
		t.Fatalf("unconfigured jobs: %d", w.Code)
	}
	if w := mapRequest(h, "GET", "/admin/map", "", nil); w.Code != 200 || !strings.Contains(w.Body.String(), "Peta Jepang") {
		t.Fatalf("admin page: %d", w.Code)
	}
}
