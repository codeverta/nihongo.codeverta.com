package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type coverageQuestion struct {
	CanonicalID      string   `json:"canonicalId"`
	Surface          string   `json:"surface"`
	Category         string   `json:"category"`
	Level            string   `json:"level"`
	Sector           string   `json:"sector"`
	Form             string   `json:"form"`
	ModuleID         string   `json:"moduleId"`
	ModuleTitle      string   `json:"moduleTitle"`
	LessonID         string   `json:"lessonId"`
	LessonTitle      string   `json:"lessonTitle"`
	QuestionID       string   `json:"questionId"`
	Ordinal          int      `json:"ordinal"`
	Kind             string   `json:"kind"`
	ImageDescription string   `json:"imageDescription"`
	Section          string   `json:"section"`
	Group            string   `json:"group"`
	Instructions     string   `json:"instructions"`
	Passage          string   `json:"passage"`
	Prompt           string   `json:"prompt"`
	Options          []string `json:"options"`
	Correct          int      `json:"correctIndex"`
	Explanation      string   `json:"explanation"`
	SourcePath       string   `json:"sourcePath"`
}

type coverageOccurrence struct {
	CanonicalID string `json:"canonicalId"`
	Surface     string `json:"surface"`
	Category    string `json:"category"`
	Level       string `json:"level"`
	Sector      string `json:"sector"`
	Form        string `json:"form"`
	ModuleID    string `json:"moduleId"`
	LessonID    string `json:"lessonId"`
	QuestionID  string `json:"questionId"`
	Ordinal     int    `json:"ordinal"`
	Section     string `json:"section"`
	Group       string `json:"group"`
	SourcePath  string `json:"sourcePath"`
}

type coverageText struct {
	ContextID   string   `json:"contextId"`
	Surface     string   `json:"surface"`
	Category    string   `json:"category"`
	Level       string   `json:"level"`
	Sector      string   `json:"sector"`
	Form        string   `json:"form"`
	ModuleID    string   `json:"moduleId"`
	LessonID    string   `json:"lessonId"`
	Field       string   `json:"field"`
	SourceText  string   `json:"sourceText"`
	SourcePath  string   `json:"sourcePath"`
	QuestionIDs []string `json:"questionIds"`
}

type coverageMaterial struct {
	ID             string   `json:"id"`
	ContextID      string   `json:"contextId"`
	Surface        string   `json:"surface"`
	Category       string   `json:"category"`
	Level          string   `json:"level"`
	Sector         string   `json:"sector"`
	ModuleID       string   `json:"moduleId"`
	LessonID       string   `json:"lessonId"`
	ItemID         string   `json:"itemId"`
	ItemType       string   `json:"itemType"`
	Field          string   `json:"field"`
	SourceText     string   `json:"sourceText"`
	SourcePath     string   `json:"sourcePath"`
	Preserve       bool     `json:"preserveJapanese"`
	AlreadyEnglish bool     `json:"alreadyEnglish"`
	QuestionIDs    []string `json:"questionIds"`
}

type coverageBundle struct {
	Questions   []coverageQuestion   `json:"questions"`
	Occurrences []coverageOccurrence `json:"occurrences"`
	SharedText  []coverageText       `json:"sharedText"`
	Materials   []coverageMaterial   `json:"materials"`
}

type curriculumAsset struct {
	Sources []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"sources"`
	Modules []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Lessons     []struct {
			ID         string   `json:"id"`
			Title      string   `json:"title"`
			Objectives []string `json:"objectives"`
			Sections   []struct {
				Title string `json:"title"`
				Body  string `json:"body"`
			} `json:"sections"`
			Steps      []string `json:"steps"`
			Caution    string   `json:"caution"`
			Scenario   string   `json:"scenario"`
			Response   string   `json:"response"`
			Vocabulary []struct {
				Japanese string `json:"japanese"`
				Reading  string `json:"reading"`
				Meaning  string `json:"meaning"`
			} `json:"vocabulary"`
			Dialogue []struct {
				Japanese string `json:"japanese"`
				Reading  string `json:"reading"`
				Meaning  string `json:"meaning"`
			} `json:"dialogue"`
			Checks []struct {
				Kind        string   `json:"kind"`
				Prompt      string   `json:"prompt"`
				Options     []string `json:"options"`
				Answer      int      `json:"answer"`
				Explanation string   `json:"explanation"`
			} `json:"checks"`
		} `json:"lessons"`
	} `json:"modules"`
}

func sourceFingerprint(prompt, passage string, options []string, answer int, explanation string) string {
	raw, _ := json.Marshal([]any{prompt, passage, options, answer, explanation})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func TestExportLocalizationCoverage(t *testing.T) {
	output := os.Getenv("COVERAGE_EXPORT_PATH")
	if output == "" {
		t.Skip("set COVERAGE_EXPORT_PATH to create the temporary inventory extract")
	}
	bundle := coverageBundle{}
	questionIndex := map[string]int{}
	bankByFingerprint := map[string]string{}
	levelMaterialQuestions := map[string][]string{}
	addQuestion := func(q coverageQuestion, occurrence coverageOccurrence) {
		if q.Options == nil {
			q.Options = []string{}
		}
		if _, ok := questionIndex[q.CanonicalID]; !ok {
			questionIndex[q.CanonicalID] = len(bundle.Questions)
			bundle.Questions = append(bundle.Questions, q)
		}
		bundle.Occurrences = append(bundle.Occurrences, occurrence)
	}
	addText := func(contextID, surface, category, level, sector, form, moduleID, lessonID, field, value, path string, ids []string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		bundle.SharedText = append(bundle.SharedText, coverageText{
			ContextID: contextID, Surface: surface, Category: category, Level: level,
			Sector: sector, Form: form, ModuleID: moduleID, LessonID: lessonID,
			Field: field, SourceText: value, SourcePath: path, QuestionIDs: ids,
		})
	}
	addMaterial := func(material coverageMaterial) {
		if strings.TrimSpace(material.SourceText) == "" {
			return
		}
		if material.QuestionIDs == nil {
			material.QuestionIDs = []string{}
		}
		bundle.Materials = append(bundle.Materials, material)
	}

	store := NewMemoryStore()
	for _, level := range store.vocabulary {
		levelPath := "backend/data/vocabulary/" + strings.ToLower(level.Level) + ".json"
		addMaterial(coverageMaterial{ID: "Vocabulary." + level.Level + ".level.title", ContextID: "vocabulary:" + level.Level, Surface: "JLPT learning material", Category: "Vocabulary", Level: level.Level, ItemID: level.Level, ItemType: "Level", Field: "title", SourceText: level.Title, SourcePath: levelPath, QuestionIDs: []string{}})
		addMaterial(coverageMaterial{ID: "Vocabulary." + level.Level + ".level.description", ContextID: "vocabulary:" + level.Level, Surface: "JLPT learning material", Category: "Vocabulary", Level: level.Level, ItemID: level.Level, ItemType: "Level", Field: "description", SourceText: level.Description, SourcePath: levelPath, QuestionIDs: []string{}})
		addMaterial(coverageMaterial{ID: "Vocabulary." + level.Level + ".level.sourceName", ContextID: "vocabulary:" + level.Level, Surface: "JLPT learning material", Category: "Vocabulary", Level: level.Level, ItemID: level.Level, ItemType: "Level", Field: "sourceName", SourceText: level.SourceName, SourcePath: levelPath, QuestionIDs: []string{}})
		addMaterial(coverageMaterial{ID: "Vocabulary." + level.Level + ".level.sourceLicense", ContextID: "vocabulary:" + level.Level, Surface: "JLPT learning material", Category: "Vocabulary", Level: level.Level, ItemID: level.Level, ItemType: "Level", Field: "sourceLicense", SourceText: level.SourceLicense, SourcePath: levelPath, QuestionIDs: []string{}})
		for _, lesson := range level.Lessons {
			attempt, err := store.StartVocabularyAttempt("coverage", lesson.ID)
			if err != nil {
				t.Fatalf("vocabulary attempt %s: %v", lesson.ID, err)
			}
			state := store.vocabAttempts[attempt.ID]
			path := "backend/data/vocabulary/" + strings.ToLower(level.Level) + ".json"
			contextID := "vocabulary:" + level.Level + ":" + lesson.ID
			ids := make([]string, 0, len(attempt.Questions))
			for i, q := range attempt.Questions {
				correctID := state.correct[q.ID]
				correct := -1
				options := make([]string, 0, len(q.Options))
				for oi, option := range q.Options {
					options = append(options, option.Text)
					if option.ID == correctID {
						correct = oi
					}
				}
				canonical := "JLPT.Vocabulary." + level.Level + "." + lesson.ID + "." + q.ID
				ids = append(ids, canonical)
				levelMaterialQuestions["Vocabulary:"+level.Level] = append(levelMaterialQuestions["Vocabulary:"+level.Level], canonical)
				question := coverageQuestion{CanonicalID: canonical, Surface: "JLPT lesson quiz", Category: "Vocabulary", Level: level.Level, LessonID: lesson.ID, LessonTitle: lesson.Title, QuestionID: q.ID, Ordinal: i + 1, Kind: q.Kind, ImageDescription: q.ImageDescription, Prompt: q.Prompt, Options: options, Correct: correct, Explanation: state.explanations[q.ID], SourcePath: path}
				occurrence := coverageOccurrence{CanonicalID: canonical, Surface: question.Surface, Category: question.Category, Level: level.Level, LessonID: lesson.ID, QuestionID: q.ID, Ordinal: i + 1, SourcePath: path}
				addQuestion(question, occurrence)
			}
			addText(contextID, "JLPT lesson quiz", "Vocabulary", level.Level, "", "", "", lesson.ID, "Lesson title", lesson.Title, path, ids)
			addText(contextID, "JLPT lesson quiz", "Vocabulary", level.Level, "", "", "", lesson.ID, "Lesson description", lesson.Description, path, ids)
			addWordMaterial := func(itemID, itemType, field, source string, preserve bool) {
				addMaterial(coverageMaterial{ID: "Vocabulary." + level.Level + "." + lesson.ID + "." + itemID + "." + field, ContextID: contextID, Surface: "JLPT learning material", Category: "Vocabulary", Level: level.Level, LessonID: lesson.ID, ItemID: itemID, ItemType: itemType, Field: field, SourceText: source, SourcePath: path, Preserve: preserve, QuestionIDs: ids})
			}
			for _, word := range lesson.Words {
				addWordMaterial(word.ID, "Word", "term", word.Term, true)
				addWordMaterial(word.ID, "Word", "reading", word.Reading, true)
				addWordMaterial(word.ID, "Word", "meaning", word.Meaning, false)
				addWordMaterial(word.ID, "Word", "example", word.Example, true)
				addWordMaterial(word.ID, "Word", "exampleMeaning", word.ExampleMeaning, false)
				addWordMaterial(word.ID, "Word", "note", word.Note, false)
				addWordMaterial(word.ID, "Word", "imageDescription", word.ImageDescription, false)
				for exampleIndex, example := range word.Examples {
					exampleID := word.ID + ".example-" + itoaCoverage(exampleIndex+1)
					addWordMaterial(exampleID, "Example", "japanese", example.Japanese, true)
					addWordMaterial(exampleID, "Example", "reading", example.Reading, true)
					addWordMaterial(exampleID, "Example", "meaning", example.Meaning, false)
				}
			}
		}
	}

	grammarPaths := map[string]string{"N5": "backend/internal/store/grammar_content.go", "N4": "backend/internal/store/grammar_n4_content.go", "N3": "backend/internal/store/grammar_n3_content.go", "N2": "backend/internal/store/grammar_n2_content.go", "N1": "backend/internal/store/grammar_n1_content.go"}
	seenGrammarCourse := map[string]bool{}
	for _, lesson := range store.grammar {
		ids := make([]string, 0, len(lesson.Questions))
		path := grammarPaths[lesson.Level]
		for i, q := range lesson.Questions {
			qid := lesson.ID + "-q" + itoaCoverage(i+1)
			canonical := "JLPT.Grammar." + lesson.Level + "." + lesson.ID + "." + qid
			ids = append(ids, canonical)
			levelMaterialQuestions["Grammar:"+lesson.Level] = append(levelMaterialQuestions["Grammar:"+lesson.Level], canonical)
			addQuestion(coverageQuestion{CanonicalID: canonical, Surface: "JLPT lesson quiz", Category: "Grammar", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, LessonTitle: lesson.Title, QuestionID: qid, Ordinal: i + 1, Prompt: q.Prompt, Options: append([]string{}, q.Choices...), Correct: q.Correct, Explanation: q.Explanation, SourcePath: path}, coverageOccurrence{CanonicalID: canonical, Surface: "JLPT lesson quiz", Category: "Grammar", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, QuestionID: qid, Ordinal: i + 1, SourcePath: path})
		}
		addText("grammar:"+lesson.Level+":"+lesson.ID, "JLPT lesson quiz", "Grammar", lesson.Level, "", "", lesson.Chapter, lesson.ID, "Lesson title", lesson.Title, path, ids)
		addText("grammar:"+lesson.Level+":"+lesson.ID, "JLPT lesson quiz", "Grammar", lesson.Level, "", "", lesson.Chapter, lesson.ID, "Chapter label", lesson.Chapter, path, ids)
		if !seenGrammarCourse[lesson.Level] {
			seenGrammarCourse[lesson.Level] = true
			addMaterial(coverageMaterial{ID: "Grammar." + lesson.Level + ".course.title", ContextID: "grammar:" + lesson.Level, Surface: "JLPT learning material", Category: "Grammar", Level: lesson.Level, ItemID: lesson.Level, ItemType: "Course", Field: "title", SourceText: "Grammar JLPT " + lesson.Level, SourcePath: path, AlreadyEnglish: true, QuestionIDs: []string{}})
		}
		addPointMaterial := func(itemID, itemType, field, source string, preserve bool) {
			addMaterial(coverageMaterial{ID: "Grammar." + lesson.Level + "." + lesson.ID + "." + itemID + "." + field, ContextID: "grammar:" + lesson.Level + ":" + lesson.ID, Surface: "JLPT learning material", Category: "Grammar", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, ItemID: itemID, ItemType: itemType, Field: field, SourceText: source, SourcePath: path, Preserve: preserve, QuestionIDs: ids})
		}
		for pointIndex, point := range lesson.Points {
			pointID := "point-" + itoaCoverage(pointIndex+1)
			addPointMaterial(pointID, "Grammar point", "pattern", point.Pattern, false)
			addPointMaterial(pointID, "Grammar point", "meaning", point.Meaning, false)
			addPointMaterial(pointID, "Grammar point", "formation", point.Formation, false)
			addPointMaterial(pointID, "Grammar point", "explanation", point.Explanation, false)
			addPointMaterial(pointID, "Grammar point", "commonMistake", point.CommonMistake, false)
			for exampleIndex, example := range point.Examples {
				exampleID := pointID + ".example-" + itoaCoverage(exampleIndex+1)
				addPointMaterial(exampleID, "Grammar example", "japanese", example.Japanese, true)
				addPointMaterial(exampleID, "Grammar example", "reading", example.Reading, true)
				addPointMaterial(exampleID, "Grammar example", "meaning", example.Meaning, false)
			}
		}
	}
	kanjiPaths := map[string]string{"N5": "backend/internal/store/kanji_content.go", "N4": "backend/internal/store/kanji_content.go", "N1": "backend/internal/store/kanji_n1.go"}
	seenKanjiCourse := map[string]bool{}
	for _, lesson := range store.kanji {
		ids := make([]string, 0, len(lesson.Questions))
		path := kanjiPaths[lesson.Level]
		for i, q := range lesson.Questions {
			qid := lesson.ID + "-q" + itoaCoverage(i+1)
			canonical := "JLPT.Kanji." + lesson.Level + "." + lesson.ID + "." + qid
			ids = append(ids, canonical)
			levelMaterialQuestions["Kanji:"+lesson.Level] = append(levelMaterialQuestions["Kanji:"+lesson.Level], canonical)
			addQuestion(coverageQuestion{CanonicalID: canonical, Surface: "JLPT lesson quiz", Category: "Kanji", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, LessonTitle: lesson.Title, QuestionID: qid, Ordinal: i + 1, Prompt: q.Prompt, Options: append([]string{}, q.Choices...), Correct: q.Correct, Explanation: q.Explanation, SourcePath: path}, coverageOccurrence{CanonicalID: canonical, Surface: "JLPT lesson quiz", Category: "Kanji", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, QuestionID: qid, Ordinal: i + 1, SourcePath: path})
		}
		addText("kanji:"+lesson.Level+":"+lesson.ID, "JLPT lesson quiz", "Kanji", lesson.Level, "", "", lesson.Chapter, lesson.ID, "Lesson title", lesson.Title, path, ids)
		addText("kanji:"+lesson.Level+":"+lesson.ID, "JLPT lesson quiz", "Kanji", lesson.Level, "", "", lesson.Chapter, lesson.ID, "Chapter label", lesson.Chapter, path, ids)
		if !seenKanjiCourse[lesson.Level] {
			seenKanjiCourse[lesson.Level] = true
			addMaterial(coverageMaterial{ID: "Kanji." + lesson.Level + ".course.title", ContextID: "kanji:" + lesson.Level, Surface: "JLPT learning material", Category: "Kanji", Level: lesson.Level, ItemID: lesson.Level, ItemType: "Course", Field: "title", SourceText: "Kanji JLPT " + lesson.Level, SourcePath: path, AlreadyEnglish: true, QuestionIDs: []string{}})
		}
		addCharacterMaterial := func(itemID, itemType, field, source string, preserve bool) {
			addMaterial(coverageMaterial{ID: "Kanji." + lesson.Level + "." + lesson.ID + "." + itemID + "." + field, ContextID: "kanji:" + lesson.Level + ":" + lesson.ID, Surface: "JLPT learning material", Category: "Kanji", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, ItemID: itemID, ItemType: itemType, Field: field, SourceText: source, SourcePath: path, Preserve: preserve, QuestionIDs: ids})
		}
		for characterIndex, character := range lesson.Characters {
			characterID := character.Character
			if characterID == "" {
				characterID = "character-" + itoaCoverage(characterIndex+1)
			}
			addCharacterMaterial(characterID, "Kanji character", "character", character.Character, true)
			addCharacterMaterial(characterID, "Kanji character", "meaning", character.Meaning, false)
			addCharacterMaterial(characterID, "Kanji character", "onReading", character.OnReading, true)
			addCharacterMaterial(characterID, "Kanji character", "kunReading", character.KunReading, true)
			addCharacterMaterial(characterID, "Kanji character", "mnemonic", character.Mnemonic, false)
			addCharacterMaterial(characterID, "Kanji character", "note", character.Note, false)
			for exampleIndex, example := range character.Examples {
				exampleID := characterID + ".example-" + itoaCoverage(exampleIndex+1)
				addCharacterMaterial(exampleID, "Kanji example", "japanese", example.Japanese, true)
				addCharacterMaterial(exampleID, "Kanji example", "reading", example.Reading, true)
				addCharacterMaterial(exampleID, "Kanji example", "meaning", example.Meaning, false)
			}
		}
	}
	readingPaths := map[string]string{"N5": "backend/internal/store/reading_content.go", "N4": "backend/internal/store/reading_content.go", "N3": "backend/internal/store/reading_content.go", "N2": "backend/internal/store/reading_content.go", "N1": "backend/internal/store/reading_content.go"}
	seenReadingCourse := map[string]bool{}
	for _, lesson := range store.reading {
		ids := make([]string, 0, len(lesson.Questions))
		path := readingPaths[lesson.Level]
		pageTexts := make([]string, 0, len(lesson.Pages))
		for _, page := range lesson.Pages {
			if page.Japanese != "" {
				pageTexts = append(pageTexts, page.Japanese)
			}
		}
		passage := strings.Join(pageTexts, "\n")
		for i, q := range lesson.Questions {
			qid := lesson.ID + "-q" + itoaCoverage(i+1)
			canonical := "JLPT.Reading." + lesson.Level + "." + lesson.ID + "." + qid
			ids = append(ids, canonical)
			levelMaterialQuestions["Reading:"+lesson.Level] = append(levelMaterialQuestions["Reading:"+lesson.Level], canonical)
			addQuestion(coverageQuestion{CanonicalID: canonical, Surface: "JLPT lesson quiz", Category: "Reading", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, LessonTitle: lesson.Title, QuestionID: qid, Ordinal: i + 1, Passage: passage, Prompt: q.Prompt, Options: append([]string{}, q.Choices...), Correct: q.Correct, Explanation: q.Explanation, SourcePath: path}, coverageOccurrence{CanonicalID: canonical, Surface: "JLPT lesson quiz", Category: "Reading", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, QuestionID: qid, Ordinal: i + 1, SourcePath: path})
		}
		addText("reading:"+lesson.Level+":"+lesson.ID, "JLPT lesson quiz", "Reading", lesson.Level, "", "", lesson.Chapter, lesson.ID, "Lesson title", lesson.Title, path, ids)
		addText("reading:"+lesson.Level+":"+lesson.ID, "JLPT lesson quiz", "Reading", lesson.Level, "", "", lesson.Chapter, lesson.ID, "Chapter label", lesson.Chapter, path, ids)
		if !seenReadingCourse[lesson.Level] {
			seenReadingCourse[lesson.Level] = true
			addMaterial(coverageMaterial{ID: "Reading." + lesson.Level + ".course.title", ContextID: "reading:" + lesson.Level, Surface: "JLPT learning material", Category: "Reading", Level: lesson.Level, ItemID: lesson.Level, ItemType: "Course", Field: "title", SourceText: "Reading JLPT " + lesson.Level, SourcePath: path, AlreadyEnglish: true, QuestionIDs: []string{}})
		}
		addReadingMaterial := func(itemID, itemType, field, source string, preserve bool) {
			addMaterial(coverageMaterial{ID: "Reading." + lesson.Level + "." + lesson.ID + "." + itemID + "." + field, ContextID: "reading:" + lesson.Level + ":" + lesson.ID, Surface: "JLPT learning material", Category: "Reading", Level: lesson.Level, ModuleID: lesson.Chapter, LessonID: lesson.ID, ItemID: itemID, ItemType: itemType, Field: field, SourceText: source, SourcePath: path, Preserve: preserve, QuestionIDs: ids})
		}
		addReadingMaterial(lesson.ID, "Reading lesson", "japaneseTitle", lesson.JapaneseTitle, true)
		addReadingMaterial(lesson.ID, "Reading lesson", "description", lesson.Description, false)
		for pageIndex, page := range lesson.Pages {
			pageID := "page-" + itoaCoverage(pageIndex+1)
			addReadingMaterial(pageID, "Reading page", "title", page.Title, false)
			addReadingMaterial(pageID, "Reading page", "japanese", page.Japanese, true)
			addReadingMaterial(pageID, "Reading page", "reading", page.Reading, true)
			addReadingMaterial(pageID, "Reading page", "meaning", page.Meaning, false)
			addReadingMaterial(pageID, "Reading page", "imageDescription", page.ImageDescription, false)
			for vocabularyIndex, word := range page.Vocabulary {
				wordID := pageID + ".vocabulary-" + itoaCoverage(vocabularyIndex+1)
				addReadingMaterial(wordID, "Reading vocabulary", "japanese", word.Japanese, true)
				addReadingMaterial(wordID, "Reading vocabulary", "reading", word.Reading, true)
				addReadingMaterial(wordID, "Reading vocabulary", "meaning", word.Meaning, false)
			}
		}
	}
	for i := range bundle.Materials {
		material := &bundle.Materials[i]
		if material.ItemType == "Level" || material.ItemType == "Course" {
			key := material.Category + ":" + material.Level
			material.QuestionIDs = append(material.QuestionIDs, levelMaterialQuestions[key]...)
		}
	}

	for i, item := range diagnosticContent() {
		id := "n5-diagnostic-q" + padCoverage(i+1, 2)
		canonical := "Diagnostic.N5." + id
		options := append([]string{}, item.choices[:]...)
		addQuestion(coverageQuestion{CanonicalID: canonical, Surface: "Diagnostic", Category: item.section, Level: "N5", QuestionID: id, Ordinal: i + 1, Section: item.section, Passage: item.passage, Prompt: item.prompt, Options: options, Correct: 0, Explanation: item.explanation, SourcePath: "backend/internal/store/diagnostic_content.go"}, coverageOccurrence{CanonicalID: canonical, Surface: "Diagnostic", Category: item.section, Level: "N5", QuestionID: id, Ordinal: i + 1, Section: item.section, SourcePath: "backend/internal/store/diagnostic_content.go"})
	}

	simSources := map[string]string{
		"N5": "backend/internal/store/simulation_content.go", "N4": "backend/internal/store/simulation_n4_content.go", "N3": "backend/internal/store/simulation_n3_content.go", "N2": "backend/internal/store/simulation_n2_content.go", "N1": "backend/internal/store/simulation_n1_content.go",
		"NAT": "backend/internal/store/simulation_nat_content.go", "NAT-2": "backend/internal/store/simulation_nat_additional_content.go", "NAT-3": "backend/internal/store/simulation_nat_additional_content.go", "NAT-4": "backend/internal/store/simulation_nat_additional_content.go", "NAT-5": "backend/internal/store/simulation_nat_additional_content.go",
		"JFT": "backend/internal/store/simulation_jft_content.go", "JFT-2": "backend/internal/store/simulation_jft_additional_content.go", "JFT-3": "backend/internal/store/simulation_jft_additional_content.go", "JFT-4": "backend/internal/store/simulation_jft_additional_content.go", "JFT-5": "backend/internal/store/simulation_jft_additional_content.go",
	}
	var simForms []string
	for _, level := range []string{"N5", "N4", "N3", "N2", "N1"} {
		simForms = append(simForms, level)
	}
	for i := 1; i <= 5; i++ {
		simForms = append(simForms, "NAT-"+itoaCoverage(i))
	}
	for i := 1; i <= 5; i++ {
		if i == 1 {
			simForms = append(simForms, "JFT")
		} else {
			simForms = append(simForms, "JFT-"+itoaCoverage(i))
		}
	}
	for _, level := range simForms {
		paper := simulationPaper(level)
		form := paper.paperID
		path := simSources[level]
		ids := make([]string, 0, len(paper.items))
		for i, item := range paper.items {
			choices := make([]string, 0, len(item.choices))
			correct := -1
			for oi, choice := range item.choices {
				if choice != "" {
					if correct < 0 {
						correct = len(choices)
					}
					choices = append(choices, choice)
				}
				_ = oi
			}
			id := strings.ToLower(level) + "-sim-01-q" + padCoverage(i+1, 2)
			canonical := "Simulation." + level + "." + form + "." + id
			ids = append(ids, canonical)
			addQuestion(coverageQuestion{CanonicalID: canonical, Surface: "Simulation", Category: "Simulation", Level: level, Form: form, QuestionID: id, Ordinal: i + 1, Section: item.section, Group: item.group, Instructions: item.instructions, Passage: item.passage, Prompt: item.prompt, Options: choices, Correct: correct, Explanation: item.explanation, SourcePath: path}, coverageOccurrence{CanonicalID: canonical, Surface: "Simulation", Category: "Simulation", Level: level, Form: form, QuestionID: id, Ordinal: i + 1, Section: item.section, Group: item.group, SourcePath: path})
		}
		addText("simulation:"+form, "Simulation", "Simulation", level, "", form, "", "", "Paper title", level+" written simulation", path, ids)
		for _, section := range paper.sections {
			sectionIDs := []string{}
			for _, q := range bundle.Questions {
				if q.Form == form && q.Section == section.ID {
					sectionIDs = append(sectionIDs, q.CanonicalID)
				}
			}
			addText("simulation:"+form, "Simulation", "Simulation", level, "", form, "", "", "Section label · "+section.ID, section.Name, path, sectionIDs)
		}
		for _, q := range paper.items {
			// Instructions, group labels, and section labels are covered once per form in Shared Text.
			_ = q
		}
		seenInstructions := map[string]bool{}
		seenGroups := map[string]bool{}
		for i, item := range paper.items {
			qid := ids[i]
			if item.instructions != "" && !seenInstructions[item.instructions] {
				seenInstructions[item.instructions] = true
				addText("simulation:"+form, "Simulation", "Simulation", level, "", form, "", "", "Instructions", item.instructions, path, ids)
			}
			if item.group != "" && !seenGroups[item.group] {
				seenGroups[item.group] = true
				addText("simulation:"+form, "Simulation", "Simulation", level, "", form, "", "", "Group label", item.group, path, []string{qid})
			}
		}
	}

	loadSSW()
	bankFingerprint := map[string]string{}
	sectors := make([]string, 0, len(sswData.Banks))
	for sector := range sswData.Banks {
		sectors = append(sectors, sector)
	}
	sort.Strings(sectors)
	for _, sector := range sectors {
		for i, q := range sswData.Banks[sector] {
			fingerprint := sourceFingerprint(q.Prompt, q.Passage, q.Options, q.Answer, q.Explanation)
			canonical := "SSW.Bank." + sector + "." + q.ID
			bankFingerprint[fingerprint] = canonical
			bankByFingerprint[fingerprint] = canonical
			addQuestion(coverageQuestion{CanonicalID: canonical, Surface: "SSW practice bank", Category: "SSW practice", Sector: sector, ModuleID: q.ModuleID, LessonID: q.LessonID, QuestionID: q.ID, Ordinal: i + 1, Kind: q.Kind, Section: q.Section, Group: q.Group, Passage: q.Passage, Prompt: q.Prompt, Options: append([]string{}, q.Options...), Correct: q.Answer, Explanation: q.Explanation, SourcePath: "backend/internal/store/ssw_resources/assessments.json"}, coverageOccurrence{CanonicalID: canonical, Surface: "SSW practice bank", Category: "SSW practice", Sector: sector, ModuleID: q.ModuleID, LessonID: q.LessonID, QuestionID: q.ID, Ordinal: i + 1, Section: q.Section, Group: q.Group, SourcePath: "backend/internal/store/ssw_resources/assessments.json"})
		}
	}

	assetsDir := filepath.Join("..", "..", "..", "assets", "content", "ssw")
	assetFiles, err := filepath.Glob(filepath.Join(assetsDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(assetFiles)
	for _, file := range assetFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var asset curriculumAsset
		if err := json.Unmarshal(data, &asset); err != nil {
			t.Fatalf("decode %s: %v", file, err)
		}
		sector := strings.TrimSuffix(filepath.Base(file), ".json")
		path := filepath.ToSlash(filepath.Join("assets/content/ssw", filepath.Base(file)))
		sectorQuestionIDs := []string{}
		for _, module := range asset.Modules {
			for _, lesson := range module.Lessons {
				ids := []string{}
				for i, check := range lesson.Checks {
					fingerprint := sourceFingerprint(check.Prompt, "", check.Options, check.Answer, check.Explanation)
					canonical := bankByFingerprint[fingerprint]
					if canonical == "" {
						qid := lesson.ID + "-check" + itoaCoverage(i+1)
						canonical = "SSW.Curriculum." + sector + "." + lesson.ID + "." + qid
						addQuestion(coverageQuestion{CanonicalID: canonical, Surface: "SSW curriculum check", Category: "SSW curriculum", Sector: sector, ModuleID: module.ID, ModuleTitle: module.Title, LessonID: lesson.ID, LessonTitle: lesson.Title, QuestionID: qid, Ordinal: i + 1, Kind: check.Kind, Prompt: check.Prompt, Options: append([]string{}, check.Options...), Correct: check.Answer, Explanation: check.Explanation, SourcePath: path}, coverageOccurrence{CanonicalID: canonical, Surface: "SSW curriculum check", Category: "SSW curriculum", Sector: sector, ModuleID: module.ID, LessonID: lesson.ID, QuestionID: qid, Ordinal: i + 1, SourcePath: path})
					} else {
						bundle.Occurrences = append(bundle.Occurrences, coverageOccurrence{CanonicalID: canonical, Surface: "SSW curriculum check", Category: "SSW curriculum", Sector: sector, ModuleID: module.ID, LessonID: lesson.ID, QuestionID: sswQuestionIDByCanonical(sector, canonical, sswData.Banks), Ordinal: i + 1, SourcePath: path})
					}
					ids = append(ids, canonical)
				}
				sectorQuestionIDs = append(sectorQuestionIDs, ids...)
				addText("ssw-curriculum:"+sector+":"+lesson.ID, "SSW curriculum", "SSW curriculum", "", sector, "", module.ID, lesson.ID, "Module title", module.Title, path, ids)
				addText("ssw-curriculum:"+sector+":"+lesson.ID, "SSW curriculum", "SSW curriculum", "", sector, "", module.ID, lesson.ID, "Lesson title", lesson.Title, path, ids)
				addSSWMaterial := func(itemID, itemType, field, source string, preserve bool) {
					addMaterial(coverageMaterial{ID: "SSW." + sector + "." + module.ID + "." + lesson.ID + "." + itemID + "." + field, ContextID: "ssw-curriculum:" + sector + ":" + lesson.ID, Surface: "SSW learning material", Category: "SSW curriculum", Sector: sector, ModuleID: module.ID, LessonID: lesson.ID, ItemID: itemID, ItemType: itemType, Field: field, SourceText: source, SourcePath: path, Preserve: preserve, QuestionIDs: ids})
				}
				addSSWMaterial(module.ID, "SSW module", "title", module.Title, false)
				addSSWMaterial(module.ID, "SSW module", "description", module.Description, false)
				addSSWMaterial(lesson.ID, "SSW lesson", "title", lesson.Title, false)
				for objectiveIndex, objective := range lesson.Objectives {
					addSSWMaterial("objective-"+itoaCoverage(objectiveIndex+1), "SSW objective", "objective", objective, false)
				}
				for sectionIndex, section := range lesson.Sections {
					sectionID := "section-" + itoaCoverage(sectionIndex+1)
					addSSWMaterial(sectionID, "SSW section", "title", section.Title, false)
					addSSWMaterial(sectionID, "SSW section", "body", section.Body, false)
				}
				for stepIndex, step := range lesson.Steps {
					addSSWMaterial("step-"+itoaCoverage(stepIndex+1), "SSW step", "step", step, false)
				}
				addSSWMaterial(lesson.ID, "SSW lesson", "caution", lesson.Caution, false)
				addSSWMaterial(lesson.ID, "SSW lesson", "scenario", lesson.Scenario, false)
				addSSWMaterial(lesson.ID, "SSW lesson", "response", lesson.Response, false)
				for vocabularyIndex, word := range lesson.Vocabulary {
					wordID := "vocabulary-" + itoaCoverage(vocabularyIndex+1)
					addSSWMaterial(wordID, "SSW vocabulary", "japanese", word.Japanese, true)
					addSSWMaterial(wordID, "SSW vocabulary", "reading", word.Reading, true)
					addSSWMaterial(wordID, "SSW vocabulary", "meaning", word.Meaning, false)
				}
				for dialogueIndex, dialogue := range lesson.Dialogue {
					dialogueID := "dialogue-" + itoaCoverage(dialogueIndex+1)
					addSSWMaterial(dialogueID, "SSW dialogue", "japanese", dialogue.Japanese, true)
					addSSWMaterial(dialogueID, "SSW dialogue", "reading", dialogue.Reading, true)
					addSSWMaterial(dialogueID, "SSW dialogue", "meaning", dialogue.Meaning, false)
				}
			}
		}
		for i, source := range asset.Sources {
			addMaterial(coverageMaterial{
				ID: "SSW." + sector + ".source-" + itoaCoverage(i+1) + ".title", ContextID: "ssw-curriculum:" + sector + ":sources",
				Surface: "SSW learning material", Category: "SSW curriculum", Sector: sector, ItemID: "source-" + itoaCoverage(i+1),
				ItemType: "SSW source", Field: "sourceTitle", SourceText: source.Title, SourcePath: path, QuestionIDs: sectorQuestionIDs,
			})
		}
	}

	paperLevels := make([]string, 0, len(sswData.Papers))
	for level := range sswData.Papers {
		paperLevels = append(paperLevels, level)
	}
	sort.Strings(paperLevels)
	for _, level := range paperLevels {
		paper := sswData.Papers[level]
		config, ok := sswSimulationPaper(level)
		if !ok {
			continue
		}
		form := config.paperID
		paperIDs := []string{}
		for i, q := range paper.Items {
			fingerprint := sourceFingerprint(q.Prompt, q.Passage, q.Options, q.Answer, q.Explanation)
			canonical := bankFingerprint[fingerprint]
			if canonical == "" {
				t.Fatalf("SSW paper %s question %s does not match a bank item", level, q.ID)
			}
			paperIDs = append(paperIDs, canonical)
			bundle.Occurrences = append(bundle.Occurrences, coverageOccurrence{CanonicalID: canonical, Surface: "SSW simulation", Category: "SSW simulation", Sector: level, Form: form, QuestionID: q.ID, Ordinal: i + 1, Section: q.Section, Group: q.Group, SourcePath: "backend/internal/store/ssw_resources/assessments.json"})
		}
		addText("ssw-paper:"+form, "SSW simulation", "SSW simulation", "", level, form, "", "", "Paper title", paper.Title, "backend/internal/store/ssw_resources/assessments.json", paperIDs)
		for _, section := range paper.Sections {
			ids := []string{}
			for i, q := range paper.Items {
				if q.Section == section.ID {
					ids = append(ids, paperIDs[i])
				}
			}
			addText("ssw-paper:"+form, "SSW simulation", "SSW simulation", "", level, form, "", "", "Section label · "+section.ID, section.Name, "backend/internal/store/ssw_resources/assessments.json", ids)
		}
		addText("ssw-paper:"+form, "SSW simulation", "SSW simulation", "", level, form, "", "", "Instructions", "Pilih satu jawaban paling tepat. Soal orisinal Nihongo Master.", "backend/internal/store/ssw.go", paperIDs)
	}

	if len(bundle.Questions) == 0 {
		t.Fatal("no questions extracted")
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("exported questions=%d occurrences=%d shared_text=%d to %s", len(bundle.Questions), len(bundle.Occurrences), len(bundle.SharedText), output)
}

func itoaCoverage(value int) string       { return strconv.Itoa(value) }
func padCoverage(value, width int) string { return fmt.Sprintf("%0*d", width, value) }
func sswQuestionIDByCanonical(sector, canonical string, banks map[string][]sswQuestion) string {
	for _, q := range banks[sector] {
		if "SSW.Bank."+sector+"."+q.ID == canonical {
			return q.ID
		}
	}
	return canonical
}
