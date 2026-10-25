package api

import (
	"encoding/json"
	"testing"
)

func TestUnknownEnglishNewsErrorUsesGenericMessage(t *testing.T) {
	localized, changed, err := translateNewsErrorJSON([]byte(`{"error":{"code":"new_error_code","message":"Pesan berbahasa Indonesia"}}`))
	if err != nil || !changed {
		t.Fatalf("translate unknown error: changed=%v err=%v", changed, err)
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(localized, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != "new_error_code" || payload.Error.Message != "The news request could not be completed." {
		t.Fatalf("unknown English errors must use the generic message: %#v", payload.Error)
	}
}
