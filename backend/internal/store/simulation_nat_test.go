package store

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSimulationNATPaper(t *testing.T) {
	items := simulationNATContent()
	if len(items) != 43 {
		t.Fatal(len(items))
	}
	expected := []int{7, 5, 6, 3, 9, 4, 4, 2, 2, 1}
	counts := map[string]int{}
	order := []string{}
	vocabulary, reading := 0, 0
	for _, item := range items {
		if counts[item.group] == 0 {
			order = append(order, item.group)
		}
		counts[item.group]++
		if item.section == "vocabulary" {
			vocabulary++
		} else if item.section != "grammar" {
			t.Fatal(item.section)
		}
		if strings.HasPrefix(item.group, "読解") {
			reading++
			if item.passage == "" {
				t.Fatal("missing reading passage")
			}
		}
		if item.prompt == "" || item.instructions == "" || item.explanation == "" {
			t.Fatal("incomplete question")
		}
		seen := map[string]bool{}
		for _, choice := range item.choices {
			if choice == "" || seen[choice] {
				t.Fatal("invalid options")
			}
			seen[choice] = true
		}
	}
	if vocabulary != 21 || reading != 5 || len(order) != len(expected) {
		t.Fatalf("bad composition: %d %d %v", vocabulary, reading, counts)
	}
	for i, group := range order {
		if counts[group] != expected[i] {
			t.Fatal(group, counts[group])
		}
	}
	// Verify the answer belongs in the displayed star slot, not merely anywhere in the sentence.
	keys := []string{"駅の", "へ", "を", "新しい"}
	for i, key := range keys {
		if items[30+i].choices[0] != key {
			t.Fatal("wrong star key", i)
		}
	}

}

func TestSimulationNATDeadlinesAndIsolation(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.StartExam("free", "nat-5q-full-1"); !errors.Is(err, ErrPremium) {
		t.Fatal("unpaid NAT access")
	}
	grantTestLifetime(t, s, "user")
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	s.simulationNow = func() time.Time { return now }
	n5, err := s.StartSimulation("user")
	if err != nil {
		t.Fatal(err)
	}
	nat, err := s.StartExam("user", "nat-5q-full-1")
	if err != nil {
		t.Fatal(err)
	}
	if nat.ID == n5.ID || nat.ExamID != "nat-5q-full-1" {
		t.Fatal("levels share session")
	}
	resumed, err := s.StartSimulation("user", "NAT")
	if err != nil || resumed.ID != nat.ID {
		t.Fatal("resume failed")
	}
	session, err := s.SimulationSession("user", nat.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Questions) != 21 || session.ExpiresAt.Sub(now) != 20*time.Minute {
		t.Fatal("wrong vocabulary deadline")
	}
	raw, _ := json.Marshal(session)
	if strings.Contains(string(raw), "correctOptionId") || strings.Contains(string(raw), "explanation") {
		t.Fatal("leaked answer key")
	}
	if _, err := s.SimulationSession("other", nat.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("ownership not checked")
	}
	if _, err := s.SubmitSimulationSection("user", nat.ID, "grammar"); !errors.Is(err, ErrInvalid) {
		t.Fatal("future section unlocked")
	}
	if s.SimulationOverview("user", "NAT")["activeAttemptId"] != nat.ID || s.SimulationOverview("user")["activeAttemptId"] != n5.ID {
		t.Fatal("overview mixes levels")
	}
	now = now.Add(30 * time.Minute)
	session, err = s.SimulationSession("user", nat.ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.ActiveSection != "grammar" || len(session.Questions) != 22 || session.ExpiresAt.Sub(now) != 30*time.Minute {
		t.Fatal("reconnection reset deadline")
	}
	old := s.simulationAttempts[nat.ID].questions[0]
	if err = s.SaveSimulationAnswer("user", nat.ID, old.ID, old.Options[0].ID, false); !errors.Is(err, ErrInvalid) {
		t.Fatal("locked section writable")
	}
	now = now.Add(30 * time.Minute)
	session, err = s.SimulationSession("user", nat.ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.Result == nil || session.Result.Total != 43 || session.Result.DurationSeconds != 3600 || len(session.Result.Reviews) != 43 || session.Result.Score != 0 {
		t.Fatal("offline expiry failed")
	}
	overview := s.SimulationOverview("user", "NAT")
	if overview["questionCount"] != 43 || overview["durationMinutes"] != 60 || len(overview["history"].([]map[string]any)) != 1 {
		t.Fatal("incorrect NAT overview")
	}
	if _, err = s.StartSimulation("user", "N2"); !errors.Is(err, ErrInvalid) {
		t.Fatal("unsupported level accepted")
	}
}

func TestSimulationNATGradesAndRetry(t *testing.T) {
	s := NewMemoryStore()
	grantTestLifetime(t, s, "user")
	attempt, err := s.StartSimulation("user", "NAT")
	if err != nil {
		t.Fatal(err)
	}
	state := s.simulationAttempts[attempt.ID]
	for _, section := range []string{"vocabulary", "grammar"} {
		for _, q := range state.questions {
			if q.Section == section {
				if err = s.SaveSimulationAnswer("user", attempt.ID, q.ID, state.correct[q.ID], false); err != nil {
					t.Fatal(err)
				}
			}
		}
		session, err := s.SubmitSimulationSection("user", attempt.ID, section)
		if err != nil {
			t.Fatal(err)
		}
		if section == "vocabulary" && session.ExpiresAt.Sub(session.SectionStartedAt) != 40*time.Minute {
			t.Fatal("wrong second timer")
		}
	}
	session, err := s.SubmitSimulationSection("user", attempt.ID, "grammar")
	if err != nil {
		t.Fatal(err)
	}
	if session.Result.Score != 100 || session.Result.Correct != 43 || session.Result.Answered != 43 || len(session.Result.Sections) != 10 {
		t.Fatal("wrong grading")
	}
	if strings.Contains(session.Result.ScoringNote, "Dokkai") {
		t.Fatal("reading incorrectly omitted")
	}
	retry, err := s.StartSimulation("user", "NAT")
	if err != nil {
		t.Fatal(err)
	}
	if retry.ID == attempt.ID || s.simulationAttempts[retry.ID].correct[state.questions[0].ID] == state.correct[state.questions[0].ID] {
		t.Fatal("retry must use new option IDs")
	}
}
