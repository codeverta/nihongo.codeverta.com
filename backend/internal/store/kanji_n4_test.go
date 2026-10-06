package store

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"unicode"
)

func TestKanjiN4ReferenceCoverageAndCatalog(t *testing.T) {
	s := NewMemoryStore()
	referenceBytes, err := os.ReadFile("../../../docs/kanji-n4-reference.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference struct {
		Characters map[string]struct {
			Strokes int `json:"strokes"`
		} `json:"characters"`
	}
	if err := json.Unmarshal(referenceBytes, &reference); err != nil {
		t.Fatal(err)
	}
	if len(reference.Characters) != 166 {
		t.Fatal("N4 reference is incomplete")
	}
	seen := map[string]bool{}
	examples, questions := 0, 0
	for _, lesson := range s.kanji {
		if lesson.Level != "N4" {
			continue
		}
		for _, c := range lesson.Characters {
			ref, exists := reference.Characters[c.Character]
			if !exists || seen[c.Character] || c.Strokes != ref.Strokes {
				t.Fatalf("invalid reference coverage or stroke count: %s", c.Character)
			}
			seen[c.Character] = true
			for _, example := range c.Examples {
				if !strings.Contains(example.Japanese, c.Character) || example.Meaning == "" {
					t.Fatalf("unrelated example for %s", c.Character)
				}
				for _, r := range example.Reading {
					if unicode.Is(unicode.Han, r) {
						t.Fatalf("example reading contains kanji: %s", example.Reading)
					}
				}
				examples++
			}
		}
		questions += len(lesson.Questions)
	}
	if len(seen) != 166 || examples != 332 || questions != 332 {
		t.Fatalf("incomplete N4: %d characters, %d examples, %d questions", len(seen), examples, questions)
	}
	for level, count := range map[string]int{"N5": 20, "N4": 34} {
		course := s.KanjiCourse("kanji-catalog-test", level)
		if course.Level != level || len(course.Lessons) != count || course.Progress != 0 {
			t.Fatalf("invalid %s course: %+v", level, course)
		}
		for i, summary := range course.Lessons {
			if summary.Order != i+1 {
				t.Fatalf("order should restart at one per level: %+v", summary)
			}
			lesson, err := s.KanjiLesson("kanji-catalog-test", summary.ID)
			if err != nil || len(lesson.Characters) != summary.CharacterCount || summary.QuestionCount != 2*summary.CharacterCount {
				t.Fatalf("invalid lesson detail: %+v %v", lesson, err)
			}
		}
	}
	if len(s.KanjiCourse("kanji-catalog-test").Lessons) != 20 {
		t.Fatal("default N5 behavior changed")
	}
}

func TestKanjiN4CompletionIsolationAndRetry(t *testing.T) {
	s := NewMemoryStore()
	const user = "kanji-n4-progress-test"
	xpBefore := s.Profile(user).XP
	for index, lesson := range s.KanjiCourse(user, "N4").Lessons {
		a, err := s.StartKanjiAttempt(user, lesson.ID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.CompleteKanjiAttempt(user, a.ID); !errors.Is(err, ErrInvalid) {
			t.Fatal("incomplete attempt was scored")
		}
		q := a.Questions[0]
		if _, err := s.AnswerKanjiQuestion("another-user", a.ID, q.ID, q.Options[0].ID); !errors.Is(err, ErrNotFound) {
			t.Fatal("another user accessed the attempt")
		}
		if _, err := s.AnswerKanjiQuestion(user, a.ID, q.ID, "missing-option"); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid answer accepted")
		}
		payload, _ := json.Marshal(a)
		if strings.Contains(string(payload), "correctOptionId") || strings.Contains(string(payload), "explanation") {
			t.Fatal("answer key exposed before submission")
		}
		for _, q := range a.Questions {
			feedback, err := s.AnswerKanjiQuestion(user, a.ID, q.ID, s.kanjiAttempts[a.ID].Correct[q.ID])
			if err != nil || !feedback.Correct || feedback.Explanation == "" {
				t.Fatalf("invalid feedback: %+v %v", feedback, err)
			}
		}
		r, err := s.CompleteKanjiAttempt(user, a.ID)
		if err != nil || r.Score != 100 || !r.Passed || r.Progress != (index+1)*100/34 {
			t.Fatalf("invalid N4 result: %+v %v", r, err)
		}
		repeated, err := s.CompleteKanjiAttempt(user, a.ID)
		if err != nil || repeated != r {
			t.Fatal("completion retry changed the result")
		}
	}
	if s.Profile(user).XP != xpBefore+34*50 || s.KanjiCourse(user, "N4").Progress != 100 {
		t.Fatal("course completion or first-pass XP failed")
	}
	if s.KanjiCourse(user, "N5").Progress != 0 || s.KanjiCourse("another-user", "N4").Progress != 0 || s.ReadingCourse(user, "N4").Progress != 0 {
		t.Fatal("progress leaked into another level, user, or subject")
	}
	// Fail a retry. Neither completion, best score, nor XP may be downgraded or duplicated.
	a, _ := s.StartKanjiAttempt(user, "kanji-n4-01")
	for _, q := range a.Questions {
		correct := s.kanjiAttempts[a.ID].Correct[q.ID]
		for _, o := range q.Options {
			if o.ID == correct {
				continue
			}
			first, _ := s.AnswerKanjiQuestion(user, a.ID, q.ID, o.ID)
			again, _ := s.AnswerKanjiQuestion(user, a.ID, q.ID, correct)
			if first != again || first.Correct {
				t.Fatal("submitted answer was changed after feedback")
			}
			break
		}
	}
	r, err := s.CompleteKanjiAttempt(user, a.ID)
	if err != nil || r.Score != 0 || r.Passed || r.BestScore != 100 || r.Progress != 100 || s.Profile(user).XP != xpBefore+34*50 {
		t.Fatalf("retry damaged progress: %+v %v", r, err)
	}
}

func TestKanjiN4PassThresholdForBothLessonSizes(t *testing.T) {
	for _, id := range []string{"kanji-n4-01", "kanji-n4-31"} {
		for _, pass := range []bool{false, true} {
			s := NewMemoryStore()
			a, _ := s.StartKanjiAttempt("threshold", id)
			correctCount := (len(a.Questions)*80 + 99) / 100
			if !pass {
				correctCount--
			}
			for i, q := range a.Questions {
				correct := s.kanjiAttempts[a.ID].Correct[q.ID]
				option := correct
				if i >= correctCount {
					for _, o := range q.Options {
						if o.ID != correct {
							option = o.ID
							break
						}
					}
				}
				s.AnswerKanjiQuestion("threshold", a.ID, q.ID, option)
			}
			r, err := s.CompleteKanjiAttempt("threshold", a.ID)
			if err != nil || r.Passed != pass {
				t.Fatalf("incorrect threshold for %s: %+v", id, r)
			}
		}
	}
}
