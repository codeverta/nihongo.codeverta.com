package domain

import "time"

type DiagnosticQuestion struct {
	ID      string       `json:"id"`
	Section string       `json:"section"`
	Passage string       `json:"passage,omitempty"`
	Prompt  string       `json:"prompt"`
	Options []QuizOption `json:"options"`
}
type DiagnosticSession struct {
	Attempt    ExamAttempt          `json:"attempt"`
	ExpiresAt  time.Time            `json:"expiresAt"`
	ServerTime time.Time            `json:"serverTime"`
	Questions  []DiagnosticQuestion `json:"questions"`
	Answers    map[string]string    `json:"answers"`
	Result     *DiagnosticResult    `json:"result,omitempty"`
}
type DiagnosticSection struct {
	Name    string `json:"name"`
	Correct int    `json:"correct"`
	Total   int    `json:"total"`
	Score   int    `json:"score"`
}
type DiagnosticReview struct {
	Question         DiagnosticQuestion `json:"question"`
	SelectedOptionID string             `json:"selectedOptionId"`
	CorrectOptionID  string             `json:"correctOptionId"`
	Correct          bool               `json:"correct"`
	Explanation      string             `json:"explanation"`
}
type DiagnosticResult struct {
	AttemptID       string              `json:"attemptId"`
	Score           int                 `json:"score"`
	Correct         int                 `json:"correct"`
	Total           int                 `json:"total"`
	Answered        int                 `json:"answered"`
	CompletedAt     time.Time           `json:"completedAt"`
	DurationSeconds int                 `json:"durationSeconds"`
	Sections        []DiagnosticSection `json:"sections"`
	Reviews         []DiagnosticReview  `json:"reviews"`
}
