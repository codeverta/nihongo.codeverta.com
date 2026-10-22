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

func jftPaperNumber(level string) int {
	switch strings.ToUpper(level) {
	case "JFT":
		return 1
	case "JFT-2":
		return 2
	case "JFT-3":
		return 3
	case "JFT-4":
		return 4
	case "JFT-5":
		return 5
	default:
		return 0
	}
}

func IsJFTLevel(level string) bool {
	return jftPaperNumber(level) != 0
}

func simulationPaper(levels ...string) simulationPaperConfig {
	level := "N5"
	if len(levels) > 0 {
		level = strings.ToUpper(levels[0])
	}
	if paper, ok := sswSimulationPaper(level); ok {
		return paper
	}
	if level == "NAT" || strings.HasPrefix(level, "NAT-") {
		paperNo := 1
		if level != "NAT" {
			if _, err := fmt.Sscanf(level, "NAT-%d", &paperNo); err != nil || paperNo < 2 || paperNo > 5 {
				paperNo = 1
			}
		}
		items := simulationNATContent()
		if paperNo > 1 {
			items = simulationNATAdditionalContent(paperNo)
		}
		count := len(items)
		vocabularyCount := count / 2
		return simulationPaperConfig{level, fmt.Sprintf("nat-5q-full-%d", paperNo), fmt.Sprintf("nat-5q-written-%02d", paperNo), []domain.SimulationSection{
			{ID: "vocabulary", Name: "文字・語彙 · Kosakata & Kanji", DurationSeconds: 1200, QuestionCount: vocabularyCount},
			{ID: "grammar", Name: "文法・読解 · Grammar & Reading", DurationSeconds: 2400, QuestionCount: count - vocabularyCount},
		}, items, []string{"listening"}}
	}
	if paperNo := jftPaperNumber(level); paperNo > 0 {
		items := simulationJFTContent()
		if paperNo > 1 {
			items = simulationJFTAdditionalContent(paperNo)
		}
		return simulationPaperConfig{level, fmt.Sprintf("jft-a2-full-%d", paperNo), fmt.Sprintf("jft-written-%02d", paperNo), []domain.SimulationSection{
			{ID: "vocabulary", Name: "文字と語彙 · Kosakata", DurationSeconds: 2700, QuestionCount: 12},
			{ID: "conversation", Name: "会話と表現 · Percakapan & Ungkapan", QuestionCount: 12},
			{ID: "reading", Name: "読解 · Reading", QuestionCount: 12},
		}, items, []string{"listening"}}
	}
	if level == "N3" {
		return simulationPaperConfig{level, "jlpt-n3-simulation-1", "n3-language-01", []domain.SimulationSection{
			{ID: "vocabulary", Name: "文字・語彙 · Kosakata & Kanji", DurationSeconds: 1800, QuestionCount: 35},
			{ID: "grammar", Name: "文法・読解 · Grammar & Reading", DurationSeconds: 4200, QuestionCount: 39},
		}, simulationN3Content(), []string{"listening"}}
	}
	if level == "N2" {
		return simulationPaperConfig{level, "jlpt-n2-simulation-1", "n2-language-01", []domain.SimulationSection{
			{ID: "written", Name: "言語知識（文字・語彙・文法）・読解 · Kosakata, Grammar & Reading", DurationSeconds: 6300, QuestionCount: 64},
		}, simulationN2Content(), []string{"listening"}}
	}
	if level == "N1" {
		return simulationPaperConfig{level, "jlpt-n1-simulation-1", "n1-language-01", []domain.SimulationSection{
			{ID: "written", Name: "言語知識（文字・語彙・文法）・読解 · Kosakata, Grammar & Reading", DurationSeconds: 6600, QuestionCount: 64},
		}, simulationN1Content(), []string{"listening"}}
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
	if paper.level != "N5" && paper.level != "N4" && paper.level != "N3" && paper.level != "N2" && paper.level != "N1" && !IsJFTLevel(paper.level) && paper.level != "NAT" && !strings.HasPrefix(paper.level, "NAT-") && !IsSSWLevel(paper.level) {
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
		session: domain.SimulationSession{Attempt: attempt, PaperID: paper.paperID, ActiveSection: paper.sections[0].ID, SectionStartedAt: now, ExpiresAt: now.Add(time.Duration(paper.sections[0].DurationSeconds) * time.Second), Sections: paper.sections, Answers: map[string]string{}},
		correct: map[string]string{}, explanations: map[string]string{}, marked: map[string]bool{},
	}
	for i, item := range paper.items {
		id := fmt.Sprintf("%s-sim-01-q%02d", strings.ToLower(paper.level), i+1)
		options := make([]domain.QuizOption, 0, 4)
		for _, choice := range item.choices {
			if choice == "" {
				continue
			}
			optionKey := make([]byte, 8)
			if _, err := rand.Read(optionKey); err != nil {
				return domain.ExamAttempt{}, err
			}
			options = append(options, domain.QuizOption{ID: hex.EncodeToString(optionKey), Text: choice})
		}
		state.correct[id] = options[0].ID
		state.explanations[id] = item.explanation
		// Cryptographic Fisher-Yates prevents a predictable answer-position pattern.
		for j := len(options) - 1; j > 0; j-- {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(j+1)))
			if err != nil {
				return domain.ExamAttempt{}, err
			}
			k := int(n.Int64())
			options[j], options[k] = options[k], options[j]
		}
		state.questions = append(state.questions, domain.SimulationQuestion{ID: id, Number: i + 1, Section: item.section, Group: item.group, Instructions: item.instructions, Passage: item.passage, Prompt: item.prompt, Options: options})
	}
	if err := s.persistSSWLocked(state); err != nil {
		return domain.ExamAttempt{}, err
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

func (s *MemoryStore) advanceSimulationLocked(state *simulationAttemptState, at time.Time) (err error) {
	if strings.HasPrefix(state.session.PaperID, "ssw-") {
		before := cloneSimulation(state)
		defer func() {
			if err == nil && state.session.Result == nil {
				err = s.persistSSWLocked(state)
			}
			if err != nil {
				*state = *before
			}
		}()
	}
	for i, section := range state.session.Sections {
		if section.ID == state.session.ActiveSection && i+1 < len(state.session.Sections) {
			next := state.session.Sections[i+1]
			state.session.ActiveSection = next.ID
			state.session.SectionStartedAt = at
			// JFT written practice shares one deadline across all three sections.
			if !strings.HasPrefix(state.session.PaperID, "jft-written-") {
				state.session.ExpiresAt = at.Add(time.Duration(next.DurationSeconds) * time.Second)
			}
			return nil
		}
	}
	result := domain.SimulationResult{AttemptID: state.session.Attempt.ID, PaperID: state.session.PaperID, Total: len(state.questions), Answered: len(state.session.Answers), CompletedAt: at, DurationSeconds: int(at.Sub(state.session.Attempt.StartedAt).Seconds()), Sections: []domain.DiagnosticSection{}, Reviews: []domain.SimulationReview{}, ScoringNote: "Skor latihan = persentase jawaban benar. Dokkai dan listening dilewati. Ini bukan skor berskala atau keputusan kelulusan resmi JLPT."}
	if state.session.PaperID == "n4-language-01" || state.session.PaperID == "n3-language-01" || state.session.PaperID == "n2-language-01" || state.session.PaperID == "n1-language-01" {
		result.ScoringNote = "Skor latihan = persentase jawaban benar. Listening dilewati. Ini bukan skor berskala atau keputusan kelulusan resmi JLPT."
	}
	if strings.HasPrefix(state.session.PaperID, "jft-written-") {
		result.ScoringNote = "Skor latihan tertulis = persentase jawaban benar. Listening belum termasuk. Soal orisinal; bukan skor berskala atau keputusan kelulusan resmi JFT-Basic."
	}
	if strings.HasPrefix(state.session.PaperID, "nat-5q-written-") {
		result.ScoringNote = "Skor latihan tertulis = persentase jawaban benar. Listening belum termasuk. Soal orisinal; bukan skor berskala atau keputusan kelulusan resmi NAT-TEST."
	}
	if strings.HasPrefix(state.session.PaperID, "ssw-") {
		result.ScoringNote = "Soal orisinal Nihongo Master. Nilai adalah persentase benar, bukan hasil resmi SSW. Penilaian praktik di aplikasi memakai skenario teks, tanpa demonstrasi alat atau seluruh variasi visual resmi."
		if strings.HasPrefix(state.session.PaperID, "ssw-agriculture-") {
			result.ScoringNote += " Adaptasi tertulis: komponen audio instruksi Jepang belum tersedia; hasil tidak mengukur kemampuan mendengar."
		}
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
	state.session.Result = &result
	state.session.Attempt.Status = "completed"
	state.session.ActiveSection = "completed"
	event := studyEvent{ID: "simulation:" + state.session.Attempt.ID, Kind: "exam", ExamID: state.session.Attempt.ExamID, Score: result.Score, Answered: result.Answered, Seconds: result.DurationSeconds, At: at}
	if strings.HasPrefix(state.session.PaperID, "ssw-") {
		event.SSWAttempt = &SSWAttemptRecord{state.session, state.questions, state.correct, state.explanations, state.marked}
	}
	if err := s.recordStudyLocked(state.session.Attempt.UserID, event); err != nil {
		state.session.Result = nil
		state.session.Attempt.Status = "in_progress"
		state.session.ActiveSection = state.session.Sections[len(state.session.Sections)-1].ID
		return err
	}
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
		before := cloneSimulation(state)
		if option == "" {
			delete(state.session.Answers, question)
		} else {
			state.session.Answers[question] = option
		}
		state.marked[question] = marked
		if err := s.persistSSWLocked(state); err != nil {
			*state = *before
			return err
		}
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
	requested, active := -1, len(state.session.Sections)
	for i, part := range state.session.Sections {
		if part.ID == section {
			requested = i
		}
		if part.ID == state.session.ActiveSection {
			active = i
		}
	}
	if requested < 0 {
		return domain.SimulationSession{}, ErrInvalid
	}
	if requested > active {
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
	title := "JLPT " + paper.level + " Simulation"
	if IsJFTLevel(paper.level) {
		title = fmt.Sprintf("JFT Exam %d · Latihan tertulis", jftPaperNumber(paper.level))
	}
	if paper.level == "NAT" {
		title = "NAT 5Q Exam 1 · Latihan tertulis"
	}
	if strings.HasPrefix(paper.level, "NAT-") {
		title = fmt.Sprintf("NAT 5Q Exam %s · Latihan tertulis", strings.TrimPrefix(paper.level, "NAT-"))
	}
	if IsSSWLevel(paper.level) {
		title = sswTitle(paper.level)
	}
	return map[string]any{"paperId": paper.paperID, "title": title, "questionCount": len(paper.items), "durationMinutes": duration / 60, "sections": paper.sections, "activeAttemptId": active, "history": history, "omittedSections": paper.omitted, "sharedTimer": IsJFTLevel(paper.level)}, nil
}
