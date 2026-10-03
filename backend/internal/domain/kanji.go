package domain

import "time"

type KanjiExample struct {
	Japanese string `json:"japanese"`
	Reading  string `json:"reading"`
	Meaning  string `json:"meaning"`
}

// Readings contain common on/kun forms, not every dictionary reading.
type KanjiCharacter struct {
	Character  string         `json:"character"`
	Meaning    string         `json:"meaning"`
	OnReading  string         `json:"onReading"`
	KunReading string         `json:"kunReading"`
	Strokes    int            `json:"strokes"`
	Mnemonic   string         `json:"mnemonic"`
	Note       string         `json:"note"`
	Examples   []KanjiExample `json:"examples"`
}

type KanjiLessonSummary struct {
	ID             string `json:"id"`
	Order          int    `json:"order"`
	Chapter        string `json:"chapter"`
	Title          string `json:"title"`
	CharacterCount int    `json:"characterCount"`
	QuestionCount  int    `json:"questionCount"`
	BestScore      int    `json:"bestScore"`
	Completed      bool   `json:"completed"`
}

type KanjiCourse struct {
	Level            string               `json:"level"`
	Title            string               `json:"title"`
	Progress         int                  `json:"progress"`
	CompletedLessons int                  `json:"completedLessons"`
	Lessons          []KanjiLessonSummary `json:"lessons"`
}

type KanjiLesson struct {
	KanjiLessonSummary
	Characters []KanjiCharacter `json:"characters"`
}

type KanjiQuestion struct {
	ID      string       `json:"id"`
	Prompt  string       `json:"prompt"`
	Options []QuizOption `json:"options"`
}

type KanjiAttempt struct {
	ID        string          `json:"id"`
	LessonID  string          `json:"lessonId"`
	Questions []KanjiQuestion `json:"questions"`
	StartedAt time.Time       `json:"startedAt"`
}

// Feedback and scoring share the same wire contract as vocabulary quizzes.
type KanjiAnswerFeedback = VocabularyAnswerFeedback
type KanjiAttemptResult = VocabularyAttemptResult
