package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nihongo-master-backend/internal/domain"
)

func TestStudyTimezoneInitializesOnceAndCanBeChanged(t *testing.T) {
	server := newServer()
	googleLogin := func(timezone string) domain.AuthSession {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/v1/auth/google", strings.NewReader(`{"idToken":"valid"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Device-Timezone", timezone)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("Google login failed: %d %s", response.Code, response.Body.String())
		}
		var payload struct {
			Data domain.AuthSession `json:"data"`
		}
		decodeResponse(t, response, &payload)
		return payload.Data
	}

	firstSession := googleLogin("America/Los_Angeles")
	if firstSession.User.Timezone != "America/Los_Angeles" {
		t.Fatalf("new account timezone = %q, want America/Los_Angeles", firstSession.User.Timezone)
	}
	secondSession := googleLogin("Asia/Tokyo")
	if secondSession.User.Timezone != "America/Los_Angeles" {
		t.Fatalf("returning account timezone changed to %q", secondSession.User.Timezone)
	}

	update := httptest.NewRequest(http.MethodPut, "/v1/me/timezone", strings.NewReader(`{"timezone":"Europe/Paris"}`))
	update.Header.Set("Content-Type", "application/json")
	update.Header.Set("Authorization", "Bearer "+firstSession.Token)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, update)
	if response.Code != http.StatusOK {
		t.Fatalf("timezone update failed: %d %s", response.Code, response.Body.String())
	}

	read := httptest.NewRequest(http.MethodGet, "/v1/me/gamification", nil)
	read.Header.Set("Authorization", "Bearer "+firstSession.Token)
	response = httptest.NewRecorder()
	server.ServeHTTP(response, read)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"timezone":"Europe/Paris"`) {
		t.Fatalf("study timezone was not persisted: %d %s", response.Code, response.Body.String())
	}

	invalid := httptest.NewRequest(http.MethodPut, "/v1/me/timezone", strings.NewReader(`{"timezone":"Mars/Olympus"}`))
	invalid.Header.Set("Content-Type", "application/json")
	invalid.Header.Set("Authorization", "Bearer "+firstSession.Token)
	response = httptest.NewRecorder()
	server.ServeHTTP(response, invalid)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid timezone status = %d, want 422: %s", response.Code, response.Body.String())
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatal(err)
	}
}
