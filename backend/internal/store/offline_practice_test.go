package store

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/testdb"
)

func TestOfflinePracticeQuestionBanksUseAuthoredLessonContent(t *testing.T) {
	s := NewMemoryStore()
	for _, tc := range []struct{ kind, lessonID string }{
		{"vocabulary", "n5-daily"},
		{"grammar", "grammar-n5-01"},
		{"kanji", "kanji-n5-01"},
		{"reading", "reading-n5-01"},
	} {
		practice, err := s.OfflinePracticePackage(database.DemoUserID, tc.kind, tc.lessonID)
		if err != nil || practice.Manifest.Kind != tc.kind || practice.Manifest.LessonID != tc.lessonID || len(practice.Questions) == 0 {
			t.Fatalf("missing authored %s practice for %s: %#v %v", tc.kind, tc.lessonID, practice, err)
		}
		for _, question := range practice.Questions {
			found := false
			for _, option := range question.Options {
				if option.ID == question.CorrectOptionID {
					found = true
				}
			}
			if question.Prompt == "" || question.Explanation == "" || !found {
				t.Fatalf("incomplete authored practice question: %#v", question)
			}
		}
	}
	if _, err := s.OfflinePracticePackage(database.DemoUserID, "simulation", "jlpt-n5"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("exam simulation was exposed as lesson practice: %v", err)
	}
	if _, err := s.OfflinePracticePackage(database.DemoUserID, "grammar", "grammar-n3-01"); !errors.Is(err, ErrPremium) {
		t.Fatalf("free account received a premium package: %v", err)
	}
}

func TestOfflinePracticeSyncIsIdempotentAndPreservesBestScore(t *testing.T) {
	s := NewMemoryStore()
	user := database.DemoUserID
	first := offlinePracticeInput(t, s, user, "grammar", "grammar-n5-01", 4)
	firstFingerprint := "fingerprint-first"
	result, err := s.SyncOfflinePracticeAttempt(user, first, firstFingerprint)
	if err != nil || !result.Verified || result.Score != 80 || result.BestScore != 80 || !result.Passed {
		t.Fatalf("first verified completion failed: %#v %v", result, err)
	}
	if xp := s.Gamification(user).XP; xp != 50 {
		t.Fatalf("server did not award lesson XP: %d", xp)
	}
	duplicate, err := s.SyncOfflinePracticeAttempt(user, first, firstFingerprint)
	if err != nil || duplicate != result || s.Gamification(user).XP != 50 {
		t.Fatalf("duplicate sync was not idempotent: %#v %v xp=%d", duplicate, err, s.Gamification(user).XP)
	}
	if _, err = s.SyncOfflinePracticeAttempt(user, first, "different-answer-fingerprint"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("attempt ID accepted changed answers: %v", err)
	}

	second := offlinePracticeInput(t, s, user, "grammar", "grammar-n5-01", 2)
	secondResult, err := s.SyncOfflinePracticeAttempt(user, second, "fingerprint-second")
	if err != nil || secondResult.Score != 40 || secondResult.BestScore != 80 || secondResult.Passed || s.Gamification(user).XP != 50 {
		t.Fatalf("lower retry replaced best score or repeated XP: %#v %v xp=%d", secondResult, err, s.Gamification(user).XP)
	}
	if progress := s.Progress(user); len(progress) != 1 || progress[0].Completion != 80 || !progress[0].Completed {
		t.Fatalf("best score or pass was not preserved: %#v", progress)
	}
}

func TestOfflinePracticeAttemptIsAccountBound(t *testing.T) {
	s := NewMemoryStore()
	input := offlinePracticeInput(t, s, database.DemoUserID, "reading", "reading-n5-01", 5)
	if _, err := s.SyncOfflinePracticeAttempt("another-user", input, "fingerprint"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("attempt crossed account boundary: %v", err)
	}
}

func TestOfflinePracticeDuplicateSurvivesEntitlementOrContentChanges(t *testing.T) {
	s := NewMemoryStore()
	user := database.DemoUserID
	input := offlinePracticeInput(t, s, user, "grammar", "grammar-n5-01", 4)
	fingerprint := "replay-after-content-change"
	want, err := s.SyncOfflinePracticeAttempt(user, input, fingerprint)
	if err != nil {
		t.Fatal(err)
	}

	// A package already accepted by the server stays idempotently replayable if
	// its lesson is later removed or the account's entitlement changes.
	s.grammar = nil
	got, err := s.SyncOfflinePracticeAttempt(user, input, fingerprint)
	if err != nil || got != want || s.Gamification(user).XP != 50 {
		t.Fatalf("verified replay depended on current lesson content: %#v %v", got, err)
	}
}

func TestOfflinePracticeSSWSyncIsVerifiedIdempotentAndDoesNotReward(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.OfflinePracticePackage("free", "ssw", "building-cleaning"); !errors.Is(err, ErrPremium) {
		t.Fatalf("free account received the SSW practice package: %v", err)
	}
	user := "ssw-practice-user"
	grantTestLifetime(t, s, user)
	practice, err := s.OfflinePracticePackage(user, "ssw", "building-cleaning")
	if err != nil || practice.Manifest.Level != "SSW" || len(practice.Questions) != 60 {
		t.Fatalf("missing authored SSW practice bank: %d questions, %v", len(practice.Questions), err)
	}
	practice.Manifest.UserID = user
	input := domain.OfflinePracticeSyncInput{AttemptID: uuid.NewString(), Package: practice}
	module := practice.Questions[0].ModuleID
	for _, question := range practice.Questions {
		if question.ModuleID != module || len(input.Answers) == 4 {
			continue
		}
		input.Answers = append(input.Answers, domain.OfflinePracticeAnswer{
			QuestionID: question.ID, OptionID: question.CorrectOptionID,
		})
	}
	if len(input.Answers) != 4 {
		t.Fatalf("fixture module had too few questions: %s", module)
	}
	fingerprint := "ssw-practice-partial-attempt"
	first, err := s.SyncOfflinePracticeAttempt(user, input, fingerprint)
	if err != nil || !first.Verified || first.Score != 100 || first.Total != 4 || first.Passed || first.Progress != 0 {
		t.Fatalf("server did not verify SSW partial practice without an exam pass: %#v %v", first, err)
	}
	if summary := s.Gamification(user); summary.XP != 0 || summary.LessonsCompleted != 0 {
		t.Fatalf("ordinary SSW practice changed rewards: %+v", summary)
	} else {
		for _, achievement := range summary.Achievements {
			if achievement.UnlockedAt != nil {
				t.Fatalf("ordinary SSW practice unlocked an achievement: %+v", achievement)
			}
		}
	}
	duplicate, err := s.SyncOfflinePracticeAttempt(user, input, fingerprint)
	if err != nil || duplicate != first {
		t.Fatalf("duplicate SSW sync changed its result: %#v %v", duplicate, err)
	}
	if xp := s.Gamification(user).XP; xp != 0 {
		t.Fatalf("duplicate SSW sync awarded XP: %d", xp)
	}
}

func TestOfflinePracticeResultSurvivesPersistenceRestart(t *testing.T) {
	dsn := testdb.DSN(t)
	user := database.DemoUserID
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	input := offlinePracticeInput(t, s, user, "kanji", "kanji-n5-01", 10)
	fingerprint := "persisted-attempt-fingerprint"
	first, err := s.SyncOfflinePracticeAttempt(user, input, fingerprint)
	if err != nil {
		t.Fatal(err)
	}
	if xp := s.Gamification(user).XP; xp != 50 {
		t.Fatalf("first verified completion failed to award XP: %d, result=%#v", xp, first)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	restored := NewMemoryStore()
	if err := restored.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	replayed, err := restored.SyncOfflinePracticeAttempt(user, input, fingerprint)
	if err != nil || replayed != first {
		t.Fatalf("persisted duplicate lost its authoritative result: %#v %v, first %#v", replayed, err, first)
	}
	if xp := restored.Gamification(user).XP; xp != 50 {
		t.Fatalf("restart replay awarded XP twice: %d", xp)
	}
}

func offlinePracticeInput(t *testing.T, s *MemoryStore, user, kind, lessonID string, correctCount int) domain.OfflinePracticeSyncInput {
	t.Helper()
	practice, err := s.OfflinePracticePackage(user, kind, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	practice.Manifest.UserID = user
	input := domain.OfflinePracticeSyncInput{AttemptID: uuid.NewString(), Package: practice}
	for i, question := range practice.Questions {
		optionID := question.CorrectOptionID
		if i >= correctCount {
			for _, option := range question.Options {
				if option.ID != question.CorrectOptionID {
					optionID = option.ID
					break
				}
			}
		}
		input.Answers = append(input.Answers, domain.OfflinePracticeAnswer{QuestionID: question.ID, OptionID: optionID})
	}
	return input
}
