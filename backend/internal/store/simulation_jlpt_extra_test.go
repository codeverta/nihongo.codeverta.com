package store

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

var jlptRubyMarkup = regexp.MustCompile(`\{([^|{}]+)\|([^{}]+)\}`)

func plainSimulationText(s string) string {
	return jlptRubyMarkup.ReplaceAllString(s, "$1")
}

func TestSimulationJLPTExtraPapers(t *testing.T) {
	cases := []struct {
		name       string
		items      []simulationItem
		total      int
		vocabulary int
		reading    int
		groups     []int
	}{
		{"N5-2", simulationN5AdditionalContent(2), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N5-3", simulationN5AdditionalContent(3), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N4-2", simulationN4AdditionalContent(2), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
		{"N4-3", simulationN4AdditionalContent(3), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
		{"N5-4", simulationN5AdditionalContent(4), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N5-5", simulationN5AdditionalContent(5), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N4-4", simulationN4AdditionalContent(4), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
		{"N4-5", simulationN4AdditionalContent(5), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
		{"N5-6", simulationN5AdditionalContent(6), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N5-7", simulationN5AdditionalContent(7), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N4-6", simulationN4AdditionalContent(6), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
		{"N4-7", simulationN4AdditionalContent(7), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
		{"N5-8", simulationN5AdditionalContent(8), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N5-9", simulationN5AdditionalContent(9), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N5-10", simulationN5AdditionalContent(10), 38, 21, 0, []int{7, 5, 6, 3, 9, 4, 4}},
		{"N4-8", simulationN4AdditionalContent(8), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
		{"N4-9", simulationN4AdditionalContent(9), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
		{"N4-10", simulationN4AdditionalContent(10), 57, 28, 8, []int{7, 5, 8, 4, 4, 13, 4, 4, 3, 3, 2}},
	}
	if simulationN5AdditionalContent(11) != nil || simulationN4AdditionalContent(1) != nil {
		t.Fatal("unexpected extra paper")
	}
	for _, c := range cases {
		if len(c.items) != c.total {
			t.Fatal(c.name, len(c.items))
		}
		counts := map[string]int{}
		order := []string{}
		vocabulary, reading := 0, 0
		prompts := map[string]bool{}
		for _, item := range c.items {
			group := plainSimulationText(item.group)
			if counts[group] == 0 {
				order = append(order, group)
			}
			counts[group]++
			switch item.section {
			case "vocabulary":
				vocabulary++
			case "grammar":
			default:
				t.Fatal(c.name, item.section)
			}
			if strings.HasPrefix(group, "読解") {
				reading++
				if item.passage == "" {
					t.Fatal(c.name, "missing reading passage")
				}
			}
			if item.prompt == "" || item.instructions == "" || item.explanation == "" {
				t.Fatal(c.name, "incomplete question")
			}
			if prompts[item.prompt+"|"+item.passage] {
				t.Fatal(c.name, "duplicate question", item.prompt)
			}
			prompts[item.prompt+"|"+item.passage] = true
			seen := map[string]bool{}
			for _, choice := range item.choices {
				if choice == "" || seen[plainSimulationText(choice)] {
					t.Fatal(c.name, "invalid options", item.prompt)
				}
				seen[plainSimulationText(choice)] = true
			}
			for _, text := range []string{item.group, item.instructions, item.passage, item.prompt, item.choices[0], item.choices[1], item.choices[2], item.choices[3]} {
				stripped := jlptRubyMarkup.ReplaceAllString(text, "")
				if strings.ContainsAny(stripped, "{}|") || strings.Count(text, "[[") != strings.Count(text, "]]") {
					t.Fatal(c.name, "unbalanced markup", text)
				}
			}
		}
		if vocabulary != c.vocabulary || reading != c.reading || len(order) != len(c.groups) {
			t.Fatalf("%s bad composition: %d %d %v", c.name, vocabulary, reading, counts)
		}
		for i, group := range order {
			if counts[group] != c.groups[i] {
				t.Fatal(c.name, group, counts[group])
			}
		}
	}
}

func TestSimulationJLPTExtraPapersAreFreeAndIsolated(t *testing.T) {
	s := NewMemoryStore()
	for _, examID := range []string{"jlpt-n5-simulation-2", "jlpt-n5-simulation-5", "jlpt-n4-simulation-1", "jlpt-n4-simulation-4", "jlpt-n4-simulation-5", "jlpt-n5-simulation-6", "jlpt-n5-simulation-7", "jlpt-n4-simulation-6", "jlpt-n4-simulation-7", "jlpt-n5-simulation-10", "jlpt-n4-simulation-10", "jlpt-n3-simulation-2", "jlpt-n3-simulation-3", "jlpt-n2-simulation-2", "jlpt-n2-simulation-3", "jlpt-n1-simulation-2", "jlpt-n1-simulation-3"} {
		if !IsFreeJLPTSimulationExam(examID) {
			t.Fatal("not free", examID)
		}
	}
	if IsFreeJLPTSimulationExam("jlpt-n3-simulation-1") || IsFreeJLPTSimulationExam("jlpt-n4-simulation-11") || IsFreeJLPTSimulationExam("jlpt-n3-simulation-11") || IsFreeJLPTSimulationExam("jlpt-n1-simulation-11") {
		t.Fatal("premium paper marked free")
	}
	if _, err := s.StartExam("free", "jlpt-n3-simulation-1"); !errors.Is(err, ErrPremium) {
		t.Fatal("N3 should stay premium")
	}
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	s.simulationNow = func() time.Time { return now }
	n5, err := s.StartExam("free", "jlpt-n5-simulation-2")
	if err != nil {
		t.Fatal(err)
	}
	n4, err := s.StartExam("free", "jlpt-n4-simulation-3")
	if err != nil {
		t.Fatal(err)
	}
	base, err := s.StartSimulation("free")
	if err != nil {
		t.Fatal(err)
	}
	if n5.ID == n4.ID || n5.ID == base.ID || n4.ID == base.ID || n4.ExamID != "jlpt-n4-simulation-3" {
		t.Fatal("papers share sessions")
	}
	session, err := s.SimulationSession("free", n4.ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.PaperID != "n4-language-03" || len(session.Questions) != 28 || session.ExpiresAt.Sub(now) != 25*time.Minute {
		t.Fatal("wrong N4 paper 03 vocabulary section", session.PaperID, len(session.Questions))
	}
	session, err = s.SimulationSession("free", n5.ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.PaperID != "n5-language-02" || len(session.Questions) != 21 || session.ExpiresAt.Sub(now) != 20*time.Minute {
		t.Fatal("wrong N5 paper 02 vocabulary section", session.PaperID, len(session.Questions))
	}
	resumed, err := s.StartSimulation("free", "N5-2")
	if err != nil || resumed.ID != n5.ID {
		t.Fatal("resume failed")
	}
	if s.SimulationOverview("free", "N5-2")["activeAttemptId"] != n5.ID || s.SimulationOverview("free", "N4-3")["activeAttemptId"] != n4.ID {
		t.Fatal("overview mixes papers")
	}
	if _, err = s.StartSimulation("free", "N5-11"); !errors.Is(err, ErrInvalid) {
		t.Fatal("unsupported paper accepted")
	}
}

func TestSimulationJLPTN3ToN1ExtraPapers(t *testing.T) {
	s := NewMemoryStore()
	for _, c := range []struct {
		level    string
		examID   string
		total    int
		sections int
	}{{"N3-2", "jlpt-n3-simulation-2", 74, 2}, {"N3-3", "jlpt-n3-simulation-3", 74, 2},
		{"N2-2", "jlpt-n2-simulation-2", 64, 1}, {"N2-3", "jlpt-n2-simulation-3", 64, 1},
		{"N1-2", "jlpt-n1-simulation-2", 64, 1}, {"N1-3", "jlpt-n1-simulation-3", 64, 1}} {
		attempt, err := s.StartExam("free", c.examID)
		if err != nil {
			t.Fatalf("%s: %v", c.examID, err)
		}
		session, err := s.SimulationSession("free", attempt.ID)
		if err != nil {
			t.Fatal(err)
		}
		if attempt.ExamID != c.examID || len(session.Sections) != c.sections {
			t.Fatalf("%s: bad session %+v", c.examID, session.Sections)
		}
		overview, err := s.SimulationOverviewChecked("free", c.level)
		if err != nil {
			t.Fatal(err)
		}
		if overview["questionCount"] != c.total {
			t.Fatalf("%s: questionCount %v, want %d", c.level, overview["questionCount"], c.total)
		}
	}
	for _, level := range []string{"N3-11", "N2-11", "N1-11", "N3-1", "N1-02"} {
		if _, err := s.StartSimulation("free", level); !errors.Is(err, ErrInvalid) {
			t.Fatalf("%s should be invalid, got %v", level, err)
		}
	}
	if _, err := s.StartSimulation("free", "N1"); !errors.Is(err, ErrPremium) {
		t.Fatal("N1 paper 01 should stay premium")
	}
}

func TestSimulationNATAndJFTPapersSixToTen(t *testing.T) {
	s := NewMemoryStore()
	grantTestLifetime(t, s, "premium")
	for n := 6; n <= 10; n++ {
		for _, c := range []struct {
			examID string
			total  int
		}{{fmt.Sprintf("nat-5q-full-%d", n), 43}, {fmt.Sprintf("jft-a2-full-%d", n), 36}} {
			if _, err := s.StartExam("free", c.examID); !errors.Is(err, ErrPremium) {
				t.Fatalf("%s should be premium, got %v", c.examID, err)
			}
			attempt, err := s.StartExam("premium", c.examID)
			if err != nil {
				t.Fatalf("%s: %v", c.examID, err)
			}
			session, err := s.SimulationSession("premium", attempt.ID)
			if err != nil {
				t.Fatal(err)
			}
			total := 0
			for _, section := range session.Sections {
				total += section.QuestionCount
			}
			if attempt.ExamID != c.examID || total != c.total {
				t.Fatalf("%s: bad session total %d", c.examID, total)
			}
		}
	}
	for _, id := range []string{"nat-5q-full-11", "nat-5q-full-06", "nat-5q-full-0"} {
		if _, err := s.StartExam("premium", id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("%s should be not found, got %v", id, err)
		}
	}
	if IsJFTLevel("JFT-11") || IsJFTLevel("JFT-1") || IsJFTLevel("JFT-06") || !IsJFTLevel("JFT-10") {
		t.Fatal("JFT level parsing is wrong")
	}
}
