package store

import (
	"encoding/json"
	"fmt"
	grammarData "nihongo-master-backend/data/grammar"
)

func seedGrammarN3() []grammarLessonContent {
	var lessons []grammarLessonContent
	if err := json.Unmarshal(grammarData.N3, &lessons); err != nil {
		panic("invalid grammar N3: " + err.Error())
	}
	if len(lessons) != 36 {
		panic("incomplete grammar N3 curriculum")
	}
	for i := range lessons {
		l := &lessons[i]
		l.ID, l.Level = fmt.Sprintf("grammar-n3-%02d", i+1), "N3"
		if l.Title == "" || l.Chapter == "" || len(l.Points) != 2 || len(l.Questions) != 5 {
			panic("incomplete grammar N3 lesson: " + l.ID)
		}
		for _, p := range l.Points {
			if p.Pattern == "" || p.Meaning == "" || p.Formation == "" || p.Explanation == "" || p.CommonMistake == "" || len(p.Examples) != 2 {
				panic("incomplete grammar N3 point: " + l.ID)
			}
			for _, e := range p.Examples {
				if e.Japanese == "" || e.Reading == "" || e.Meaning == "" {
					panic("incomplete grammar N3 example: " + l.ID)
				}
			}
		}
		for _, q := range l.Questions {
			if len(q.Choices) != 4 || q.Correct < 0 || q.Correct >= 4 || q.Prompt == "" || q.Explanation == "" {
				panic("invalid grammar N3 question: " + l.ID)
			}
			seen := map[string]bool{}
			for _, c := range q.Choices {
				if c == "" || seen[c] {
					panic("duplicate grammar N3 choice: " + l.ID)
				}
				seen[c] = true
			}
		}
	}
	return lessons
}
