package api

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestEnglishSSWSourceTitleCatalogACoversAssignedSourceRows(t *testing.T) {
	wantCounts := map[string]int{
		"accommodation":          4,
		"agriculture":            3,
		"automobile-maintenance": 4,
		"aviation":               7,
		"building-cleaning":      4,
		"construction":           4,
		"fishery":                4,
		"food-manufacturing":     6,
		"food-service":           8,
		"forestry":               6,
	}
	seen := make(map[sswSourceTitleEnglishKey]bool)
	rows := 0
	for sector, wantCount := range wantCounts {
		path := filepath.Join("..", "..", "..", "assets", "content", "ssw", sector+".json")
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s source list: %v", sector, err)
		}
		var document struct {
			Sources []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				URL   string `json:"url"`
			} `json:"sources"`
		}
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatalf("decode %s source list: %v", sector, err)
		}
		if len(document.Sources) != wantCount {
			t.Errorf("%s has %d source rows, want %d; review the inventory", sector, len(document.Sources), wantCount)
		}
		for _, source := range document.Sources {
			rows++
			key := sswSourceTitleEnglishKey{sector: sector, source: source.Title}
			if source.Title == "" {
				t.Errorf("%s contains a source row with an empty title", sector)
			}
			parsedURL, err := url.ParseRequestURI(source.URL)
			if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
				t.Errorf("%s source %q has invalid original URL %q", sector, source.Title, source.URL)
			}
			translation, ok := englishSSWSourceTitleCatalogA[key]
			if !ok || translation == "" {
				t.Errorf("%s source title has no exact English mapping: %q (id %q, URL %q)", sector, source.Title, source.ID, source.URL)
				continue
			}
			if seen[key] {
				t.Errorf("duplicate source title in %s: %q", sector, source.Title)
			}
			seen[key] = true
		}
	}
	if rows != 50 {
		t.Errorf("enumerated %d assigned source rows, want 50", rows)
	}
	if len(englishSSWSourceTitleCatalogA) != len(seen) {
		t.Errorf("source inventory uses %d unique catalog keys, catalog contains %d entries", len(seen), len(englishSSWSourceTitleCatalogA))
	}
	for key := range englishSSWSourceTitleCatalogA {
		if !seen[key] {
			t.Errorf("catalog contains a key absent from its assigned source row: sector %q, source %q", key.sector, key.source)
		}
	}
}
