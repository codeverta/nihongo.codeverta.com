package api

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnglishSSWRailwayCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "railway" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		key := examSSWEnglishKey{field: "ssw." + field, source: source}
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := sswCurriculumRailwayEnglish[key]; translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := sswCurriculumRailwayEnglish[key]
		if !ok {
			t.Errorf("railway material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		if strings.TrimSpace(translated) == "" || translated == source {
			t.Errorf("railway material %s (%s) has an empty or unchanged translation: %q", id, field, translated)
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("railway material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[key] = true
	}
	if len(seenIDs) != 278 {
		t.Errorf("railway curriculum has %d translated source IDs, want 278", len(seenIDs))
	}
	if len(seenPairs) != 237 {
		t.Errorf("railway curriculum has %d distinct translated source-field pairs, want 237", len(seenPairs))
	}
	if len(sswCurriculumRailwayEnglish) != 237 {
		t.Errorf("railway catalog contains %d entries, want exactly 237", len(sswCurriculumRailwayEnglish))
	}
}
