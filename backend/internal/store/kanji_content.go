package store

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"nihongo-master-backend/data/kanji"
)

func seedKanji() []kanjiLessonContent {
	lessons := seedKanjiLevel("N5", kanji.N5, 20, 100)
	lessons = append(lessons, seedKanjiLevel("N4", kanji.N4, 34, 166)...)
	lessons = append(lessons, seedKanjiLevel("N3", kanji.N3, 77, 367)...)
	lessons = append(lessons, seedKanjiLevel("N2", kanji.N2, 74, 367)...)
	return append(lessons, seedKanjiN1()...)
}

func seedKanjiLevel(level string, content []byte, lessonCount, characterCount int) []kanjiLessonContent {
	var lessons []kanjiLessonContent
	if err := json.Unmarshal(content, &lessons); err != nil {
		panic("invalid embedded kanji curriculum: " + err.Error())
	}
	seen := map[string]bool{}
	idWidth := 2
	if lessonCount >= 100 {
		idWidth = 3
	}
	for i, lesson := range lessons {
		lessons[i].Level = level
		if lesson.ID != fmt.Sprintf("kanji-%s-%0*d", strings.ToLower(level), idWidth, i+1) || lesson.Title == "" || lesson.Chapter == "" || len(lesson.Characters) < 4 || len(lesson.Characters) > 5 || len(lesson.Questions) != 2*len(lesson.Characters) {
			panic("invalid kanji lesson structure: " + lesson.ID)
		}
		for _, c := range lesson.Characters {
			if seen[c.Character] || len([]rune(c.Character)) != 1 || !unicode.Is(unicode.Han, []rune(c.Character)[0]) || c.Strokes < 1 || len(c.Examples) < 1 || len(c.Examples) > 2 || c.Meaning == "" || c.OnReading == "" || c.KunReading == "" || c.Mnemonic == "" || c.Note == "" {
				panic("invalid kanji character: " + c.Character)
			}
			seen[c.Character] = true
			for _, example := range c.Examples {
				if !strings.Contains(example.Japanese, c.Character) || example.Reading == "" || example.Meaning == "" {
					panic("invalid kanji example: " + c.Character)
				}
			}
		}
		for _, q := range lesson.Questions {
			choices := map[string]bool{}
			for _, choice := range q.Choices {
				if choice == "" || choices[choice] {
					panic("invalid kanji question choices: " + lesson.ID)
				}
				choices[choice] = true
			}
			if len(q.Choices) != 4 || q.Correct < 0 || q.Correct >= len(q.Choices) || q.Prompt == "" || q.Explanation == "" {
				panic("invalid kanji question: " + lesson.ID)
			}
		}
	}
	if len(lessons) != lessonCount || len(seen) != characterCount {
		panic("incomplete kanji " + level + " curriculum")
	}
	return lessons
}
