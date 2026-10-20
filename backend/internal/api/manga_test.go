package api_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/manga"
	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/testdb"
)

func TestMangaLifetimeCacheAndEditorialRoutes(t *testing.T) {
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	repo, err := manga.New(db, nil)
	if err != nil {
		t.Fatal(err)
	}
	memory := store.NewMemoryStore()
	buyer, err := memory.AuthenticateGoogle(domain.GoogleIdentity{Subject: "manga-buyer", Email: "manga@example.com", DisplayName: "Manga Buyer"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := memory.CreateSession(buyer.ID)
	if err != nil {
		t.Fatal(err)
	}
	accountToken, err := memory.BillingAccountToken(buyer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := memory.SaveVerifiedPurchase(buyer.ID, domain.Purchase{ID: "manga-purchase", Platform: "android", AccountToken: accountToken, State: "purchased", CheckedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	s := api.NewServer(memory, fakeGoogleVerifier{})
	s.ConfigureManga(repo, "test-key")
	handler := s.Handler()
	raw := "https://rawkuma.net/manga/test-story/"
	id := manga.ID(raw)
	series := manga.Series{ID: id, URL: raw, Title: "Test Story", Enabled: true, Reviewed: true, Language: "ja", Status: "ready", Genres: []string{"Sports"}, CoverURL: "https://our-cos.example/manga/cover.png", NextFetchAt: time.Now().Add(manga.RefreshInterval)}
	if err = db.Create(&series).Error; err != nil {
		t.Fatal(err)
	}
	ch := manga.Chapter{ID: manga.ID(raw + "chapter-1.100/"), SeriesID: id, URL: raw + "chapter-1.100/", Title: "Chapter 1", Number: "1", Status: "ready", Pages: []manga.Page{{Index: 1, URL: "https://our-cos.example/manga/page.png", Width: 800, Height: 1200}}}
	if err = db.Create(&ch).Error; err != nil {
		t.Fatal(err)
	}
	request := func(method, path, body, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/v1/manga", "/v1/manga?genre=Sports", "/v1/manga/" + id, "/v1/manga/chapters/" + ch.ID} {
		w := request("GET", path, "", session)
		if w.Code != 200 || !strings.Contains(w.Body.String(), "our-cos.example") {
			t.Fatal(path, w.Code, w.Body.String())
		}
	}
	if w := request("GET", "/v1/manga?limit=1000", "", session); w.Code != 400 {
		t.Fatal("unbounded catalog", w.Code)
	}
	for _, method := range []string{"GET", "POST"} {
		path := "/v1/admin/manga"
		if method == "POST" {
			path += "/series"
		}
		if w := request(method, path, "{}", ""); w.Code != 401 {
			t.Fatal("unprotected admin", w.Code)
		}
	}
	if w := request("POST", "/v1/admin/manga/series", `{"sourceUrl":"https://localhost/private","enabled":true,"japaneseVerified":true,"nonAdultReviewed":true}`, "test-key"); w.Code != 400 {
		t.Fatal("arbitrary URL accepted", w.Code)
	}
	if w := request("POST", "/v1/admin/manga/series", `{"sourceUrl":"`+raw+`","enabled":false}`, "test-key"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request("GET", "/v1/manga/chapters/"+ch.ID, "", session); w.Code != 404 {
		t.Fatal("hidden chapter public", w.Code)
	}
	if _, err := repo.Detail(context.Background(), id); err == nil {
		t.Fatal("hidden detail public")
	}
	isolated := api.NewMangaHandler(repo, "test-key", nil)
	for _, path := range []string{"/v1/auth/me", "/v1/billing", "/v1/profile"} {
		w := httptest.NewRecorder()
		isolated.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatal("standalone service exposed another domain", path, w.Code)
		}
	}
}
