package domain

import "time"

type GrammarExample struct {
	Japanese string `json:"japanese"`
	Reading  string `json:"reading"`
	Meaning  string `json:"meaning"`
}

type GrammarPoint struct {
	Pattern       string           `json:"pattern"`
	Meaning       string           `json:"meaning"`
	Formation     string           `json:"formation"`
	Explanation   string           `json:"explanation"`
	CommonMistake string           `json:"commonMistake"`
	Examples      []GrammarExample `json:"examples"`
}

type GrammarLessonSummary struct {
	Patterns      []string `json:"patterns"`
	Level         string   `json:"level"`
	ID            string   `json:"id"`
	Order         int      `json:"order"`
	Chapter       string   `json:"chapter"`
	Title         string   `json:"title"`
	PointCount    int      `json:"pointCount"`
	QuestionCount int      `json:"questionCount"`
	BestScore     int      `json:"bestScore"`
	Completed     bool     `json:"completed"`
}

type GrammarCourse struct {
	Level            string                 `json:"level"`
	Title            string                 `json:"title"`
	Progress         int                    `json:"progress"`
	CompletedLessons int                    `json:"completedLessons"`
	Lessons          []GrammarLessonSummary `json:"lessons"`
}

type GrammarLesson struct {
	GrammarLessonSummary
	Points []GrammarPoint `json:"points"`
}

type GrammarQuestion struct {
	ID      string       `json:"id"`
	Prompt  string       `json:"prompt"`
	Options []QuizOption `json:"options"`
}

type GrammarAttempt struct {
	ID        string            `json:"id"`
	LessonID  string            `json:"lessonId"`
	Questions []GrammarQuestion `json:"questions"`
	StartedAt time.Time         `json:"startedAt"`
}

// Feedback and scoring share the same wire contract as vocabulary quizzes.
type GrammarAnswerFeedback = VocabularyAnswerFeedback
type GrammarAttemptResult = VocabularyAttemptResult
