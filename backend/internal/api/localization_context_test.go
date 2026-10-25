package api

import (
	"encoding/json"
	"testing"
)

func TestLearningJSONUsesKanjiAndReadingGlossariesByRoute(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		field  string
		source string
		want   string
	}{
		{name: "Kanji choice", path: "/v1/kanji/lessons/kanji-n3-15", field: "text", source: "pemandangan", want: "scenery; view"},
		{name: "Kanji scoped meaning", path: "/v1/kanji/lessons/kanji-n4-23", field: "meaning", source: "berat", want: "heavy"},
		{name: "Kanji N3 scoped prompt", path: "/v1/kanji/lessons/kanji-n3-38", field: "prompt", source: "Apa arti utama kanji 報 pada kartu ini?", want: "What is the main meaning of 報 on this card?"},
		{name: "Kanji N3 early material", path: "/v1/kanji/lessons/kanji-n3-01", field: "mnemonic", source: "Hubungkan 王 dengan 王様 (raja). Bayangkan adegan: Dalam cerita muncul seorang raja yang baik hati.", want: "Connect 王 with 王様 (king). Picture this scene: A kind-hearted king appears in the story."},
		{name: "Kanji N3 later material", path: "/v1/kanji/lessons/kanji-n3-39", field: "meaning", source: "lupa", want: "to forget"},
		{name: "Kanji N2 scoped mnemonic", path: "/v1/kanji/lessons/kanji-n2-09", field: "mnemonic", source: "Pelajari 準 melalui kata nyata 基準 (standard; basis; criterion; norm; reference; datum).", want: "Learn 準 through the real word 基準 (standard; basis; criterion; norm; reference; datum)."},
		{name: "Kanji N2 later material", path: "/v1/kanji/lessons/kanji-n2-45", field: "mnemonic", source: "Pelajari 群 melalui kata nyata 群 (group (math)).", want: "Learn 群 through the real word 群 (group (math))."},
		{name: "Kanji N2 middle material", path: "/v1/kanji/lessons/kanji-n2-39", field: "mnemonic", source: "Pelajari 籍 melalui kata nyata 国籍 (nationality).", want: "Learn 籍 through the real word 国籍 (nationality)."},
		{name: "Reading meaning", path: "/v1/reading/lessons/reading-n5-12", field: "meaning", source: "merapikan", want: "tidy up"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localized, err := translateLearningJSONForPath([]byte(`{"data":{"`+tt.field+`":"`+tt.source+`"}}`), tt.path)
			if err != nil {
				t.Fatal(err)
			}
			var payload struct {
				Data struct {
					Meaning  string `json:"meaning"`
					Text     string `json:"text"`
					Prompt   string `json:"prompt"`
					Mnemonic string `json:"mnemonic"`
				} `json:"data"`
			}
			if err := json.Unmarshal(localized, &payload); err != nil {
				t.Fatal(err)
			}
			got := payload.Data.Meaning
			if tt.field == "text" {
				got = payload.Data.Text
			} else if tt.field == "prompt" {
				got = payload.Data.Prompt
			} else if tt.field == "mnemonic" {
				got = payload.Data.Mnemonic
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
