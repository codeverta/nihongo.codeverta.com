package api

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJLPTKanjiN2Lessons09To18MnemonicNoteCoverage(t *testing.T) {
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
		if !strings.HasPrefix(contextID, "kanji:N2:kanji-n2-") || (field != "mnemonic" && field != "note") {
			continue
		}
		lesson := strings.TrimPrefix(contextID, "kanji:N2:kanji-n2-")
		if len(lesson) != 2 || lesson < "09" || lesson > "18" {
			continue
		}
		if value(row, "surface") != "JLPT learning material" || value(row, "category") != "Kanji" || value(row, "level") != "N2" {
			t.Errorf("unexpected inventory row %s: surface=%q category=%q level=%q", value(row, "material_id"), value(row, "surface"), value(row, "category"), value(row, "level"))
			continue
		}
		key := jlptKanjiMaterialEnglishKey{contextID: contextID, field: field, source: value(row, "source_text")}
		translated, ok := jlptKanjiN2Lessons09To18English[key]
		if !ok || strings.TrimSpace(translated) == "" {
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
	if len(seenIDs) != 100 {
		t.Errorf("covered %d N2 Kanji mnemonic/note source IDs, want 100", len(seenIDs))
	}
	if len(seenKeys) != 100 {
		t.Errorf("covered %d context-scoped mnemonic/note sources, want 100", len(seenKeys))
	}
	if len(jlptKanjiN2Lessons09To18English) != 100 {
		t.Errorf("catalog has %d contextual mnemonic/note entries, want 100", len(jlptKanjiN2Lessons09To18English))
	}
}
