package api_test

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/store"
	"strings"
	"testing"
)

func TestAnkiAPIResourceDetailsAuthAndReviewContract(t *testing.T) {
	memory := store.NewMemoryStore()
	token, _ := memory.CreateSession(database.DemoUserID)
	handler := api.NewServer(memory, fakeGoogleVerifier{}).Handler()
	request := func(method, path, body string, auth bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if auth {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/v1/anki/decks", "/v1/anki/credits", "/v1/anki/decks/wakayama"} {
		if w := request("GET", path, "", false); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := request("GET", "/v1/anki/decks/wakayama/study", "", false); w.Code != 401 {
		t.Fatal("guest study allowed")
	}
	if w := request("GET", "/v1/anki/decks/missing", "", false); w.Code != 404 {
		t.Fatal("unknown deck accepted")
	}
	if w := request("GET", "/v1/anki/decks/n3/study", "", true); w.Code != 402 {
		t.Fatal("premium deck allowed", w.Code)
	}
	if w := request("GET", "/v1/anki/decks/wakayama/study?limit=999", "", true); w.Code != 422 {
		t.Fatal("invalid page accepted")
	}
	if w := request("GET", "/v1/anki/decks/strokes_kanken/study", "", true); w.Code != 422 {
		t.Fatal("placeholder stroke answers exposed")
	}
	w := request("GET", "/v1/anki/decks/wakayama/study?limit=1", "", true)
	var session struct {
		Data struct {
			RequiredTier string `json:"requiredTier"`
			Cards        []struct {
				Index   int `json:"index"`
				Version int `json:"version"`
			} `json:"cards"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil || len(session.Data.Cards) != 1 {
		t.Fatal(err, w.Body.String())
	}
	if session.Data.RequiredTier != "free" {
		t.Fatal("study session lacks authoritative free classification", w.Body.String())
	}
	body := `{"index":0,"version":0,"rating":"good","requestId":"` + uuid.NewString() + `"}`
	for i := 0; i < 2; i++ {
		w = request("POST", "/v1/anki/decks/wakayama/review", body, true)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w = request("POST", "/v1/anki/decks/wakayama/review", strings.Replace(body, "good", "invalid", 1), true); w.Code != 422 {
		t.Fatal("invalid rating accepted")
	}
	if w = request("POST", "/v1/anki/decks/wakayama/review", `{"index":0,"version":0,"rating":"easy","requestId":"`+uuid.NewString()+`"}`, true); w.Code != 409 {
		t.Fatal("stale version accepted", w.Code)
	}
	if w = request("GET", "/v1/anki/decks/wakayama", "", true); !strings.Contains(w.Body.String(), `"seen":1`) {
		t.Fatal("progress missing")
	}
}
