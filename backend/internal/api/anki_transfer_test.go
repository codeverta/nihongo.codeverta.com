package api_test

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"mime/multipart"
	"net/http/httptest"
	"nihongo-master-backend/internal/ankitransfer"
	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/store"
	"testing"
)

func TestAnkiTransferAPIImportExportAndIsolation(t *testing.T) {
	memory := store.NewMemoryStore()
	token, _ := memory.CreateSession(database.DemoUserID)
	handler := api.NewServer(memory, fakeGoogleVerifier{}).Handler()
	doc := ankitransfer.Document{Version: 1, Title: "日本語", Cards: []ankitransfer.Card{{Front: "猫", Answers: []string{"ねこ"}}}}
	raw, _ := json.Marshal(doc)
	upload := func(path string, fields map[string]string, authenticated bool) *httptest.ResponseRecorder {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		file, _ := writer.CreateFormFile("file", "deck.json")
		file.Write(raw)
		for k, v := range fields {
			writer.WriteField(k, v)
		}
		writer.Close()
		req := httptest.NewRequest("POST", path, &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		if authenticated {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	request := func(method, path string, authenticated bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		if authenticated {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	if w := upload("/v1/anki/import/preview", nil, false); w.Code != 401 {
		t.Fatal(w.Code)
	}
	w := upload("/v1/anki/import/preview", nil, true)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var preview struct {
		Data struct {
			SHA256    string `json:"sha256"`
			CardCount int    `json:"cardCount"`
		}
	}
	json.Unmarshal(w.Body.Bytes(), &preview)
	if preview.Data.CardCount != 1 || len(preview.Data.SHA256) != 64 {
		t.Fatal(w.Body.String())
	}
	if w := upload("/v1/anki/import", map[string]string{"sha256": "wrong", "requestId": uuid.NewString()}, true); w.Code != 409 {
		t.Fatal("changed file accepted", w.Code)
	}
	fields := map[string]string{"sha256": preview.Data.SHA256, "requestId": uuid.NewString(), "title": "My deck"}
	var id string
	for i := 0; i < 2; i++ {
		w = upload("/v1/anki/import", fields, true)
		if w.Code != 201 {
			t.Fatal(w.Code, w.Body.String())
		}
		var result struct {
			Data struct {
				ID string `json:"id"`
			}
		}
		json.Unmarshal(w.Body.Bytes(), &result)
		if i == 0 {
			id = result.Data.ID
		}
		if id == "" || id != result.Data.ID {
			t.Fatal("retry duplicated deck")
		}
	}
	if w := request("GET", "/v1/anki/decks/"+id+"/export", false); w.Code != 401 {
		t.Fatal("guest exported private deck")
	}
	for _, format := range []string{"apkg", "json", "csv"} {
		w = request("GET", "/v1/anki/decks/"+id+"/export?format="+format, true)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		exported, err := ankitransfer.Parse("deck."+format, w.Body.Bytes())
		if err != nil || exported.Title != "My deck" || len(exported.Cards) != 1 {
			t.Fatal("bad export", err)
		}
	}
	w = request("GET", "/v1/anki/decks/wakayama/export?format=json", true)
	if w.Code != 200 {
		t.Fatal("built-in export failed", w.Code, w.Body.String())
	}
	exported, err := ankitransfer.Parse("deck.json", w.Body.Bytes())
	if err != nil || len(exported.Cards) <= 20 || exported.SourceID != "wakayama" {
		t.Fatal("catalog export truncated or lost attribution", err)
	}
	if w := request("GET", "/v1/anki/decks/n3/export", true); w.Code != 402 {
		t.Fatal("premium export not gated", w.Code)
	}
	if w := request("DELETE", "/v1/anki/decks/"+id, true); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request("GET", "/v1/anki/decks/"+id, true); w.Code != 404 {
		t.Fatal("deleted deck still visible", w.Code)
	}
}
