package store

import (
	"errors"
	"fmt"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/testdb"
	"sync"
	"testing"
	"time"
)

func passStudyLesson(t *testing.T, s *MemoryStore, user, subject, id string) {
	t.Helper()
	switch subject {
	case "grammar":
		a, e := s.StartGrammarAttempt(user, id)
		if e != nil {
			t.Fatal(e)
		}
		for _, q := range a.Questions {
			if _, e = s.AnswerGrammarQuestion(user, a.ID, q.ID, s.grammarAttempts[a.ID].Correct[q.ID]); e != nil {
				t.Fatal(e)
			}
		}
		if _, e = s.CompleteGrammarAttempt(user, a.ID); e != nil {
			t.Fatal(e)
		}
		if _, e = s.CompleteGrammarAttempt(user, a.ID); e != nil {
			t.Fatal(e)
		}
	case "kanji":
		a, e := s.StartKanjiAttempt(user, id)
		if e != nil {
			t.Fatal(e)
		}
		for _, q := range a.Questions {
			if _, e = s.AnswerKanjiQuestion(user, a.ID, q.ID, s.kanjiAttempts[a.ID].Correct[q.ID]); e != nil {
				t.Fatal(e)
			}
		}
		if _, e = s.CompleteKanjiAttempt(user, a.ID); e != nil {
			t.Fatal(e)
		}
	case "reading":
		a, e := s.StartReadingAttempt(user, id)
		if e != nil {
			t.Fatal(e)
		}
		for _, q := range a.Questions {
			if _, e = s.AnswerReadingQuestion(user, a.ID, q.ID, s.readingAttempts[a.ID].Correct[q.ID]); e != nil {
				t.Fatal(e)
			}
		}
		if _, e = s.CompleteReadingAttempt(user, a.ID); e != nil {
			t.Fatal(e)
		}
	case "vocabulary":
		a, e := s.StartVocabularyAttempt(user, id)
		if e != nil {
			t.Fatal(e)
		}
		for _, q := range a.Questions {
			if _, e = s.AnswerVocabularyQuestion(user, a.ID, q.ID, s.vocabAttempts[a.ID].correct[q.ID]); e != nil {
				t.Fatal(e)
			}
		}
		if _, e = s.CompleteVocabularyAttempt(user, a.ID); e != nil {
			t.Fatal(e)
		}
	}
}

func achievementByID(t *testing.T, s GamificationSummary, id string) domain.Achievement {
	t.Helper()
	for _, a := range s.Achievements {
		if a.ID == id {
			return a
		}
	}
	t.Fatal("missing achievement", id)
	return domain.Achievement{}
}

func TestGamificationRealLessonRewardsPersistAndDoNotRepeat(t *testing.T) {
	path := testdb.DSN(t)
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(path); err != nil {
		t.Fatal(err)
	}
	lessons := map[string]string{"grammar": "grammar-n4-01", "kanji": "kanji-n4-01", "reading": "reading-n2-01", "vocabulary": "n5-daily"}
	for kind, id := range lessons {
		passStudyLesson(t, s, "c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", kind, id)
	}
	summary := s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec")
	if summary.XP != 200 || summary.LessonsCompleted != 4 || summary.StreakDays != 1 || summary.Level != 1 {
		t.Fatalf("bad summary: %+v", summary)
	}
	first := achievementByID(t, summary, "first-lesson")
	if first.UnlockedAt == nil || achievementByID(t, summary, "all-subjects").UnlockedAt == nil {
		t.Fatal("missing achievement unlock")
	}
	if s.Gamification("926ea9c7-89bb-496e-bb5c-c9d4c7d2df6f").XP != 0 {
		t.Fatal("account leak")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := NewMemoryStore()
	if err := reopened.EnableGamificationPersistence(path); err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	for kind, id := range lessons {
		passStudyLesson(t, reopened, "c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", kind, id)
	}
	after := reopened.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec")
	if after.XP != 200 || after.LessonsCompleted != 4 || !achievementByID(t, after, "first-lesson").UnlockedAt.Equal(*first.UnlockedAt) {
		t.Fatal("restart duplicated rewards or lost progress")
	}
	if reopened.Profile("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec").XP != 200 || !reopened.GrammarCourse("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", "N4").Lessons[0].Completed {
		t.Fatal("profile/course not restored")
	}
	if _, err := reopened.SaveProgress("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", domain.LessonProgress{LessonID: "n5-daily", Completion: 100}); !errors.Is(err, ErrInvalid) {
		t.Fatal("client can forge vocabulary completion")
	}
	if _, err := reopened.SaveProgress("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", domain.LessonProgress{LessonID: "fake", Completion: 100}); err != nil {
		t.Fatal(err)
	}
	if reopened.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec").LessonsCompleted != 4 {
		t.Fatal("manual progress awarded achievement")
	}
}

func TestGamificationStreakJakartaMidnightGapAndPermanentBadge(t *testing.T) {
	s := NewMemoryStore()
	base := time.Date(2026, 10, 1, 16, 59, 0, 0, time.UTC)
	now := base
	s.gamificationNow = func() time.Time { return now }
	record := func(id string, at time.Time, answered int) {
		t.Helper()
		s.mu.Lock()
		defer s.mu.Unlock()
		if err := s.recordStudyLocked("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", studyEvent{ID: id, Kind: "exam", ExamID: "diagnostic", Score: 0, Answered: answered, At: at, Seconds: 60}); err != nil {
			t.Fatal(err)
		}
	}
	record("first", now, 1)
	now = base.Add(time.Minute)
	record("midnight", now, 1)
	if s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec").StreakDays != 2 {
		t.Fatal("midnight boundary must use Jakarta")
	}
	for i := 1; i <= 5; i++ {
		now = base.Add(time.Minute).AddDate(0, 0, i)
		record(fmt.Sprint(i), now, 1)
	}
	summary := s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec")
	stamp := achievementByID(t, summary, "streak-7").UnlockedAt
	if summary.StreakDays != 7 || stamp == nil {
		t.Fatal("seven day reward missing")
	}
	record("same-day", now, 1)
	record("same-day", now, 1)
	if s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec").MinutesToday != 2 {
		t.Fatal("duplicate completion counted time twice")
	}
	now = now.AddDate(0, 0, 2)
	record("empty-exam", now, 0)
	if summary = s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec"); summary.StreakDays != 0 || summary.BestStreak != 7 || !achievementByID(t, summary, "streak-7").UnlockedAt.Equal(*stamp) {
		t.Fatal("expired streak removed badge or empty exam created activity")
	}
	record("new", now, 1)
	if s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec").StreakDays != 1 {
		t.Fatal("gap did not reset current streak")
	}
}

func TestGamificationConcurrentRewardsAndDatabaseFailure(t *testing.T) {
	s := NewMemoryStore()
	path := testdb.DSN(t)
	if err := s.EnableGamificationPersistence(path); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.mu.Lock()
			defer s.mu.Unlock()
			p := domain.LessonProgress{LessonID: "grammar-n5-01", Completed: true, Completion: 100}
			if err := s.recordStudyLocked("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", studyEvent{ID: fmt.Sprint(i), Kind: "lesson", Subject: "grammar", LessonID: p.LessonID, Score: 100, Answered: 5, At: time.Now(), Progress: &p}); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec").XP != 50 {
		t.Fatal("concurrent duplicate reward")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	a, err := s.StartGrammarAttempt("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", "grammar-n5-02")
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range a.Questions {
		if _, err = s.AnswerGrammarQuestion("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", a.ID, q.ID, s.grammarAttempts[a.ID].Correct[q.ID]); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.CompleteGrammarAttempt("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", a.ID); err == nil {
		t.Fatal("closed database reported completion success")
	}
	if s.grammarAttempts[a.ID].Result != nil || s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec").XP != 50 || s.progress["c5056e4a-dc47-4f74-bdb9-44b8c794c9ec"]["grammar-n5-02"].Completed {
		t.Fatal("disk failure partially awarded progress")
	}
}

func TestGamificationExamRewardsAndEmptyAttempts(t *testing.T) {
	s := NewMemoryStore()
	grantTestLifetime(t, s, "c5056e4a-dc47-4f74-bdb9-44b8c794c9ec")
	for i := 0; i < 2; i++ {
		a, err := s.StartSimulation("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", "N4")
		if err != nil {
			t.Fatal(err)
		}
		state := s.simulationAttempts[a.ID]
		for _, section := range []string{"vocabulary", "grammar"} {
			for _, q := range state.questions {
				if q.Section == section {
					if err = s.SaveSimulationAnswer("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", a.ID, q.ID, state.correct[q.ID], false); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err = s.SubmitSimulationSection("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec", a.ID, section); err != nil {
				t.Fatal(err)
			}
		}
	}
	if summary := s.Gamification("c5056e4a-dc47-4f74-bdb9-44b8c794c9ec"); summary.XP != 100 || summary.ExamsCompleted != 1 || achievementByID(t, summary, "exam-ready").UnlockedAt == nil {
		t.Fatal("exam reward duplicated or missing")
	}
	grantTestLifetime(t, s, "926ea9c7-89bb-496e-bb5c-c9d4c7d2df6f")
	empty, err := s.StartSimulation("926ea9c7-89bb-496e-bb5c-c9d4c7d2df6f", "N4")
	if err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"vocabulary", "grammar"} {
		if _, err = s.SubmitSimulationSection("926ea9c7-89bb-496e-bb5c-c9d4c7d2df6f", empty.ID, section); err != nil {
			t.Fatal(err)
		}
	}
	if summary := s.Gamification("926ea9c7-89bb-496e-bb5c-c9d4c7d2df6f"); summary.XP != 0 || summary.ActiveDays != 0 || summary.ExamsCompleted != 0 {
		t.Fatal("unanswered exam earned progress")
	}
}

func TestGamificationPersistenceRejectsSecondWriter(t *testing.T) {
	path := testdb.DSN(t)
	first := NewMemoryStore()
	if err := first.EnableGamificationPersistence(path); err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second := NewMemoryStore()
	if err := second.EnableGamificationPersistence(path); err == nil {
		second.Close()
		t.Fatal("second process can overwrite cached learning state")
	}
}
