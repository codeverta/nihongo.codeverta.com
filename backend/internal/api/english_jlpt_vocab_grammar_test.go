package api

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"unicode"

	"nihongo-master-backend/internal/store"
)

func TestEnglishJLPTVocabularyAndGrammarCatalog(t *testing.T) {
	cases := []struct {
		field, source, want string
	}{
		{"vocabulary.title", "Perkenalan diri", "Self-introductions"},
		{"vocabulary.meaning", "pelajar", "student"},
		{"vocabulary.note", "Perhatikan partikel dan pasangan kata pada dua contoh. Gunakan cara baca yang tertera ketika mengucapkan kata.", "Pay attention to the particles and word combinations in the two examples. Use the provided reading when saying the word."},
		{"grammar.meaning", "harus melakukan …", "must do …"},
		{"grammar.commonMistake", "行かなくてもいい berarti tidak perlu pergi, bukan tidak boleh pergi.", "行かなくてもいい means you do not need to go, not that you must not go."},
		{"grammar.exampleMeaning", "Hari ini saya tidak pergi ke sekolah.", "I am not going to school today."},
		{"grammar.prompt", "Apa makna kalimat berikut?\n今日は学校へ行かない。", "What does the following sentence mean?\n今日は学校へ行かない。"},
	}
	for _, tc := range cases {
		got, ok := englishJLPTVocabularyGrammar(tc.field, tc.source)
		if !ok || got != tc.want {
			t.Errorf("translation %q (%q) = %q, %v; want %q", tc.field, tc.source, got, ok, tc.want)
		}
	}
	for _, japanese := range []string{"私", "日本語を勉強します。", "行かなかった"} {
		if _, ok := englishJLPTVocabularyGrammar("vocabulary.meaning", japanese); ok {
			t.Errorf("Japanese source text must remain unmapped: %q", japanese)
		}
	}
}

func TestN5GeneratedVocabularyQuizInventory(t *testing.T) {
	s := store.NewMemoryStore()
	const userID = "localization-inventory"
	var questions, prompts, explanations, englishChoices int
	seenPrompts, seenExplanations := map[string]struct{}{}, map[string]struct{}{}
	unmappedPrompts, unmappedExplanations := map[string]struct{}{}, map[string]struct{}{}
	unmappedOther := map[jlptEnglishSource]struct{}{}
	for _, level := range s.VocabularyLevels(userID) {
		if level.Level != "N5" {
			continue
		}
		for _, summary := range level.Lessons {
			attempt, err := s.StartVocabularyAttempt(userID, summary.ID)
			if err != nil {
				t.Fatal(err)
			}
			for _, q := range attempt.Questions {
				questions++
				if _, ok := englishJLPTVocabularyGrammar("vocabulary.quiz.kind", q.Kind); !ok {
					unmappedOther[jlptEnglishSource{field: "vocabulary.quiz.kind", source: q.Kind}] = struct{}{}
				}
				if q.ImageDescription != "" {
					if _, ok := englishJLPTVocabularyGrammar("vocabulary.quiz.imageDescription", q.ImageDescription); !ok {
						unmappedOther[jlptEnglishSource{field: "vocabulary.quiz.imageDescription", source: q.ImageDescription}] = struct{}{}
					}
				}
				if _, ok := englishJLPTVocabularyGrammar("vocabulary.quiz.prompt", q.Prompt); !ok {
					prompts++
					unmappedPrompts[q.Prompt] = struct{}{}
				}
				seenPrompts[q.Prompt] = struct{}{}
				for _, option := range q.Options {
					if !hasJapaneseText(option.Text) {
						if _, ok := englishJLPTVocabularyGrammar("vocabulary.meaning", option.Text); !ok {
							englishChoices++
						}
					}
				}
				feedback, err := s.AnswerVocabularyQuestion(userID, attempt.ID, q.ID, q.Options[0].ID)
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := englishJLPTVocabularyGrammar("vocabulary.quiz.explanation", feedback.Explanation); !ok {
					explanations++
					unmappedExplanations[feedback.Explanation] = struct{}{}
				}
				seenExplanations[feedback.Explanation] = struct{}{}
			}
		}
	}
	t.Logf("generated N5 quiz inventory: %d questions; %d unique prompts (%d unmapped occurrences), %d unmapped Indonesian choices, %d unique explanations (%d unmapped occurrences)", questions, len(seenPrompts), prompts, englishChoices, len(seenExplanations), explanations)
	if questions != 480 {
		t.Fatalf("expected 480 generated N5 questions, got %d", questions)
	}
	for source := range unmappedPrompts {
		t.Errorf("unmapped vocabulary.quiz.prompt: %q", source)
	}
	for source := range unmappedExplanations {
		t.Errorf("unmapped vocabulary.quiz.explanation: %q", source)
	}
	for source := range unmappedOther {
		t.Errorf("unmapped %s: %q", source.field, source.source)
	}
}

func TestN4GeneratedVocabularyQuizInventory(t *testing.T) {
	s := store.NewMemoryStore()
	const userID = "localization-inventory-n4"
	var questions int
	unmapped := map[jlptEnglishSource]struct{}{}
	missingComponents := map[jlptEnglishSource]struct{}{}
	for _, level := range s.VocabularyLevels(userID) {
		if level.Level != "N4" {
			continue
		}
		for _, summary := range level.Lessons {
			attempt, err := s.StartVocabularyAttempt(userID, summary.ID)
			if err != nil {
				t.Fatal(err)
			}
			for _, q := range attempt.Questions {
				questions++
				for _, fieldSource := range []jlptEnglishSource{
					{field: "kind", source: q.Kind},
					{field: "prompt", source: q.Prompt},
				} {
					if coverageReportTranslate(fieldSource.field, fieldSource.source).Missing {
						unmapped[jlptEnglishSource{field: fieldSource.field, source: fieldSource.source}] = struct{}{}
					}
				}
				for _, option := range q.Options {
					if !hasJapaneseText(option.Text) {
						if coverageReportTranslate("vocabulary.meaning", option.Text).Missing {
							unmapped[jlptEnglishSource{field: "vocabulary.meaning", source: option.Text}] = struct{}{}
						}
					}
				}
				feedback, err := s.AnswerVocabularyQuestion(userID, attempt.ID, q.ID, q.Options[0].ID)
				if err != nil {
					t.Fatal(err)
				}
				if coverageReportTranslate("explanation", feedback.Explanation).Missing {
					unmapped[jlptEnglishSource{field: "vocabulary.quiz.explanation", source: feedback.Explanation}] = struct{}{}
					for _, component := range generatedVocabularyExplanationComponents(feedback.Explanation) {
						if coverageReportTranslate(component.field, component.source).Missing {
							missingComponents[component] = struct{}{}
						}
					}
				}
			}
		}
	}
	t.Logf("generated N4 quiz inventory: %d questions; %d unique unmapped source strings", questions, len(unmapped))
	for field := range map[string]struct{}{"vocabulary.meaning": {}, "vocabulary.note": {}, "vocabulary.exampleMeaning": {}} {
		count := 0
		for component := range missingComponents {
			if component.field == field {
				count++
			}
		}
		t.Logf("generated N4 explanation component coverage: %s has %d unique unmapped sources", field, count)
	}
	if questions != 630 {
		t.Fatalf("expected 630 generated N4 questions, got %d", questions)
	}
	for source := range unmapped {
		t.Errorf("unmapped %s: %q", source.field, source.source)
	}
}

func TestN3VocabularyCatalogCoverage(t *testing.T) {
	testVocabularyCatalogCoverage(t, "n3", 0)
}

func TestN2VocabularyCatalogCoverage(t *testing.T) {
	testVocabularyCatalogCoverage(t, "n2", 0)
}

func TestN1VocabularyCatalogCoverage(t *testing.T) {
	testVocabularyCatalogCoverage(t, "n1", 0)
}

func generatedVocabularyExplanationComponents(source string) []jlptEnglishSource {
	meaningStart := strings.Index(source, " berarti ")
	exampleStart := strings.LastIndex(source, " Contoh: ")
	if meaningStart <= 0 || exampleStart <= meaningStart {
		return nil
	}
	meaningEnd := strings.Index(source[meaningStart+len(" berarti "):exampleStart], ". ")
	if meaningEnd < 0 {
		return nil
	}
	meaningEnd += meaningStart + len(" berarti ")
	noteStart := meaningEnd + 2
	exampleAndMeaning := source[exampleStart+len(" Contoh: "):]
	separator := strings.LastIndex(exampleAndMeaning, " — ")
	if separator < 0 {
		return nil
	}
	return []jlptEnglishSource{
		{field: "vocabulary.meaning", source: source[meaningStart+len(" berarti ") : meaningEnd]},
		{field: "vocabulary.note", source: source[noteStart:exampleStart]},
		{field: "vocabulary.exampleMeaning", source: exampleAndMeaning[separator+len(" — "):]},
	}
}

func hasJapaneseText(value string) bool {
	for _, r := range value {
		if (r >= 0x3040 && r <= 0x30ff) || (r >= 0x3400 && r <= 0x9fff) {
			return true
		}
	}
	return false
}

func isJapaneseOnly(value string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) && r < 0x3000 {
			return false
		}
	}
	return true
}

func TestN5GrammarCatalogCoverage(t *testing.T) {
	testGrammarCatalogCoverage(t, "N5", 40)
}

func TestN4GrammarCatalogCoverage(t *testing.T) {
	testGrammarCatalogCoverage(t, "N4", 0)
}

func TestN3GrammarCatalogCoverage(t *testing.T) {
	testGrammarCatalogCoverage(t, "N3", 0)
}

func TestN2GrammarCatalogCoverage(t *testing.T) {
	testGrammarCatalogCoverage(t, "N2", 0)
}

func TestN1GrammarCatalogCoverage(t *testing.T) {
	testGrammarCatalogCoverage(t, "N1", 0)
}

func testGrammarCatalogCoverage(t *testing.T, level string, expectedLessons int) {
	t.Helper()
	s := store.NewMemoryStore()
	userID := "grammar-localization-inventory-" + level
	sources := map[jlptEnglishSource]struct{}{}
	add := func(field, source string) {
		if source != "" && !isJapaneseOnly(source) {
			sources[jlptEnglishSource{field: field, source: source}] = struct{}{}
		}
	}
	course := s.GrammarCourse(userID, level)
	for _, summary := range course.Lessons {
		lesson, err := s.GrammarLesson(userID, summary.ID)
		if err != nil {
			t.Fatal(err)
		}
		add("grammar.chapter", lesson.Chapter)
		add("grammar.title", lesson.Title)
		for _, point := range lesson.Points {
			add("grammar.meaning", point.Meaning)
			add("grammar.formation", point.Formation)
			add("grammar.explanation", point.Explanation)
			add("grammar.commonMistake", point.CommonMistake)
			for _, example := range point.Examples {
				add("grammar.exampleMeaning", example.Meaning)
			}
		}
		attempt, err := s.StartGrammarAttempt(userID, summary.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, q := range attempt.Questions {
			add("grammar.prompt", q.Prompt)
			for _, option := range q.Options {
				add("grammar.choice", option.Text)
			}
			feedback, err := s.AnswerGrammarQuestion(userID, attempt.ID, q.ID, q.Options[0].ID)
			if err != nil {
				t.Fatal(err)
			}
			add("grammar.explanation", feedback.Explanation)
		}
	}
	covered := 0
	fieldTotals := map[string]int{}
	fieldCovered := map[string]int{}
	for source := range sources {
		fieldTotals[source.field]++
		if _, ok := englishJLPTVocabularyGrammar(source.field, source.source); ok {
			covered++
			fieldCovered[source.field]++
		} else {
			t.Logf("unmapped %s: %q", source.field, source.source)
		}
	}
	t.Logf("%s grammar coverage: %d/%d unique Indonesian field/source pairs", level, covered, len(sources))
	for field, total := range fieldTotals {
		t.Logf("  %s: %d/%d", field, fieldCovered[field], total)
	}
	if len(course.Lessons) == 0 || (expectedLessons > 0 && len(course.Lessons) != expectedLessons) {
		t.Fatalf("unexpected %s grammar lesson count: %d", level, len(course.Lessons))
	}
	if covered != len(sources) {
		t.Errorf("%s grammar source coverage is incomplete: covered %d of %d unique Indonesian field/source pairs", level, covered, len(sources))
	}
}

func TestN5VocabularyCatalogCoverage(t *testing.T) {
	testVocabularyCatalogCoverage(t, "n5", 160)
}

func TestN4VocabularyCatalogCoverage(t *testing.T) {
	testVocabularyCatalogCoverage(t, "n4", 0)
}

func TestVocabularyGrammarSharedTextCoverage(t *testing.T) {
	s := store.NewMemoryStore()
	userID := "shared-text-localization-inventory"
	missing := map[jlptEnglishSource]struct{}{}
	check := func(field, level, source string) {
		if source == "" || isJapaneseOnly(source) {
			return
		}
		shared := coverageReportSharedText{Field: field, Level: level, SourceText: source}
		if coverageReportSharedTextTranslation(shared).Missing {
			missing[jlptEnglishSource{field: coverageReportSharedTextField(shared), source: source}] = struct{}{}
		}
	}
	for _, level := range s.VocabularyLevels(userID) {
		for _, lesson := range level.Lessons {
			check("Lesson title", level.Level, lesson.Title)
			check("Lesson description", level.Level, lesson.Description)
		}
	}
	for _, level := range []string{"N5", "N4", "N3", "N2", "N1"} {
		course := s.GrammarCourse(userID, level)
		for _, summary := range course.Lessons {
			lesson, err := s.GrammarLesson(userID, summary.ID)
			if err != nil {
				t.Fatal(err)
			}
			check("Chapter label", level, lesson.Chapter)
			check("Lesson title", level, lesson.Title)
		}
	}
	t.Logf("vocabulary/grammar shared-text coverage: %d unique field/source pairs missing", len(missing))
	for source := range missing {
		t.Errorf("unmapped shared text %s: %q", source.field, source.source)
	}
}

func testVocabularyCatalogCoverage(t *testing.T, level string, expectedMeanings int) {
	t.Helper()
	type example struct {
		Meaning string `json:"meaning"`
	}
	type word struct {
		Meaning          string    `json:"meaning"`
		ExampleMeaning   string    `json:"exampleMeaning"`
		Note             string    `json:"note"`
		ImageDescription string    `json:"imageDescription"`
		Examples         []example `json:"examples"`
	}
	type lesson struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Words       []word `json:"words"`
	}
	data, err := os.ReadFile("../../data/vocabulary/" + level + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var lessons []lesson
	if err := json.Unmarshal(data, &lessons); err != nil {
		t.Fatal(err)
	}
	fields := map[string]map[string]struct{}{}
	add := func(field, source string) {
		if source == "" {
			return
		}
		if fields[field] == nil {
			fields[field] = map[string]struct{}{}
		}
		fields[field][source] = struct{}{}
	}
	for _, l := range lessons {
		add("vocabulary.title", l.Title)
		add("vocabulary.description", l.Description)
		for _, w := range l.Words {
			add("vocabulary.meaning", w.Meaning)
			add("vocabulary.exampleMeaning", w.ExampleMeaning)
			add("vocabulary.note", w.Note)
			add("vocabulary.imageDescription", w.ImageDescription)
			for _, e := range w.Examples {
				add("vocabulary.exampleMeaning", e.Meaning)
			}
		}
	}
	total, covered := 0, 0
	meaningTotal, meaningCovered := 0, 0
	fieldTotals, fieldCovered := map[string]int{}, map[string]int{}
	for field, sources := range fields {
		for source := range sources {
			total++
			fieldTotals[field]++
			_, ok := englishJLPTVocabularyGrammar(field, source)
			if ok {
				covered++
				fieldCovered[field]++
			} else {
				t.Logf("unmapped %s: %q", field, source)
			}
			if field == "vocabulary.meaning" {
				meaningTotal++
				if ok {
					meaningCovered++
				}
			}
		}
	}
	t.Logf("%s vocabulary catalog coverage: %d/%d unique field/source pairs", level, covered, total)
	t.Logf("%s vocabulary meaning coverage: %d/%d unique meanings", level, meaningCovered, meaningTotal)
	for field, fieldTotal := range fieldTotals {
		t.Logf("  %s: %d/%d", field, fieldCovered[field], fieldTotal)
	}
	t.Logf("Catalog size: %d unique field/source pairs", len(jlptEnglishCatalog))
	if covered == 0 || total == 0 {
		t.Fatalf("unexpected %s vocabulary inventory: %d covered of %d", level, covered, total)
	}
	if covered != total {
		t.Errorf("%s vocabulary source coverage is incomplete: covered %d of %d unique field/source pairs", level, covered, total)
	}
	if meaningCovered != meaningTotal || (expectedMeanings > 0 && meaningTotal != expectedMeanings) {
		t.Errorf("all %d %s vocabulary meanings must be translated; covered %d/%d", expectedMeanings, level, meaningCovered, meaningTotal)
	}
}
func TestJLPTVocabularyGrammarMaterialInventoryCoverage(t *testing.T) {
	bundlePath := os.Getenv("ENGLISH_COVERAGE_BUNDLE_PATH")
	if bundlePath == "" {
		t.Skip("set ENGLISH_COVERAGE_BUNDLE_PATH to verify JLPT learning materials")
	}
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	var bundle coverageReportBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
	}
	levels := map[string]bool{"N1": true, "N2": true, "N3": true, "N4": true, "N5": true}
	var total, missing int
	for _, material := range bundle.Materials {
		if material.Surface != "JLPT learning material" || !levels[material.Level] || (material.Category != "Vocabulary" && material.Category != "Grammar") {
			continue
		}
		total++
		translation := coverageReportMaterialTranslation(material)
		if translation.Status == "Needs translation" {
			missing++
			t.Errorf("unmapped JLPT %s %s material %s field %s: %q", material.Category, material.Level, material.ItemID, material.Field, material.SourceText)
		}
	}
	if total == 0 {
		t.Fatal("JLPT vocabulary and grammar material inventory is empty")
	}
	t.Logf("JLPT vocabulary and grammar materials: %d total, %d needing translation", total, missing)
}
