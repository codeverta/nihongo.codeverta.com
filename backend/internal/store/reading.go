package store

import (
	"fmt"
	"time"

	"nihongo-master-backend/internal/domain"
)

type readingQuestionContent struct {
	Prompt      string
	Choices     []string
	Correct     int
	Explanation string
}
type readingLessonContent struct {
	Level                                  string
	ID, Chapter, Title                     string
	JapaneseTitle, Description, ImageAsset string
	Minutes                                int
	Pages                                  []domain.ReadingPage
	Questions                              []readingQuestionContent
}
type readingAttemptState struct {
	Attempt      domain.ReadingAttempt
	UserID       string
	Correct      map[string]string
	Explanations map[string]string
	Answers      map[string]domain.ReadingAnswerFeedback
	Result       *domain.ReadingAttemptResult
}

func (s *MemoryStore) ReadingCourse(userID string, levels ...string) domain.ReadingCourse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	level := "N5"
	if len(levels) > 0 {
		level = levels[0]
	}
	course := domain.ReadingCourse{Level: level, Title: "Reading JLPT " + level, Lessons: make([]domain.ReadingLessonSummary, 0, 12)}
	for i := range s.reading {
		if s.reading[i].Level != level {
			continue
		}
		item := s.readingSummaryLocked(userID, i)
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
func (s *MemoryStore) readingSummaryLocked(userID string, index int) domain.ReadingLessonSummary {
	lesson := s.reading[index]
	order := 0
	for i := 0; i <= index; i++ {
		if s.reading[i].Level == lesson.Level {
			order++
		}
	}
	progress := s.progress[userID][lesson.ID]
	return domain.ReadingLessonSummary{JapaneseTitle: lesson.JapaneseTitle, Description: lesson.Description, Minutes: lesson.Minutes, ImageAsset: lesson.ImageAsset, ImageURL: s.assetURLLocked(lesson.ImageAsset), ID: lesson.ID, Order: order, Chapter: lesson.Chapter, Title: lesson.Title, PageCount: len(lesson.Pages), QuestionCount: len(lesson.Questions), BestScore: progress.Completion, Completed: progress.Completed}
}
func (s *MemoryStore) ReadingLesson(userID, lessonID string) (domain.ReadingLesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i, lesson := range s.reading {
		if lesson.ID == lessonID {
			pages := append([]domain.ReadingPage(nil), lesson.Pages...)
			for j := range pages {
				pages[j].ImageURL = s.assetURLLocked(pages[j].ImageAsset)
			}
			return domain.ReadingLesson{ReadingLessonSummary: s.readingSummaryLocked(userID, i), Pages: pages}, nil
		}
	}
	return domain.ReadingLesson{}, ErrNotFound
}
func (s *MemoryStore) StartReadingAttempt(userID, lessonID string) (domain.ReadingAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, lesson := range s.reading {
		if lesson.ID == lessonID {
			attempt := domain.ReadingAttempt{ID: fmt.Sprintf("reading-attempt-%d", s.attemptID.Add(1)), LessonID: lessonID, StartedAt: time.Now().UTC(), Questions: make([]domain.ReadingQuestion, 0, len(lesson.Questions))}
			state := &readingAttemptState{UserID: userID, Correct: map[string]string{}, Explanations: map[string]string{}, Answers: map[string]domain.ReadingAnswerFeedback{}}
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
				attempt.Questions = append(attempt.Questions, domain.ReadingQuestion{ID: qid, Prompt: q.Prompt, Options: options})
			}
			state.Attempt = attempt
			s.readingAttempts[attempt.ID] = state
			return attempt, nil
		}
	}
	return domain.ReadingAttempt{}, ErrNotFound
}
func (s *MemoryStore) AnswerReadingQuestion(userID, attemptID, questionID, optionID string) (domain.ReadingAnswerFeedback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.readingAttempts[attemptID]
	if !ok || state.UserID != userID {
		return domain.ReadingAnswerFeedback{}, ErrNotFound
	}
	correct, ok := state.Correct[questionID]
	if !ok {
		return domain.ReadingAnswerFeedback{}, ErrInvalid
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
		return domain.ReadingAnswerFeedback{}, ErrInvalid
	}
	if answer, ok := state.Answers[questionID]; ok {
		return answer, nil
	}
	answer := domain.ReadingAnswerFeedback{QuestionID: questionID, Correct: correct == optionID, CorrectOptionID: correct, Explanation: state.Explanations[questionID]}
	state.Answers[questionID] = answer
	return answer, nil
}
func (s *MemoryStore) CompleteReadingAttempt(userID, attemptID string) (domain.ReadingAttemptResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.readingAttempts[attemptID]
	if !ok || state.UserID != userID {
		return domain.ReadingAttemptResult{}, ErrNotFound
	}
	if state.Result != nil {
		return *state.Result, nil
	}
	if len(state.Answers) != len(state.Attempt.Questions) || len(state.Answers) == 0 {
		return domain.ReadingAttemptResult{}, ErrInvalid
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
	progress := domain.LessonProgress{LessonID: state.Attempt.LessonID, Completion: best, Completed: previous.Completed || passed, LastPosition: len(state.Answers) * 60, UpdatedAt: time.Now().UTC()}
	now := s.gamificationNow().UTC()
	progress.UpdatedAt = now
	if err := s.recordStudyLocked(userID, studyEvent{ID: "reading:" + attemptID, Kind: "lesson", Subject: "reading", LessonID: state.Attempt.LessonID, Score: score, Answered: len(state.Answers), Seconds: int(now.Sub(state.Attempt.StartedAt).Seconds()), At: now, Progress: &progress}); err != nil {
		return domain.ReadingAttemptResult{}, err
	}
	level := ""
	for _, l := range s.reading {
		if l.ID == state.Attempt.LessonID {
			level = l.Level
			break
		}
	}
	completed, total := 0, 0
	for _, l := range s.reading {
		if l.Level != level {
			continue
		}
		total++
		if s.progress[userID][l.ID].Completed {
			completed++
		}
	}
	result := domain.ReadingAttemptResult{AttemptID: attemptID, LessonID: state.Attempt.LessonID, Score: score, Correct: correct, Total: len(state.Answers), Passed: passed, BestScore: best, Progress: completed * 100 / total, CanRetry: true}
	state.Result = &result
	return result, nil
}
