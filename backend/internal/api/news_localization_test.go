package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/news"
	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/testdb"
)

type localizedNewsFixtureFetcher struct {
	article news.Article
}

func (f localizedNewsFixtureFetcher) Fetch(context.Context, news.Source, time.Time) (news.FetchResult, error) {
	return news.FetchResult{Articles: []news.Article{f.article}}, nil
}

func TestPublicNewsAPITranslatesEnglishMetadataAndPreservesJapaneseSource(t *testing.T) {
	t.Setenv("PUBLIC_ENGLISH_CONTENT_ENABLED", "false")
	t.Setenv("ENGLISH_QA_PREVIEW_ENABLED", "true")

	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	repo, err := news.New(db)
	if err != nil {
		t.Fatal(err)
	}
	article := news.Article{
		URL:         "https://jp.globalvoices.org/2026/10/09/123/",
		Title:       "東京の祭り",
		TitleTokens: []news.Token{{Start: 0, End: 2, Surface: "東京", Base: "東京", Reading: "とうきょう", ReadingSource: "automatic"}},
		Summary:     "日本の夏祭りを紹介する記事です。",
		Paragraphs: []news.Paragraph{{
			Text:   "東京で新しい文化の催しが始まり、多くの人が祭りを楽しんでいます。",
			Kind:   "paragraph",
			Tokens: []news.Token{{Start: 0, End: 2, Surface: "東京", Base: "東京", Reading: "とうきょう", ReadingSource: "automatic"}},
		}},
		Changes:     "Furigana ditambahkan oleh Nihongo Master; media eksternal tidak disertakan.",
		PublishedAt: time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC),
		DateType:    "published",
		Credits:     []news.Credit{{Name: "Writer", Role: "author"}},
		References:  []news.Credit{},
		License:     "CC BY 3.0",
		LicenseURL:  "https://creativecommons.org/licenses/by/3.0/",
	}
	fetcher := localizedNewsFixtureFetcher{article: article}
	if _, err := repo.Sync(context.Background(), fetcher, "global-voices", false); err != nil {
		t.Fatal(err)
	}
	server := api.NewServer(store.NewMemoryStore(), fakeGoogleVerifier{})
	if err = server.ConfigureNews(repo, fetcher, "test-key", t.TempDir(), context.Background()); err != nil {
		t.Fatal(err)
	}
	handler := server.Handler()
	articlePage, err := repo.List(context.Background(), 1, 20, "", false)
	if err != nil || len(articlePage.Items) != 1 {
		t.Fatalf("load fixture article: %v, items=%d", err, len(articlePage.Items))
	}
	id := articlePage.Items[0].ID
	request := func(acceptLanguage, path string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if acceptLanguage != "" {
			req.Header.Set("Accept-Language", acceptLanguage)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	type articleEnvelope struct {
		Data news.Article `json:"data"`
	}
	decodeArticle := func(response *httptest.ResponseRecorder) news.Article {
		t.Helper()
		var payload articleEnvelope
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode article: %v: %s", err, response.Body.String())
		}
		if response.Code != http.StatusOK {
			t.Fatalf("news request failed: %d %s", response.Code, response.Body.String())
		}
		return payload.Data
	}

	englishResponse := request("en-US,en;q=0.9", "/v1/news/"+id)
	english := decodeArticle(englishResponse)
	if english.Changes != "Furigana added by Nihongo Master; external media is not included." {
		t.Fatalf("English metadata was not translated: %q", english.Changes)
	}
	if englishResponse.Header().Get("Content-Language") != "en" || !strings.Contains(englishResponse.Header().Get("Vary"), "Accept-Language") || !strings.Contains(englishResponse.Header().Get("Cache-Control"), "no-store") {
		t.Fatalf("localized response headers are incomplete: %#v", englishResponse.Header())
	}
	if english.Title != article.Title || english.Summary != article.Summary || english.DateType != article.DateType || english.License != article.License {
		t.Fatalf("source article text or metadata changed: %#v", english)
	}
	if !reflect.DeepEqual(english.TitleTokens, article.TitleTokens) || len(english.Paragraphs) != 1 || english.Paragraphs[0].Text != article.Paragraphs[0].Text || !reflect.DeepEqual(english.Paragraphs[0].Tokens, article.Paragraphs[0].Tokens) {
		t.Fatalf("Japanese article text and token data must remain intact: %#v", english)
	}

	type errorEnvelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	assertEnglishError := func(response *httptest.ResponseRecorder, status int, code, message string) {
		t.Helper()
		var payload errorEnvelope
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode API error: %v: %s", err, response.Body.String())
		}
		if response.Code != status || payload.Error.Code != code || payload.Error.Message != message || response.Header().Get("Content-Language") != "en" {
			t.Fatalf("English API error mismatch: status=%d payload=%#v headers=%#v", response.Code, payload.Error, response.Header())
		}
	}
	assertEnglishError(request("en", "/v1/news/missing-article"), http.StatusNotFound, "news_not_found", "The article or source could not be found.")
	assertEnglishError(request("en", "/v1/news?limit=500"), http.StatusBadRequest, "invalid_news", "The news request is invalid. Check the settings or article details.")

	var listPayload struct {
		Data struct {
			Items []news.Article `json:"items"`
		} `json:"data"`
	}
	listResponse := request("en", "/v1/news")
	if listResponse.Code != http.StatusOK || json.Unmarshal(listResponse.Body.Bytes(), &listPayload) != nil || len(listPayload.Data.Items) != 1 || listPayload.Data.Items[0].Changes != english.Changes {
		t.Fatalf("news list metadata was not localized: %d %s", listResponse.Code, listResponse.Body.String())
	}

	indonesian := decodeArticle(request("", "/v1/news/"+id))
	if indonesian.Changes != article.Changes {
		t.Fatalf("requests without a locale must keep Indonesian metadata: %q", indonesian.Changes)
	}

	t.Setenv("ENGLISH_QA_PREVIEW_ENABLED", "false")
	closedGateResponse := request("en", "/v1/news/"+id)
	closedGate := decodeArticle(closedGateResponse)
	if closedGate.Changes != article.Changes || closedGateResponse.Header().Get("Content-Language") != "id" {
		t.Fatalf("English news metadata must stay behind the existing release gate: %q (%q)", closedGate.Changes, closedGateResponse.Header().Get("Content-Language"))
	}
}
