package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/jobs"
	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/testdb"
)

func TestJobsGuestCatalogAndErrors(t *testing.T) {
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	repo, err := jobs.New(db, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	server := api.NewServer(store.NewMemoryStore(), fakeGoogleVerifier{})
	server.ConfigureJobs(repo)
	handler := server.Handler()
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/v1/jobs/meta", 200}, {"/v1/jobs/sources", 200}, {"/v1/jobs", 200},
		{"/v1/jobs?salaryMin=NaN&salaryPeriod=year", 400},
		{"/v1/jobs?limit=10000", 400}, {"/v1/jobs?category=invalid", 400},
		{"/v1/jobs/not-a-uuid", 404},
		{"/v1/jobs/123e4567-e89b-42d3-a456-426614174000", 404},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != tc.status {
				t.Fatalf("%d: %s", response.Code, response.Body)
			}
			var envelope map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if tc.status == 200 && envelope["data"] == nil {
				t.Fatal("missing data envelope")
			}
		})
	}
}

func TestJobsUnconfiguredReturnsServiceUnavailable(t *testing.T) {
	handler := newServer()
	for _, path := range []string{"/v1/jobs", "/v1/jobs/sources", "/v1/jobs/123e4567-e89b-42d3-a456-426614174000"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s: %d", path, response.Code)
		}
	}
}
