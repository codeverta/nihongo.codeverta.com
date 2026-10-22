package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/news"
	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/testdb"
)

type newsFixtureFetcher struct{ calls atomic.Int32 }

func (f *newsFixtureFetcher) Fetch(context.Context, news.Source, time.Time) (news.FetchResult, error) {
	f.calls.Add(1)
	return news.FetchResult{Articles: []news.Article{{URL: "https://jp.globalvoices.org/2026/10/09/123/", Title: "日本のニュース", TitleTokens: []news.Token{}, Paragraphs: []news.Paragraph{{Text: "日本の文化について読むためのニュースです。新しい催しが始まります。", Tokens: []news.Token{}}}, Credits: []news.Credit{{Name: "Writer", Role: "author"}}, References: []news.Credit{}, PublishedAt: time.Now().UTC().Add(-time.Hour), License: "CC BY 3.0", LicenseURL: "https://creativecommons.org/licenses/by/3.0/"}}}, nil
}
func TestNewsGuestReadsAdminAndImages(t *testing.T) {
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	repo, err := news.New(db)
	if err != nil {
		t.Fatal(err)
	}
	fetcher := &newsFixtureFetcher{}
	if _, err = repo.Sync(context.Background(), fetcher, "global-voices", false); err != nil {
		t.Fatal(err)
	}
	s := api.NewServer(store.NewMemoryStore(), fakeGoogleVerifier{})
	if err = s.ConfigureNews(repo, fetcher, "test-key", t.TempDir(), context.Background()); err != nil {
		t.Fatal(err)
	}
	handler := s.Handler()
	request := func(method, path, body, key string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if key != "" {
			req.Header.Set("Authorization", "Bearer "+key)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	for _, path := range []string{"/v1/news", "/v1/news?limit=3"} {
		if w := request("GET", path, "", ""); w.Code != 200 {
			t.Fatal(path, w.Code, w.Body.String())
		}
	}
	page, _ := repo.List(context.Background(), 1, 20, "", false)
	id := page.Items[0].ID
	if w := request("GET", "/v1/news/"+id, "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "paragraphs") {
		t.Fatal(w.Code, w.Body.String())
	}
	if fetcher.calls.Load() != 1 {
		t.Fatal("guest read triggered upstream fetch")
	}
	for _, route := range []struct{ method, path string }{{"GET", "/v1/admin/news"}, {"GET", "/v1/admin/news/" + id}, {"PUT", "/v1/admin/news/" + id}, {"DELETE", "/v1/admin/news/" + id}, {"PUT", "/v1/admin/news/sources/osaka"}, {"POST", "/v1/admin/news/sources/osaka/fetch"}, {"POST", "/v1/admin/news/images"}} {
		for _, key := range []string{"", "wrong"} {
			if w := request(route.method, route.path, "{}", key); w.Code != 401 {
				t.Fatal("admin route unprotected", route, w.Code)
			}
		}
	}
	if w := request("GET", "/v1/news?limit=500", "", ""); w.Code != 400 {
		t.Fatal("unbounded list")
	}
	if w := request("PUT", "/v1/admin/news/sources/osaka", `{"timezone":"Asia/Jakarta","days":[0,4],"hour":9,"minute":30,"enabled":false}`, "test-key"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request("PUT", "/v1/admin/news/"+id, `{"hidden":true,"imageUrl":"","imageCredit":"","imageLicense":"","readings":[]}`, "test-key"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := request("GET", "/v1/news/"+id, "", ""); w.Code != 404 {
		t.Fatal("hidden news public")
	}
	if w := request("GET", "/v1/admin/news/"+id, "", "test-key"); w.Code != 200 {
		t.Fatal("admin cannot inspect hidden article")
	}
	if w := request("PUT", "/v1/admin/news/"+id, `{"imageUrl":"https://publisher.example/photo.jpg"}`, "test-key"); w.Code != 400 {
		t.Fatal("remote hotlink accepted")
	}
	var upload bytes.Buffer
	writer := multipart.NewWriter(&upload)
	part, _ := writer.CreateFormFile("file", "own.png")
	if err := png.Encode(part, image.NewRGBA(image.Rect(0, 0, 12, 12))); err != nil {
		t.Fatal(err)
	}
	writer.Close()
	req := httptest.NewRequest("POST", "/v1/admin/news/images", &upload)
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var payload struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if w := request("GET", payload.Data.URL, "", ""); w.Code != 200 || w.Header().Get("Content-Type") != "image/jpeg" || w.Body.Len() > 200*1024 {
		t.Fatal("uploaded image not served", w.Code)
	}
	if w := request("DELETE", "/v1/admin/news/"+id, "", "test-key"); w.Code != 204 {
		t.Fatal(w.Code, w.Body.String())
	}
	if fetcher.calls.Load() != 1 {
		t.Fatal("admin editing triggered fetch")
	}
	if w := request(http.MethodGet, "/admin/news", "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Kunci admin") {
		t.Fatal("admin UI missing")
	}
}
