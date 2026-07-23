package domain

import "time"

type VocabularyWord struct {
	ID             string `json:"id"`
	Term           string `json:"term"`
	Reading        string `json:"reading"`
	Meaning        string `json:"meaning"`
	Example        string `json:"example"`
	ExampleMeaning string `json:"exampleMeaning"`
	Note           string `json:"note"`
}

type VocabularyLessonContent struct {
	ID          string           `json:"id"`
	Order       int              `json:"order"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Words       []VocabularyWord `json:"words"`
}

type VocabularyLevelContent struct {
	Level        string                    `json:"level"`
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	RequiredTier AccessTier                `json:"requiredTier"`
	Lessons      []VocabularyLessonContent `json:"lessons"`
}

type VocabularyLessonSummary struct {
	ID          string `json:"id"`
	Order       int    `json:"order"`
	Title       string `json:"title"`
	Description string `json:"description"`
	WordCount   int    `json:"wordCount"`
	BestScore   int    `json:"bestScore"`
	Completed   bool   `json:"completed"`
}

type VocabularyLevel struct {
	Level        string                    `json:"level"`
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	RequiredTier AccessTier                `json:"requiredTier"`
	Progress     int                       `json:"progress"`
	Lessons      []VocabularyLessonSummary `json:"lessons"`
}

type VocabularyLesson struct {
	ID          string           `json:"id"`
	Level       string           `json:"level"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	BestScore   int              `json:"bestScore"`
	Completed   bool             `json:"completed"`
	Words       []VocabularyWord `json:"words"`
}

type QuizOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type VocabularyQuestion struct {
	ID      string       `json:"id"`
	Prompt  string       `json:"prompt"`
	Options []QuizOption `json:"options"`
}

type VocabularyAttempt struct {
	ID        string               `json:"id"`
	LessonID  string               `json:"lessonId"`
	Questions []VocabularyQuestion `json:"questions"`
	StartedAt time.Time            `json:"startedAt"`
}

type VocabularyAnswerFeedback struct {
	QuestionID      string `json:"questionId"`
	Correct         bool   `json:"correct"`
	CorrectOptionID string `json:"correctOptionId"`
	Explanation     string `json:"explanation"`
}

type VocabularyAttemptResult struct {
	AttemptID string `json:"attemptId"`
	LessonID  string `json:"lessonId"`
	Score     int    `json:"score"`
	Correct   int    `json:"correctAnswers"`
	Total     int    `json:"totalQuestions"`
	Passed    bool   `json:"passed"`
	BestScore int    `json:"bestScore"`
	Progress  int    `json:"progress"`
	CanRetry  bool   `json:"canRetry"`
}
