package store

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
)

//go:embed ssw_resources/assessments.json
var sswJSON []byte

type sswQuestion struct {
	ID          string   `json:"id"`
	ModuleID    string   `json:"moduleId"`
	Group       string   `json:"group"`
	LessonID    string   `json:"lessonId"`
	Prompt      string   `json:"prompt"`
	Options     []string `json:"options"`
	Answer      int      `json:"answer"`
	Explanation string   `json:"explanation"`
	Section     string   `json:"section"`
	Passage     string   `json:"passage"`
	Track       string   `json:"track"`
	Kind        string   `json:"kind"`
}
type sswPaper struct {
	Title    string                     `json:"title"`
	Sections []domain.SimulationSection `json:"sections"`
	Items    []sswQuestion              `json:"items"`
	Omitted  []string                   `json:"omitted"`
}

var sswData struct {
	Banks  map[string][]sswQuestion `json:"banks"`
	Papers map[string]sswPaper      `json:"papers"`
}
var sswOnce sync.Once

func loadSSW() {
	sswOnce.Do(func() {
		if err := json.Unmarshal(sswJSON, &sswData); err != nil {
			panic(err)
		}
	})
}
func sswSimulationPaper(level string) (simulationPaperConfig, bool) {
	loadSSW()
	p, ok := sswData.Papers[level]
	if !ok {
		return simulationPaperConfig{}, false
	}
	items := make([]simulationItem, 0, len(p.Items))
	for _, q := range p.Items {
		choices := [4]string{q.Options[q.Answer]}
		n := 1
		for i, option := range q.Options {
			if i != q.Answer {
				choices[n] = option
				n++
			}
		}
		items = append(items, simulationItem{q.Section, q.Group, "Pilih satu jawaban paling tepat. Soal orisinal Nihongo Master.", q.Passage, q.Prompt, choices, q.Explanation})
	}
	omitted := p.Omitted
	if omitted == nil {
		omitted = []string{}
	}
	return simulationPaperConfig{level, strings.ToLower(level) + "-simulation-1", strings.ToLower(level) + "-01", p.Sections, items, omitted}, true
}

func SSWLevels() []string {
	loadSSW()
	levels := make([]string, 0, len(sswData.Papers))
	for level := range sswData.Papers {
		levels = append(levels, level)
	}
	sort.Strings(levels)
	return levels
}
func IsSSWLevel(level string) bool { _, ok := sswSimulationPaper(strings.ToUpper(level)); return ok }
func sswTitle(level string) string { loadSSW(); return sswData.Papers[level].Title }

// Practice intentionally reveals feedback after each response, grants no XP,
// and does not claim to be a secured exam or practical certification.
func (s *MemoryStore) SSWPractice(user, sector, module string) ([]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !openLearningAccess && s.entitlementLocked(user).Tier != domain.TierPremium {
		return nil, ErrPremium
	}
	loadSSW()
	bank, ok := sswData.Banks[sector]
	if !ok {
		return nil, ErrNotFound
	}
	result := []map[string]any{}
	for _, q := range bank {
		if module == "" || module == q.ModuleID {
			result = append(result, map[string]any{"id": q.ID, "moduleId": q.ModuleID, "group": q.Group, "lessonId": q.LessonID, "prompt": q.Prompt, "options": q.Options})
		}
	}
	if len(result) == 0 {
		return nil, ErrNotFound
	}
	return result, nil
}
func (s *MemoryStore) SSWPracticeCheck(user, sector, id string, answer int) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !openLearningAccess && s.entitlementLocked(user).Tier != domain.TierPremium {
		return nil, ErrPremium
	}
	loadSSW()
	for _, q := range sswData.Banks[sector] {
		if q.ID == id {
			if answer < 0 || answer >= len(q.Options) {
				return nil, ErrInvalid
			}
			return map[string]any{"correct": answer == q.Answer, "answer": q.Answer, "explanation": q.Explanation, "lessonId": q.LessonID}, nil
		}
	}
	return nil, ErrNotFound
}

// Durable SSW sessions include the hidden key. Only simulationSnapshot may be
// exposed by the API. The account's existing MariaDB row owns all this data.
type SSWAttemptRecord struct {
	Session      domain.SimulationSession
	Questions    []domain.SimulationQuestion
	Correct      map[string]string
	Explanations map[string]string
	Marked       map[string]bool
}

func (s *MemoryStore) persistSSWLocked(state *simulationAttemptState) error {
	if !strings.HasPrefix(state.session.PaperID, "ssw-") {
		return nil
	}
	user := state.session.Attempt.UserID
	base := s.gamification[user]
	if base == nil {
		base = newGamificationState()
	}
	raw, err := json.Marshal(base)
	if err != nil {
		return err
	}
	next := newGamificationState()
	if err = json.Unmarshal(raw, next); err != nil {
		return err
	}
	if next.SSWAttempts == nil {
		next.SSWAttempts = map[string]SSWAttemptRecord{}
	}
	record := SSWAttemptRecord{state.session, state.questions, state.correct, state.explanations, state.marked}
	// JSON clone detaches maps from the live session before cache publication.
	raw, err = json.Marshal(record)
	if err != nil {
		return err
	}
	var detached SSWAttemptRecord
	if err = json.Unmarshal(raw, &detached); err != nil {
		return err
	}
	next.SSWAttempts[state.session.Attempt.ID] = detached
	if s.gamificationDB != nil {
		raw, err = json.Marshal(next)
		if err != nil {
			return err
		}
		if err = s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state"})}).Create(&database.GamificationUser{UserID: user, State: raw}).Error; err != nil {
			return err
		}
	}
	s.gamification[user] = next
	return nil
}

func cloneSimulation(state *simulationAttemptState) *simulationAttemptState {
	raw, _ := json.Marshal(SSWAttemptRecord{state.session, state.questions, state.correct, state.explanations, state.marked})
	var r SSWAttemptRecord
	if err := json.Unmarshal(raw, &r); err != nil {
		panic(err)
	}
	return &simulationAttemptState{r.Session, r.Questions, r.Correct, r.Explanations, r.Marked}
}
