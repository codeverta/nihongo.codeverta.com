package domain

import "time"

type ReadingWord struct {
	Japanese string `json:"japanese"`
	Reading  string `json:"reading"`
	Meaning  string `json:"meaning"`
}
type ReadingPage struct {
	Title            string        `json:"title"`
	Japanese         string        `json:"japanese"`
	Reading          string        `json:"reading"`
	Meaning          string        `json:"meaning"`
	ImageAsset       string        `json:"imageAsset"`
	ImageURL         string        `json:"imageUrl"`
	ImagePanel       int           `json:"imagePanel"`
	ImageDescription string        `json:"imageDescription"`
	Vocabulary       []ReadingWord `json:"vocabulary"`
}

type ReadingLessonSummary struct {
	JapaneseTitle string `json:"japaneseTitle"`
	Description   string `json:"description"`
	Minutes       int    `json:"minutes"`
	ImageAsset    string `json:"imageAsset"`
	ImageURL      string `json:"imageUrl"`
	ID            string `json:"id"`
	Order         int    `json:"order"`
	Chapter       string `json:"chapter"`
	Title         string `json:"title"`
	PageCount     int    `json:"pageCount"`
	QuestionCount int    `json:"questionCount"`
	BestScore     int    `json:"bestScore"`
	Completed     bool   `json:"completed"`
}

type ReadingCourse struct {
	Level            string                 `json:"level"`
	Title            string                 `json:"title"`
	Progress         int                    `json:"progress"`
	CompletedLessons int                    `json:"completedLessons"`
	Lessons          []ReadingLessonSummary `json:"lessons"`
}

type ReadingLesson struct {
	ReadingLessonSummary
	Pages []ReadingPage `json:"pages"`
}

type ReadingQuestion struct {
	ID      string       `json:"id"`
	Prompt  string       `json:"prompt"`
	Options []QuizOption `json:"options"`
}

type ReadingAttempt struct {
	ID        string            `json:"id"`
	LessonID  string            `json:"lessonId"`
	Questions []ReadingQuestion `json:"questions"`
	StartedAt time.Time         `json:"startedAt"`
}

// Feedback and scoring share the same wire contract as vocabulary quizzes.
type ReadingAnswerFeedback = VocabularyAnswerFeedback
type ReadingAttemptResult = VocabularyAttemptResult
