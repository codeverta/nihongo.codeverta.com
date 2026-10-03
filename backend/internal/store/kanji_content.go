package store

import (
	"encoding/json"
	"fmt"

	"nihongo-master-backend/data/kanji"
)

func seedKanji() []kanjiLessonContent {
	var lessons []kanjiLessonContent
	if err := json.Unmarshal(kanji.N5, &lessons); err != nil {
		panic("invalid embedded kanji curriculum: " + err.Error())
	}
	seen := map[string]bool{}
	for i, lesson := range lessons {
		if lesson.ID != fmt.Sprintf("kanji-n5-%02d", i+1) || len(lesson.Characters) != 5 || len(lesson.Questions) != 10 {
			panic("invalid kanji lesson structure: " + lesson.ID)
		}
		for _, c := range lesson.Characters {
			if seen[c.Character] || len([]rune(c.Character)) != 1 || c.Strokes < 1 || len(c.Examples) != 2 || c.Meaning == "" || c.OnReading == "" || c.KunReading == "" || c.Mnemonic == "" {
				panic("invalid kanji character: " + c.Character)
			}
			seen[c.Character] = true
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
	if len(lessons) != 20 || len(seen) != 100 {
		panic("incomplete kanji N5 curriculum")
	}
	return lessons
}
