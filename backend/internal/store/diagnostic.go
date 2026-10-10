package store

import (
	"fmt"
	"time"

	"nihongo-master-backend/internal/domain"
)

type diagnosticAttemptState struct {
	session      domain.DiagnosticSession
	correct      map[string]string
	explanations map[string]string
}

func (s *MemoryStore) DiagnosticOverview(userID string) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var exam domain.Exam
	for _, e := range s.exams {
		if e.ID == "jlpt-n5-diagnostic" {
			exam = e
		}
	}
	history := make([]domain.DiagnosticResult, 0)
	active := ""
	for i := len(s.diagnosticOrder) - 1; i >= 0; i-- {
		a := s.diagnosticAttempts[s.diagnosticOrder[i]].session
		if a.Attempt.UserID != userID {
			continue
		}
		if a.Result == nil && active == "" {
			active = a.Attempt.ID
		}
		if a.Result != nil && len(history) < 10 {
			history = append(history, *a.Result)
		}
	}
	counts := map[string]int{}
	for _, item := range diagnosticContent() {
		counts[item.section]++
	}
	return map[string]any{"exam": exam, "sectionCounts": counts, "history": history, "activeAttemptId": active}
}

func (s *MemoryStore) startDiagnostic(userID string) domain.ExamAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Return an unfinished attempt on repeated taps or network retries.
	for i := len(s.diagnosticOrder) - 1; i >= 0; i-- {
		a := s.diagnosticAttempts[s.diagnosticOrder[i]].session
		if a.Attempt.UserID == userID && a.Result == nil {
			return a.Attempt
		}
	}
	now := time.Now().UTC()
	attempt := domain.ExamAttempt{ID: fmt.Sprintf("diagnostic-%d", s.attemptID.Add(1)), ExamID: "jlpt-n5-diagnostic", UserID: userID, Status: "in_progress", StartedAt: now}
	state := &diagnosticAttemptState{session: domain.DiagnosticSession{Attempt: attempt, ExpiresAt: now.Add(30 * time.Minute), Questions: make([]domain.DiagnosticQuestion, 0, 25), Answers: map[string]string{}}, correct: map[string]string{}, explanations: map[string]string{}}
	for i, item := range diagnosticContent() {
		id := fmt.Sprintf("n5-diagnostic-q%02d", i+1)
		options := make([]domain.QuizOption, 4)
		for j, choice := range item.choices {
			options[j] = domain.QuizOption{ID: fmt.Sprintf("%s-o%d", id, j+1), Text: choice}
		}
		state.correct[id] = options[0].ID
		state.explanations[id] = item.explanation
		shiftOptions(options, i%4)
		state.session.Questions = append(state.session.Questions, domain.DiagnosticQuestion{ID: id, Section: item.section, Passage: item.passage, Prompt: item.prompt, Options: options})
	}
	s.diagnosticAttempts[attempt.ID] = state
	s.diagnosticOrder = append(s.diagnosticOrder, attempt.ID)
	return attempt
}

func (s *MemoryStore) DiagnosticSession(userID, id string) (domain.DiagnosticSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.diagnosticAttempts[id]
	if !ok || a.session.Attempt.UserID != userID {
		return domain.DiagnosticSession{}, ErrNotFound
	}
	result := a.session
	result.ServerTime = time.Now().UTC()
	result.Answers = make(map[string]string, len(a.session.Answers))
	for q, o := range a.session.Answers {
		result.Answers[q] = o
	}
	return result, nil
}

func (s *MemoryStore) SaveDiagnosticAnswer(userID, id, questionID, optionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.diagnosticAttempts[id]
	if !ok || a.session.Attempt.UserID != userID {
		return ErrNotFound
	}
	if a.session.Result != nil || !time.Now().Before(a.session.ExpiresAt) {
		return ErrInvalid
	}
	for _, q := range a.session.Questions {
		if q.ID != questionID {
			continue
		}
		for _, option := range q.Options {
			if option.ID == optionID {
				a.session.Answers[questionID] = optionID
				return nil
			}
		}
	}
	return ErrInvalid
}

func (s *MemoryStore) CompleteDiagnostic(userID, id string) (domain.DiagnosticResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.diagnosticAttempts[id]
	if !ok || a.session.Attempt.UserID != userID {
		return domain.DiagnosticResult{}, ErrNotFound
	}
	if a.session.Result != nil {
		return *a.session.Result, nil
	}
	now := time.Now().UTC()
	elapsed := int(now.Sub(a.session.Attempt.StartedAt).Seconds())
	if elapsed > 1800 {
		elapsed = 1800
	}
	result := domain.DiagnosticResult{AttemptID: id, Total: len(a.session.Questions), Answered: len(a.session.Answers), CompletedAt: now, DurationSeconds: elapsed, Sections: []domain.DiagnosticSection{}, Reviews: []domain.DiagnosticReview{}}
	sections := map[string]*domain.DiagnosticSection{}
	for _, name := range []string{"Kosakata", "Kanji", "Grammar", "Reading"} {
		sections[name] = &domain.DiagnosticSection{Name: name}
	}
	for _, q := range a.session.Questions {
		selected := a.session.Answers[q.ID]
		correct := selected == a.correct[q.ID]
		section := sections[q.Section]
		section.Total++
		if correct {
			result.Correct++
			section.Correct++
		}
		result.Reviews = append(result.Reviews, domain.DiagnosticReview{Question: q, SelectedOptionID: selected, CorrectOptionID: a.correct[q.ID], Correct: correct, Explanation: a.explanations[q.ID]})
	}
	result.Score = result.Correct * 100 / result.Total
	for _, name := range []string{"Kosakata", "Kanji", "Grammar", "Reading"} {
		section := sections[name]
		section.Score = section.Correct * 100 / section.Total
		result.Sections = append(result.Sections, *section)
	}
	if err := s.recordStudyLocked(userID, studyEvent{ID: "diagnostic:" + id, Kind: "exam", ExamID: a.session.Attempt.ExamID, Score: result.Score, Answered: result.Answered, Seconds: result.DurationSeconds, At: now}); err != nil {
		return domain.DiagnosticResult{}, err
	}
	a.session.Attempt.Status = "completed"
	a.session.Result = &result
	return result, nil
}
