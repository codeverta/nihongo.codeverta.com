package store

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestJFTPaper(t *testing.T) {
	paper := simulationPaper("JFT")
	if len(paper.items) != 36 || len(paper.sections) != 3 {
		t.Fatal("incorrect paper")
	}
	counts := map[string]int{}
	for _, item := range paper.items {
		counts[item.section]++
		if item.prompt == "" || item.explanation == "" || item.instructions == "" {
			t.Fatal("incomplete question")
		}
		if item.section == "reading" && item.passage == "" {
			t.Fatal("missing reading passage")
		}
		choices := map[string]bool{}
		for _, choice := range item.choices {
			if choice == "" || choices[choice] {
				t.Fatal("invalid options")
			}
			choices[choice] = true
		}
	}
	for _, section := range paper.sections {
		if counts[section.ID] != 12 {
			t.Fatal(counts)
		}
	}
}

func TestJFTSharedDeadlineAndGrading(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.StartExam("free", "jft-a2-full-1"); !errors.Is(err, ErrPremium) {
		t.Fatal("unpaid access")
	}
	grantTestLifetime(t, s, "user")
	now := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	s.simulationNow = func() time.Time { return now }
	attempt, err := s.StartExam("user", "jft-a2-full-1")
	if err != nil {
		t.Fatal(err)
	}
	state := s.simulationAttempts[attempt.ID]
	deadline := now.Add(45 * time.Minute)
	resumed, err := s.StartSimulation("user", "JFT")
	if err != nil || resumed.ID != attempt.ID {
		t.Fatal("resume failed")
	}
	if _, err := s.SimulationSession("other", attempt.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("ownership missing")
	}
	if _, err := s.SubmitSimulationSection("user", attempt.ID, "reading"); !errors.Is(err, ErrInvalid) {
		t.Fatal("future section writable")
	}
	for _, part := range []string{"vocabulary", "conversation", "reading"} {
		session, err := s.SimulationSession("user", attempt.ID)
		if err != nil || session.ActiveSection != part || len(session.Questions) != 12 || !session.ExpiresAt.Equal(deadline) {
			t.Fatal("wrong stage or deadline", session, err)
		}
		raw, _ := json.Marshal(session)
		if strings.Contains(string(raw), "correctOptionId") || strings.Contains(string(raw), "explanation") {
			t.Fatal("answer key leaked")
		}
		for _, q := range session.Questions {
			if err := s.SaveSimulationAnswer("user", attempt.ID, q.ID, state.correct[q.ID], false); err != nil {
				t.Fatal(err)
			}
		}
		now = now.Add(5 * time.Minute)
		submitted, err := s.SubmitSimulationSection("user", attempt.ID, part)
		if err != nil {
			t.Fatal(err)
		}
		if part != "reading" && !submitted.ExpiresAt.Equal(deadline) {
			t.Fatal("section transition reset timer")
		}
		// Retrying the previous request cannot skip another section.
		retried, err := s.SubmitSimulationSection("user", attempt.ID, part)
		if err != nil || retried.ActiveSection != submitted.ActiveSection {
			t.Fatal("retry advanced stage")
		}
		if err := s.SaveSimulationAnswer("user", attempt.ID, session.Questions[0].ID, state.correct[session.Questions[0].ID], false); !errors.Is(err, ErrInvalid) {
			t.Fatal("closed stage writable")
		}
	}
	result := state.session.Result
	if result == nil || result.Score != 100 || result.Correct != 36 || len(result.Reviews) != 36 || result.DurationSeconds != 900 {
		t.Fatal("incorrect grading", result)
	}
	if !strings.Contains(result.ScoringNote, "JFT-Basic") || strings.Contains(result.ScoringNote, "JLPT") {
		t.Fatal(result.ScoringNote)
	}
	if s.SimulationOverview("user", "JFT")["activeAttemptId"] != "" || s.SimulationOverview("user", "N5")["questionCount"] != 38 {
		t.Fatal("histories mixed")
	}
	retry, err := s.StartSimulation("user", "JFT")
	if err != nil || retry.ID == attempt.ID {
		t.Fatal("no new attempt")
	}
	now = now.Add(46 * time.Minute)
	expired, err := s.SimulationSession("user", retry.ID)
	if err != nil || expired.Result == nil || expired.Result.Total != 36 || expired.Result.Score != 0 || expired.Result.DurationSeconds != 2700 {
		t.Fatal("offline expiry failed", expired, err)
	}
}
