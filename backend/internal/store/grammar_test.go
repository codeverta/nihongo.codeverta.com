package store

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"nihongo-master-backend/internal/domain"
)

func TestGrammarCurriculumAndQuizIntegrity(t *testing.T) {
	lessons := seedGrammar()
	if len(lessons) != 40 {
		t.Fatalf("expected 40 N5 lessons, got %d", len(lessons))
	}
	seen := map[string]bool{}
	points, examples, questions := 0, 0, 0
	for i, l := range lessons {
		if seen[l.ID] || l.ID == "" || l.Title == "" || l.Chapter == "" {
			t.Fatalf("invalid lesson %d: %+v", i, l)
		}
		seen[l.ID] = true
		if len(l.Points) != 2 || len(l.Questions) != 5 {
			t.Fatalf("incomplete content for %s", l.ID)
		}
		for _, p := range l.Points {
			if p.Pattern == "" || p.Meaning == "" || p.Formation == "" || p.Explanation == "" || p.CommonMistake == "" || len(p.Examples) != 2 {
				t.Fatalf("incomplete grammar point in %s", l.ID)
			}
			for _, e := range p.Examples {
				if e.Japanese == "" || e.Reading == "" || e.Meaning == "" {
					t.Fatalf("incomplete example in %s", l.ID)
				}
				examples++
			}
			points++
		}
		for _, q := range l.Questions {
			if q.Prompt == "" || q.Explanation == "" || len(q.Choices) != 4 || q.Correct < 0 || q.Correct >= len(q.Choices) {
				t.Fatalf("invalid question in %s", l.ID)
			}
			choices := map[string]bool{}
			for _, c := range q.Choices {
				if strings.TrimSpace(c) == "" || choices[c] {
					t.Fatalf("duplicate/empty choices in %s: %v", l.ID, q.Choices)
				}
				choices[c] = true
			}
			questions++
		}
	}
	if points != 80 || examples != 160 || questions != 200 {
		t.Fatalf("unexpected curriculum: %d patterns, %d examples, %d questions", points, examples, questions)
	}
}

func TestGrammarFullCourseCompletionAndResultRetry(t *testing.T) {
	s := NewMemoryStore()
	const user = "demo-user"
	for _, lesson := range s.grammar {
		if lesson.Level != "N5" {
			continue
		}
		attempt, err := s.StartGrammarAttempt(user, lesson.ID)
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(attempt)
		if strings.Contains(string(encoded), "correct") || strings.Contains(string(encoded), "explanation") {
			t.Fatal("attempt leaks answers")
		}
		if _, err = s.CompleteGrammarAttempt(user, attempt.ID); !errors.Is(err, ErrInvalid) {
			t.Fatal("unfinished quiz must not be graded")
		}
		if _, err = s.AnswerGrammarQuestion("another-user", attempt.ID, attempt.Questions[0].ID, attempt.Questions[0].Options[0].ID); !errors.Is(err, ErrNotFound) {
			t.Fatal("attempt must belong to user")
		}
		if _, err = s.AnswerGrammarQuestion(user, attempt.ID, attempt.Questions[0].ID, "invalid-option"); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid choice accepted")
		}
		for _, q := range attempt.Questions {
			feedback, err := s.AnswerGrammarQuestion(user, attempt.ID, q.ID, s.grammarAttempts[attempt.ID].Correct[q.ID])
			if err != nil || !feedback.Correct || feedback.Explanation == "" {
				t.Fatalf("invalid feedback: %+v %v", feedback, err)
			}
		}
		result, err := s.CompleteGrammarAttempt(user, attempt.ID)
		if err != nil || result.Score != 100 || !result.Passed {
			t.Fatalf("unexpected result: %+v %v", result, err)
		}
		xp := s.Profile(user).XP
		repeated, err := s.CompleteGrammarAttempt(user, attempt.ID)
		if err != nil || repeated != result || s.Profile(user).XP != xp {
			t.Fatal("completion retry must be idempotent")
		}
	}
	course := s.GrammarCourse(user)
	if course.Progress != 100 || course.CompletedLessons != 40 || s.Profile(user).XP != 2000 {
		t.Fatalf("invalid course progress %+v", course)
	}
	if s.GrammarCourse("other").Progress != 0 {
		t.Fatal("user progress leaked")
	}
	if _, err := s.SaveProgress(user, domain.LessonProgress{LessonID: s.grammar[0].ID, Completion: 100}); !errors.Is(err, ErrInvalid) {
		t.Fatal("grammar progress must be graded by server")
	}
}

func TestGrammarFailedRetryPreservesBestScoreAndAnswerIsImmutable(t *testing.T) {
	s := NewMemoryStore()
	const user = "demo-user"
	lessonID := s.grammar[0].ID
	for run := 0; run < 2; run++ {
		a, _ := s.StartGrammarAttempt(user, lessonID)
		for i, q := range a.Questions {
			correct := s.grammarAttempts[a.ID].Correct[q.ID]
			choice := correct
			// First attempt: 4/5=80%; second attempt: 0/5.
			if run == 1 || i == 4 {
				for _, o := range q.Options {
					if o.ID != correct {
						choice = o.ID
						break
					}
				}
			}
			first, err := s.AnswerGrammarQuestion(user, a.ID, q.ID, choice)
			if err != nil {
				t.Fatal(err)
			}
			repeat, err := s.AnswerGrammarQuestion(user, a.ID, q.ID, correct)
			if err != nil || repeat != first {
				t.Fatal("answers can be changed after feedback")
			}
		}
		result, err := s.CompleteGrammarAttempt(user, a.ID)
		if err != nil {
			t.Fatal(err)
		}
		if result.BestScore != 80 || result.Passed != (run == 0) {
			t.Fatalf("bad result: %+v", result)
		}
	}
	lesson, _ := s.GrammarLesson(user, lessonID)
	if !lesson.Completed || lesson.BestScore != 80 || s.Profile(user).XP != 50 {
		t.Fatal("retry overwrote completion or awarded XP twice")
	}
	if _, err := s.GrammarLesson(user, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal("missing lesson must be 404")
	}
}
