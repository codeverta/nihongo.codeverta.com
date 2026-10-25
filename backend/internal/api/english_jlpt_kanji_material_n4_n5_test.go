package api

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
)

func TestJLPTKanjiN4N5ContextualMaterialCoverage(t *testing.T) {
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

	wantIDs := map[string]bool{
		"Kanji.N5.kanji-n5-04.土.meaning":           true,
		"Kanji.N5.kanji-n5-04.土.example-1.meaning": true,
		"Kanji.N5.kanji-n5-09.足.example-2.meaning": true,
		"Kanji.N5.kanji-n5-17.食.example-1.meaning": true,
		"Kanji.N4.kanji-n4-03.目.example-1.meaning": true,
		"Kanji.N4.kanji-n4-04.足.example-2.meaning": true,
		"Kanji.N4.kanji-n4-04.歩.meaning":           true,
		"Kanji.N4.kanji-n4-04.歩.example-1.meaning": true,
		"Kanji.N4.kanji-n4-04.歩.example-2.meaning": true,
		"Kanji.N4.kanji-n4-10.仕.example-1.meaning": true,
		"Kanji.N4.kanji-n4-10.事.example-1.meaning": true,
		"Kanji.N4.kanji-n4-21.不.example-2.meaning": true,
		"Kanji.N4.kanji-n4-23.重.example-1.meaning": true,
		"Kanji.N4.kanji-n4-25.死.example-2.meaning": true,
		"Kanji.N4.kanji-n4-28.持.example-2.meaning": true,
		"Kanji.N4.kanji-n4-29.服.meaning":           true,
		"Kanji.N4.kanji-n4-29.服.example-1.meaning": true,
		"Kanji.N4.kanji-n4-30.広.meaning":           true,
		"Kanji.N4.kanji-n4-30.広.example-1.meaning": true,
		"Kanji.N4.kanji-n4-31.終.example-1.meaning": true,
	}
	seenIDs := make(map[string]bool, len(wantIDs))
	seenKeys := make(map[jlptKanjiMaterialEnglishKey]bool)
	for _, row := range records[1:] {
		id := value(row, "material_id")
		if !wantIDs[id] {
			continue
		}
		if value(row, "surface") != "JLPT learning material" || value(row, "category") != "Kanji" || value(row, "field") != "meaning" {
			t.Errorf("unexpected source row for %s: surface=%q category=%q field=%q", id, value(row, "surface"), value(row, "category"), value(row, "field"))
			continue
		}
		key := jlptKanjiMaterialEnglishKey{
			contextID: value(row, "context_id"),
			field:     value(row, "field"),
			source:    value(row, "source_text"),
		}
		if translated, ok := jlptKanjiMaterialEnglishCatalog[key]; !ok || translated == "" {
			t.Errorf("Kanji material %s has no exact contextual translation for %q", id, key.source)
		}
		seenIDs[id] = true
		seenKeys[key] = true
	}
	for id := range wantIDs {
		if !seenIDs[id] {
			t.Errorf("authoritative Kanji source ID %s is missing from materials.csv", id)
		}
	}
	if len(seenIDs) != 20 {
		t.Errorf("covered %d N4/N5 Kanji source IDs, want 20", len(seenIDs))
	}
	if len(seenKeys) != 15 {
		t.Errorf("covered %d context-scoped meaning sources, want 15", len(seenKeys))
	}
	if len(jlptKanjiMaterialEnglishCatalog) != 15 {
		t.Errorf("catalog has %d contextual meanings, want 15", len(jlptKanjiMaterialEnglishCatalog))
	}
}
