package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/prefinfo"
	"nihongo-master-backend/internal/store"
)

func TestPrefectureLivingIsPublicAndAdminEditable(t *testing.T) {
	memory := store.NewMemoryStore()
	defer func() { _ = memory.Close() }()
	server := api.NewServer(memory, nil)
	server.ConfigureMap(nil, "map-token", nil)
	living, err := prefinfo.Load("")
	if err != nil {
		t.Fatal(err)
	}
	server.ConfigurePrefectureLiving(living)
	handler := server.Handler()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/map/prefecture-living", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("public: %d", response.Code)
	}
	var payload struct{ Data prefinfo.Data }
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil || len(payload.Data.Prefectures) != 47 {
		t.Fatalf("payload: %v %d", err, len(payload.Data.Prefectures))
	}

	put := func(token string, body []byte) int {
		request := httptest.NewRequest(http.MethodPut, "/v1/admin/prefecture-living", strings.NewReader(string(body)))
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Code
	}
	edited := payload.Data
	edited.Prefectures[12].MinimumWage = 1300
	body, _ := json.Marshal(edited)
	if code := put("wrong", body); code != http.StatusUnauthorized {
		t.Fatalf("wrong token: %d", code)
	}
	if code := put("map-token", body); code != http.StatusOK || living.Get().Prefectures[12].MinimumWage != 1300 {
		t.Fatalf("valid edit: %d", code)
	}
	edited.Prefectures = edited.Prefectures[:10]
	bad, _ := json.Marshal(edited)
	if code := put("map-token", bad); code != http.StatusUnprocessableEntity {
		t.Fatalf("incomplete edit: %d", code)
	}
}

func TestPrefectureProfilesAreCachedWithETag(t *testing.T) {
	memory := store.NewMemoryStore()
	defer func() { _ = memory.Close() }()
	handler := api.NewServer(memory, nil).Handler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/map/prefecture-profiles", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status %d", response.Code)
	}
	var payload struct {
		Data struct {
			Prefectures []struct {
				Code string
				Text map[string]*struct{ Intro string }
			}
		}
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil || len(payload.Data.Prefectures) != 47 {
		t.Fatalf("payload: %v %d", err, len(payload.Data.Prefectures))
	}
	if payload.Data.Prefectures[12].Code != "13" || payload.Data.Prefectures[12].Text["id"] == nil {
		t.Fatalf("Tokyo profile missing: %+v", payload.Data.Prefectures[12])
	}
	etag := response.Header().Get("ETag")
	again := httptest.NewRequest(http.MethodGet, "/v1/map/prefecture-profiles", nil)
	again.Header.Set("If-None-Match", etag)
	cached := httptest.NewRecorder()
	handler.ServeHTTP(cached, again)
	if etag == "" || cached.Code != http.StatusNotModified || cached.Body.Len() != 0 {
		t.Fatalf("revalidation: etag=%q status=%d", etag, cached.Code)
	}
}
