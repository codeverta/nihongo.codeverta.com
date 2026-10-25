package api

import (
	"encoding/json"
	"testing"
	"unicode"

	"nihongo-master-backend/data/kanji"
	"nihongo-master-backend/data/reading"
)

func TestTranslateJLPTKanjiReadingText(t *testing.T) {
	tests := []struct {
		field, source, want string
	}{
		{"chapter", "Keseharian", "Daily life"},
		{"title", "Pagi bersama Hana", "A morning with Hana"},
		{"description", "Ikuti Hana dari bangun tidur hingga berangkat sekolah.", "Follow Hana from waking up to leaving for school."},
		{"mnemonic", "Satu garis mewakili satu.", "One line represents one."},
		{"prompt", "Apa arti kanji 一?", "What does the kanji 一 mean?"},
		{"explanation", "Kalimat 毎朝七時に起きます menyebut Hana bangun pukul tujuh.", "The sentence 毎朝七時に起きます says that Hana wakes up at seven."},
	}
	for _, test := range tests {
		if got := translateJLPTKanjiReadingText(test.field, test.source); got != test.want {
			t.Errorf("translateJLPTKanjiReadingText(%q, %q) = %q, want %q", test.field, test.source, got, test.want)
		}
	}
}

func TestN5KanjiCoreContentHasReviewedEnglish(t *testing.T) {
	type example struct {
		Meaning string `json:"meaning"`
	}
	type character struct {
		Meaning  string    `json:"meaning"`
		Mnemonic string    `json:"mnemonic"`
		Note     string    `json:"note"`
		Examples []example `json:"examples"`
	}
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Title      string      `json:"Title"`
		Chapter    string      `json:"Chapter"`
		Characters []character `json:"Characters"`
		Questions  []question  `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N5, &lessons); err != nil {
		t.Fatal(err)
	}
	check := func(field, source string) {
		t.Helper()
		_, ok := jlptKanjiReadingEnglish[field][source]
		if !ok {
			t.Errorf("missing reviewed N5 Kanji translation for %s %q", field, source)
		}
	}
	for _, l := range lessons {
		check("title", l.Title)
		check("chapter", l.Chapter)
		for _, c := range l.Characters {
			check("meaning", c.Meaning)
			check("mnemonic", c.Mnemonic)
			check("note", c.Note)
			for _, e := range c.Examples {
				check("meaning", e.Meaning)
			}
		}
		for _, q := range l.Questions {
			check("prompt", q.Prompt)
			check("explanation", q.Explanation)
			for _, choice := range q.Choices {
				if _, ok := jlptKanjiReadingEnglish["choices"][choice]; ok {
					continue
				}
				containsJapanese := false
				for _, r := range choice {
					if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
						containsJapanese = true
						break
					}
				}
				if !containsJapanese || translateJLPTKanjiReadingText("choices", choice) != choice {
					t.Errorf("untranslated N5 Kanji choice is not preserved Japanese: %q", choice)
				}
			}
		}
	}
}

func TestN4KanjiCoreAndQuizContentHasReviewedEnglish(t *testing.T) {
	preservedMeaning := map[string]struct{}{"Tokyo": {}, "Kyoto": {}, "kanji": {}, "bank": {}}
	type example struct {
		Meaning string `json:"meaning"`
	}
	type character struct {
		Meaning  string    `json:"meaning"`
		Note     string    `json:"note"`
		Mnemonic string    `json:"mnemonic"`
		Examples []example `json:"examples"`
	}
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Title      string      `json:"Title"`
		Chapter    string      `json:"Chapter"`
		Characters []character `json:"Characters"`
		Questions  []question  `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N4, &lessons); err != nil {
		t.Fatal(err)
	}
	for _, lesson := range lessons {
		for field, source := range map[string]string{"title": lesson.Title, "chapter": lesson.Chapter} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N4 Kanji translation for %s %q", field, source)
			}
		}
		for _, character := range lesson.Characters {
			checkMeaning := func(source string) {
				translated, ok := jlptKanjiReadingEnglish["meaning"][source]
				if !ok {
					t.Errorf("missing reviewed N4 Kanji translation or preserve decision for meaning %q", source)
					return
				}
				if translated == source {
					if _, preserve := preservedMeaning[source]; !preserve {
						t.Errorf("N4 Kanji meaning unexpectedly unchanged without a preserve decision: %q", source)
					}
				}
			}
			checkMeaning(character.Meaning)
			for _, example := range character.Examples {
				checkMeaning(example.Meaning)
			}
			if translated, ok := jlptKanjiReadingEnglish["mnemonic"][character.Mnemonic]; !ok || translated == character.Mnemonic {
				t.Errorf("missing reviewed N4 Kanji translation for mnemonic %q", character.Mnemonic)
			}
			if character.Note == "" {
				continue
			}
			translated, ok := jlptKanjiReadingEnglish["note"][character.Note]
			if !ok || translated == character.Note {
				t.Errorf("missing reviewed N4 Kanji translation for note %q", character.Note)
			}
		}
		for _, question := range lesson.Questions {
			if translated, ok := jlptKanjiReadingEnglish["prompt"][question.Prompt]; !ok || translated == question.Prompt {
				t.Errorf("missing reviewed N4 Kanji prompt translation for %q", question.Prompt)
			}
			if translated, ok := jlptKanjiReadingEnglish["explanation"][question.Explanation]; !ok || translated == question.Explanation {
				t.Errorf("missing reviewed N4 Kanji explanation translation for %q", question.Explanation)
			}
			for _, choice := range question.Choices {
				if translated, ok := jlptKanjiReadingEnglish["choices"][choice]; ok {
					if translated == choice {
						t.Errorf("N4 Kanji answer choice was marked translated but unchanged: %q", choice)
					}
					continue
				}
				containsJapanese := false
				for _, r := range choice {
					if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
						containsJapanese = true
						break
					}
				}
				if !containsJapanese || translateJLPTKanjiReadingText("choices", choice) != choice {
					t.Errorf("untranslated N4 Kanji choice is not preserved Japanese: %q", choice)
				}
			}
		}
	}
}

func TestN3KanjiLessonLabelsHaveReviewedEnglish(t *testing.T) {
	type lesson struct {
		Title   string `json:"Title"`
		Chapter string `json:"Chapter"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	titles, chapters := map[string]struct{}{}, map[string]struct{}{}
	for _, l := range lessons {
		for field, source := range map[string]string{"title": l.Title, "chapter": l.Chapter} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		titles[l.Title] = struct{}{}
		chapters[l.Chapter] = struct{}{}
	}
	if len(titles) != 77 || len(chapters) != 9 {
		t.Fatalf("unexpected N3 Kanji label inventory: %d unique titles, %d unique chapters", len(titles), len(chapters))
	}
}

func TestN3KanjiEveryLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	missing := map[string]map[string]struct{}{}
	check := func(field, source string) {
		t.Helper()
		if translated, ok := jlptKanjiReadingEnglish[field][source]; !ok || translated == source {
			if missing[field] == nil {
				missing[field] = map[string]struct{}{}
			}
			missing[field][source] = struct{}{}
		}
	}
	questionCount := 0
	for _, l := range lessons {
		for _, q := range l.Questions {
			questionCount++
			check("prompt", q.Prompt)
			check("explanation", q.Explanation)
			for _, choice := range q.Choices {
				if _, translated := jlptKanjiReadingEnglish["choices"][choice]; translated {
					continue
				}
				containsJapanese := false
				for _, r := range choice {
					if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
						containsJapanese = true
						break
					}
				}
				if !containsJapanese {
					check("choices", choice)
				}
			}
		}
	}
	if len(lessons) != 77 || questionCount != 734 {
		t.Errorf("N3 Kanji coverage has %d lessons and %d questions", len(lessons), questionCount)
	}
	for field, sources := range missing {
		sample := make([]string, 0, 5)
		for source := range sources {
			if len(sample) == cap(sample) {
				break
			}
			sample = append(sample, source)
		}
		t.Errorf("N3 Kanji has %d untranslated %s sources; examples: %q", len(sources), field, sample)
	}
}

func TestN3KanjiEveryCharacterAndExampleHasReviewedEnglish(t *testing.T) {
	type example struct {
		Meaning string `json:"meaning"`
	}
	type character struct {
		Meaning  string    `json:"meaning"`
		Mnemonic string    `json:"mnemonic"`
		Note     string    `json:"note"`
		Examples []example `json:"examples"`
	}
	type lesson struct {
		Characters []character `json:"Characters"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	missing := map[string]map[string]struct{}{}
	check := func(field, source string) {
		t.Helper()
		if translated, ok := jlptKanjiReadingEnglish[field][source]; !ok || translated == source {
			if missing[field] == nil {
				missing[field] = map[string]struct{}{}
			}
			missing[field][source] = struct{}{}
		}
	}
	characters, examples := 0, 0
	for _, l := range lessons {
		for _, c := range l.Characters {
			characters++
			check("meaning", c.Meaning)
			check("mnemonic", c.Mnemonic)
			check("note", c.Note)
			for _, e := range c.Examples {
				examples++
				check("meaning", e.Meaning)
			}
		}
	}
	if len(lessons) != 77 || characters != 367 || examples != 734 {
		t.Errorf("N3 Kanji material coverage has %d lessons, %d characters, and %d examples", len(lessons), characters, examples)
	}
	for field, sources := range missing {
		sample := make([]string, 0, 5)
		for source := range sources {
			if len(sample) == cap(sample) {
				break
			}
			sample = append(sample, source)
		}
		t.Errorf("N3 Kanji has %d untranslated %s sources; examples: %q", len(sources), field, sample)
	}
}

func TestN3KanjiFirstEightLessonsQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 8 || len(lessons[0].Questions) != 10 || len(lessons[1].Questions) != 10 || len(lessons[2].Questions) != 10 || len(lessons[3].Questions) != 8 || len(lessons[4].Questions) != 10 || len(lessons[5].Questions) != 10 || len(lessons[6].Questions) != 10 || len(lessons[7].Questions) != 10 {
		t.Fatalf("unexpected question counts among the first eight N3 Kanji lessons")
	}
	for _, lesson := range lessons[:8] {
		for _, q := range lesson.Questions {
			for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
				translated, ok := jlptKanjiReadingEnglish[field][source]
				if !ok || translated == source {
					t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
				}
			}
			for _, choice := range q.Choices {
				containsJapanese := false
				for _, r := range choice {
					if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
						containsJapanese = true
						break
					}
				}
				if containsJapanese {
					if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
						t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
					}
				} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
					t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
				}
			}
		}
	}
}

func TestN3KanjiNinthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 9 || len(lessons[8].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the ninth N3 Kanji lesson")
	}
	for _, q := range lessons[8].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiTenthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 10 || len(lessons[9].Questions) != 8 {
		t.Fatalf("expected 8 quiz questions in the tenth N3 Kanji lesson")
	}
	for _, q := range lessons[9].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiEleventhLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 11 || len(lessons[10].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the eleventh N3 Kanji lesson")
	}
	for _, q := range lessons[10].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiTwelfthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 12 || len(lessons[11].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the twelfth N3 Kanji lesson")
	}
	for _, q := range lessons[11].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiThirteenthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 13 || len(lessons[12].Questions) != 8 {
		t.Fatalf("expected 8 quiz questions in the thirteenth N3 Kanji lesson")
	}
	for _, q := range lessons[12].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiFourteenthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 14 || len(lessons[13].Questions) != 8 {
		t.Fatalf("expected 8 quiz questions in the fourteenth N3 Kanji lesson")
	}
	for _, q := range lessons[13].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiFifteenthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 15 || len(lessons[14].Questions) != 8 {
		t.Fatalf("expected 8 quiz questions in the fifteenth N3 Kanji lesson")
	}
	for _, q := range lessons[14].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiSixteenthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 16 || len(lessons[15].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the sixteenth N3 Kanji lesson")
	}
	for _, q := range lessons[15].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiSeventeenthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 17 || len(lessons[16].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the seventeenth N3 Kanji lesson")
	}
	for _, q := range lessons[16].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiEighteenthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 18 || len(lessons[17].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the eighteenth N3 Kanji lesson")
	}
	for _, q := range lessons[17].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiNineteenthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 19 || len(lessons[18].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the nineteenth N3 Kanji lesson")
	}
	for _, q := range lessons[18].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiTwentiethLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 20 || len(lessons[19].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the twentieth N3 Kanji lesson")
	}
	for _, q := range lessons[19].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiTwentyFirstLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 21 || len(lessons[20].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the twenty-first N3 Kanji lesson")
	}
	for _, q := range lessons[20].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiTwentySecondLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 22 || len(lessons[21].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the twenty-second N3 Kanji lesson")
	}
	for _, q := range lessons[21].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiTwentyThirdLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 23 || len(lessons[22].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the twenty-third N3 Kanji lesson")
	}
	for _, q := range lessons[22].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiTwentyFourthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 24 || len(lessons[23].Questions) != 10 {
		t.Fatalf("expected 10 quiz questions in the twenty-fourth N3 Kanji lesson")
	}
	for _, q := range lessons[23].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiLessons25And26QuizHaveReviewedEnglishAndPreserveJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 26 {
		t.Fatal("expected at least 26 N3 Kanji lessons")
	}
	for index := 24; index <= 25; index++ {
		if len(lessons[index].Questions) != 10 {
			t.Errorf("expected 10 quiz questions in N3 Kanji lesson %d", index+1)
		}
		for _, q := range lessons[index].Questions {
			for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
				translated, ok := jlptKanjiReadingEnglish[field][source]
				if !ok || translated == source {
					t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
				}
			}
			for _, choice := range q.Choices {
				containsJapanese := false
				for _, r := range choice {
					if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
						containsJapanese = true
						break
					}
				}
				if containsJapanese {
					if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
						t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
					}
				} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
					t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
				}
			}
		}
	}
}

func TestN3KanjiLessons27And28QuizHaveReviewedEnglishAndPreserveJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 28 {
		t.Fatal("expected at least 28 N3 Kanji lessons")
	}
	for index := 26; index <= 27; index++ {
		if len(lessons[index].Questions) != 8 {
			t.Errorf("expected 8 quiz questions in N3 Kanji lesson %d", index+1)
		}
		for _, q := range lessons[index].Questions {
			for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
				translated, ok := jlptKanjiReadingEnglish[field][source]
				if !ok || translated == source {
					t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
				}
			}
			for _, choice := range q.Choices {
				containsJapanese := false
				for _, r := range choice {
					if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
						containsJapanese = true
						break
					}
				}
				if containsJapanese {
					if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
						t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
					}
				} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
					t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
				}
			}
		}
	}
}

func TestN3KanjiLessons29And30QuizHaveReviewedEnglishAndPreserveJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 30 {
		t.Fatal("expected at least 30 N3 Kanji lessons")
	}
	for index, count := range map[int]int{28: 8, 29: 10} {
		if len(lessons[index].Questions) != count {
			t.Errorf("expected %d quiz questions in N3 Kanji lesson %d", count, index+1)
		}
		for _, q := range lessons[index].Questions {
			for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
				translated, ok := jlptKanjiReadingEnglish[field][source]
				if !ok || translated == source {
					t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
				}
			}
			for _, choice := range q.Choices {
				containsJapanese := false
				for _, r := range choice {
					if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
						containsJapanese = true
						break
					}
				}
				if containsJapanese {
					if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
						t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
					}
				} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
					t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
				}
			}
		}
	}
}

func TestN3KanjiThirtyFirstLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 31 || len(lessons[30].Questions) != 10 {
		t.Fatal("expected 10 quiz questions in the thirty-first N3 Kanji lesson")
	}
	for _, q := range lessons[30].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiThirtySecondLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 32 || len(lessons[31].Questions) != 10 {
		t.Fatal("expected 10 quiz questions in the thirty-second N3 Kanji lesson")
	}
	for _, q := range lessons[31].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiThirtyThirdLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 33 || len(lessons[32].Questions) != 10 {
		t.Fatal("expected 10 quiz questions in the thirty-third N3 Kanji lesson")
	}
	for _, q := range lessons[32].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiThirtyFourthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 34 || len(lessons[33].Questions) != 10 {
		t.Fatal("expected 10 quiz questions in the thirty-fourth N3 Kanji lesson")
	}
	for _, q := range lessons[33].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiThirtyFifthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 35 || len(lessons[34].Questions) != 10 {
		t.Fatal("expected 10 quiz questions in the thirty-fifth N3 Kanji lesson")
	}
	for _, q := range lessons[34].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiThirtySixthLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 36 || len(lessons[35].Questions) != 8 {
		t.Fatal("expected 8 quiz questions in the thirty-sixth N3 Kanji lesson")
	}
	for _, q := range lessons[35].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN3KanjiThirtySeventhLessonQuizHasReviewedEnglishAndPreservesJapanese(t *testing.T) {
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Questions []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(kanji.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	if len(lessons) < 37 || len(lessons[36].Questions) != 8 {
		t.Fatal("expected 8 quiz questions in the thirty-seventh N3 Kanji lesson")
	}
	for _, q := range lessons[36].Questions {
		for field, source := range map[string]string{"prompt": q.Prompt, "explanation": q.Explanation} {
			translated, ok := jlptKanjiReadingEnglish[field][source]
			if !ok || translated == source {
				t.Errorf("missing reviewed N3 Kanji %s translation for %q", field, source)
			}
		}
		for _, choice := range q.Choices {
			containsJapanese := false
			for _, r := range choice {
				if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
					containsJapanese = true
					break
				}
			}
			if containsJapanese {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Kanji answer choice changed: %q => %q", choice, got)
				}
			} else if got := translateJLPTKanjiReadingText("choices", choice); got == choice {
				t.Errorf("missing reviewed N3 Kanji choice translation for %q", choice)
			}
		}
	}
}

func TestN5ReadingContentHasReviewedEnglishForEveryLearnerFacingField(t *testing.T) {
	type word struct {
		Meaning string `json:"meaning"`
	}
	type page struct {
		Title            string `json:"title"`
		Meaning          string `json:"meaning"`
		ImageDescription string `json:"imageDescription"`
		Vocabulary       []word `json:"vocabulary"`
	}
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Title       string     `json:"Title"`
		Chapter     string     `json:"Chapter"`
		Description string     `json:"Description"`
		Pages       []page     `json:"Pages"`
		Questions   []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(reading.N5, &lessons); err != nil {
		t.Fatal(err)
	}
	check := func(field, source string) {
		t.Helper()
		translated, ok := jlptKanjiReadingEnglish[field][source]
		if !ok || translated == source {
			t.Errorf("missing reviewed N5 Reading translation for %s %q", field, source)
		}
	}
	for _, l := range lessons {
		check("title", l.Title)
		check("chapter", l.Chapter)
		check("description", l.Description)
		for _, p := range l.Pages {
			check("title", p.Title)
			check("meaning", p.Meaning)
			check("imageDescription", p.ImageDescription)
			for _, v := range p.Vocabulary {
				check("meaning", v.Meaning)
			}
		}
		for _, q := range l.Questions {
			check("prompt", q.Prompt)
			check("explanation", q.Explanation)
			for _, choice := range q.Choices {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N5 Reading answer choice changed: %q => %q", choice, got)
				}
			}
		}
	}
}

func TestN4ReadingContentHasReviewedEnglishForEveryLearnerFacingField(t *testing.T) {
	type word struct {
		Meaning string `json:"meaning"`
	}
	type page struct {
		Title            string `json:"title"`
		Meaning          string `json:"meaning"`
		ImageDescription string `json:"imageDescription"`
		Vocabulary       []word `json:"vocabulary"`
	}
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Title       string     `json:"Title"`
		Chapter     string     `json:"Chapter"`
		Description string     `json:"Description"`
		Pages       []page     `json:"Pages"`
		Questions   []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(reading.N4, &lessons); err != nil {
		t.Fatal(err)
	}
	check := func(field, source string) {
		t.Helper()
		translated, ok := jlptKanjiReadingEnglish[field][source]
		if !ok || translated == source {
			t.Errorf("missing reviewed N4 Reading translation for %s %q", field, source)
		}
	}
	for _, l := range lessons {
		check("title", l.Title)
		check("chapter", l.Chapter)
		check("description", l.Description)
		for _, p := range l.Pages {
			check("title", p.Title)
			check("meaning", p.Meaning)
			check("imageDescription", p.ImageDescription)
			for _, v := range p.Vocabulary {
				check("meaning", v.Meaning)
			}
		}
		for _, q := range l.Questions {
			check("prompt", q.Prompt)
			check("explanation", q.Explanation)
			for _, choice := range q.Choices {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N4 Reading answer choice changed: %q", choice)
				}
			}
		}
	}
}

func TestN3ReadingContentHasReviewedEnglishForEveryLearnerFacingField(t *testing.T) {
	type word struct {
		Meaning string `json:"meaning"`
	}
	type page struct {
		Title            string `json:"title"`
		Meaning          string `json:"meaning"`
		ImageDescription string `json:"imageDescription"`
		Vocabulary       []word `json:"vocabulary"`
	}
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Title       string     `json:"Title"`
		Chapter     string     `json:"Chapter"`
		Description string     `json:"Description"`
		Pages       []page     `json:"Pages"`
		Questions   []question `json:"Questions"`
	}
	var lessons []lesson
	if err := json.Unmarshal(reading.N3, &lessons); err != nil {
		t.Fatal(err)
	}
	check := func(field, source string) {
		t.Helper()
		translated, mapped := jlptKanjiReadingEnglish[field][source]
		if mapped {
			if translated == source {
				t.Errorf("N3 Reading translation is unchanged for %s %q", field, source)
			}
			return
		}
		for _, r := range source {
			if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
				if got := translateJLPTKanjiReadingText(field, source); got != source {
					t.Errorf("Japanese N3 Reading source changed for %s: %q => %q", field, source, got)
				}
				return
			}
		}
		t.Errorf("missing reviewed N3 Reading translation for %s %q", field, source)
	}
	for _, l := range lessons {
		check("title", l.Title)
		check("chapter", l.Chapter)
		check("description", l.Description)
		for _, p := range l.Pages {
			check("title", p.Title)
			check("meaning", p.Meaning)
			check("imageDescription", p.ImageDescription)
			for _, v := range p.Vocabulary {
				check("meaning", v.Meaning)
			}
		}
		for _, q := range l.Questions {
			check("prompt", q.Prompt)
			check("explanation", q.Explanation)
			for _, choice := range q.Choices {
				if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
					t.Errorf("Japanese N3 Reading answer choice changed: %q", choice)
				}
			}
		}
	}
}

func TestN2N1ReadingLessonLabelsHaveReviewedEnglish(t *testing.T) {
	type lesson struct {
		Title   string `json:"Title"`
		Chapter string `json:"Chapter"`
	}
	for level, source := range map[string][]byte{"N2": reading.N2, "N1": reading.N1} {
		var lessons []lesson
		if err := json.Unmarshal(source, &lessons); err != nil {
			t.Fatalf("decode %s Reading lessons: %v", level, err)
		}
		if len(lessons) != 12 {
			t.Errorf("expected 12 %s Reading lessons, got %d", level, len(lessons))
		}
		for _, item := range lessons {
			for field, value := range map[string]string{"title": item.Title, "chapter": item.Chapter} {
				translated, ok := jlptKanjiReadingEnglish[field][value]
				if !ok || translated == value {
					t.Errorf("missing reviewed %s Reading %s translation for %q", level, field, value)
				}
			}
		}
	}
}

func TestN1KanjiChapterLabelHasReviewedEnglish(t *testing.T) {
	if got := translateJLPTKanjiReadingText("chapter", "Kanji N1"); got != "N1 Kanji" {
		t.Fatalf("N1 Kanji chapter label = %q, want %q", got, "N1 Kanji")
	}
}

func TestN1N2KanjiLearnerFacingContentHasReviewedEnglish(t *testing.T) {
	type example struct {
		Meaning string `json:"meaning"`
	}
	type character struct {
		Meaning  string    `json:"meaning"`
		Mnemonic string    `json:"mnemonic"`
		Note     string    `json:"note"`
		Examples []example `json:"examples"`
	}
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Title      string      `json:"Title"`
		Chapter    string      `json:"Chapter"`
		Characters []character `json:"Characters"`
		Questions  []question  `json:"Questions"`
	}
	for level, source := range map[string][]byte{"N1": kanji.N1, "N2": kanji.N2} {
		var lessons []lesson
		if err := json.Unmarshal(source, &lessons); err != nil {
			t.Fatalf("decode %s Kanji: %v", level, err)
		}
		questions := 0
		missing := map[string]map[string]struct{}{}
		check := func(field, text string) {
			t.Helper()
			if translated, ok := jlptKanjiReadingEnglish[field][text]; !ok || translated == text {
				if missing[field] == nil {
					missing[field] = map[string]struct{}{}
				}
				missing[field][text] = struct{}{}
			}
		}
		for _, l := range lessons {
			check("title", l.Title)
			check("chapter", l.Chapter)
			for _, c := range l.Characters {
				if got := translateJLPTKanjiReadingText("meaning", c.Meaning); got != c.Meaning {
					t.Errorf("%s Kanji English gloss changed: %q => %q", level, c.Meaning, got)
				}
				check("mnemonic", c.Mnemonic)
				check("note", c.Note)
				for _, example := range c.Examples {
					if got := translateJLPTKanjiReadingText("meaning", example.Meaning); got != example.Meaning {
						t.Errorf("%s Kanji English example gloss changed: %q => %q", level, example.Meaning, got)
					}
				}
			}
			for _, q := range l.Questions {
				questions++
				check("prompt", q.Prompt)
				check("explanation", q.Explanation)
				for _, choice := range q.Choices {
					if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
						t.Errorf("%s Kanji English gloss or Japanese reading changed: %q => %q", level, choice, got)
					}
				}
			}
		}
		if level == "N1" && (len(lessons) != 247 || questions != 2464) {
			t.Errorf("N1 Kanji coverage has %d lessons and %d questions", len(lessons), questions)
		}
		if level == "N2" && (len(lessons) != 74 || questions != 734) {
			t.Errorf("N2 Kanji coverage has %d lessons and %d questions", len(lessons), questions)
		}
		for field, sources := range missing {
			sample := make([]string, 0, 5)
			for text := range sources {
				if len(sample) == cap(sample) {
					break
				}
				sample = append(sample, text)
			}
			t.Errorf("%s Kanji has %d untranslated %s sources; examples: %q", level, len(sources), field, sample)
		}
	}
}

func TestN1N2ReadingLearnerFacingContentHasReviewedEnglish(t *testing.T) {
	type word struct {
		Meaning string `json:"meaning"`
	}
	type page struct {
		Title            string `json:"title"`
		Meaning          string `json:"meaning"`
		ImageDescription string `json:"imageDescription"`
		Vocabulary       []word `json:"vocabulary"`
	}
	type question struct {
		Prompt      string   `json:"Prompt"`
		Choices     []string `json:"Choices"`
		Explanation string   `json:"Explanation"`
	}
	type lesson struct {
		Title       string     `json:"Title"`
		Chapter     string     `json:"Chapter"`
		Description string     `json:"Description"`
		Pages       []page     `json:"Pages"`
		Questions   []question `json:"Questions"`
	}
	for level, source := range map[string][]byte{"N1": reading.N1, "N2": reading.N2} {
		var lessons []lesson
		if err := json.Unmarshal(source, &lessons); err != nil {
			t.Fatalf("decode %s Reading: %v", level, err)
		}
		questions := 0
		missing := map[string]map[string]struct{}{}
		check := func(field, text string) {
			t.Helper()
			if translated, ok := jlptKanjiReadingEnglish[field][text]; !ok || translated == text {
				if missing[field] == nil {
					missing[field] = map[string]struct{}{}
				}
				missing[field][text] = struct{}{}
			}
		}
		checkMappedOrPreserved := func(field, text string) {
			t.Helper()
			if _, mapped := jlptKanjiReadingEnglish[field][text]; mapped {
				check(field, text)
				return
			}
			if got := translateJLPTKanjiReadingText(field, text); got != text {
				t.Errorf("Japanese %s Reading %s changed: %q => %q", level, field, text, got)
			}
		}
		for _, l := range lessons {
			check("title", l.Title)
			check("chapter", l.Chapter)
			check("description", l.Description)
			for _, p := range l.Pages {
				if level == "N2" {
					check("title", p.Title)
				} else if got := translateJLPTKanjiReadingText("title", p.Title); got != p.Title {
					t.Errorf("Japanese N1 Reading page title changed: %q => %q", p.Title, got)
				}
				check("meaning", p.Meaning)
				if p.ImageDescription != "" {
					check("imageDescription", p.ImageDescription)
				}
				for _, v := range p.Vocabulary {
					check("meaning", v.Meaning)
				}
			}
			for _, q := range l.Questions {
				questions++
				checkMappedOrPreserved("prompt", q.Prompt)
				for _, choice := range q.Choices {
					if got := translateJLPTKanjiReadingText("choices", choice); got != choice {
						t.Errorf("Japanese %s Reading choice changed: %q", level, choice)
					}
				}
				if level == "N2" {
					check("explanation", q.Explanation)
				} else {
					checkMappedOrPreserved("explanation", q.Explanation)
				}
			}
		}
		if len(lessons) != 12 || questions != 60 {
			t.Errorf("%s Reading coverage has %d lessons and %d questions", level, len(lessons), questions)
		}
		for field, sources := range missing {
			sample := make([]string, 0, 5)
			for text := range sources {
				if len(sample) == cap(sample) {
					break
				}
				sample = append(sample, text)
			}
			t.Errorf("%s Reading has %d untranslated %s sources; examples: %q", level, len(sources), field, sample)
		}
	}
}

func TestTranslateJLPTKanjiReadingTextPreservesJapaneseAndUnknownText(t *testing.T) {
	const japanese = "はなさんは毎朝七時に起きます。"
	if got := translateJLPTKanjiReadingText("japanese", japanese); got != japanese {
		t.Fatalf("Japanese source was changed: got %q, want %q", got, japanese)
	}
	const mixedPrompt = "はなさんは何時に起きますか。\nHana bangun pukul berapa?"
	const translated = "はなさんは何時に起きますか。\nWhat time does Hana wake up?"
	if got := translateJLPTKanjiReadingText("prompt", mixedPrompt); got != translated {
		t.Fatalf("Japanese portion of prompt was not preserved: got %q, want %q", got, translated)
	}
	const unreviewed = "belum ada terjemahan"
	if got := translateJLPTKanjiReadingText("meaning", unreviewed); got != unreviewed {
		t.Fatalf("unreviewed text changed: got %q, want %q", got, unreviewed)
	}
}
