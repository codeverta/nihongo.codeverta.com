package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"nihongo-master-backend/internal/aichat"
	"nihongo-master-backend/internal/domain"
)

// maxLessonContextRunes bounds the lesson notes sent to the model so a very
// large lesson can never blow up the token bill.
const maxLessonContextRunes = 12000

var aiLessonIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

var aiLessonTypes = map[string]bool{"grammar": true, "vocabulary": true, "kanji": true, "reading": true}

// aiLessonPath is the public lesson route, reused for premium gating and the
// English content translation so the AI sees exactly what the app shows.
func aiLessonPath(lessonType, lessonID string) string {
	return "/v1/" + lessonType + "/lessons/" + lessonID
}

// aiLessonContext builds the compact, server-owned lesson notes for the
// prompt. Results are memoized per lesson and locale: lesson content is
// static for the life of the process.
func (s *Server) aiLessonContext(userID, lessonType, lessonID, locale string) (aichat.Lesson, error) {
	key := lessonType + ":" + lessonID
	cacheKey := key + "|" + locale
	if cached, ok := s.ai.lessons.Load(cacheKey); ok {
		return cached.(aichat.Lesson), nil
	}
	var (
		lesson any
		err    error
	)
	switch lessonType {
	case "grammar":
		lesson, err = s.store.GrammarLesson(userID, lessonID)
	case "vocabulary":
		lesson, err = s.store.VocabularyLesson(userID, lessonID)
	case "kanji":
		lesson, err = s.store.KanjiLesson(userID, lessonID)
	case "reading":
		lesson, err = s.store.ReadingLesson(userID, lessonID)
	default:
		return aichat.Lesson{}, fmt.Errorf("unknown lesson type %q", lessonType)
	}
	if err != nil {
		return aichat.Lesson{}, err
	}
	body, err := json.Marshal(map[string]any{"data": lesson})
	if err != nil {
		return aichat.Lesson{}, err
	}
	if locale == localeEnglish {
		if translated, terr := translateLearningJSONForPath(body, aiLessonPath(lessonType, lessonID)); terr == nil {
			body = translated
		}
	}
	var title, content string
	switch lessonType {
	case "grammar":
		var v struct{ Data domain.GrammarLesson }
		if err = json.Unmarshal(body, &v); err == nil {
			title, content = v.Data.Title, compactGrammar(v.Data)
		}
	case "vocabulary":
		var v struct{ Data domain.VocabularyLesson }
		if err = json.Unmarshal(body, &v); err == nil {
			title, content = v.Data.Title, compactVocabulary(v.Data)
		}
	case "kanji":
		var v struct{ Data domain.KanjiLesson }
		if err = json.Unmarshal(body, &v); err == nil {
			title, content = v.Data.Title, compactKanji(v.Data)
		}
	case "reading":
		var v struct{ Data domain.ReadingLesson }
		if err = json.Unmarshal(body, &v); err == nil {
			title, content = v.Data.Title, compactReading(v.Data)
		}
	}
	if err != nil {
		return aichat.Lesson{}, err
	}
	result := aichat.Lesson{Key: key, Type: lessonType, Title: title, Content: truncateRunes(content, maxLessonContextRunes)}
	s.ai.lessons.Store(cacheKey, result)
	return result, nil
}

type contextWriter struct{ strings.Builder }

func (w *contextWriter) line(parts ...string) {
	var kept []string
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			kept = append(kept, p)
		}
	}
	if len(kept) > 0 {
		w.WriteString(strings.Join(kept, " "))
		w.WriteByte('\n')
	}
}

func (w *contextWriter) field(label, value string) {
	if value = strings.TrimSpace(value); value != "" {
		w.WriteString(label + ": " + value + "\n")
	}
}

func (w *contextWriter) example(japanese, reading, meaning string) {
	japanese = strings.TrimSpace(japanese)
	if japanese == "" {
		return
	}
	text := "- " + japanese
	if reading = strings.TrimSpace(reading); reading != "" && reading != japanese {
		text += " (" + reading + ")"
	}
	if meaning = strings.TrimSpace(meaning); meaning != "" {
		text += " = " + meaning
	}
	w.WriteString(text + "\n")
}

func compactGrammar(l domain.GrammarLesson) string {
	var w contextWriter
	w.line("Level", l.Level, "·", l.Chapter, "·", l.Title)
	for _, p := range l.Points {
		w.WriteString("\n## " + strings.TrimSpace(p.Pattern))
		if m := strings.TrimSpace(p.Meaning); m != "" {
			w.WriteString(" — " + m)
		}
		w.WriteByte('\n')
		w.field("Formation", p.Formation)
		w.field("Explanation", p.Explanation)
		w.field("Common mistake", p.CommonMistake)
		for _, e := range p.Examples {
			w.example(e.Japanese, e.Reading, e.Meaning)
		}
	}
	return w.String()
}

func compactVocabulary(l domain.VocabularyLesson) string {
	var w contextWriter
	w.line("Level", l.Level, "·", l.Title)
	w.line(l.Description)
	for _, word := range l.Words {
		w.example(word.Term, word.Reading, word.Meaning)
		if word.Example != "" {
			w.WriteString("  e.g. " + strings.TrimSpace(word.Example))
			if word.ExampleMeaning != "" {
				w.WriteString(" = " + strings.TrimSpace(word.ExampleMeaning))
			}
			w.WriteByte('\n')
		}
		if n := strings.TrimSpace(word.Note); n != "" {
			w.WriteString("  note: " + n + "\n")
		}
	}
	return w.String()
}

func compactKanji(l domain.KanjiLesson) string {
	var w contextWriter
	w.line(l.Chapter, "·", l.Title)
	for _, c := range l.Characters {
		w.WriteString("\n## " + c.Character + " — " + strings.TrimSpace(c.Meaning) + "\n")
		w.field("On", c.OnReading)
		w.field("Kun", c.KunReading)
		if c.Strokes > 0 {
			w.field("Strokes", fmt.Sprint(c.Strokes))
		}
		w.field("Mnemonic", c.Mnemonic)
		w.field("Note", c.Note)
		for _, e := range c.Examples {
			w.example(e.Japanese, e.Reading, e.Meaning)
		}
	}
	return w.String()
}

func compactReading(l domain.ReadingLesson) string {
	var w contextWriter
	w.line(l.Chapter, "·", l.Title, l.JapaneseTitle)
	w.line(l.Description)
	for i, p := range l.Pages {
		w.WriteString(fmt.Sprintf("\n## Page %d %s\n", i+1, strings.TrimSpace(p.Title)))
		w.field("Japanese", p.Japanese)
		w.field("Reading", p.Reading)
		w.field("Meaning", p.Meaning)
		for _, v := range p.Vocabulary {
			w.example(v.Japanese, v.Reading, v.Meaning)
		}
	}
	return w.String()
}

func truncateRunes(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + "\n…"
}
