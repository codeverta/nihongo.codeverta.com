package api

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

func bannerRequest(h http.Handler, method, path, token string, body []byte) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestHomeBannersPersistencePublicOrderingAndEmptyCarousel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home-banners.json")
	s := &Server{store: store.NewMemoryStore()}
	if err := s.ConfigureHomeBanners(path, "banner-admin"); err != nil {
		t.Fatal(err)
	}
	h := s.Handler()
	for _, route := range []struct{ method, path string }{
		{"GET", "/v1/admin/home-banners"}, {"PUT", "/v1/admin/home-banners"}, {"POST", "/v1/admin/home-banners/images"},
	} {
		if got := bannerRequest(h, route.method, route.path, "wrong", nil).Code; got != 401 {
			t.Fatalf("%s %s: %d", route.method, route.path, got)
		}
	}
	config := defaultHomeBanners()
	config.Banners[0].Order = 8
	config.Banners[0].Title = "Materi yang diperbarui"
	config.Banners[1].Order = 2
	config.Banners[2].Enabled = false
	data, _ := json.Marshal(config)
	if w := bannerRequest(h, "PUT", "/v1/admin/home-banners", "banner-admin", data); w.Code != 200 {
		t.Fatalf("save: %d %s", w.Code, w.Body.String())
	}
	w := bannerRequest(h, "GET", "/v1/home", "", nil)
	var dashboard struct {
		Data domain.Dashboard `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &dashboard); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || len(dashboard.Data.Banners) != 2 || dashboard.Data.Banners[0].Action != "ssw" || dashboard.Data.Banners[1].Title != "Materi yang diperbarui" {
		t.Fatalf("public carousel: %d %s", w.Code, w.Body.String())
	}
	var admin struct {
		Data homeBannerConfig `json:"data"`
	}
	w = bannerRequest(h, "GET", "/v1/admin/home-banners", "banner-admin", nil)
	if err := json.Unmarshal(w.Body.Bytes(), &admin); err != nil || len(admin.Data.Banners) != 3 {
		t.Fatalf("disabled banner lost: %s %v", w.Body.String(), err)
	}
	restored := &Server{}
	if err := restored.ConfigureHomeBanners(path, "banner-admin"); err != nil {
		t.Fatal(err)
	}
	if got := restored.homeBanners.publicBanners(); len(got) != 2 || got[0].ID != "ssw" {
		t.Fatalf("restart lost state: %+v", got)
	}
	if w := bannerRequest(h, "PUT", "/v1/admin/home-banners", "banner-admin", []byte(`{"banners":[]}`)); w.Code != 200 {
		t.Fatalf("clear: %s", w.Body.String())
	}
	if err := restored.ConfigureHomeBanners(path, "banner-admin"); err != nil {
		t.Fatal(err)
	}
	if got := restored.homeBanners.publicBanners(); got == nil || len(got) != 0 {
		t.Fatalf("empty carousel must persist as []: %+v", got)
	}
}

func TestHomeBannersRejectInvalidEditsWithoutChangingPublishedState(t *testing.T) {
	s := &Server{}
	if err := s.ConfigureHomeBanners(filepath.Join(t.TempDir(), "banners.json"), "key"); err != nil {
		t.Fatal(err)
	}
	h := s.Handler()
	for _, mutate := range []func(*homeBannerConfig){
		func(c *homeBannerConfig) { c.Banners[1].ID = c.Banners[0].ID },
		func(c *homeBannerConfig) { c.Banners[0].ImageURL = "javascript:alert(1)" },
		func(c *homeBannerConfig) { c.Banners[0].ImageURL = "/home-banner-images/../../branding.json" },
		func(c *homeBannerConfig) { c.Banners[0].Action = "/admin/" },
		func(c *homeBannerConfig) { c.Banners[0].Title = "" },
	} {
		c := defaultHomeBanners()
		mutate(&c)
		data, _ := json.Marshal(c)
		if w := bannerRequest(h, "PUT", "/v1/admin/home-banners", "key", data); w.Code != 422 {
			t.Fatalf("accepted bad edit: %d %s", w.Code, w.Body.String())
		}
	}
	if w := bannerRequest(h, "PUT", "/v1/admin/home-banners", "key", []byte(`{}`)); w.Code != 422 {
		t.Fatalf("missing list: %d", w.Code)
	}
	if got := s.homeBanners.publicBanners(); len(got) != 3 || got[0].ID != "learn" {
		t.Fatalf("bad edits changed live config: %+v", got)
	}
	noAdmin := (&Server{}).Handler()
	if w := bannerRequest(noAdmin, "GET", "/v1/admin/home-banners", "key", nil); w.Code != 401 {
		t.Fatalf("unconfigured admin: %d", w.Code)
	}
}

func uploadBannerImage(t *testing.T, h http.Handler, image []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	f, err := mw.CreateFormFile("image", "../../private.env")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(image); err != nil {
		t.Fatal(err)
	}
	if err = mw.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/v1/admin/home-banners/images", &body)
	r.Header.Set("Authorization", "Bearer key")
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestHomeBannerUploadPersistsServesImageAndRejectsNonImages(t *testing.T) {
	path := filepath.Join(t.TempDir(), "banners.json")
	s := &Server{}
	if err := s.ConfigureHomeBanners(path, "key"); err != nil {
		t.Fatal(err)
	}
	h := s.Handler()
	var picture bytes.Buffer
	if err := png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 24, 12))); err != nil {
		t.Fatal(err)
	}
	w := uploadBannerImage(t, h, picture.Bytes())
	if w.Code != 201 {
		t.Fatalf("upload: %d %s", w.Code, w.Body.String())
	}
	var result struct {
		Data struct {
			ImageURL string `json:"imageUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result.Data.ImageURL, "/home-banner-images/") || !strings.HasSuffix(result.Data.ImageURL, ".png") {
		t.Fatal(result.Data.ImageURL)
	}
	config := defaultHomeBanners()
	config.Banners[0].ImageURL = result.Data.ImageURL
	data, _ := json.Marshal(config)
	if w := bannerRequest(h, "PUT", "/v1/admin/home-banners", "key", data); w.Code != 200 {
		t.Fatalf("save upload reference: %s", w.Body.String())
	}
	restored := &Server{}
	if err := restored.ConfigureHomeBanners(path, "key"); err != nil {
		t.Fatal(err)
	}
	w = bannerRequest(restored.Handler(), "GET", result.Data.ImageURL, "", nil)
	if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), picture.Bytes()) || w.Header().Get("Content-Type") != "image/png" || w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("image after restart: %d %v", w.Code, w.Header())
	}
	for _, invalid := range [][]byte{[]byte(`<svg onload="alert(1)"></svg>`), picture.Bytes()[:40]} {
		if w := uploadBannerImage(t, h, invalid); w.Code != 422 {
			t.Fatalf("non-image accepted: %d %s", w.Code, w.Body.String())
		}
	}
	if w := uploadBannerImage(t, h, bytes.Repeat([]byte{'x'}, maxBannerImageBytes+1)); w.Code != 413 {
		t.Fatalf("oversize accepted: %d", w.Code)
	}
	if w := bannerRequest(h, "GET", "/home-banner-images/banners.json", "", nil); w.Code != 404 {
		t.Fatalf("non-image exposed: %d", w.Code)
	}
}

func TestDefaultBannerAssetsAreServed(t *testing.T) {
	s := &Server{}
	if err := s.ConfigureAssets("../../assets"); err != nil {
		t.Fatal(err)
	}
	h := s.Handler()
	for _, imageURL := range defaultBannerImages {
		w := bannerRequest(h, "HEAD", imageURL, "", nil)
		if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" {
			t.Fatalf("missing banner asset %s: %d %v", imageURL, w.Code, w.Header())
		}
	}
}
