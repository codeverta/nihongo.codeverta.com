package api

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJLPTKanjiN3Lessons39To77MaterialCoverage(t *testing.T) {
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
	seenKeys := make(map[jlptKanjiMaterialEnglishKey]bool)
	for _, row := range records[1:] {
		contextID, field := value(row, "context_id"), value(row, "field")
		if !strings.HasPrefix(contextID, "kanji:N3:kanji-n3-") || (field != "meaning" && field != "mnemonic" && field != "note") {
			continue
		}
		lesson := strings.TrimPrefix(contextID, "kanji:N3:kanji-n3-")
		if len(lesson) != 2 || lesson < "39" || lesson > "77" {
			continue
		}
		if value(row, "surface") != "JLPT learning material" || value(row, "category") != "Kanji" || value(row, "level") != "N3" {
			t.Errorf("unexpected inventory row %s: surface=%q category=%q level=%q", value(row, "material_id"), value(row, "surface"), value(row, "category"), value(row, "level"))
			continue
		}
		key := jlptKanjiMaterialEnglishKey{contextID: contextID, field: field, source: value(row, "source_text")}
		translated, ok := jlptKanjiN3Lessons39To77English[key]
		if !ok {
			continue
		}
		if strings.TrimSpace(translated) == "" {
			t.Errorf("Kanji material %s has no exact contextual %s translation for %q", value(row, "material_id"), field, key.source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(key.source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("Kanji material %s translation dropped Japanese study text %q: %q", value(row, "material_id"), japaneseRun, translated)
			}
		}
		seenIDs[value(row, "material_id")] = true
		seenKeys[key] = true
	}
	if len(seenIDs) != 878 {
		t.Errorf("covered %d N3 Kanji material source IDs, want 878", len(seenIDs))
	}
	if len(seenKeys) != 849 {
		t.Errorf("covered %d context-scoped source keys, want 849", len(seenKeys))
	}
	if len(jlptKanjiN3Lessons39To77English) != 849 {
		t.Errorf("catalog has %d contextual entries, want 849", len(jlptKanjiN3Lessons39To77English))
	}
}
