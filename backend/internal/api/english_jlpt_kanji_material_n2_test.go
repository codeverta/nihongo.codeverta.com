package api

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJLPTKanjiN2MnemonicNoteCoverage(t *testing.T) {
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
	wantedContexts := map[string]bool{
		"kanji:N2:kanji-n2-01": true,
		"kanji:N2:kanji-n2-02": true,
		"kanji:N2:kanji-n2-03": true,
		"kanji:N2:kanji-n2-04": true,
		"kanji:N2:kanji-n2-05": true,
		"kanji:N2:kanji-n2-06": true,
		"kanji:N2:kanji-n2-07": true,
		"kanji:N2:kanji-n2-08": true,
	}
	seenIDs := make(map[string]bool)
	seenKeys := make(map[jlptKanjiMaterialEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "JLPT learning material" || value(row, "category") != "Kanji" || value(row, "level") != "N2" {
			continue
		}
		contextID := value(row, "context_id")
		if !wantedContexts[contextID] {
			continue
		}
		field := value(row, "field")
		if field != "mnemonic" && field != "note" {
			continue
		}
		id := value(row, "material_id")
		if seenIDs[id] {
			t.Errorf("duplicate authoritative N2 material ID %s", id)
		}
		seenIDs[id] = true
		source := value(row, "source_text")
		key := jlptKanjiMaterialEnglishKey{contextID: contextID, field: field, source: source}
		translation, ok := jlptKanjiN2MnemonicNoteEnglish[key]
		if source == "" || !ok || translation == "" {
			t.Errorf("N2 material %s has no exact contextual %s translation for %q", id, field, source)
			continue
		}
		for _, run := range japaneseStudyRuns(source) {
			if !strings.Contains(translation, run) {
				t.Errorf("translation for %s changed Japanese study text %q", id, run)
			}
		}
		seenKeys[key] = true
	}
	if len(seenIDs) != 80 {
		t.Errorf("covered %d authoritative N2 material IDs in lessons 01–08, want 80", len(seenIDs))
	}
	if len(seenKeys) != 80 {
		t.Errorf("covered %d exact contextual sources in lessons 01–08, want 80", len(seenKeys))
	}
	if len(jlptKanjiN2MnemonicNoteEnglish) != 80 {
		t.Errorf("N2 catalog contains %d translations, want 80", len(jlptKanjiN2MnemonicNoteEnglish))
	}
}
