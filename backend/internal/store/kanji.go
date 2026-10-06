package store

import (
	"fmt"
	"time"

	"nihongo-master-backend/internal/domain"
)

type kanjiQuestionContent struct {
	Prompt      string
	Choices     []string
	Correct     int
	Explanation string
}
type kanjiLessonContent struct {
	Level              string
	ID, Chapter, Title string
	Characters         []domain.KanjiCharacter
	Questions          []kanjiQuestionContent
}
type kanjiAttemptState struct {
	Attempt      domain.KanjiAttempt
	UserID       string
	Correct      map[string]string
	Explanations map[string]string
	Answers      map[string]domain.KanjiAnswerFeedback
	Result       *domain.KanjiAttemptResult
}

func (s *MemoryStore) KanjiCourse(userID string, levels ...string) domain.KanjiCourse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	level := "N5"
	if len(levels) > 0 {
		level = levels[0]
	}
	course := domain.KanjiCourse{Level: level, Title: "Kanji JLPT " + level, Lessons: make([]domain.KanjiLessonSummary, 0)}
	for i := range s.kanji {
		if s.kanji[i].Level != level {
			continue
		}
		item := s.kanjiSummaryLocked(userID, i)
		course.Lessons = append(course.Lessons, item)
		if item.Completed {
			course.CompletedLessons++
		}
	}
	if len(course.Lessons) > 0 {
		course.Progress = course.CompletedLessons * 100 / len(course.Lessons)
	}
	return course
}
func (s *MemoryStore) kanjiSummaryLocked(userID string, index int) domain.KanjiLessonSummary {
	lesson := s.kanji[index]
	order := 0
	for i := 0; i <= index; i++ {
		if s.kanji[i].Level == lesson.Level {
			order++
		}
	}
	progress := s.progress[userID][lesson.ID]
	return domain.KanjiLessonSummary{ID: lesson.ID, Order: order, Chapter: lesson.Chapter, Title: lesson.Title, CharacterCount: len(lesson.Characters), QuestionCount: len(lesson.Questions), BestScore: progress.Completion, Completed: progress.Completed}
}
func (s *MemoryStore) KanjiLesson(userID, lessonID string) (domain.KanjiLesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i, lesson := range s.kanji {
		if lesson.ID == lessonID {
			return domain.KanjiLesson{KanjiLessonSummary: s.kanjiSummaryLocked(userID, i), Characters: lesson.Characters}, nil
		}
	}
	return domain.KanjiLesson{}, ErrNotFound
}
func (s *MemoryStore) StartKanjiAttempt(userID, lessonID string) (domain.KanjiAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, lesson := range s.kanji {
		if lesson.ID == lessonID {
			attempt := domain.KanjiAttempt{ID: fmt.Sprintf("kanji-attempt-%d", s.attemptID.Add(1)), LessonID: lessonID, StartedAt: time.Now().UTC(), Questions: make([]domain.KanjiQuestion, 0, len(lesson.Questions))}
			state := &kanjiAttemptState{UserID: userID, Correct: map[string]string{}, Explanations: map[string]string{}, Answers: map[string]domain.KanjiAnswerFeedback{}}
			for i, q := range lesson.Questions {
				qid := fmt.Sprintf("%s-q%d", lessonID, i+1)
				options := make([]domain.QuizOption, len(q.Choices))
				for j, text := range q.Choices {
					options[j] = domain.QuizOption{ID: fmt.Sprintf("%s-o%d", qid, j+1), Text: text}
				}
				state.Correct[qid] = options[q.Correct].ID
				state.Explanations[qid] = q.Explanation
				// Rotate positions so correct answers do not always appear first.
				shiftOptions(options, i%len(options))
				attempt.Questions = append(attempt.Questions, domain.KanjiQuestion{ID: qid, Prompt: q.Prompt, Options: options})
			}
			state.Attempt = attempt
			s.kanjiAttempts[attempt.ID] = state
			return attempt, nil
		}
	}
	return domain.KanjiAttempt{}, ErrNotFound
}
func (s *MemoryStore) AnswerKanjiQuestion(userID, attemptID, questionID, optionID string) (domain.KanjiAnswerFeedback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.kanjiAttempts[attemptID]
	if !ok || state.UserID != userID {
		return domain.KanjiAnswerFeedback{}, ErrNotFound
	}
	correct, ok := state.Correct[questionID]
	if !ok {
		return domain.KanjiAnswerFeedback{}, ErrInvalid
	}
	valid := false
	for _, q := range state.Attempt.Questions {
		if q.ID == questionID {
			for _, o := range q.Options {
				if o.ID == optionID {
					valid = true
				}
			}
		}
	}
	if !valid {
		return domain.KanjiAnswerFeedback{}, ErrInvalid
	}
	if answer, ok := state.Answers[questionID]; ok {
		return answer, nil
	}
	answer := domain.KanjiAnswerFeedback{QuestionID: questionID, Correct: correct == optionID, CorrectOptionID: correct, Explanation: state.Explanations[questionID]}
	state.Answers[questionID] = answer
	return answer, nil
}
func (s *MemoryStore) CompleteKanjiAttempt(userID, attemptID string) (domain.KanjiAttemptResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.kanjiAttempts[attemptID]
	if !ok || state.UserID != userID {
		return domain.KanjiAttemptResult{}, ErrNotFound
	}
	if state.Result != nil {
		return *state.Result, nil
	}
	if len(state.Answers) != len(state.Attempt.Questions) || len(state.Answers) == 0 {
		return domain.KanjiAttemptResult{}, ErrInvalid
	}
	correct := 0
	for _, a := range state.Answers {
		if a.Correct {
			correct++
		}
	}
	score := correct * 100 / len(state.Answers)
	passed := score >= 80
	if s.progress[userID] == nil {
		s.progress[userID] = map[string]domain.LessonProgress{}
	}
	previous := s.progress[userID][state.Attempt.LessonID]
	best := previous.Completion
	if score > best {
		best = score
	}
	s.progress[userID][state.Attempt.LessonID] = domain.LessonProgress{LessonID: state.Attempt.LessonID, Completion: best, Completed: previous.Completed || passed, LastPosition: len(state.Answers) * 60, UpdatedAt: time.Now().UTC()}
	if passed && !previous.Completed {
		profile := s.ensureProfileLocked(userID)
		profile.XP += 50
		profile.Level = 1 + profile.XP/500
		s.profiles[userID] = profile
	}
	level := ""
	for _, l := range s.kanji {
		if l.ID == state.Attempt.LessonID {
			level = l.Level
			break
		}
	}
	completed, total := 0, 0
	for _, l := range s.kanji {
		if l.Level != level {
			continue
		}
		total++
		if s.progress[userID][l.ID].Completed {
			completed++
		}
	}
	result := domain.KanjiAttemptResult{AttemptID: attemptID, LessonID: state.Attempt.LessonID, Score: score, Correct: correct, Total: len(state.Answers), Passed: passed, BestScore: best, Progress: completed * 100 / total, CanRetry: true}
	state.Result = &result
	return result, nil
}
