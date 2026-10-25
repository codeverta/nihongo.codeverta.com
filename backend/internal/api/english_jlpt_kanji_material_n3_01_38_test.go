package api

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJLPTKanjiN3Lessons01To38MaterialCoverage(t *testing.T) {
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
	translatedIDs := make(map[string]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "JLPT learning material" || value(row, "category") != "Kanji" || value(row, "level") != "N3" {
			continue
		}
		contextID := value(row, "context_id")
		if !strings.HasPrefix(contextID, "kanji:N3:kanji-n3-") {
			continue
		}
		lessonID := strings.TrimPrefix(contextID, "kanji:N3:kanji-n3-")
		var lessonNumber int
		if _, err := fmt.Sscanf(lessonID, "%d", &lessonNumber); err != nil || lessonNumber < 1 || lessonNumber > 38 {
			continue
		}
		field := value(row, "field")
		if field != "meaning" && field != "mnemonic" && field != "note" {
			continue
		}
		id := value(row, "material_id")
		if seenIDs[id] {
			t.Errorf("duplicate authoritative N3 material ID %s", id)
		}
		seenIDs[id] = true
		key := jlptKanjiMaterialEnglishKey{contextID: contextID, field: field, source: value(row, "source_text")}
		translation, ok := jlptKanjiN3EarlyMaterialEnglish[key]
		if value(row, "status") == "Needs translation" && (!ok || translation == "") {
			t.Errorf("N3 material %s has no exact contextual %s translation for %q", id, field, key.source)
			continue
		}
		if ok {
			if value(row, "source_text") == "" || translation == "" {
				t.Errorf("N3 material %s has an empty source or translation", id)
			}
			for _, run := range japaneseStudyRuns(key.source) {
				if !strings.Contains(translation, run) {
					t.Errorf("translation for %s changed Japanese study text %q", id, run)
				}
			}
			seenKeys[key] = true
			translatedIDs[id] = true
		}
	}
	if len(seenIDs) != 895 {
		t.Errorf("enumerated %d authoritative N3 material IDs for lessons 01–38, want 895", len(seenIDs))
	}
	if len(translatedIDs) != 823 {
		t.Errorf("catalog covers %d N3 material rows in lessons 01–38, want 823", len(translatedIDs))
	}
	if len(jlptKanjiN3EarlyMaterialEnglish) != 782 {
		t.Errorf("N3 early material catalog contains %d contextual translations, want 782", len(jlptKanjiN3EarlyMaterialEnglish))
	}
	if len(seenKeys) != len(jlptKanjiN3EarlyMaterialEnglish) {
		t.Errorf("materials.csv references %d catalog keys, catalog contains %d", len(seenKeys), len(jlptKanjiN3EarlyMaterialEnglish))
	}
}
