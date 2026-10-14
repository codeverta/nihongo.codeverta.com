package store

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSimulationN3Paper(t *testing.T) {
	items := simulationN3Content()
	if len(items) != 74 {
		t.Fatal(len(items))
	}
	expected := []int{8, 6, 11, 5, 5, 13, 5, 5, 4, 6, 4, 2}
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
	if vocabulary != 35 || reading != 16 || len(order) != len(expected) {
		t.Fatalf("bad composition: %d %d %v", vocabulary, reading, counts)
	}
	for i, group := range order {
		if counts[group] != expected[i] {
			t.Fatal(group, counts[group])
		}
	}
	// Verify the answer belongs in the displayed star slot, not merely anywhere in the sentence.
	keys := []string{"ように", "勉強を", "交通も", "説明した", "会議を延期する"}
	for i, key := range keys {
		if items[48+i].choices[0] != key {
			t.Fatal("wrong star key", i)
		}
	}

}

func TestSimulationN3DeadlinesAndIsolation(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.StartExam("free-user", "jlpt-n3-simulation-1"); !errors.Is(err, ErrPremium) {
		t.Fatal("N3 start must require premium")
	}
	grantTestLifetime(t, s, "user")
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	s.simulationNow = func() time.Time { return now }
	n5, err := s.StartSimulation("user")
	if err != nil {
		t.Fatal(err)
	}
	n3, err := s.StartExam("user", "jlpt-n3-simulation-1")
	if err != nil {
		t.Fatal(err)
	}
	if n3.ID == n5.ID || n3.ExamID != "jlpt-n3-simulation-1" {
		t.Fatal("levels share session")
	}
	resumed, err := s.StartSimulation("user", "N3")
	if err != nil || resumed.ID != n3.ID {
		t.Fatal("resume failed")
	}
	session, err := s.SimulationSession("user", n3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Questions) != 35 || session.ExpiresAt.Sub(now) != 30*time.Minute {
		t.Fatal("wrong vocabulary deadline")
	}
	raw, _ := json.Marshal(session)
	if strings.Contains(string(raw), "correctOptionId") || strings.Contains(string(raw), "explanation") {
		t.Fatal("leaked answer key")
	}
	if _, err := s.SimulationSession("other", n3.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("ownership not checked")
	}
	if _, err := s.SubmitSimulationSection("user", n3.ID, "grammar"); !errors.Is(err, ErrInvalid) {
		t.Fatal("future section unlocked")
	}
	if s.SimulationOverview("user", "N3")["activeAttemptId"] != n3.ID || s.SimulationOverview("user")["activeAttemptId"] != n5.ID {
		t.Fatal("overview mixes levels")
	}
	now = now.Add(35 * time.Minute)
	session, err = s.SimulationSession("user", n3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.ActiveSection != "grammar" || len(session.Questions) != 39 || session.ExpiresAt.Sub(now) != 65*time.Minute {
		t.Fatal("reconnection reset deadline")
	}
	old := s.simulationAttempts[n3.ID].questions[0]
	if err = s.SaveSimulationAnswer("user", n3.ID, old.ID, old.Options[0].ID, false); !errors.Is(err, ErrInvalid) {
		t.Fatal("locked section writable")
	}
	now = now.Add(65 * time.Minute)
	session, err = s.SimulationSession("user", n3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.Result == nil || session.Result.Total != 74 || session.Result.DurationSeconds != 6000 || len(session.Result.Reviews) != 74 || session.Result.Score != 0 {
		t.Fatal("offline expiry failed")
	}
	overview := s.SimulationOverview("user", "N3")
	if overview["questionCount"] != 74 || overview["durationMinutes"] != 100 || len(overview["history"].([]map[string]any)) != 1 {
		t.Fatal("incorrect N3 overview")
	}
	if _, err = s.StartSimulation("user", "N2"); !errors.Is(err, ErrInvalid) {
		t.Fatal("unsupported level accepted")
	}
}

func TestSimulationN3GradesAndRetry(t *testing.T) {
	s := NewMemoryStore()
	grantTestLifetime(t, s, "user")
	attempt, err := s.StartSimulation("user", "N3")
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
		if section == "vocabulary" && session.ExpiresAt.Sub(session.SectionStartedAt) != 70*time.Minute {
			t.Fatal("wrong second timer")
		}
	}
	session, err := s.SubmitSimulationSection("user", attempt.ID, "grammar")
	if err != nil {
		t.Fatal(err)
	}
	if session.Result.Score != 100 || session.Result.Correct != 74 || session.Result.Answered != 74 || len(session.Result.Sections) != 12 {
		t.Fatal("wrong grading")
	}
	if strings.Contains(session.Result.ScoringNote, "Dokkai") {
		t.Fatal("reading incorrectly omitted")
	}
	retry, err := s.StartSimulation("user", "N3")
	if err != nil {
		t.Fatal(err)
	}
	if retry.ID == attempt.ID || s.simulationAttempts[retry.ID].correct[state.questions[0].ID] == state.correct[state.questions[0].ID] {
		t.Fatal("retry must use new option IDs")
	}
}
