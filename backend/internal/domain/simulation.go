package domain

import "time"

type SimulationQuestion struct {
	ID           string       `json:"id"`
	Number       int          `json:"number"`
	Section      string       `json:"section"`
	Group        string       `json:"group"`
	Instructions string       `json:"instructions"`
	Passage      string       `json:"passage,omitempty"`
	Prompt       string       `json:"prompt"`
	Options      []QuizOption `json:"options"`
}

type SimulationSection struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DurationSeconds int    `json:"durationSeconds"`
	QuestionCount   int    `json:"questionCount"`
}

type SimulationSession struct {
	Attempt          ExamAttempt          `json:"attempt"`
	PaperID          string               `json:"paperId"`
	ActiveSection    string               `json:"activeSection"`
	SectionStartedAt time.Time            `json:"sectionStartedAt"`
	ExpiresAt        time.Time            `json:"expiresAt"`
	ServerTime       time.Time            `json:"serverTime"`
	Sections         []SimulationSection  `json:"sections"`
	Questions        []SimulationQuestion `json:"questions"`
	Answers          map[string]string    `json:"answers"`
	Marked           []string             `json:"marked"`
	Result           *SimulationResult    `json:"result,omitempty"`
}

type SimulationReview struct {
	Question         SimulationQuestion `json:"question"`
	SelectedOptionID string             `json:"selectedOptionId"`
	CorrectOptionID  string             `json:"correctOptionId"`
	Correct          bool               `json:"correct"`
	Explanation      string             `json:"explanation"`
}

type SimulationResult struct {
	AttemptID       string              `json:"attemptId"`
	PaperID         string              `json:"paperId"`
	Score           int                 `json:"score"`
	Correct         int                 `json:"correct"`
	Total           int                 `json:"total"`
	Answered        int                 `json:"answered"`
	DurationSeconds int                 `json:"durationSeconds"`
	CompletedAt     time.Time           `json:"completedAt"`
	Sections        []DiagnosticSection `json:"sections"`
	Reviews         []SimulationReview  `json:"reviews"`
	ScoringNote     string              `json:"scoringNote"`
}
