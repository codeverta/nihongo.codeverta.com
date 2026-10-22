package store

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSimulationN4Paper(t *testing.T) {
	items := simulationN4Content()
	if len(items) != 57 {
		t.Fatal(len(items))
	}
	expected := []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}
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
	if vocabulary != 28 || reading != 8 || len(order) != len(expected) {
		t.Fatalf("bad composition: %d %d %v", vocabulary, reading, counts)
	}
	for i, group := range order {
		if counts[group] != expected[i] {
			t.Fatal(group, counts[group])
		}
	}
	// Verify the answer belongs in the displayed star slot, not merely anywhere in the sentence.
	sentences := [][]string{{"京都", "に", "行った", "ことが"}, {"前に", "歯", "を", "磨いて"}, {"友達", "が", "くれた", "もの"}, {"ために", "毎日", "日本語を", "勉強して"}}
	for i, sentence := range sentences {
		if items[41+i].choices[0] != sentence[2] {
			t.Fatal("wrong star key", i)
		}
	}
}

func TestSimulationN4DeadlinesAndIsolation(t *testing.T) {
	s := NewMemoryStore()
	grantTestLifetime(t, s, "user")
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	s.simulationNow = func() time.Time { return now }
	n5, err := s.StartSimulation("user")
	if err != nil {
		t.Fatal(err)
	}
	n4, err := s.StartExam("user", "jlpt-n4-simulation-1")
	if err != nil {
		t.Fatal(err)
	}
	if n4.ID == n5.ID || n4.ExamID != "jlpt-n4-simulation-1" {
		t.Fatal("levels share session")
	}
	resumed, err := s.StartSimulation("user", "N4")
	if err != nil || resumed.ID != n4.ID {
		t.Fatal("resume failed")
	}
	session, err := s.SimulationSession("user", n4.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Questions) != 28 || session.ExpiresAt.Sub(now) != 25*time.Minute {
		t.Fatal("wrong vocabulary deadline")
	}
	raw, _ := json.Marshal(session)
	if strings.Contains(string(raw), "correctOptionId") || strings.Contains(string(raw), "explanation") {
		t.Fatal("leaked answer key")
	}
	if _, err := s.SimulationSession("other", n4.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("ownership not checked")
	}
	if _, err := s.SubmitSimulationSection("user", n4.ID, "grammar"); !errors.Is(err, ErrInvalid) {
		t.Fatal("future section unlocked")
	}
	if s.SimulationOverview("user", "N4")["activeAttemptId"] != n4.ID || s.SimulationOverview("user")["activeAttemptId"] != n5.ID {
		t.Fatal("overview mixes levels")
	}
	now = now.Add(30 * time.Minute)
	session, err = s.SimulationSession("user", n4.ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.ActiveSection != "grammar" || len(session.Questions) != 29 || session.ExpiresAt.Sub(now) != 50*time.Minute {
		t.Fatal("reconnection reset deadline")
	}
	old := s.simulationAttempts[n4.ID].questions[0]
	if err = s.SaveSimulationAnswer("user", n4.ID, old.ID, old.Options[0].ID, false); !errors.Is(err, ErrInvalid) {
		t.Fatal("locked section writable")
	}
	now = now.Add(50 * time.Minute)
	session, err = s.SimulationSession("user", n4.ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.Result == nil || session.Result.Total != 57 || session.Result.DurationSeconds != 4800 || len(session.Result.Reviews) != 57 || session.Result.Score != 0 {
		t.Fatal("offline expiry failed")
	}
	overview := s.SimulationOverview("user", "N4")
	if overview["questionCount"] != 57 || overview["durationMinutes"] != 80 || len(overview["history"].([]map[string]any)) != 1 {
		t.Fatal("incorrect N4 overview")
	}
	if _, err = s.StartSimulation("user", "N6"); !errors.Is(err, ErrInvalid) {
		t.Fatal("unsupported level accepted")
	}
}

func TestSimulationN4GradesAndRetry(t *testing.T) {
	s := NewMemoryStore()
	grantTestLifetime(t, s, "user")
	attempt, err := s.StartSimulation("user", "N4")
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
		if section == "vocabulary" && session.ExpiresAt.Sub(session.SectionStartedAt) != 55*time.Minute {
			t.Fatal("wrong second timer")
		}
	}
	session, err := s.SubmitSimulationSection("user", attempt.ID, "grammar")
	if err != nil {
		t.Fatal(err)
	}
	if session.Result.Score != 100 || session.Result.Correct != 57 || session.Result.Answered != 57 || len(session.Result.Sections) != 11 {
		t.Fatal("wrong grading")
	}
	if strings.Contains(session.Result.ScoringNote, "Dokkai") {
		t.Fatal("reading incorrectly omitted")
	}
	retry, err := s.StartSimulation("user", "N4")
	if err != nil {
		t.Fatal(err)
	}
	if retry.ID == attempt.ID || s.simulationAttempts[retry.ID].correct[state.questions[0].ID] == state.correct[state.questions[0].ID] {
		t.Fatal("retry must use new option IDs")
	}
}
