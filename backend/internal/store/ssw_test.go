package store

import (
	"encoding/json"
	"errors"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/testdb"
	"strings"
	"testing"
	"time"
)

func TestSSWContentCoverageAndPracticeGrading(t *testing.T) {
	loadSSW()
	for sector, count := range map[string]int{"nursing-care": 87, "building-cleaning": 60} {
		bank := sswData.Banks[sector]
		if len(bank) != count {
			t.Fatal(sector, len(bank))
		}
		seen := map[string]bool{}
		for _, q := range bank {
			if seen[q.ID] || q.Prompt == "" || q.Explanation == "" || q.LessonID == "" || q.Answer < 0 || q.Answer >= len(q.Options) {
				t.Fatal("incomplete question", q.ID)
			}
			seen[q.ID] = true
			options := map[string]bool{}
			for _, o := range q.Options {
				if o == "" || options[o] {
					t.Fatal("invalid options", q.ID)
				}
				options[o] = true
			}
		}
	}
	for level, count := range map[string]int{"SSW-KAIGO": 45, "SSW-KAIGO-JP": 15, "SSW-CLEANING": 50} {
		p, ok := sswSimulationPaper(level)
		if !ok || len(p.items) != count {
			t.Fatal(level, len(p.items))
		}
		for _, part := range p.sections {
			n := 0
			for _, q := range p.items {
				if q.section == part.ID {
					n++
				}
			}
			if n != part.QuestionCount {
				t.Fatal("section mismatch", level, part.ID, n)
			}
		}
	}
	s := NewMemoryStore()
	if _, err := s.SSWPractice("free", "building-cleaning", ""); !errors.Is(err, ErrPremium) {
		t.Fatal("free practice permitted")
	}
	grantTestLifetime(t, s, "user")
	list, err := s.SSWPractice("user", "building-cleaning", "chemicals")
	if err != nil || len(list) != 6 {
		t.Fatal(len(list), err)
	}
	raw, _ := json.Marshal(list)
	if strings.Contains(string(raw), `"answer"`) || strings.Contains(string(raw), `"explanation"`) {
		t.Fatal("practice key exposed before check")
	}
	for _, q := range sswData.Banks["building-cleaning"] {
		result, err := s.SSWPracticeCheck("user", "building-cleaning", q.ID, q.Answer)
		if err != nil || result["correct"] != true {
			t.Fatal(q.ID, err)
		}
	}
	if _, err := s.SSWPracticeCheck("user", "building-cleaning", "cleaning-01-1", 99); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid choice accepted")
	}
	if s.Gamification("user").XP != 0 {
		t.Fatal("practice awarded XP")
	}
}

func TestSSWSimulationSectionsClockOwnershipAndIdempotence(t *testing.T) {
	s := NewMemoryStore()
	grantTestLifetime(t, s, "user")
	now := time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)
	s.simulationNow = func() time.Time { return now }
	for _, level := range []string{"SSW-KAIGO", "SSW-KAIGO-JP", "SSW-CLEANING"} {
		if _, err := s.StartSimulation("free", level); !errors.Is(err, ErrPremium) {
			t.Fatal("free exam allowed")
		}
		a, err := s.StartSimulation("user", level)
		if err != nil {
			t.Fatal(err)
		}
		again, _ := s.StartExam("user", strings.ToLower(level)+"-simulation-1")
		if again.ID != a.ID {
			t.Fatal("resume made duplicate")
		}
		if _, err = s.SimulationSession("other", a.ID); !errors.Is(err, ErrNotFound) {
			t.Fatal("owner leak")
		}
		session, err := s.SimulationSession("user", a.ID)
		if err != nil || session.Result != nil {
			t.Fatal(err)
		}
		for _, q := range session.Questions {
			for _, o := range q.Options {
				if o.Text == "" {
					t.Fatal("blank choice")
				}
			}
		}
		raw, _ := json.Marshal(session)
		if strings.Contains(string(raw), "correctOptionId") || strings.Contains(string(raw), "explanation") {
			t.Fatal("hidden key exposed")
		}
		q := session.Questions[0]
		if err = s.SaveSimulationAnswer("user", a.ID, q.ID, s.simulationAttempts[a.ID].correct[q.ID], true); err != nil {
			t.Fatal(err)
		}
		if err = s.SaveSimulationAnswer("user", a.ID, q.ID, "not-an-option", false); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid choice accepted")
		}
		first := session.ActiveSection
		out, err := s.SubmitSimulationSection("user", a.ID, first)
		if err != nil {
			t.Fatal(err)
		}
		if level == "SSW-CLEANING" {
			if out.ActiveSection != "grammar" || len(out.Questions) != 30 || out.ExpiresAt.Sub(now) != 30*time.Minute {
				t.Fatal("bad practical boundary", out.ActiveSection)
			}
			if err = s.SaveSimulationAnswer("user", a.ID, q.ID, q.Options[0].ID, false); !errors.Is(err, ErrInvalid) {
				t.Fatal("locked theory editable")
			}
			retry, _ := s.SubmitSimulationSection("user", a.ID, first)
			if retry.ActiveSection != "grammar" {
				t.Fatal("retry advanced twice")
			}
			now = now.Add(31 * time.Minute)
			out, err = s.SimulationSession("user", a.ID)
		}
		if err != nil || out.Result == nil || out.Result.Correct != 1 || out.Result.Answered != 1 {
			t.Fatal("bad result", err)
		}
		if !strings.Contains(out.Result.ScoringNote, "SSW") {
			t.Fatal("wrong scoring note")
		}
		overview, _ := s.SimulationOverviewChecked("user", level)
		if len(overview["history"].([]map[string]any)) != 1 {
			t.Fatal("missing history")
		}
	}
	if err := s.DeleteAccount("user"); err != nil {
		t.Fatal(err)
	}
	for _, a := range s.simulationAttempts {
		if a.session.Attempt.UserID == "user" {
			t.Fatal("deleted exam retained")
		}
	}
}

func TestSSWMariaDBResumeAndFailedWrite(t *testing.T) {
	dsn := testdb.DSN(t)
	user := database.DemoUserID
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	grantTestLifetime(t, s, user)
	a, err := s.StartSimulation(user, "SSW-CLEANING")
	if err != nil {
		t.Fatal(err)
	}
	session, _ := s.SimulationSession(user, a.ID)
	q := session.Questions[0]
	choice := s.simulationAttempts[a.ID].correct[q.ID]
	if err = s.SaveSimulationAnswer(user, a.ID, q.ID, choice, true); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	restored := NewMemoryStore()
	if err = restored.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	grantTestLifetime(t, restored, user)
	got, err := restored.SimulationSession(user, a.ID)
	if err != nil || got.Answers[q.ID] != choice || len(got.Marked) != 1 || !got.ExpiresAt.Equal(session.ExpiresAt) {
		t.Fatal("restart lost session", err)
	}
	after, err := restored.SubmitSimulationSection(user, a.ID, "vocabulary")
	if err != nil || after.ActiveSection != "grammar" {
		t.Fatal(err)
	}
	if err = restored.persistenceConn.Close(); err != nil {
		t.Fatal(err)
	}
	next := after.Questions[0]
	if err = restored.SaveSimulationAnswer(user, a.ID, next.ID, next.Options[0].ID, false); err == nil {
		t.Fatal("database error hidden")
	}
	if len(restored.simulationAttempts[a.ID].session.Answers) != 1 {
		t.Fatal("failed write changed memory")
	}
	if _, err = restored.SubmitSimulationSection(user, a.ID, "grammar"); err == nil {
		t.Fatal("failed completion write hidden")
	}
	if restored.simulationAttempts[a.ID].session.Result != nil || restored.Gamification(user).ExamsCompleted != 0 {
		t.Fatal("failed completion published result or reward")
	}
}
