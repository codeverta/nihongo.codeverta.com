package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnglishSSWSourceTitleCatalogBCoversAssignedSources(t *testing.T) {
	sectors := []string{
		"industrial-manufacturing", "kaigo", "linen-supply", "logistics-warehouse",
		"railway", "resource-recycling", "road-transport", "shipbuilding", "wood-industry",
	}
	type source struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	type curriculum struct {
		Sources []source `json:"sources"`
	}
	seen := make(map[sswSourceTitleEnglishKey]bool)
	rowCount := 0
	for _, sector := range sectors {
		path := filepath.Join("..", "..", "..", "assets", "content", "ssw", sector+".json")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var content curriculum
		if err := json.Unmarshal(data, &content); err != nil {
			t.Fatalf("decode %s: %v", sector, err)
		}
		for index, entry := range content.Sources {
			rowCount++
			key := sswSourceTitleEnglishKey{sector: sector, source: entry.Title}
			translated, ok := englishSSWSourceTitleCatalogB[key]
			if !ok {
				t.Errorf("%s source %d (%s) has no exact title translation: %q", sector, index, entry.URL, entry.Title)
				continue
			}
			seen[key] = true
			if strings.TrimSpace(translated) == "" || translated == entry.Title {
				t.Errorf("%s source %d has empty or unchanged translation %q", sector, index, translated)
			}
			parsedURL, err := url.ParseRequestURI(entry.URL)
			if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
				t.Errorf("%s source %d has invalid source URL %q", sector, index, entry.URL)
			}
			// Keep Japanese organization names in official source titles intact.
			if strings.ContainsAny(entry.Title, "国土交通省出入国在留管理庁一般社団法人全国木材組合連合会") {
				organization := strings.SplitN(entry.Title, " — ", 2)[0]
				if strings.ContainsAny(organization, "国土交通省出入国在留管理庁一般社団法人全国木材組合連合会") && !strings.Contains(translated, organization) {
					t.Errorf("%s source %d translation dropped official organization name %q: %q", sector, index, organization, translated)
				}
			}
		}
	}
	if rowCount != 40 {
		t.Errorf("assigned source lists contain %d rows, want 40", rowCount)
	}
	if len(seen) != 40 {
		t.Errorf("assigned source lists contain %d distinct sector/title keys, want 40", len(seen))
	}
	if len(englishSSWSourceTitleCatalogB) != len(seen) {
		t.Errorf("catalog has %d entries but inventory uses %d keys; it contains missing or out-of-scope entries", len(englishSSWSourceTitleCatalogB), len(seen))
	}
	for key := range englishSSWSourceTitleCatalogB {
		if !seen[key] {
			t.Errorf("catalog contains out-of-scope source title for %s: %q", key.sector, key.source)
		}
	}
}

func Example_englishSSWSourceTitleCatalogB() {
	key := sswSourceTitleEnglishKey{sector: "railway", source: `国土交通省 — 鉄道分野の業務区分（業務内容の詳細）`}
	fmt.Println(englishSSWSourceTitleCatalogB[key])
	// Output: 国土交通省 — Railway Sector Job Categories (Detailed Duties)
}
