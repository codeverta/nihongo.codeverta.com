package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func accountLocalizationRequest(t *testing.T, path, language string, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	t.Setenv(englishQAPreviewFlag, "true")
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if language != "" {
		request.Header.Set("Accept-Language", language)
	}
	recorder := httptest.NewRecorder()
	localizedAccountResponses(handler).ServeHTTP(recorder, request)
	return recorder
}

func TestAccountLocalizationTranslatesErrorsForEnglish(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		writeError(w, http.StatusUnauthorized, "unauthorized", "Sesi tidak valid atau sudah berakhir.")
	}
	recorder := accountLocalizationRequest(t, "/v1/me/profile", "en", handler)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", recorder.Code)
	}
	var payload struct {
		Error struct{ Code, Message string }
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != "unauthorized" || payload.Error.Message == "" || strings.Contains(payload.Error.Message, "Sesi") {
		t.Fatalf("error not localized: %+v", payload.Error)
	}
	if recorder.Header().Get("Content-Language") != "en" {
		t.Fatalf("Content-Language = %q", recorder.Header().Get("Content-Language"))
	}
}

func TestAccountLocalizationLeavesIndonesianAndOtherPathsUntouched(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		writeError(w, http.StatusUnauthorized, "unauthorized", "Sesi tidak valid atau sudah berakhir.")
	}
	for _, tc := range []struct{ path, language string }{
		{"/v1/me/profile", "id"},
		{"/v1/me/profile", ""},
		{"/v1/admin/news", "en"},
		{"/v1/dictionary/search", "en"},
	} {
		recorder := accountLocalizationRequest(t, tc.path, tc.language, handler)
		if !strings.Contains(recorder.Body.String(), "Sesi tidak valid") {
			t.Fatalf("%s (%q) changed: %s", tc.path, tc.language, recorder.Body.String())
		}
	}
}

func TestAccountLocalizationStreamsNonJSONUntouched(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write([]byte("binary-deck"))
	}
	recorder := accountLocalizationRequest(t, "/v1/anki/decks/x/export", "en", handler)
	if recorder.Body.String() != "binary-deck" || recorder.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("download altered: %q %q", recorder.Body.String(), recorder.Header().Get("Content-Type"))
	}
}

func TestAppUpdatePolicyUsesEnglishMessage(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		writeData(w, http.StatusOK, AppUpdatePolicy{
			Android: AppUpdatePlatform{Message: "Perbarui aplikasi.", MessageEn: "Please update the app."},
			IOS:     AppUpdatePlatform{Message: "Catatan Indonesia."},
		})
	}
	recorder := accountLocalizationRequest(t, "/v1/app-update-policy", "en", handler)
	var payload struct {
		Data struct {
			Android struct{ Message string }
			IOS     struct{ Message string }
		}
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Android.Message != "Please update the app." || payload.Data.IOS.Message != "" {
		t.Fatalf("unexpected messages: %+v", payload.Data)
	}
	indonesian := accountLocalizationRequest(t, "/v1/app-update-policy", "id", handler)
	if !strings.Contains(indonesian.Body.String(), "Perbarui aplikasi.") {
		t.Fatal("Indonesian policy changed")
	}
}
