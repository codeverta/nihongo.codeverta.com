package api

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestJLPTKanjiN2Lessons19To44MnemonicNoteCoverage(t *testing.T) {
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
	if len(records) == 0 {
		t.Fatal("materials inventory is empty")
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
		if value(row, "surface") != "JLPT learning material" || value(row, "category") != "Kanji" || value(row, "level") != "N2" || (field != "mnemonic" && field != "note") {
			continue
		}
		lessonPart := strings.TrimPrefix(contextID, "kanji:N2:kanji-n2-")
		lesson, err := strconv.Atoi(lessonPart)
		if err != nil || lesson < 19 || lesson > 44 {
			continue
		}
		id, source := value(row, "material_id"), value(row, "source_text")
		if seenIDs[id] {
			t.Errorf("duplicate authoritative N2 material ID %s", id)
		}
		seenIDs[id] = true
		key := jlptKanjiMaterialEnglishKey{contextID: contextID, field: field, source: source}
		translated, ok := jlptKanjiN2Lessons19To44English[key]
		if source == "" || !ok || strings.TrimSpace(translated) == "" {
			t.Errorf("Kanji material %s has no exact contextual %s translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("Kanji material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenKeys[key] = true
	}
	if len(seenIDs) != 260 {
		t.Errorf("covered %d N2 Kanji mnemonic/note source IDs for lessons 19–44, want 260", len(seenIDs))
	}
	if len(seenKeys) != 260 {
		t.Errorf("covered %d exact contextual mnemonic/note sources for lessons 19–44, want 260", len(seenKeys))
	}
	if len(jlptKanjiN2Lessons19To44English) != 260 {
		t.Errorf("catalog has %d contextual mnemonic/note entries, want 260", len(jlptKanjiN2Lessons19To44English))
	}
}
