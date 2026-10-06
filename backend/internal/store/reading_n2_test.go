package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode"
)

func TestReadingN2Curriculum(t *testing.T) {
	s := NewMemoryStore()
	for _, level := range []string{"N5", "N4", "N3", "N2"} {
		course := s.ReadingCourse("reading-n2-test", level)
		if len(course.Lessons) != 12 || course.Level != level || course.Progress != 0 {
			t.Fatalf("invalid %s catalog: %+v", level, course)
		}
	}
	count := 0
	titles := map[string]bool{}
	pages := map[string]bool{}
	for _, lesson := range s.reading {
		if lesson.Level != "N2" {
			continue
		}
		count++
		if lesson.ID != fmt.Sprintf("reading-n2-%02d", count) || titles[lesson.Title] || lesson.Description == "" || lesson.Chapter == "" {
			t.Fatalf("invalid N2 metadata: %+v", lesson)
		}
		titles[lesson.Title] = true
		if len(lesson.Pages) != 3 || len(lesson.Questions) != 5 {
			t.Fatalf("incomplete %s", lesson.ID)
		}
		passageLength := 0
		for _, page := range lesson.Pages {
			pageLength := len([]rune(strings.ReplaceAll(page.Japanese, "\n", "")))
			passageLength += pageLength
			if pages[page.Japanese] || pageLength < 250 || len(page.Vocabulary) != 2 {
				t.Fatalf("duplicate or incomplete passage in %s", lesson.ID)
			}
			if !strings.Contains(page.Japanese, "\n\n") || !strings.Contains(page.Reading, "\n\n") || !strings.Contains(page.Meaning, "\n\n") {
				t.Fatalf("long passage must have paragraph breaks and matching assistance in %s", lesson.ID)
			}
			pages[page.Japanese] = true
			for _, r := range page.Reading {
				if unicode.Is(unicode.Han, r) {
					t.Fatalf("kana assistance contains kanji in %s: %s", lesson.ID, page.Reading)
				}
			}
		}
		if passageLength < 800 {
			t.Fatalf("N2 lesson %s is too short: %d characters", lesson.ID, passageLength)
		}
		payload, err := s.ReadingLesson("reading-n2-test", lesson.ID)
		if err != nil || payload.Order != count || payload.QuestionCount != 5 || payload.PageCount != 3 {
			t.Fatalf("invalid detail: %+v %v", payload, err)
		}
		encoded, err := json.Marshal(payload)
		if err != nil || strings.Contains(string(encoded), "Correct") || strings.Contains(string(encoded), "correctOptionId") {
			t.Fatal("lesson details expose quiz answers")
		}
	}
	if count != 12 {
		t.Fatalf("got %d N2 lessons", count)
	}
}

func TestReadingN2CompletionAndRetry(t *testing.T) {
	s := NewMemoryStore()
	const user = "reading-n2-progress-test"
	initialXP := s.Profile(user).XP
	for index, summary := range s.ReadingCourse(user, "N2").Lessons {
		attempt, err := s.StartReadingAttempt(user, summary.ID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.CompleteReadingAttempt(user, attempt.ID); !errors.Is(err, ErrInvalid) {
			t.Fatal("incomplete attempt was scored")
		}
		if _, err := s.AnswerReadingQuestion("other-user", attempt.ID, attempt.Questions[0].ID, attempt.Questions[0].Options[0].ID); !errors.Is(err, ErrNotFound) {
			t.Fatal("attempt ownership not enforced")
		}
		encoded, _ := json.Marshal(attempt)
		if strings.Contains(string(encoded), "correctOptionId") || strings.Contains(string(encoded), "explanation") {
			t.Fatal("attempt exposes answer key")
		}
		for _, q := range attempt.Questions {
			feedback, err := s.AnswerReadingQuestion(user, attempt.ID, q.ID, s.readingAttempts[attempt.ID].Correct[q.ID])
			if err != nil || !feedback.Correct || feedback.Explanation == "" {
				t.Fatalf("bad feedback: %+v %v", feedback, err)
			}
		}
		result, err := s.CompleteReadingAttempt(user, attempt.ID)
		if err != nil || !result.Passed || result.Score != 100 || result.Progress != (index+1)*100/12 {
			t.Fatalf("invalid progress result: %+v %v", result, err)
		}
		repeat, err := s.CompleteReadingAttempt(user, attempt.ID)
		if err != nil || repeat != result {
			t.Fatal("completion is not idempotent")
		}
	}
	if s.ReadingCourse(user, "N2").Progress != 100 || s.Profile(user).XP != initialXP+600 {
		t.Fatal("full course completion or one-time XP failed")
	}
	for _, level := range []string{"N5", "N4", "N3"} {
		if s.ReadingCourse(user, level).Progress != 0 {
			t.Fatal("progress leaked to another level")
		}
	}
	if s.ReadingCourse("other-user", "N2").Progress != 0 {
		t.Fatal("progress leaked to another user")
	}
	// A new retry can fail without replacing the previous best score/completion.
	a, err := s.StartReadingAttempt(user, "reading-n2-01")
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range a.Questions {
		correct := s.readingAttempts[a.ID].Correct[q.ID]
		for _, option := range q.Options {
			if option.ID == correct {
				continue
			}
			feedback, err := s.AnswerReadingQuestion(user, a.ID, q.ID, option.ID)
			if err != nil || feedback.Correct {
				t.Fatal("wrong answer not graded correctly")
			}
			repeated, err := s.AnswerReadingQuestion(user, a.ID, q.ID, correct)
			if err != nil || repeated != feedback {
				t.Fatal("answer changed after feedback")
			}
			break
		}
	}
	r, err := s.CompleteReadingAttempt(user, a.ID)
	if err != nil || r.Passed || r.Score != 0 || r.BestScore != 100 || r.Progress != 100 || !r.CanRetry || s.Profile(user).XP != initialXP+600 {
		t.Fatalf("retry damaged progress: %+v %v", r, err)
	}
}
