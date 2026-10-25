package api

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestJLPTKanjiN3QuizContextualCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "questions.csv")
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

	seenQuestionIDs := make(map[string]bool)
	seenCatalogKeys := make(map[jlptKanjiMaterialEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "JLPT lesson quiz" || value(row, "category") != "Kanji" || value(row, "level") != "N3" {
			continue
		}
		lessonID := value(row, "lesson_id")
		var lessonNumber int
		if _, err := fmt.Sscanf(lessonID, "kanji-n3-%d", &lessonNumber); err != nil || lessonNumber < 38 || lessonNumber > 77 {
			continue
		}
		questionID := value(row, "question_id")
		if seenQuestionIDs[questionID] {
			t.Errorf("duplicate authoritative Kanji N3 quiz ID %s", questionID)
		}
		seenQuestionIDs[questionID] = true

		optionCount, err := strconv.Atoi(value(row, "option_count"))
		if err != nil || optionCount < 2 {
			t.Errorf("question %s has invalid option_count %q", questionID, value(row, "option_count"))
		}
		correctIndex, err := strconv.Atoi(value(row, "correct_index"))
		if err != nil || correctIndex < 0 || correctIndex >= optionCount {
			t.Errorf("question %s has invalid correct_index %q for %d options", questionID, value(row, "correct_index"), optionCount)
		}

		contextID := "kanji:N3:" + lessonID
		for _, field := range []string{"prompt", "explanation"} {
			source := value(row, field+"_source")
			status := value(row, field+"_status")
			if source == "" {
				t.Errorf("question %s has an empty %s source", questionID, field)
				continue
			}
			key := jlptKanjiMaterialEnglishKey{contextID: contextID, field: field, source: source}
			translation, ok := jlptKanjiN3QuizEnglishCatalog[key]
			if status == "Needs translation" && !ok {
				t.Errorf("question %s has no exact contextual %s translation for %q", questionID, field, source)
				continue
			}
			if ok {
				if translation == "" {
					t.Errorf("question %s has an empty contextual %s translation for %q", questionID, field, source)
					continue
				}
				seenCatalogKeys[key] = true
			}
		}
	}

	if len(seenQuestionIDs) != 384 {
		t.Errorf("covered %d authoritative N3 quiz IDs in lessons 38–77, want 384", len(seenQuestionIDs))
	}
	if len(seenCatalogKeys) != len(jlptKanjiN3QuizEnglishCatalog) {
		t.Errorf("CSV references %d catalog keys, catalog contains %d", len(seenCatalogKeys), len(jlptKanjiN3QuizEnglishCatalog))
	}
	if len(jlptKanjiN3QuizEnglishCatalog) != 760 {
		t.Errorf("N3 quiz catalog contains %d contextual prompt/explanation translations, want 760", len(jlptKanjiN3QuizEnglishCatalog))
	}
}
