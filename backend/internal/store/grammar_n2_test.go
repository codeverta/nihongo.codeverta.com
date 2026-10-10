package store

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestGrammarN2ContentAndFullCompletion(t *testing.T) {
	s := NewMemoryStore()
	lessons := seedGrammarN2()
	patterns := map[string]bool{}
	positions := map[int]bool{}
	for i, l := range lessons {
		if l.ID != fmt.Sprintf("grammar-n2-%02d", i+1) || l.Level != "N2" {
			t.Fatal("unstable curriculum IDs")
		}
		for _, p := range l.Points {
			if patterns[p.Pattern] {
				t.Fatalf("duplicate pattern: %s", p.Pattern)
			}
			patterns[p.Pattern] = true
		}
		for _, q := range l.Questions {
			positions[q.Correct] = true
		}
		if !strings.Contains(l.Questions[1].Prompt, "（　）") {
			t.Fatal("missing sentence grammar practice")
		}
		a, err := s.StartGrammarAttempt("n2-user", l.ID)
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(a)
		if strings.Contains(string(encoded), "correct") || strings.Contains(string(encoded), "explanation") {
			t.Fatal("quiz leaks answers")
		}
		for _, q := range a.Questions {
			feedback, err := s.AnswerGrammarQuestion("n2-user", a.ID, q.ID, s.grammarAttempts[a.ID].Correct[q.ID])
			if err != nil || !feedback.Correct || feedback.Explanation == "" {
				t.Fatalf("bad feedback: %+v %v", feedback, err)
			}
		}
		result, err := s.CompleteGrammarAttempt("n2-user", a.ID)
		if err != nil || result.Score != 100 || !result.Passed {
			t.Fatalf("bad completion: %+v %v", result, err)
		}
		xp := s.Profile("n2-user").XP
		repeated, err := s.CompleteGrammarAttempt("n2-user", a.ID)
		if err != nil || repeated != result || s.Profile("n2-user").XP != xp {
			t.Fatal("retry double awards XP")
		}
	}
	if len(patterns) != 120 || len(positions) != 4 {
		t.Fatalf("incomplete content: %d patterns, %d answer positions", len(patterns), len(positions))
	}
	c := s.GrammarCourse("n2-user", "N2")
	if len(c.Lessons) != 60 || c.CompletedLessons != 60 || c.Progress != 100 || s.Profile("n2-user").XP != 3000 {
		t.Fatalf("bad course progress: %+v", c)
	}
	for _, level := range []string{"N5", "N4", "N3"} {
		if s.GrammarCourse("n2-user", level).Progress != 0 {
			t.Fatal("progress leaked to another level")
		}
	}
	if s.GrammarCourse("other-user", "N2").Progress != 0 {
		t.Fatal("progress leaked to another user")
	}
}
