package store

import (
	"fmt"
	"nihongo-master-backend/internal/domain"
	"time"
)

type grammarQuestionContent struct {
	Prompt      string
	Choices     []string
	Correct     int
	Explanation string
}
type grammarLessonContent struct {
	ID, Level, Chapter, Title string
	Points                    []domain.GrammarPoint
	Questions                 []grammarQuestionContent
}
type grammarAttemptState struct {
	Attempt      domain.GrammarAttempt
	UserID       string
	Correct      map[string]string
	Explanations map[string]string
	Answers      map[string]domain.GrammarAnswerFeedback
	Result       *domain.GrammarAttemptResult
}

func (s *MemoryStore) GrammarCourse(userID string, requestedLevel ...string) domain.GrammarCourse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	level := "N5"
	if len(requestedLevel) > 0 {
		level = requestedLevel[0]
	}
	course := domain.GrammarCourse{Level: level, Title: "Grammar JLPT " + level, Lessons: make([]domain.GrammarLessonSummary, 0)}
	for i := range s.grammar {
		if s.grammar[i].Level != level {
			continue
		}
		item := s.grammarSummaryLocked(userID, i)
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
func (s *MemoryStore) grammarSummaryLocked(userID string, index int) domain.GrammarLessonSummary {
	lesson := s.grammar[index]
	progress := s.progress[userID][lesson.ID]
	patterns := make([]string, 0, len(lesson.Points))
	for _, point := range lesson.Points {
		patterns = append(patterns, point.Pattern)
	}
	order := 0
	for i := 0; i <= index; i++ {
		if s.grammar[i].Level == lesson.Level {
			order++
		}
	}
	return domain.GrammarLessonSummary{ID: lesson.ID, Level: lesson.Level, Patterns: patterns, Order: order, Chapter: lesson.Chapter, Title: lesson.Title, PointCount: len(lesson.Points), QuestionCount: len(lesson.Questions), BestScore: progress.Completion, Completed: progress.Completed}
}
func (s *MemoryStore) GrammarLesson(userID, lessonID string) (domain.GrammarLesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i, lesson := range s.grammar {
		if lesson.ID == lessonID {
			return domain.GrammarLesson{GrammarLessonSummary: s.grammarSummaryLocked(userID, i), Points: lesson.Points}, nil
		}
	}
	return domain.GrammarLesson{}, ErrNotFound
}
func (s *MemoryStore) StartGrammarAttempt(userID, lessonID string) (domain.GrammarAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, lesson := range s.grammar {
		if lesson.ID == lessonID {
			attempt := domain.GrammarAttempt{ID: fmt.Sprintf("grammar-attempt-%d", s.attemptID.Add(1)), LessonID: lessonID, StartedAt: time.Now().UTC(), Questions: make([]domain.GrammarQuestion, 0, len(lesson.Questions))}
			state := &grammarAttemptState{UserID: userID, Correct: map[string]string{}, Explanations: map[string]string{}, Answers: map[string]domain.GrammarAnswerFeedback{}}
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
				attempt.Questions = append(attempt.Questions, domain.GrammarQuestion{ID: qid, Prompt: q.Prompt, Options: options})
			}
			state.Attempt = attempt
			s.grammarAttempts[attempt.ID] = state
			return attempt, nil
		}
	}
	return domain.GrammarAttempt{}, ErrNotFound
}
func (s *MemoryStore) AnswerGrammarQuestion(userID, attemptID, questionID, optionID string) (domain.GrammarAnswerFeedback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.grammarAttempts[attemptID]
	if !ok || state.UserID != userID {
		return domain.GrammarAnswerFeedback{}, ErrNotFound
	}
	correct, ok := state.Correct[questionID]
	if !ok {
		return domain.GrammarAnswerFeedback{}, ErrInvalid
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
		return domain.GrammarAnswerFeedback{}, ErrInvalid
	}
	if answer, ok := state.Answers[questionID]; ok {
		return answer, nil
	}
	answer := domain.GrammarAnswerFeedback{QuestionID: questionID, Correct: correct == optionID, CorrectOptionID: correct, Explanation: state.Explanations[questionID]}
	state.Answers[questionID] = answer
	return answer, nil
}
func (s *MemoryStore) CompleteGrammarAttempt(userID, attemptID string) (domain.GrammarAttemptResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.grammarAttempts[attemptID]
	if !ok || state.UserID != userID {
		return domain.GrammarAttemptResult{}, ErrNotFound
	}
	if state.Result != nil {
		return *state.Result, nil
	}
	if len(state.Answers) != len(state.Attempt.Questions) || len(state.Answers) == 0 {
		return domain.GrammarAttemptResult{}, ErrInvalid
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
	completed, total := 0, 0
	level := ""
	for _, l := range s.grammar {
		if l.ID == state.Attempt.LessonID {
			level = l.Level
			break
		}
	}
	for _, l := range s.grammar {
		if l.Level != level {
			continue
		}
		total++
		if s.progress[userID][l.ID].Completed {
			completed++
		}
	}
	result := domain.GrammarAttemptResult{AttemptID: attemptID, LessonID: state.Attempt.LessonID, Score: score, Correct: correct, Total: len(state.Answers), Passed: passed, BestScore: best, Progress: completed * 100 / total, CanRetry: true}
	state.Result = &result
	return result, nil
}
