package store

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/domain"
)

// OfflinePracticePackage returns the authored lesson quiz together with its
// answers for client-side feedback. The API signs the immutable package before
// returning it. Premium checks happen here as well as in the HTTP middleware so
// direct store callers cannot mint a restricted package.
func (s *MemoryStore) OfflinePracticePackage(user, kind, lessonID string) (domain.OfflinePracticePackage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	level, exists := s.offlinePracticeLessonLevelLocked(kind, lessonID)
	if !exists {
		return domain.OfflinePracticePackage{}, ErrNotFound
	}
	if !openLearningAccess && (kind == "ssw" || PremiumLevel(level)) && s.entitlementLocked(user).Tier != domain.TierPremium {
		return domain.OfflinePracticePackage{}, ErrPremium
	}

	questions, err := s.offlinePracticeQuestionsLocked(kind, lessonID)
	if err != nil {
		return domain.OfflinePracticePackage{}, err
	}
	return domain.OfflinePracticePackage{
		Manifest:  domain.OfflinePracticeManifest{Kind: kind, LessonID: lessonID, Level: level},
		Questions: questions,
	}, nil
}

func (s *MemoryStore) offlinePracticeLessonLevelLocked(kind, lessonID string) (string, bool) {
	switch kind {
	case "vocabulary":
		level, lesson := s.findVocabularyLessonLocked(lessonID)
		if level != nil && lesson != nil {
			return level.Level, true
		}
	case "grammar":
		for _, lesson := range s.grammar {
			if lesson.ID == lessonID {
				return lesson.Level, true
			}
		}
	case "kanji":
		for _, lesson := range s.kanji {
			if lesson.ID == lessonID {
				return lesson.Level, true
			}
		}
	case "reading":
		for _, lesson := range s.reading {
			if lesson.ID == lessonID {
				return lesson.Level, true
			}
		}
	case "ssw":
		loadSSW()
		if len(sswData.Banks[lessonID]) > 0 {
			return "SSW", true
		}
	}
	return "", false
}

func (s *MemoryStore) offlinePracticeQuestionsLocked(kind, lessonID string) ([]domain.OfflinePracticeQuestion, error) {
	switch kind {
	case "vocabulary":
		level, lesson := s.findVocabularyLessonLocked(lessonID)
		if lesson == nil || level == nil {
			return nil, ErrNotFound
		}
		var source []domain.VocabularyQuestion
		var correct map[string]string
		var explanations map[string]string
		if level.Level == "N5" {
			source, correct, explanations = s.n5VocabularyQuestionsLocked(lesson)
		} else {
			options := make([]domain.QuizOption, 0, len(lesson.Words))
			for _, word := range lesson.Words {
				options = append(options, domain.QuizOption{ID: word.ID, Text: word.Meaning})
			}
			source = make([]domain.VocabularyQuestion, 0, len(lesson.Words))
			correct, explanations = map[string]string{}, map[string]string{}
			for index, word := range lesson.Words {
				questionID := fmt.Sprintf("%s-q%d", lesson.ID, index+1)
				rotated := append([]domain.QuizOption(nil), options...)
				shiftOptions(rotated, index)
				source = append(source, domain.VocabularyQuestion{
					ID: questionID, Prompt: fmt.Sprintf("Apa arti %s (%s)?", word.Term, word.Reading), Options: rotated,
				})
				correct[questionID] = word.ID
				explanations[questionID] = fmt.Sprintf("%s berarti %s. %s Contoh: %s — %s", word.Term, word.Meaning, word.Note, word.Example, word.ExampleMeaning)
			}
		}
		result := make([]domain.OfflinePracticeQuestion, 0, len(source))
		for _, question := range source {
			result = append(result, domain.OfflinePracticeQuestion{
				Kind: question.Kind, ImageAsset: question.ImageAsset,
				ImageURL: question.ImageURL, ImagePanel: question.ImagePanel,
				ImageDescription: question.ImageDescription,
				ID:               question.ID, Prompt: question.Prompt, Options: question.Options,
				CorrectOptionID: correct[question.ID], Explanation: explanations[question.ID],
			})
		}
		return result, nil
	case "grammar":
		for _, lesson := range s.grammar {
			if lesson.ID != lessonID {
				continue
			}
			result := make([]domain.OfflinePracticeQuestion, 0, len(lesson.Questions))
			for i, question := range lesson.Questions {
				id := fmt.Sprintf("%s-q%d", lesson.ID, i+1)
				options := make([]domain.QuizOption, len(question.Choices))
				for j, choice := range question.Choices {
					options[j] = domain.QuizOption{ID: fmt.Sprintf("%s-o%d", id, j+1), Text: choice}
				}
				if question.Correct < 0 || question.Correct >= len(options) {
					return nil, ErrInvalid
				}
				correct := options[question.Correct].ID
				shiftOptions(options, i%len(options))
				result = append(result, domain.OfflinePracticeQuestion{
					ID: id, Prompt: question.Prompt, Options: options,
					CorrectOptionID: correct, Explanation: question.Explanation,
				})
			}
			return result, nil
		}
	case "kanji":
		for _, lesson := range s.kanji {
			if lesson.ID != lessonID {
				continue
			}
			result := make([]domain.OfflinePracticeQuestion, 0, len(lesson.Questions))
			for i, question := range lesson.Questions {
				id := fmt.Sprintf("%s-q%d", lesson.ID, i+1)
				options := make([]domain.QuizOption, len(question.Choices))
				for j, choice := range question.Choices {
					options[j] = domain.QuizOption{ID: fmt.Sprintf("%s-o%d", id, j+1), Text: choice}
				}
				if question.Correct < 0 || question.Correct >= len(options) {
					return nil, ErrInvalid
				}
				correct := options[question.Correct].ID
				shiftOptions(options, i%len(options))
				result = append(result, domain.OfflinePracticeQuestion{
					ID: id, Prompt: question.Prompt, Options: options,
					CorrectOptionID: correct, Explanation: question.Explanation,
				})
			}
			return result, nil
		}
	case "reading":
		for _, lesson := range s.reading {
			if lesson.ID != lessonID {
				continue
			}
			result := make([]domain.OfflinePracticeQuestion, 0, len(lesson.Questions))
			for i, question := range lesson.Questions {
				id := fmt.Sprintf("%s-q%d", lesson.ID, i+1)
				options := make([]domain.QuizOption, len(question.Choices))
				for j, choice := range question.Choices {
					options[j] = domain.QuizOption{ID: fmt.Sprintf("%s-o%d", id, j+1), Text: choice}
				}
				if question.Correct < 0 || question.Correct >= len(options) {
					return nil, ErrInvalid
				}
				correct := options[question.Correct].ID
				shiftOptions(options, i%len(options))
				result = append(result, domain.OfflinePracticeQuestion{
					ID: id, Prompt: question.Prompt, Options: options,
					CorrectOptionID: correct, Explanation: question.Explanation,
				})
			}
			return result, nil
		}
	case "ssw":
		loadSSW()
		bank, exists := sswData.Banks[lessonID]
		if !exists {
			return nil, ErrNotFound
		}
		result := make([]domain.OfflinePracticeQuestion, 0, len(bank))
		for _, question := range bank {
			if question.ID == "" || question.Prompt == "" || question.Explanation == "" || question.Answer < 0 || question.Answer >= len(question.Options) || len(question.Options) < 2 {
				return nil, ErrInvalid
			}
			options := make([]domain.QuizOption, len(question.Options))
			for i, option := range question.Options {
				options[i] = domain.QuizOption{ID: fmt.Sprintf("%s-o%d", question.ID, i+1), Text: option}
			}
			result = append(result, domain.OfflinePracticeQuestion{
				ModuleID: question.ModuleID, Group: question.Group, LessonID: question.LessonID,
				ID: question.ID, Prompt: question.Prompt, Options: options,
				CorrectOptionID: options[question.Answer].ID, Explanation: question.Explanation,
			})
		}
		return result, nil
	default:
		return nil, ErrNotFound
	}
	return nil, ErrNotFound
}

type offlinePracticeAttemptRecord struct {
	Fingerprint string                       `json:"fingerprint"`
	ContentHash string                       `json:"contentHash"`
	Result      domain.OfflinePracticeResult `json:"result"`
}

// SyncOfflinePracticeAttempt grades only question/option IDs from a package
// whose manifest has already been verified by the API. Its durable record and
// gamification event are written in the same account-state commit.
func (s *MemoryStore) SyncOfflinePracticeAttempt(user string, input domain.OfflinePracticeSyncInput, fingerprint string) (domain.OfflinePracticeResult, error) {
	if _, err := uuid.Parse(input.AttemptID); err != nil || input.AttemptID == uuid.Nil.String() || fingerprint == "" {
		return domain.OfflinePracticeResult{}, ErrInvalid
	}
	manifest := input.Package.Manifest
	if manifest.UserID != user {
		return domain.OfflinePracticeResult{}, ErrNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.gamification[user]
	if state != nil {
		if existing, ok := state.OfflinePracticeAttempts[input.AttemptID]; ok {
			if existing.Fingerprint != fingerprint {
				return domain.OfflinePracticeResult{}, ErrInvalid
			}
			return existing.Result, nil
		}
		if state.Events["offline-practice:"+input.AttemptID] {
			return domain.OfflinePracticeResult{}, errors.New("offline practice result record is missing")
		}
	}

	level, exists := s.offlinePracticeLessonLevelLocked(manifest.Kind, manifest.LessonID)
	if !exists || level != manifest.Level {
		return domain.OfflinePracticeResult{}, ErrInvalid
	}
	if !openLearningAccess && (manifest.Kind == "ssw" || PremiumLevel(level)) && s.entitlementLocked(user).Tier != domain.TierPremium {
		return domain.OfflinePracticeResult{}, ErrPremium
	}
	partialPractice := manifest.Kind == "ssw"
	if len(input.Package.Questions) == 0 || len(input.Answers) == 0 || len(input.Answers) > len(input.Package.Questions) || (!partialPractice && len(input.Answers) != len(input.Package.Questions)) {
		return domain.OfflinePracticeResult{}, ErrInvalid
	}
	answers := make(map[string]string, len(input.Answers))
	for _, answer := range input.Answers {
		if answer.QuestionID == "" || answer.OptionID == "" {
			return domain.OfflinePracticeResult{}, ErrInvalid
		}
		if _, duplicate := answers[answer.QuestionID]; duplicate {
			return domain.OfflinePracticeResult{}, ErrInvalid
		}
		answers[answer.QuestionID] = answer.OptionID
	}
	correct := 0
	seenQuestions := make(map[string]bool, len(input.Package.Questions))
	for _, question := range input.Package.Questions {
		if question.ID == "" || seenQuestions[question.ID] || question.CorrectOptionID == "" {
			return domain.OfflinePracticeResult{}, ErrInvalid
		}
		seenQuestions[question.ID] = true
		selected, answered := answers[question.ID]
		if !answered {
			if partialPractice {
				continue
			}
			return domain.OfflinePracticeResult{}, ErrInvalid
		}
		validOption := false
		correctOption := false
		for _, option := range question.Options {
			if option.ID == selected {
				validOption = true
			}
			if option.ID == question.CorrectOptionID {
				correctOption = true
			}
		}
		if !validOption || !correctOption {
			return domain.OfflinePracticeResult{}, ErrInvalid
		}
		if selected == question.CorrectOptionID {
			correct++
		}
	}
	for questionID := range answers {
		if !seenQuestions[questionID] {
			return domain.OfflinePracticeResult{}, ErrInvalid
		}
	}

	total := len(input.Package.Questions)
	if partialPractice {
		total = len(input.Answers)
	}
	score := correct * 100 / total
	passed := !partialPractice && score >= 80
	now := s.gamificationNow().UTC()
	result := domain.OfflinePracticeResult{
		AttemptID: input.AttemptID, LessonID: manifest.LessonID,
		Score: score, Correct: correct, Total: total, Passed: passed,
		BestScore: score,
		CanRetry:  true, Verified: true,
	}
	var progress *domain.LessonProgress
	if !partialPractice {
		if s.progress[user] == nil {
			s.progress[user] = map[string]domain.LessonProgress{}
		}
		previous := s.progress[user][manifest.LessonID]
		best := previous.Completion
		if score > best {
			best = score
		}
		lessonProgress := domain.LessonProgress{
			LessonID: manifest.LessonID, Completion: best,
			LastPosition: total * 60, Completed: previous.Completed || passed,
			UpdatedAt: now,
		}
		progress = &lessonProgress
		result.BestScore = best
		result.Progress = s.offlinePracticeProgressLocked(user, manifest.Kind, level, manifest.LessonID, previous.Completed || passed)
	}
	record := offlinePracticeAttemptRecord{Fingerprint: fingerprint, ContentHash: manifest.ContentHash, Result: result}
	event := studyEvent{
		ID: "offline-practice:" + input.AttemptID, Kind: "lesson", Subject: manifest.Kind,
		LessonID: manifest.LessonID, Score: score, Answered: total,
		Seconds: 0, At: now, Progress: progress,
		OfflinePracticeAttemptID: input.AttemptID,
		OfflinePracticeRecord:    &record,
	}
	if partialPractice {
		event.Kind = "ssw-practice"
		event.Answered = 0
	}
	if err := s.recordStudyLocked(user, event); err != nil {
		return domain.OfflinePracticeResult{}, err
	}
	return result, nil
}

func (s *MemoryStore) offlinePracticeProgressLocked(user, kind, level, lessonID string, willBeCompleted bool) int {
	completed, total := 0, 0
	switch kind {
	case "vocabulary":
		for _, course := range s.vocabulary {
			if course.Level != level {
				continue
			}
			for _, lesson := range course.Lessons {
				total++
				if s.progress[user][lesson.ID].Completed || (lesson.ID == lessonID && willBeCompleted) {
					completed++
				}
			}
		}
	case "grammar":
		for _, lesson := range s.grammar {
			if lesson.Level != level {
				continue
			}
			total++
			if s.progress[user][lesson.ID].Completed || (lesson.ID == lessonID && willBeCompleted) {
				completed++
			}
		}
	case "kanji":
		for _, lesson := range s.kanji {
			if lesson.Level != level {
				continue
			}
			total++
			if s.progress[user][lesson.ID].Completed || (lesson.ID == lessonID && willBeCompleted) {
				completed++
			}
		}
	case "reading":
		for _, lesson := range s.reading {
			if lesson.Level != level {
				continue
			}
			total++
			if s.progress[user][lesson.ID].Completed || (lesson.ID == lessonID && willBeCompleted) {
				completed++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return completed * 100 / total
}
