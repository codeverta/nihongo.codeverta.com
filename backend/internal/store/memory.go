package store

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"nihongo-master-backend/internal/domain"
)

var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalid            = errors.New("invalid input")
	ErrPremium            = errors.New("premium entitlement required")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSessionNotFound    = errors.New("session not found")
)

type account struct {
	ID           string
	Email        string
	DisplayName  string
	PhotoURL     string
	PasswordHash []byte
	GoogleSub    string
}

type MemoryStore struct {
	mu            sync.RWMutex
	profiles      map[string]domain.Profile
	progress      map[string]map[string]domain.LessonProgress
	achievements  map[string][]domain.Achievement
	categories    []domain.LearningCategory
	tracks        []domain.LearningTrack
	exams         []domain.Exam
	plans         []domain.Plan
	attemptID     atomic.Uint64
	accounts      map[string]account
	sessions      map[string]string
	vocabulary    []domain.VocabularyLevelContent
	vocabAttempts map[string]*vocabularyAttemptState
}

func NewMemoryStore() *MemoryStore {
	result := &MemoryStore{
		profiles:      map[string]domain.Profile{},
		progress:      map[string]map[string]domain.LessonProgress{},
		achievements:  map[string][]domain.Achievement{},
		categories:    seedCategories(),
		tracks:        seedTracks(),
		exams:         seedExams(),
		plans:         seedPlans(),
		accounts:      map[string]account{},
		sessions:      map[string]string{},
		vocabulary:    seedVocabulary(),
		vocabAttempts: map[string]*vocabularyAttemptState{},
	}
	result.seedDemoAccount()
	return result
}

func (s *MemoryStore) Dashboard(userID string) domain.Dashboard {
	dashboard := domain.Dashboard{
		IsAuthenticated: userID != "",
		DailyGoal:       domain.DailyGoal{TargetMinutes: 20},
		Categories:      s.Categories(),
		Tracks:          s.Tracks(""),
		Exams:           s.Exams(""),
		Plans:           s.Plans(),
	}
	if userID == "" {
		return dashboard
	}
	profile := s.Profile(userID)
	dashboard.Profile = &profile
	dashboard.DailyGoal.StreakDays = profile.StreakDays
	dashboard.DailyGoal.MinutesLearned = todayMinutes(s.Progress(userID))
	dashboard.Achievements = s.Achievements(userID)
	return dashboard
}

func todayMinutes(items []domain.LessonProgress) int {
	const secondsPerMinute = 60
	minutes := 0
	today := time.Now().UTC()
	for _, item := range items {
		if item.UpdatedAt.Year() == today.Year() && item.UpdatedAt.YearDay() == today.YearDay() {
			minutes += item.LastPosition / secondsPerMinute
		}
	}
	return minutes
}

func (s *MemoryStore) Categories() []domain.LearningCategory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.LearningCategory(nil), s.categories...)
}

func (s *MemoryStore) Tracks(system string) []domain.LearningTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return filterTracks(s.tracks, system)
}

func (s *MemoryStore) Exams(system string) []domain.Exam {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if system == "" {
		return append([]domain.Exam(nil), s.exams...)
	}
	var result []domain.Exam
	for _, exam := range s.exams {
		if strings.EqualFold(exam.System, system) {
			result = append(result, exam)
		}
	}
	return result
}

func (s *MemoryStore) Plans() []domain.Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Plan(nil), s.plans...)
}

func (s *MemoryStore) Profile(userID string) domain.Profile {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ensureProfileLocked(userID)
}

func (s *MemoryStore) UpdateProfile(userID, displayName, target string) (domain.Profile, error) {
	if strings.TrimSpace(displayName) == "" || strings.TrimSpace(target) == "" {
		return domain.Profile{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	profile := s.ensureProfileLocked(userID)
	profile.DisplayName = strings.TrimSpace(displayName)
	profile.Target = strings.ToUpper(strings.TrimSpace(target))
	s.profiles[userID] = profile
	return profile, nil
}

func (s *MemoryStore) Progress(userID string) []domain.LessonProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.progress[userID]
	result := make([]domain.LessonProgress, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}

func (s *MemoryStore) SaveProgress(userID string, input domain.LessonProgress) (domain.LessonProgress, error) {
	if input.LessonID == "" || input.Completion < 0 || input.Completion > 100 || input.LastPosition < 0 {
		return domain.LessonProgress{}, ErrInvalid
	}
	input.Completed = input.Completion == 100
	input.UpdatedAt = time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.progress[userID] == nil {
		s.progress[userID] = map[string]domain.LessonProgress{}
	}
	previous := s.progress[userID][input.LessonID]
	s.progress[userID][input.LessonID] = input
	if input.Completed && !previous.Completed {
		profile := s.ensureProfileLocked(userID)
		profile.XP += 50
		profile.Level = 1 + profile.XP/500
		s.profiles[userID] = profile
	}
	return input, nil
}

func (s *MemoryStore) Achievements(userID string) []domain.Achievement {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.achievements[userID]; ok {
		return append([]domain.Achievement(nil), existing...)
	}
	items := seedAchievements()
	s.achievements[userID] = items
	return append([]domain.Achievement(nil), items...)
}

func (s *MemoryStore) StartExam(userID, examID string) (domain.ExamAttempt, error) {
	s.mu.RLock()
	var selected *domain.Exam
	for i := range s.exams {
		if s.exams[i].ID == examID {
			copy := s.exams[i]
			selected = &copy
			break
		}
	}
	profile := s.profiles[userID]
	s.mu.RUnlock()
	if selected == nil {
		return domain.ExamAttempt{}, ErrNotFound
	}
	if selected.RequiredTier == domain.TierPremium && profile.Entitlement.Tier != domain.TierPremium {
		return domain.ExamAttempt{}, ErrPremium
	}
	id := s.attemptID.Add(1)
	return domain.ExamAttempt{
		ID:        fmt.Sprintf("attempt-%d", id),
		ExamID:    examID,
		UserID:    userID,
		Status:    "in_progress",
		StartedAt: time.Now().UTC(),
	}, nil
}

func (s *MemoryStore) ensureProfileLocked(userID string) domain.Profile {
	if profile, ok := s.profiles[userID]; ok {
		return profile
	}
	profile := domain.Profile{
		ID:          userID,
		DisplayName: "Nihongo Learner",
		Email:       userID + "@example.local",
		Target:      "JLPT N5",
		Level:       1,
		XP:          0,
		StreakDays:  0,
		Entitlement: domain.Entitlement{Tier: domain.TierFree},
	}
	s.profiles[userID] = profile
	return profile
}

func filterTracks(tracks []domain.LearningTrack, system string) []domain.LearningTrack {
	if system == "" {
		return append([]domain.LearningTrack(nil), tracks...)
	}
	var result []domain.LearningTrack
	for _, track := range tracks {
		if strings.EqualFold(track.System, system) {
			result = append(result, track)
		}
	}
	return result
}

func seedCategories() []domain.LearningCategory {
	return []domain.LearningCategory{
		{ID: "vocabulary", Name: "Vocabulary", Description: "Kosakata tematik dengan spaced repetition", LessonCount: 240, RequiredTier: domain.TierFree},
		{ID: "grammar", Name: "Grammar", Description: "Pola tata bahasa dan contoh kalimat", LessonCount: 186, RequiredTier: domain.TierFree},
		{ID: "kanji", Name: "Kanji", Description: "Makna, onyomi, kunyomi, dan latihan tulis", LessonCount: 650, RequiredTier: domain.TierFree},
		{ID: "listening", Name: "Listening", Description: "Dialog native speaker dan latihan soal", LessonCount: 92, RequiredTier: domain.TierPremium},
		{ID: "reading", Name: "Reading", Description: "Bacaan bertingkat sesuai kemampuan", LessonCount: 68, RequiredTier: domain.TierPremium},
		{ID: "speaking", Name: "Speaking AI", Description: "Latihan pelafalan dan percakapan", LessonCount: 80, RequiredTier: domain.TierPremium},
	}
}

func seedTracks() []domain.LearningTrack {
	return []domain.LearningTrack{
		{ID: "jlpt-n5", System: "JLPT", Level: "N5", Title: "JLPT N5 Foundation", Description: "Fondasi kana, 100 kanji, dan grammar dasar", RequiredTier: domain.TierFree},
		{ID: "jlpt-n4", System: "JLPT", Level: "N4", Title: "JLPT N4 Elementary", Description: "Pemahaman bahasa Jepang sehari-hari", RequiredTier: domain.TierFree},
		{ID: "jlpt-n3", System: "JLPT", Level: "N3", Title: "JLPT N3 Intermediate", Description: "Jembatan menuju bahasa Jepang tingkat lanjut", RequiredTier: domain.TierPremium},
		{ID: "jlpt-n2", System: "JLPT", Level: "N2", Title: "JLPT N2 Advanced", Description: "Bahasa akademik dan profesional", RequiredTier: domain.TierPremium},
		{ID: "jlpt-n1", System: "JLPT", Level: "N1", Title: "JLPT N1 Mastery", Description: "Penguasaan bahasa Jepang tingkat tinggi", RequiredTier: domain.TierPremium},
		{ID: "nat-5q", System: "NAT", Level: "5Q", Title: "NAT-TEST 5Q", Description: "Persiapan NAT level dasar", RequiredTier: domain.TierFree},
		{ID: "jft-a2", System: "JFT", Level: "A2", Title: "JFT-Basic A2", Description: "Bahasa Jepang untuk kehidupan dan kerja", RequiredTier: domain.TierPremium},
	}
}

func seedExams() []domain.Exam {
	return []domain.Exam{
		{ID: "jlpt-n5-diagnostic", System: "JLPT", Level: "N5", Title: "JLPT N5 Diagnostic", QuestionCount: 25, DurationMin: 30, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-full-1", System: "JLPT", Level: "N5", Title: "JLPT N5 Full Simulation", QuestionCount: 80, DurationMin: 105, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-1", System: "NAT", Level: "5Q", Title: "NAT 5Q Simulation", QuestionCount: 70, DurationMin: 90, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-1", System: "JFT", Level: "A2", Title: "JFT-Basic A2 Simulation", QuestionCount: 60, DurationMin: 60, RequiredTier: domain.TierPremium},
	}
}

func seedPlans() []domain.Plan {
	features := []string{"Semua level N5–N1", "JLPT, NAT, dan JFT mock exam", "Materi offline", "Analisis belajar lengkap"}
	return []domain.Plan{
		{ID: "free", Name: "Free", BillingType: "free", PriceIDR: 0, Features: []string{"JLPT N5–N4 dasar", "Daily practice", "Progress dasar"}},
		{ID: "premium-monthly", Name: "Premium Bulanan", BillingType: "monthly", PriceIDR: 59000, Features: features},
		{ID: "premium-yearly", Name: "Premium Tahunan", BillingType: "yearly", PriceIDR: 499000, Features: features, Recommended: true},
		{ID: "premium-lifetime", Name: "Lifetime Access", BillingType: "lifetime", PriceIDR: 1499000, Features: append(features, "Akses selamanya")},
	}
}

func seedAchievements() []domain.Achievement {
	return []domain.Achievement{
		{ID: "first-lesson", Name: "First Step", Description: "Selesaikan pelajaran pertama", Icon: "footprints", Progress: 0, Target: 1},
		{ID: "streak-7", Name: "7-Day Streak", Description: "Belajar tujuh hari berturut-turut", Icon: "flame", Progress: 0, Target: 7},
		{ID: "kana-master", Name: "Kana Master", Description: "Kuasai hiragana dan katakana", Icon: "kana", Progress: 0, Target: 92},
		{ID: "exam-ready", Name: "Exam Ready", Description: "Dapatkan skor minimal 80%", Icon: "trophy", Progress: 0, Target: 1},
	}
}
