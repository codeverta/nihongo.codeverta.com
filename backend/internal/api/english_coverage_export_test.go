package api

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

// TestExportEnglishQuestionCoverage writes a compact, editable coverage register
// from the shared question inventory. It is opt-in because it creates files.
func TestExportEnglishQuestionCoverage(t *testing.T) {
	inputPath := os.Getenv("ENGLISH_COVERAGE_BUNDLE_PATH")
	outputDir := os.Getenv("ENGLISH_COVERAGE_REPORT_DIR")
	if inputPath == "" || outputDir == "" {
		t.Skip("set ENGLISH_COVERAGE_BUNDLE_PATH and ENGLISH_COVERAGE_REPORT_DIR to write the coverage register")
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	var bundle coverageReportBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
	}
	if len(bundle.Questions) == 0 {
		t.Fatal("question inventory is empty")
	}
	seenQuestionIDs := make(map[string]struct{}, len(bundle.Questions))
	for _, question := range bundle.Questions {
		if _, duplicate := seenQuestionIDs[question.CanonicalID]; duplicate {
			t.Fatalf("question inventory contains a duplicate canonical ID: %s", question.CanonicalID)
		}
		seenQuestionIDs[question.CanonicalID] = struct{}{}
	}
	if len(bundle.SharedText) == 0 {
		t.Fatal("shared learner-facing context inventory is empty")
	}
	for _, shared := range bundle.SharedText {
		if len(shared.QuestionIDs) == 0 {
			t.Fatalf("shared text has no linked questions: context=%s field=%s", shared.ContextID, shared.Field)
		}
		for _, questionID := range shared.QuestionIDs {
			if _, exists := seenQuestionIDs[questionID]; !exists {
				t.Fatalf("shared text links to an unknown question: context=%s question=%s", shared.ContextID, questionID)
			}
		}
	}
	if len(bundle.Materials) == 0 {
		t.Fatal("learning material inventory is empty")
	}
	seenMaterialIDs := make(map[string]struct{}, len(bundle.Materials))
	for _, material := range bundle.Materials {
		if material.ID == "" || strings.TrimSpace(material.SourceText) == "" {
			t.Fatalf("learning material is missing its identity or source text: %#v", material)
		}
		if _, duplicate := seenMaterialIDs[material.ID]; duplicate {
			t.Fatalf("learning material inventory contains a duplicate ID: %s", material.ID)
		}
		seenMaterialIDs[material.ID] = struct{}{}
		for _, questionID := range material.QuestionIDs {
			if _, exists := seenQuestionIDs[questionID]; !exists {
				t.Fatalf("learning material links to an unknown question: material=%s question=%s", material.ID, questionID)
			}
		}
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeQuestionCoverage(filepath.Join(outputDir, "questions.csv"), bundle); err != nil {
		t.Fatal(err)
	}
	if err := writeOptionCoverage(filepath.Join(outputDir, "options.csv"), bundle); err != nil {
		t.Fatal(err)
	}
	if err := writePartCoverage(filepath.Join(outputDir, "parts.csv"), bundle); err != nil {
		t.Fatal(err)
	}
	if err := writeSharedTextCoverage(filepath.Join(outputDir, "shared_text.csv"), bundle); err != nil {
		t.Fatal(err)
	}
	if err := writeMaterialsCoverage(filepath.Join(outputDir, "materials.csv"), bundle); err != nil {
		t.Fatal(err)
	}
}

type coverageReportQuestion struct {
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

type coverageReportOccurrence struct {
	CanonicalID string `json:"canonicalId"`
}

type coverageReportSharedText struct {
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

type coverageReportMaterial struct {
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

type coverageReportPart struct {
	ID                string
	Name              string
	Questions         []coverageReportQuestion
	Occurrences       int
	Complete          int
	Missing           int
	MissingSharedText int
	MissingMaterials  int
	SourceSets        map[string]struct{}
	QuestionPartIDs   map[string]string
}

type coverageReportBundle struct {
	Questions   []coverageReportQuestion   `json:"questions"`
	Occurrences []coverageReportOccurrence `json:"occurrences"`
	SharedText  []coverageReportSharedText `json:"sharedText"`
	Materials   []coverageReportMaterial   `json:"materials"`
}

type coverageReportField struct {
	Source  string
	English string
	Status  string
	Missing bool
}

func coverageReportTranslate(field, source string) coverageReportField {
	if strings.TrimSpace(source) == "" {
		return coverageReportField{Status: "Not present"}
	}
	if field == "preserveJapanese" {
		if coverageReportJapaneseOnly(source) {
			return coverageReportField{Source: source, English: source, Status: "Preserve Japanese"}
		}
		return coverageReportField{Source: source, Status: "Needs translation", Missing: true}
	}
	if translated, ok := translateReviewedLearningText(field, source); ok {
		return coverageReportField{Source: source, English: translated, Status: "Exact mapping"}
	}
	if translated, ok := learningTranslations[field][source]; ok {
		return coverageReportField{Source: source, English: translated, Status: "Exact mapping"}
	}
	if coverageReportJapaneseOnly(source) {
		return coverageReportField{Source: source, English: source, Status: "Preserve Japanese"}
	}
	return coverageReportField{Source: source, Status: "Needs translation", Missing: true}
}

func coverageReportKind(question coverageReportQuestion) coverageReportField {
	if question.Category == "Vocabulary" {
		return coverageReportTranslate("kind", question.Kind)
	}
	if strings.TrimSpace(question.Kind) == "" {
		return coverageReportField{Status: "Not present"}
	}
	return coverageReportField{Source: question.Kind, Status: "Internal metadata"}
}

func coverageReportJapaneseOnly(value string) bool {
	hasJapanese := false
	for _, char := range value {
		if unicode.IsLetter(char) {
			if unicode.In(char, unicode.Han, unicode.Hiragana, unicode.Katakana) || char == 'ー' {
				hasJapanese = true
			} else {
				return false
			}
		}
	}
	return hasJapanese
}

func coverageReportContainsJapanese(value string) bool {
	for _, char := range value {
		if unicode.In(char, unicode.Han, unicode.Hiragana, unicode.Katakana) || char == 'ー' {
			return true
		}
	}
	return false
}

func coverageReportFields(question coverageReportQuestion) []coverageReportField {
	return []coverageReportField{
		coverageReportQuestionTranslate(question, "title", question.ModuleTitle),
		coverageReportQuestionTranslate(question, "title", question.LessonTitle),
		coverageReportKind(question),
		coverageReportQuestionTranslate(question, "section", question.Section),
		coverageReportQuestionTranslate(question, "group", question.Group),
		coverageReportQuestionTranslate(question, "instructions", question.Instructions),
		coverageReportQuestionTranslate(question, "passage", question.Passage),
		coverageReportQuestionTranslate(question, "prompt", question.Prompt),
		coverageReportQuestionTranslate(question, "explanation", question.Explanation),
	}
}

func coverageReportQuestionTranslate(question coverageReportQuestion, field, source string) coverageReportField {
	if question.Surface == "SSW curriculum check" {
		if translated, ok := translateSSWCurriculumMaterialText(question.Sector, field, source); ok {
			return coverageReportField{Source: source, English: translated, Status: "Exact mapping"}
		}
	}
	if question.Category == "Kanji" {
		contextID := "kanji:" + question.Level + ":" + question.LessonID
		key := jlptKanjiMaterialEnglishKey{contextID: contextID, field: field, source: source}
		if translated, ok := jlptKanjiN3QuizEnglishCatalog[key]; ok {
			return coverageReportField{Source: source, English: translated, Status: "Exact mapping"}
		}
		if translated, ok := jlptKanjiReadingEnglish[field][source]; ok {
			return coverageReportField{Source: source, English: translated, Status: "Exact mapping"}
		}
	}
	return coverageReportTranslate(field, source)
}

func coverageReportQuestionStatus(question coverageReportQuestion) string {
	for index, field := range coverageReportFields(question) {
		if index == 2 && question.Category != "Vocabulary" { // Other assessment kinds are internal metadata.
			continue
		}
		if field.Missing {
			return "In progress"
		}
	}
	if question.Category == "Vocabulary" && coverageReportTranslate("imageDescription", question.ImageDescription).Missing {
		return "In progress"
	}
	for _, option := range question.Options {
		if coverageReportOption(question, option).Missing {
			return "In progress"
		}
	}
	return "Mapped"
}

func coverageReportOptionField(question coverageReportQuestion) string {
	if question.Category == "Vocabulary" {
		if question.Kind == "Arti kata" || (question.Kind == "" && isVocabularyMeaningQuestion(question.Prompt)) {
			return "vocabulary.meaning"
		}
		return "preserveJapanese"
	}
	if strings.HasPrefix(question.Surface, "SSW ") {
		return "options"
	}
	return "text"
}

func coverageReportOption(question coverageReportQuestion, source string) coverageReportField {
	if question.Category == "Simulation" && coverageReportContainsJapanese(source) {
		return coverageReportField{Source: source, English: source, Status: "Preserve Japanese"}
	}
	if question.Category == "Kanji" {
		if translated, ok := jlptKanjiReadingEnglish["choices"][source]; ok {
			return coverageReportField{Source: source, English: translated, Status: "Exact mapping"}
		}
	}
	if coverageReportEnglishLabel(source) {
		return coverageReportField{Source: source, English: source, Status: "Already English"}
	}
	englishGlossPrompt := (question.Level == "N1" && strings.HasPrefix(question.Prompt, "Apa gloss sumber untuk kanji ")) ||
		(question.Level == "N2" && strings.HasPrefix(question.Prompt, "Apa arti kamus ringkas "))
	if question.Category == "Kanji" && englishGlossPrompt && !coverageReportJapaneseOnly(source) {
		// N1/N2 sourced gloss quizzes store their answer choices in English;
		// the reading questions at those levels remain Japanese.
		return coverageReportField{Source: source, English: source, Status: "Already English"}
	}
	return coverageReportTranslate(coverageReportOptionField(question), source)
}

func coverageReportEnglishLabel(source string) bool {
	if len(source) == 0 || len(source) > 6 {
		return false
	}
	for _, char := range source {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

func TestCoverageReportOptionPreservesChoiceLabels(t *testing.T) {
	for _, label := range []string{"A", "B", "C", "D"} {
		if got := coverageReportOption(coverageReportQuestion{Category: "Simulation"}, label); got.Missing || got.Status != "Already English" || got.English != label {
			t.Errorf("choice label %q should be preserved as already English: %#v", label, got)
		}
	}
	if got := coverageReportOption(coverageReportQuestion{Category: "Simulation"}, "ATM"); got.Missing || got.Status != "Already English" || got.English != "ATM" {
		t.Fatalf("English acronym should be preserved: %#v", got)
	}
	if got := coverageReportOption(coverageReportQuestion{Category: "Simulation"}, "C はじめての書道"); got.Missing || got.Status != "Preserve Japanese" || got.English != "C はじめての書道" {
		t.Fatalf("Japanese simulation choice should be preserved: %#v", got)
	}
}

func TestVocabularyQuizCoverageIncludesAllLearnerFacingFields(t *testing.T) {
	question := coverageReportQuestion{
		Surface:          "JLPT lesson quiz",
		Category:         "Vocabulary",
		Kind:             "Arti kata",
		ImageDescription: "Ilustrasi petunjuk konteks kosakata",
		Prompt:           "Apa arti うさぎ?",
		Options:          []string{"saya"},
		Explanation:      "私（わたし）berarti saya.\n私は学生です。\nわたしはがくせいです。\nSaya adalah pelajar.",
	}

	if kind := coverageReportKind(question); kind.Missing || kind.English != "Word meaning" {
		t.Fatalf("vocabulary quiz kind should be covered: %#v", kind)
	}
	if imageDescription := coverageReportTranslate("imageDescription", question.ImageDescription); imageDescription.Missing {
		t.Fatalf("vocabulary quiz image description should be covered: %#v", imageDescription)
	}
	if got := coverageReportQuestionStatus(question); got != "Mapped" {
		t.Fatalf("all mapped vocabulary quiz fields should yield complete coverage, got %q", got)
	}
	if got := coverageReportTranslate("preserveJapanese", "コーヒー"); got.Missing || got.English != "コーヒー" || got.Status != "Preserve Japanese" {
		t.Fatalf("Japanese reading choices should be preserved even when they contain a prolonged sound mark: %#v", got)
	}
	n1Gloss := coverageReportQuestion{Category: "Kanji", Level: "N1", Prompt: "Apa gloss sumber untuk kanji 刀?"}
	if got := coverageReportOption(n1Gloss, "Sword; Saber; Knife"); got.Missing || got.English != "Sword; Saber; Knife" || got.Status != "Already English" {
		t.Fatalf("sourced English N1 gloss choices should be recognized as English: %#v", got)
	}
	n1Reading := coverageReportQuestion{Category: "Kanji", Level: "N1", Prompt: "Bagaimana membaca「刀」?"}
	if got := coverageReportOption(n1Reading, "かたな"); got.Missing || got.English != "かたな" || got.Status != "Preserve Japanese" {
		t.Fatalf("Japanese N1 reading choices should be preserved: %#v", got)
	}
	n2Gloss := coverageReportQuestion{Category: "Kanji", Level: "N2", Prompt: "Apa arti kamus ringkas 党?"}
	if got := coverageReportOption(n2Gloss, "Party; Faction"); got.Missing || got.English != "Party; Faction" || got.Status != "Already English" {
		t.Fatalf("sourced English N2 gloss choices should be recognized as English: %#v", got)
	}

	unknown := coverageReportQuestion{
		Category: "Vocabulary",
		Prompt:   "Apa arti rekaan-XYZ (ふりがな未登録)?",
		Options:  []string{"saya"},
	}
	if got := coverageReportQuestionStatus(unknown); got != "In progress" {
		t.Fatalf("an unreviewed generated question must remain uncovered: %q", got)
	}
	payload := []byte(`{"questions":[{"prompt":"Apa arti rekaan-XYZ (ふりがな未登録)?","options":[{"text":"saya"}]}]}`)
	localized, err := translateLearningJSON(payload)
	if err != nil {
		t.Fatal(err)
	}
	var untranslated struct {
		Questions []struct {
			Prompt string `json:"prompt"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(localized, &untranslated); err != nil {
		t.Fatal(err)
	}
	if len(untranslated.Questions) != 1 || untranslated.Questions[0].Prompt != "Apa arti rekaan-XYZ (ふりがな未登録)?" {
		t.Fatalf("unreviewed prompt must remain unchanged until its exact translation is added: %s", localized)
	}
}

func TestCoverageReportIncludesSharedContextInPartCompleteness(t *testing.T) {
	question := coverageReportQuestion{
		CanonicalID: "Simulation.N5.form-1.q1",
		Surface:     "Simulation",
		Category:    "Simulation",
		Level:       "N5",
		Form:        "form-1",
	}
	bundle := coverageReportBundle{
		Questions:   []coverageReportQuestion{question},
		Occurrences: []coverageReportOccurrence{{CanonicalID: question.CanonicalID}},
		SharedText: []coverageReportSharedText{{
			ContextID: "simulation:form-1", Surface: "Simulation", Category: "Simulation", Level: "N5",
			Form: "form-1", Field: "Instructions", SourceText: "Instruksi latihan ini belum diterjemahkan.",
			SourcePath: "simulation.go", QuestionIDs: []string{question.CanonicalID},
		}},
		Materials: []coverageReportMaterial{{
			ID: "material-1", ContextID: "grammar:N5:lesson-1", Surface: "JLPT learning material", Category: "Grammar", Level: "N5",
			ItemID: "point-1", ItemType: "Grammar point", Field: "explanation", SourceText: "Materi grammar ini belum diterjemahkan.",
			SourcePath: "grammar.go", QuestionIDs: []string{question.CanonicalID},
		}},
	}
	parts := coverageReportParts(bundle)
	if len(parts) != 1 || parts[0].Missing != 0 || parts[0].MissingSharedText != 1 || parts[0].MissingMaterials != 1 {
		t.Fatalf("unmapped shared context and learning material should make their question part incomplete: %#v", parts)
	}
	if paperTitle := coverageReportSharedTextTranslation(coverageReportSharedText{Field: "Paper title", SourceText: "N5 written simulation"}); paperTitle.Missing || paperTitle.Status != "Already English" {
		t.Fatalf("an already-English paper title should be covered: %#v", paperTitle)
	}
	if japanese := coverageReportMaterialTranslation(coverageReportMaterial{SourceText: "日本語の例文です。", Preserve: true}); japanese.Missing || japanese.Status != "Preserve Japanese" {
		t.Fatalf("Japanese study text should be preserved in the material inventory: %#v", japanese)
	}
	if source := coverageReportMaterialTranslation(coverageReportMaterial{SourceText: "Grammar JLPT N5", AlreadyEnglish: true}); source.Missing || source.Status != "Already English" {
		t.Fatalf("already-English course labels should be covered: %#v", source)
	}
	if readingMeaning := coverageReportMaterialTranslation(coverageReportMaterial{Category: "Reading", Field: "meaning", SourceText: "tiket"}); readingMeaning.Missing || readingMeaning.English != "ticket" {
		t.Fatalf("Reading material should use its exact glossary translation: %#v", readingMeaning)
	}
	kanjiMeaning := coverageReportMaterialTranslation(coverageReportMaterial{Category: "Kanji", ContextID: "kanji:N4:kanji-n4-23", Field: "meaning", SourceText: "berat"})
	if kanjiMeaning.Missing || kanjiMeaning.English != "heavy" {
		t.Fatalf("Kanji material should use its lesson-scoped glossary translation: %#v", kanjiMeaning)
	}
	kanjiPrompt := coverageReportQuestionTranslate(coverageReportQuestion{Category: "Kanji", Level: "N3", LessonID: "kanji-n3-38"}, "prompt", "Apa arti utama kanji 報 pada kartu ini?")
	if kanjiPrompt.Missing || kanjiPrompt.English != "What is the main meaning of 報 on this card?" {
		t.Fatalf("Kanji quiz should use its lesson-scoped prompt translation: %#v", kanjiPrompt)
	}
	for _, material := range []coverageReportMaterial{
		{Category: "Kanji", Level: "N1", Field: "meaning", ItemType: "Kanji character", SourceText: "Sword; Saber; Knife"},
		{Category: "Kanji", Level: "N2", Field: "meaning", ItemType: "Kanji example", SourceText: "association, society, organization"},
	} {
		if got := coverageReportMaterialTranslation(material); got.Missing || got.Status != "Already English" || got.English != material.SourceText {
			t.Errorf("N1/N2 KANJIDIC meanings should be preserved as English: %#v", got)
		}
	}
}

func coverageReportParts(bundle coverageReportBundle) []coverageReportPart {
	groupOrder := make([]string, 0)
	groups := make(map[string][]coverageReportQuestion)
	for _, question := range bundle.Questions {
		key := coverageReportGroupKey(question)
		if _, exists := groups[key]; !exists {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], question)
	}
	parts := make([]coverageReportPart, 0)
	questionPartIDs := make(map[string]string, len(bundle.Questions))
	for _, key := range groupOrder {
		group := groups[key]
		for start, inGroup := 0, 1; start < len(group); inGroup++ {
			end := start + 200
			if end > len(group) {
				end = len(group)
			}
			if end < len(group) && group[end-1].LessonID == group[end].LessonID {
				for boundary := end - 1; boundary > start; boundary-- {
					if group[boundary-1].LessonID != group[boundary].LessonID {
						end = boundary
						break
					}
				}
			}
			chunk := append([]coverageReportQuestion(nil), group[start:end]...)
			id := fmt.Sprintf("PART-%03d", len(parts)+1)
			name := coverageReportPartName(chunk[0], inGroup)
			part := coverageReportPart{ID: id, Name: name, Questions: chunk, SourceSets: map[string]struct{}{}, QuestionPartIDs: questionPartIDs}
			for _, question := range chunk {
				questionPartIDs[question.CanonicalID] = id
				part.SourceSets[question.SourcePath] = struct{}{}
				if coverageReportQuestionStatus(question) == "Mapped" {
					part.Complete++
				} else {
					part.Missing++
				}
			}
			parts = append(parts, part)
			start = end
		}
	}
	for _, occurrence := range bundle.Occurrences {
		if partID := questionPartIDs[occurrence.CanonicalID]; partID != "" {
			for index := range parts {
				if parts[index].ID == partID {
					parts[index].Occurrences++
					break
				}
			}
		}
	}
	partIndexes := make(map[string]int, len(parts))
	for index, part := range parts {
		partIndexes[part.ID] = index
	}
	for _, shared := range bundle.SharedText {
		if !coverageReportSharedTextTranslation(shared).Missing {
			continue
		}
		counted := make(map[string]struct{})
		for _, questionID := range shared.QuestionIDs {
			partID := questionPartIDs[questionID]
			if partID == "" {
				continue
			}
			if _, seen := counted[partID]; seen {
				continue
			}
			counted[partID] = struct{}{}
			parts[partIndexes[partID]].MissingSharedText++
		}
	}
	for _, material := range bundle.Materials {
		if !coverageReportMaterialTranslation(material).Missing {
			continue
		}
		counted := make(map[string]struct{})
		for _, questionID := range material.QuestionIDs {
			partID := questionPartIDs[questionID]
			if partID == "" {
				continue
			}
			if _, seen := counted[partID]; seen {
				continue
			}
			counted[partID] = struct{}{}
			parts[partIndexes[partID]].MissingMaterials++
		}
	}
	assigned := 0
	for _, part := range parts {
		if len(part.Questions) > 200 {
			panic(fmt.Sprintf("coverage part exceeds 200 questions: %s", part.ID))
		}
		assigned += len(part.Questions)
	}
	if assigned != len(bundle.Questions) {
		panic(fmt.Sprintf("coverage part assignment lost questions: assigned %d of %d", assigned, len(bundle.Questions)))
	}
	return parts
}

func coverageReportPartName(question coverageReportQuestion, partInGroup int) string {
	labels := []string{question.Surface, question.Category, question.Level}
	if question.Surface == "Simulation" {
		labels = append(labels, question.Form)
	}
	if strings.HasPrefix(question.Surface, "SSW ") {
		labels = append(labels, question.Sector)
	}
	kept := labels[:0]
	for _, label := range labels {
		if label != "" {
			kept = append(kept, label)
		}
	}
	return strings.Join(kept, " / ") + fmt.Sprintf(" / Part %02d", partInGroup)
}

func coverageReportGroupKey(question coverageReportQuestion) string {
	labels := []string{question.Surface, question.Category, question.Level}
	if question.Surface == "Simulation" {
		labels = append(labels, question.Form)
	}
	if strings.HasPrefix(question.Surface, "SSW ") {
		labels = append(labels, question.Sector)
	}
	return strings.Join(labels, "\x00")
}

func coverageReportFingerprint(question coverageReportQuestion) string {
	value, _ := json.Marshal([]any{question.Prompt, question.Passage, question.Instructions, question.Options, question.Correct, question.Explanation})
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func writeQuestionCoverage(path string, bundle coverageReportBundle) error {
	parts := coverageReportParts(bundle)
	questionPartIDs := make(map[string]string, len(bundle.Questions))
	for _, part := range parts {
		for _, question := range part.Questions {
			questionPartIDs[question.CanonicalID] = part.ID
		}
	}
	return coverageReportCSV(path, []string{
		"part_id", "canonical_id", "source_fingerprint", "surface", "category", "level", "sector", "form", "module_id",
		"module_title_source", "module_title_english", "module_title_status", "lesson_id", "lesson_title_source", "lesson_title_english", "lesson_title_status",
		"question_id", "ordinal", "kind_source", "kind_english", "kind_status", "section_source", "section_english", "section_status", "group_source", "group_english", "group_status",
		"instructions_source", "instructions_english", "instructions_status", "passage_source", "passage_english", "passage_status", "prompt_source", "prompt_english", "prompt_status",
		"correct_index", "explanation_source", "explanation_english", "explanation_status", "option_count", "coverage_status", "source_path",
		"image_description_source", "image_description_english", "image_description_status",
	}, func(writer *csv.Writer) error {
		for _, question := range bundle.Questions {
			fields := coverageReportFields(question)
			moduleTitle, lessonTitle, kind, section, group, instructions, passage, prompt, explanation := fields[0], fields[1], fields[2], fields[3], fields[4], fields[5], fields[6], fields[7], fields[8]
			row := []string{
				questionPartIDs[question.CanonicalID], question.CanonicalID, coverageReportFingerprint(question), question.Surface, question.Category, question.Level, question.Sector, question.Form, question.ModuleID,
				moduleTitle.Source, moduleTitle.English, moduleTitle.Status, question.LessonID, lessonTitle.Source, lessonTitle.English, lessonTitle.Status,
				question.QuestionID, fmt.Sprint(question.Ordinal), kind.Source, kind.English, kind.Status, section.Source, section.English, section.Status, group.Source, group.English, group.Status,
				instructions.Source, instructions.English, instructions.Status, passage.Source, passage.English, passage.Status, prompt.Source, prompt.English, prompt.Status,
				fmt.Sprint(question.Correct), explanation.Source, explanation.English, explanation.Status, fmt.Sprint(len(question.Options)), coverageReportQuestionStatus(question), question.SourcePath,
				coverageReportTranslate("imageDescription", question.ImageDescription).Source, coverageReportTranslate("imageDescription", question.ImageDescription).English, coverageReportTranslate("imageDescription", question.ImageDescription).Status,
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
		return nil
	})
}

func writeOptionCoverage(path string, bundle coverageReportBundle) error {
	parts := coverageReportParts(bundle)
	questionPartIDs := make(map[string]string, len(bundle.Questions))
	for _, part := range parts {
		for _, question := range part.Questions {
			questionPartIDs[question.CanonicalID] = part.ID
		}
	}
	return coverageReportCSV(path, []string{"part_id", "canonical_id", "question_id", "question_ordinal", "option_index", "source_text", "english_text", "status", "is_correct", "source_path"}, func(writer *csv.Writer) error {
		for _, question := range bundle.Questions {
			for index, option := range question.Options {
				translated := coverageReportOption(question, option)
				correct := index == question.Correct
				if err := writer.Write([]string{
					questionPartIDs[question.CanonicalID], question.CanonicalID, question.QuestionID, fmt.Sprint(question.Ordinal), fmt.Sprint(index),
					option, translated.English, translated.Status, fmt.Sprint(correct), question.SourcePath,
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func writePartCoverage(path string, bundle coverageReportBundle) error {
	parts := coverageReportParts(bundle)
	return coverageReportCSV(path, []string{"part_id", "part_name", "question_count", "display_occurrences", "mapped_questions", "questions_needing_translation", "shared_text_needing_translation", "materials_needing_translation", "coverage_status", "source_file_count"}, func(writer *csv.Writer) error {
		for _, part := range parts {
			status := "Incomplete"
			if part.Missing == 0 && part.MissingSharedText == 0 && part.MissingMaterials == 0 {
				status = "Mapped"
			}
			if err := writer.Write([]string{
				part.ID, part.Name, fmt.Sprint(len(part.Questions)), fmt.Sprint(part.Occurrences), fmt.Sprint(part.Complete), fmt.Sprint(part.Missing), fmt.Sprint(part.MissingSharedText), fmt.Sprint(part.MissingMaterials), status, fmt.Sprint(len(part.SourceSets)),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func coverageReportSharedTextField(shared coverageReportSharedText) string {
	switch {
	case shared.Field == "Lesson title", shared.Field == "Module title", shared.Field == "Paper title":
		return "title"
	case shared.Field == "Lesson description":
		return "description"
	case strings.HasPrefix(shared.Field, "Chapter label"):
		return "chapter"
	case strings.HasPrefix(shared.Field, "Section label"):
		return "section"
	case shared.Field == "Instructions":
		return "instructions"
	case shared.Field == "Group label":
		return "group"
	default:
		return "text"
	}
}

func coverageReportSharedTextTranslation(shared coverageReportSharedText) coverageReportField {
	if shared.Field == "Paper title" && isEnglishSimulationPaperTitle(shared.SourceText) {
		return coverageReportField{Source: shared.SourceText, English: shared.SourceText, Status: "Already English"}
	}
	if shared.Surface == "SSW curriculum" {
		field := coverageReportSharedTextField(shared)
		if translated, ok := translateSSWCurriculumMaterialText(shared.Sector, field, shared.SourceText); ok {
			return coverageReportField{Source: shared.SourceText, English: translated, Status: "Exact mapping"}
		}
	}
	return coverageReportTranslate(coverageReportSharedTextField(shared), shared.SourceText)
}

func isEnglishSimulationPaperTitle(source string) bool {
	switch source {
	case "N1 written simulation", "N2 written simulation", "N3 written simulation", "N4 written simulation", "N5 written simulation",
		"NAT-1 written simulation", "NAT-2 written simulation", "NAT-3 written simulation", "NAT-4 written simulation", "NAT-5 written simulation",
		"JFT written simulation", "JFT-2 written simulation", "JFT-3 written simulation", "JFT-4 written simulation", "JFT-5 written simulation":
		return true
	default:
		return false
	}
}

func coverageReportMaterialTranslation(material coverageReportMaterial) coverageReportField {
	if strings.TrimSpace(material.SourceText) == "" {
		return coverageReportField{Status: "Not present"}
	}
	if material.Preserve {
		return coverageReportField{Source: material.SourceText, English: material.SourceText, Status: "Preserve Japanese"}
	}
	if material.AlreadyEnglish {
		return coverageReportField{Source: material.SourceText, English: material.SourceText, Status: "Already English"}
	}
	if material.Category == "Kanji" && (material.Level == "N1" || material.Level == "N2") && material.Field == "meaning" &&
		(material.ItemType == "Kanji character" || material.ItemType == "Kanji example") {
		// N1/N2 glossary meanings are English KANJIDIC values in the source
		// curriculum; preserve these glosses verbatim rather than translating them.
		return coverageReportField{Source: material.SourceText, English: material.SourceText, Status: "Already English"}
	}
	if material.Surface == "SSW learning material" {
		if translated, ok := translateSSWCurriculumMaterialText(material.Sector, material.Field, material.SourceText); ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
		if coverageReportJapaneseOnly(material.SourceText) {
			return coverageReportField{Source: material.SourceText, English: material.SourceText, Status: "Preserve Japanese"}
		}
		return coverageReportField{Source: material.SourceText, Status: "Needs translation", Missing: true}
	}
	if material.Category == "Reading" {
		if translated, ok := jlptKanjiReadingEnglish[material.Field][material.SourceText]; ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
	}
	if material.Category == "Kanji" {
		key := jlptKanjiMaterialEnglishKey{contextID: material.ContextID, field: material.Field, source: material.SourceText}
		if translated, ok := jlptKanjiN2Lessons19To44English[key]; ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
		if translated, ok := jlptKanjiN2Lessons45To74English[key]; ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
		if translated, ok := jlptKanjiN3EarlyMaterialEnglish[key]; ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
		if translated, ok := jlptKanjiN3Lessons39To77English[key]; ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
		if translated, ok := jlptKanjiN2MnemonicNoteEnglish[key]; ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
		if translated, ok := jlptKanjiN2Lessons09To18English[key]; ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
		if translated, ok := jlptKanjiMaterialEnglishCatalog[key]; ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
	}
	if material.Category == "Vocabulary" || material.Category == "Grammar" {
		domain := strings.ToLower(material.Category)
		if translated, ok := englishJLPTVocabularyGrammar(domain+"."+material.Field, material.SourceText); ok {
			return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
		}
		if material.Category == "Grammar" && material.Field == "meaning" {
			if translated, ok := englishJLPTVocabularyGrammar("grammar.exampleMeaning", material.SourceText); ok {
				return coverageReportField{Source: material.SourceText, English: translated, Status: "Exact mapping"}
			}
		}
	}
	return coverageReportTranslate(material.Field, material.SourceText)
}

func writeSharedTextCoverage(path string, bundle coverageReportBundle) error {
	parts := coverageReportParts(bundle)
	questionPartIDs := make(map[string]string, len(bundle.Questions))
	for _, part := range parts {
		for _, question := range part.Questions {
			questionPartIDs[question.CanonicalID] = part.ID
		}
	}
	return coverageReportCSV(path, []string{
		"context_id", "part_ids", "surface", "category", "level", "sector", "form", "module_id", "lesson_id",
		"field", "source_text", "english_text", "status", "question_count", "question_ids", "source_path",
	}, func(writer *csv.Writer) error {
		for _, shared := range bundle.SharedText {
			translated := coverageReportSharedTextTranslation(shared)
			partIDs := make([]string, 0)
			seenParts := make(map[string]struct{})
			for _, questionID := range shared.QuestionIDs {
				partID := questionPartIDs[questionID]
				if partID == "" {
					continue
				}
				if _, seen := seenParts[partID]; seen {
					continue
				}
				seenParts[partID] = struct{}{}
				partIDs = append(partIDs, partID)
			}
			if err := writer.Write([]string{
				shared.ContextID, strings.Join(partIDs, ";"), shared.Surface, shared.Category, shared.Level, shared.Sector, shared.Form, shared.ModuleID, shared.LessonID,
				shared.Field, translated.Source, translated.English, translated.Status, fmt.Sprint(len(shared.QuestionIDs)), strings.Join(shared.QuestionIDs, ";"), shared.SourcePath,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func writeMaterialsCoverage(path string, bundle coverageReportBundle) error {
	parts := coverageReportParts(bundle)
	questionPartIDs := make(map[string]string, len(bundle.Questions))
	for _, part := range parts {
		for _, question := range part.Questions {
			questionPartIDs[question.CanonicalID] = part.ID
		}
	}
	return coverageReportCSV(path, []string{
		"material_id", "context_id", "part_ids", "surface", "category", "level", "sector", "module_id", "lesson_id", "item_id", "item_type",
		"field", "source_text", "english_text", "status", "question_count", "question_ids", "source_path",
	}, func(writer *csv.Writer) error {
		for _, material := range bundle.Materials {
			translated := coverageReportMaterialTranslation(material)
			partIDs := make([]string, 0)
			seenParts := make(map[string]struct{})
			for _, questionID := range material.QuestionIDs {
				partID := questionPartIDs[questionID]
				if partID == "" {
					continue
				}
				if _, seen := seenParts[partID]; seen {
					continue
				}
				seenParts[partID] = struct{}{}
				partIDs = append(partIDs, partID)
			}
			if err := writer.Write([]string{
				material.ID, material.ContextID, strings.Join(partIDs, ";"), material.Surface, material.Category, material.Level, material.Sector, material.ModuleID, material.LessonID, material.ItemID, material.ItemType,
				material.Field, translated.Source, translated.English, translated.Status, fmt.Sprint(len(material.QuestionIDs)), strings.Join(material.QuestionIDs, ";"), material.SourcePath,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func coverageReportCSV(path string, header []string, rows func(*csv.Writer) error) (err error) {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	if _, err := file.Write([]byte{0xef, 0xbb, 0xbf}); err != nil {
		return err
	}
	writer := csv.NewWriter(file)
	if err := writer.Write(header); err != nil {
		return err
	}
	if err := rows(writer); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}
