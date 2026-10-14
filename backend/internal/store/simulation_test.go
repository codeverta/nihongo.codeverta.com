package store

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSimulationPaperIntegrity(t *testing.T) {
	items := simulationContent()
	if len(items) != 38 {
		t.Fatalf("expected revised 38-item language paper, got %d", len(items))
	}
	counts := map[string]int{}
	groups := map[string]int{}
	for _, item := range items {
		counts[item.section]++
		groups[item.group]++
		if item.section != "grammar" && item.section != "vocabulary" {
			t.Fatal("unexpected reading/listening section")
		}
		if item.instructions == "" || item.prompt == "" || item.explanation == "" {
			t.Fatal("incomplete authored question")
		}
		seen := map[string]bool{}
		for _, choice := range item.choices {
			if choice == "" || seen[choice] {
				t.Fatalf("invalid choices for %s", item.prompt)
			}
			seen[choice] = true
		}
	}
	if counts["vocabulary"] != 21 || counts["grammar"] != 17 || len(groups) != 7 {
		t.Fatalf("incorrect composition: %v %v", counts, groups)
	}
	expected := []int{7, 5, 6, 3, 9, 4, 4}
	seen := map[string]bool{}
	index := 0
	for _, item := range items {
		if !seen[item.group] {
			seen[item.group] = true
			if groups[item.group] != expected[index] {
				t.Fatal("incorrect mondai size")
			}
			index++
		}
	}
}

func TestSimulationSectionLocksGradingAndRetry(t *testing.T) {
	s := NewMemoryStore()
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	s.simulationNow = func() time.Time { return now }
	attempt, err := s.StartSimulation("a5090412-9ab4-5852-8229-c7d19eb5cb1c")
	if err != nil {
		t.Fatal(err)
	}
	repeat, _ := s.StartSimulation("a5090412-9ab4-5852-8229-c7d19eb5cb1c")
	if repeat.ID != attempt.ID {
		t.Fatal("start retry must resume")
	}
	session, err := s.SimulationSession("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Questions) != 21 || session.ActiveSection != "vocabulary" {
		t.Fatal("grammar released early")
	}
	serialized, _ := json.Marshal(session)
	for _, key := range []string{"correctOptionId", "explanation", "score"} {
		if strings.Contains(string(serialized), key) {
			t.Fatalf("exam leaked %s", key)
		}
	}
	if _, err = s.SimulationSession("other", attempt.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("ownership must be enforced")
	}
	if _, err = s.SubmitSimulationSection("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, "grammar"); !errors.Is(err, ErrInvalid) {
		t.Fatal("cannot skip vocabulary")
	}
	if err = s.SaveSimulationAnswer("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, "n5-sim-01-q22", "bad", false); !errors.Is(err, ErrInvalid) {
		t.Fatal("cannot answer future section")
	}
	state := s.simulationAttempts[attempt.ID]
	for _, q := range session.Questions {
		if err = s.SaveSimulationAnswer("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, q.ID, state.correct[q.ID], q.Number == 1); err != nil {
			t.Fatal(err)
		}
	}
	// Detached snapshots must not allow mutation of the live question or answers.
	snapshot, _ := s.SimulationSession("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID)
	snapshot.Answers[session.Questions[0].ID] = "tampered"
	snapshot.Questions[0].Options[0].Text = "tampered"
	check, _ := s.SimulationSession("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID)
	if check.Answers[session.Questions[0].ID] == "tampered" || check.Questions[0].Options[0].Text == "tampered" || len(check.Marked) != 1 {
		t.Fatal("snapshot mutated live state or lost flags")
	}
	now = now.Add(3 * time.Minute)
	grammar, err := s.SubmitSimulationSection("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, "vocabulary")
	if err != nil {
		t.Fatal(err)
	}
	if grammar.ActiveSection != "grammar" || len(grammar.Questions) != 17 || len(grammar.Answers) != 0 || grammar.Result != nil {
		t.Fatal("invalid section transition")
	}
	retry, _ := s.SubmitSimulationSection("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, "vocabulary")
	if retry.ActiveSection != "grammar" || retry.ExpiresAt != grammar.ExpiresAt || retry.Result != nil {
		t.Fatal("duplicate vocabulary submission must not submit grammar")
	}
	if err = s.SaveSimulationAnswer("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, session.Questions[0].ID, state.correct[session.Questions[0].ID], false); !errors.Is(err, ErrInvalid) {
		t.Fatal("closed section changed")
	}
	for i, q := range grammar.Questions {
		if i == 0 {
			continue
		} // An unanswered item is wrong.
		option := state.correct[q.ID]
		if i == 1 {
			for _, o := range q.Options {
				if o.ID != option {
					option = o.ID
					break
				}
			}
		}
		if err = s.SaveSimulationAnswer("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, q.ID, option, false); err != nil {
			t.Fatal(err)
		}
	}
	now = now.Add(5 * time.Minute)
	done, err := s.SubmitSimulationSection("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, "grammar")
	if err != nil {
		t.Fatal(err)
	}
	r := done.Result
	if r == nil || r.Total != 38 || r.Correct != 36 || r.Answered != 37 || r.Score != 94 || len(r.Sections) != 7 || len(r.Reviews) != 38 || r.DurationSeconds != 480 {
		t.Fatalf("wrong raw grade: %+v", r)
	}
	again, _ := s.SubmitSimulationSection("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, "grammar")
	if again.Result.CompletedAt != r.CompletedAt {
		t.Fatal("completion retry changed result")
	}
	if err = s.SaveSimulationAnswer("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, grammar.Questions[0].ID, state.correct[grammar.Questions[0].ID], false); !errors.Is(err, ErrInvalid) {
		t.Fatal("completed answer changed")
	}
	next, _ := s.StartSimulation("a5090412-9ab4-5852-8229-c7d19eb5cb1c")
	if next.ID == attempt.ID {
		t.Fatal("retry needs a new attempt")
	}
}

func TestSimulationExpiryAdvancesWithoutGrantingExtraTime(t *testing.T) {
	s := NewMemoryStore()
	start := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	now := start
	s.simulationNow = func() time.Time { return now }
	attempt, _ := s.StartSimulation("a5090412-9ab4-5852-8229-c7d19eb5cb1c")
	session, _ := s.SimulationSession("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID)
	now = start.Add(20 * time.Minute)
	if err := s.SaveSimulationAnswer("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID, session.Questions[0].ID, session.Questions[0].Options[0].ID, false); !errors.Is(err, ErrInvalid) {
		t.Fatal("deadline answer must be rejected")
	}
	now = start.Add(25 * time.Minute)
	grammar, _ := s.SimulationSession("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID)
	if grammar.ExpiresAt != start.Add(40*time.Minute) || grammar.SectionStartedAt != start.Add(20*time.Minute) {
		t.Fatal("reconnection extended timer")
	}
	now = start.Add(50 * time.Minute)
	done, _ := s.SimulationSession("a5090412-9ab4-5852-8229-c7d19eb5cb1c", attempt.ID)
	if done.Result == nil || done.Result.Score != 0 || done.Result.DurationSeconds != 2400 || done.Result.CompletedAt != start.Add(40*time.Minute) {
		t.Fatal("expired exam not finalized at original deadline")
	}
	overview := s.SimulationOverview("a5090412-9ab4-5852-8229-c7d19eb5cb1c")
	if overview["activeAttemptId"] != "" || len(overview["history"].([]map[string]any)) != 1 {
		t.Fatal("expired attempt missing from history")
	}
}
