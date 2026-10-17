package api_test

import (
	"encoding/json"
	"net/http/httptest"
	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
	"strings"
	"testing"
	"time"
)

func TestSSWAPIAuthEntitlementValidationAndRevocation(t *testing.T) {
	memory := store.NewMemoryStore()
	user := database.DemoUserID
	token, _ := memory.CreateSession(user)
	h := api.NewServer(memory, fakeGoogleVerifier{}).Handler()
	request := func(method, path, body string, auth bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if auth {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/v1/ssw/building-cleaning/practice", "/v1/simulation/ssw-kaigo", "/v1/simulation/ssw-kaigo-jp", "/v1/simulation/ssw-cleaning"} {
		if w := request("GET", path, "", false); w.Code != 401 {
			t.Fatal("guest allowed", path, w.Code)
		}
		if w := request("GET", path, "", true); w.Code != 402 {
			t.Fatal("unpaid allowed", path, w.Code)
		}
	}
	account, _ := memory.BillingAccountToken(user)
	purchase := domain.Purchase{Platform: "android", ID: "ssw-api-test", ProductID: "nihongo_master_lifetime", AccountToken: account, Proof: "ssw-api-test", State: "purchased", CheckedAt: time.Now()}
	if err := memory.SaveVerifiedPurchase(user, purchase); err != nil {
		t.Fatal(err)
	}
	w := request("GET", "/v1/ssw/building-cleaning/practice?module=chemicals", "", true)
	var response struct {
		Data struct {
			Questions []map[string]any `json:"questions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || w.Code != 200 || len(response.Data.Questions) != 6 {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	if strings.Contains(w.Body.String(), `"answer"`) || w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("practice response leak")
	}
	for _, body := range []string{`{"questionId":"cleaning-01-1"}`, `{"questionId":"cleaning-01-1","answer":null}`} {
		if w = request("POST", "/v1/ssw/building-cleaning/practice/check", body, true); w.Code != 400 {
			t.Fatal("missing answer accepted", w.Code)
		}
	}
	if w = request("POST", "/v1/ssw/building-cleaning/practice/check", `{"questionId":"missing","answer":0}`, true); w.Code != 404 {
		t.Fatal(w.Code)
	}
	if w = request("POST", "/v1/ssw/building-cleaning/practice/check", `{"questionId":"cleaning-01-1","answer":0}`, true); w.Code != 200 || !strings.Contains(w.Body.String(), "explanation") {
		t.Fatal(w.Code, w.Body.String())
	}
	w = request("POST", "/v1/exam-attempts", `{"examId":"ssw-cleaning-simulation-1"}`, true)
	var attempt struct {
		Data domain.ExamAttempt `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &attempt)
	if w.Code != 201 || attempt.Data.ID == "" {
		t.Fatal(w.Code, w.Body.String())
	}
	path := "/v1/simulation/attempts/" + attempt.Data.ID
	if w = request("GET", path, "", true); w.Code != 200 || strings.Contains(w.Body.String(), "correctOptionId") {
		t.Fatal(w.Code, w.Body.String())
	}
	purchase.State = "revoked"
	if err := memory.UpdateVerifiedPurchase(purchase); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{path, "/v1/ssw/building-cleaning/practice"} {
		if w = request("GET", path, "", true); w.Code != 402 {
			t.Fatal("revocation bypass", w.Code)
		}
	}
	if w = request("POST", "/v1/exam-attempts", `{"examId":"ssw-kaigo-simulation-1"}`, true); w.Code != 402 {
		t.Fatal("unpaid start", w.Code)
	}
}
