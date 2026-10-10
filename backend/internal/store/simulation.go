package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"nihongo-master-backend/internal/domain"
)

type simulationAttemptState struct {
	session      domain.SimulationSession
	questions    []domain.SimulationQuestion
	correct      map[string]string
	explanations map[string]string
	marked       map[string]bool
}

func simulationSections() []domain.SimulationSection {
	return []domain.SimulationSection{
		{ID: "vocabulary", Name: "文字・語彙 · Kosakata & Kanji", DurationSeconds: 1200, QuestionCount: 21},
		{ID: "grammar", Name: "文法 · Grammar", DurationSeconds: 1200, QuestionCount: 17},
	}
}

type simulationPaperConfig struct {
	level, examID, paperID string
	sections               []domain.SimulationSection
	items                  []simulationItem
	omitted                []string
}

func simulationPaper(levels ...string) simulationPaperConfig {
	level := "N5"
	if len(levels) > 0 {
		level = strings.ToUpper(levels[0])
	}
	if level == "N4" {
		return simulationPaperConfig{level, "jlpt-n4-simulation-1", "n4-language-01", []domain.SimulationSection{
			{ID: "vocabulary", Name: "文字・語彙 · Kosakata & Kanji", DurationSeconds: 1500, QuestionCount: 28},
			{ID: "grammar", Name: "文法・読解 · Grammar & Reading", DurationSeconds: 3300, QuestionCount: 29},
		}, simulationN4Content(), []string{"listening"}}
	}
	return simulationPaperConfig{level, "jlpt-n5-simulation-1", "n5-language-01", simulationSections(), simulationContent(), []string{"dokkai", "listening"}}
}

func (s *MemoryStore) StartSimulation(userID string, levels ...string) (domain.ExamAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.simulationNow().UTC()
	paper := simulationPaper(levels...)
	if paper.level != "N5" && paper.level != "N4" {
		return domain.ExamAttempt{}, ErrInvalid
	}
	if paper.level != "N5" && s.entitlementLocked(userID).Tier != domain.TierPremium {
		return domain.ExamAttempt{}, ErrPremium
	}
	for i := len(s.simulationOrder) - 1; i >= 0; i-- {
		state := s.simulationAttempts[s.simulationOrder[i]]
		if state.session.Attempt.UserID != userID || state.session.Attempt.ExamID != paper.examID {
			continue
		}
		if err := s.syncSimulationLocked(state, now); err != nil {
			return domain.ExamAttempt{}, err
		}
		if state.session.Result == nil {
			return state.session.Attempt, nil
		}
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return domain.ExamAttempt{}, err
	}
	attempt := domain.ExamAttempt{ID: "simulation-" + hex.EncodeToString(raw), ExamID: paper.examID, UserID: userID, Status: "in_progress", StartedAt: now}
	state := &simulationAttemptState{
		session: domain.SimulationSession{Attempt: attempt, PaperID: paper.paperID, ActiveSection: "vocabulary", SectionStartedAt: now, ExpiresAt: now.Add(time.Duration(paper.sections[0].DurationSeconds) * time.Second), Sections: paper.sections, Answers: map[string]string{}},
		correct: map[string]string{}, explanations: map[string]string{}, marked: map[string]bool{},
	}
	for i, item := range paper.items {
		id := fmt.Sprintf("%s-sim-01-q%02d", strings.ToLower(paper.level), i+1)
		options := make([]domain.QuizOption, 4)
		for j, choice := range item.choices {
			optionKey := make([]byte, 8)
			if _, err := rand.Read(optionKey); err != nil {
				return domain.ExamAttempt{}, err
			}
			options[j] = domain.QuizOption{ID: hex.EncodeToString(optionKey), Text: choice}
		}
		state.correct[id] = options[0].ID
		state.explanations[id] = item.explanation
		// Cryptographic Fisher-Yates prevents a predictable answer-position pattern.
		for j := 3; j > 0; j-- {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(j+1)))
			if err != nil {
				return domain.ExamAttempt{}, err
			}
			k := int(n.Int64())
			options[j], options[k] = options[k], options[j]
		}
		state.questions = append(state.questions, domain.SimulationQuestion{ID: id, Number: i + 1, Section: item.section, Group: item.group, Instructions: item.instructions, Passage: item.passage, Prompt: item.prompt, Options: options})
	}
	s.simulationAttempts[attempt.ID] = state
	s.simulationOrder = append(s.simulationOrder, attempt.ID)
	return attempt, nil
}

// Expired sections advance at their original deadline, not at reconnection time.
func (s *MemoryStore) syncSimulationLocked(state *simulationAttemptState, now time.Time) error {
	for state.session.Result == nil && !now.Before(state.session.ExpiresAt) {
		if err := s.advanceSimulationLocked(state, state.session.ExpiresAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryStore) advanceSimulationLocked(state *simulationAttemptState, at time.Time) error {
	if state.session.ActiveSection == "vocabulary" {
		state.session.ActiveSection = "grammar"
		state.session.SectionStartedAt = at
		state.session.ExpiresAt = at.Add(time.Duration(state.session.Sections[1].DurationSeconds) * time.Second)
		return nil
	}
	result := domain.SimulationResult{AttemptID: state.session.Attempt.ID, PaperID: state.session.PaperID, Total: len(state.questions), Answered: len(state.session.Answers), CompletedAt: at, DurationSeconds: int(at.Sub(state.session.Attempt.StartedAt).Seconds()), Sections: []domain.DiagnosticSection{}, Reviews: []domain.SimulationReview{}, ScoringNote: "Skor latihan = persentase jawaban benar. Dokkai dan listening dilewati. Ini bukan skor berskala atau keputusan kelulusan resmi JLPT."}
	if state.session.PaperID == "n4-language-01" {
		result.ScoringNote = "Skor latihan = persentase jawaban benar. Listening dilewati. Ini bukan skor berskala atau keputusan kelulusan resmi JLPT."
	}
	groups := map[string]*domain.DiagnosticSection{}
	for _, q := range state.questions {
		if groups[q.Group] == nil {
			groups[q.Group] = &domain.DiagnosticSection{Name: q.Group}
		}
		group := groups[q.Group]
		group.Total++
		selected := state.session.Answers[q.ID]
		correct := selected == state.correct[q.ID]
		if correct {
			result.Correct++
			group.Correct++
		}
		result.Reviews = append(result.Reviews, domain.SimulationReview{Question: q, SelectedOptionID: selected, CorrectOptionID: state.correct[q.ID], Correct: correct, Explanation: state.explanations[q.ID]})
	}
	seen := map[string]bool{}
	for _, q := range state.questions {
		if !seen[q.Group] {
			seen[q.Group] = true
			group := groups[q.Group]
			group.Score = group.Correct * 100 / group.Total
			result.Sections = append(result.Sections, *group)
		}
	}
	result.Score = result.Correct * 100 / result.Total
	if err := s.recordStudyLocked(state.session.Attempt.UserID, studyEvent{ID: "simulation:" + state.session.Attempt.ID, Kind: "exam", ExamID: state.session.Attempt.ExamID, Score: result.Score, Answered: result.Answered, Seconds: result.DurationSeconds, At: at}); err != nil {
		return err
	}
	state.session.Result = &result
	state.session.Attempt.Status = "completed"
	state.session.ActiveSection = "completed"
	return nil
}

// Return detached objects: callers cannot mutate live answers or option slices.
func simulationSnapshot(state *simulationAttemptState, now time.Time) domain.SimulationSession {
	out := state.session
	out.ServerTime = now
	out.Answers = map[string]string{}
	out.Questions = []domain.SimulationQuestion{}
	out.Marked = []string{}
	for _, q := range state.questions {
		if out.Result == nil && q.Section != out.ActiveSection {
			continue
		}
		q.Options = append([]domain.QuizOption(nil), q.Options...)
		out.Questions = append(out.Questions, q)
		if option, ok := state.session.Answers[q.ID]; ok {
			out.Answers[q.ID] = option
		}
		if state.marked[q.ID] {
			out.Marked = append(out.Marked, q.ID)
		}
	}
	out.Sections = append([]domain.SimulationSection(nil), out.Sections...)
	if out.Result != nil {
		r := *out.Result
		r.Sections = append([]domain.DiagnosticSection(nil), r.Sections...)
		r.Reviews = append([]domain.SimulationReview(nil), r.Reviews...)
		for i := range r.Reviews {
			r.Reviews[i].Question.Options = append([]domain.QuizOption(nil), r.Reviews[i].Question.Options...)
		}
		out.Result = &r
	}
	return out
}

func (s *MemoryStore) SimulationSession(userID, id string) (domain.SimulationSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.simulationAttempts[id]
	if !ok || state.session.Attempt.UserID != userID {
		return domain.SimulationSession{}, ErrNotFound
	}
	now := s.simulationNow().UTC()
	if err := s.syncSimulationLocked(state, now); err != nil {
		return domain.SimulationSession{}, err
	}
	return simulationSnapshot(state, now), nil
}

func (s *MemoryStore) SaveSimulationAnswer(userID, id, question, option string, marked bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.simulationAttempts[id]
	if !ok || state.session.Attempt.UserID != userID {
		return ErrNotFound
	}
	now := s.simulationNow().UTC()
	if err := s.syncSimulationLocked(state, now); err != nil {
		return err
	}
	if state.session.Result != nil {
		return ErrInvalid
	}
	for _, q := range state.questions {
		if q.ID != question || q.Section != state.session.ActiveSection {
			continue
		}
		valid := option == ""
		for _, o := range q.Options {
			if o.ID == option {
				valid = true
			}
		}
		if !valid {
			return ErrInvalid
		}
		if option == "" {
			delete(state.session.Answers, question)
		} else {
			state.session.Answers[question] = option
		}
		state.marked[question] = marked
		return nil
	}
	return ErrInvalid
}

// The expected section makes retries idempotent across the vocabulary boundary.
func (s *MemoryStore) SubmitSimulationSection(userID, id, section string) (domain.SimulationSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.simulationAttempts[id]
	if !ok || state.session.Attempt.UserID != userID {
		return domain.SimulationSession{}, ErrNotFound
	}
	now := s.simulationNow().UTC()
	if err := s.syncSimulationLocked(state, now); err != nil {
		return domain.SimulationSession{}, err
	}
	if section != "vocabulary" && section != "grammar" {
		return domain.SimulationSession{}, ErrInvalid
	}
	if section == "grammar" && state.session.ActiveSection == "vocabulary" {
		return domain.SimulationSession{}, ErrInvalid
	}
	if state.session.Result == nil && state.session.ActiveSection == section {
		if err := s.advanceSimulationLocked(state, now); err != nil {
			return domain.SimulationSession{}, err
		}
	}
	return simulationSnapshot(state, now), nil
}

func (s *MemoryStore) SimulationOverview(userID string, levels ...string) map[string]any {
	result, _ := s.SimulationOverviewChecked(userID, levels...)
	return result
}
func (s *MemoryStore) SimulationOverviewChecked(userID string, levels ...string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.simulationNow().UTC()
	paper := simulationPaper(levels...)
	history := []map[string]any{}
	active := ""
	for i := len(s.simulationOrder) - 1; i >= 0; i-- {
		state := s.simulationAttempts[s.simulationOrder[i]]
		if state.session.Attempt.UserID != userID || state.session.Attempt.ExamID != paper.examID {
			continue
		}
		if err := s.syncSimulationLocked(state, now); err != nil {
			return nil, err
		}
		if r := state.session.Result; r != nil {
			if len(history) < 10 {
				history = append(history, map[string]any{"attemptId": r.AttemptID, "paperId": r.PaperID, "score": r.Score, "correct": r.Correct, "total": r.Total, "completedAt": r.CompletedAt})
			}
		} else if active == "" {
			active = state.session.Attempt.ID
		}
	}
	duration := 0
	for _, section := range paper.sections {
		duration += section.DurationSeconds
	}
	return map[string]any{"paperId": paper.paperID, "title": "JLPT " + paper.level + " Simulation", "questionCount": len(paper.items), "durationMinutes": duration / 60, "sections": paper.sections, "activeAttemptId": active, "history": history, "omittedSections": paper.omitted}, nil
}
