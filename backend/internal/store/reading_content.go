package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"nihongo-master-backend/data/reading"
)

func seedReading() []readingLessonContent {
	lessons := seedReadingLevel("N5", reading.N5)
	lessons = append(lessons, seedReadingLevel("N4", reading.N4)...)
	lessons = append(lessons, seedReadingLevel("N3", reading.N3)...)
	lessons = append(lessons, seedReadingLevel("N2", reading.N2)...)
	return append(lessons, seedReadingLevel("N1", reading.N1)...)
}

func seedReadingLevel(level string, content []byte) []readingLessonContent {
	var lessons []readingLessonContent
	if err := json.Unmarshal(content, &lessons); err != nil {
		panic("invalid embedded reading curriculum: " + err.Error())
	}
	if len(lessons) != 12 {
		panic("incomplete reading curriculum")
	}
	for i, l := range lessons {
		lessons[i].Level = level
		if l.ID != fmt.Sprintf("reading-%s-%02d", strings.ToLower(level), i+1) || len(l.Pages) != 3 || len(l.Questions) != 5 || l.Title == "" || l.JapaneseTitle == "" || l.Minutes < 1 {
			panic("invalid reading lesson: " + l.ID)
		}
		for j, p := range l.Pages {
			if p.Japanese == "" || p.Reading == "" || p.Meaning == "" || p.Title == "" || p.ImagePanel != j || p.ImageAsset != l.ImageAsset || !strings.HasPrefix(p.ImageAsset, "assets/images/reading/") || len(p.Vocabulary) != 2 || p.ImageDescription == "" {
				panic("invalid reading page: " + l.ID)
			}
			for _, v := range p.Vocabulary {
				if v.Japanese == "" || v.Reading == "" || v.Meaning == "" {
					panic("invalid reading word: " + l.ID)
				}
			}
		}
		for _, q := range l.Questions {
			if q.Prompt == "" || q.Explanation == "" || len(q.Choices) != 4 || q.Correct < 0 || q.Correct >= len(q.Choices) {
				panic("invalid reading question: " + l.ID)
			}
			seen := map[string]bool{}
			for _, c := range q.Choices {
				if c == "" || seen[c] {
					panic("invalid reading choices: " + l.ID)
				}
				seen[c] = true
			}
		}
	}
	return lessons
}
