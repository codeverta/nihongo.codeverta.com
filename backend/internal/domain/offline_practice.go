package domain

import "time"

// OfflinePracticeManifest identifies a server-issued, immutable lesson quiz.
// The signature authenticates every field except Signature itself and the
// content hash covers the complete question bank returned with the manifest.
type OfflinePracticeManifest struct {
	SchemaVersion int       `json:"schemaVersion"`
	PackageID     string    `json:"packageId"`
	Version       string    `json:"version"`
	ContentHash   string    `json:"contentHash"`
	Kind          string    `json:"kind"`
	LessonID      string    `json:"lessonId"`
	Level         string    `json:"level"`
	UserID        string    `json:"userId"`
	IssuedAt      time.Time `json:"issuedAt"`
	Signature     string    `json:"signature"`
}

// OfflinePracticeQuestion is safe to cache on the learner's device: lesson
// practice already reveals the correct option and explanation after each
// answer. Official completion and rewards still require server verification.
type OfflinePracticeQuestion struct {
	Kind             string       `json:"kind,omitempty"`
	ModuleID         string       `json:"moduleId,omitempty"`
	Group            string       `json:"group,omitempty"`
	LessonID         string       `json:"lessonId,omitempty"`
	ImageAsset       string       `json:"imageAsset,omitempty"`
	ImageURL         string       `json:"imageUrl,omitempty"`
	ImagePanel       int          `json:"imagePanel,omitempty"`
	ImageDescription string       `json:"imageDescription,omitempty"`
	ID               string       `json:"id"`
	Prompt           string       `json:"prompt"`
	Options          []QuizOption `json:"options"`
	CorrectOptionID  string       `json:"correctOptionId"`
	Explanation      string       `json:"explanation"`
}

type OfflinePracticePackage struct {
	Manifest  OfflinePracticeManifest   `json:"manifest"`
	Questions []OfflinePracticeQuestion `json:"questions"`
}

type OfflinePracticeAnswer struct {
	QuestionID string `json:"questionId"`
	OptionID   string `json:"optionId"`
}

type OfflinePracticeSyncInput struct {
	AttemptID string                  `json:"attemptId"`
	Package   OfflinePracticePackage  `json:"package"`
	Answers   []OfflinePracticeAnswer `json:"answers"`
}

type OfflinePracticeResult struct {
	AttemptID string `json:"attemptId"`
	LessonID  string `json:"lessonId"`
	Score     int    `json:"score"`
	Correct   int    `json:"correctAnswers"`
	Total     int    `json:"totalQuestions"`
	Passed    bool   `json:"passed"`
	BestScore int    `json:"bestScore"`
	Progress  int    `json:"progress"`
	CanRetry  bool   `json:"canRetry"`
	Verified  bool   `json:"verified"`
}
