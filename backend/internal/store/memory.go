package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
)

var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalid            = errors.New("invalid input")
	ErrPremium            = errors.New("premium entitlement required")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrReviewEmailLogin   = errors.New("play review account must use email and password")
	ErrSessionNotFound    = errors.New("session not found")
)

type account struct {
	ID             string
	Email          string
	DisplayName    string
	PhotoURL       string
	CustomPhotoURL string
	Timezone       string
	PasswordHash   []byte
	GoogleSub      string
}

type MemoryStore struct {
	purchases            map[string]domain.Purchase
	billingTokens        map[string]string
	billingEnabled       bool
	mu                   sync.RWMutex
	gamification         map[string]*gamificationState
	userAnkiDecks        map[string]*userAnkiDeckRecord
	userAnkiDeckRequests map[string]string
	gamificationDB       *sql.DB
	ormDB                *gorm.DB
	ormWriter            *gorm.DB
	persistenceConn      *sql.Conn
	persistenceLockName  string
	persistenceClosed    bool
	persistenceStop      chan struct{}
	gamificationNow      func() time.Time
	profiles             map[string]domain.Profile
	progress             map[string]map[string]domain.LessonProgress
	categories           []domain.LearningCategory
	tracks               []domain.LearningTrack
	exams                []domain.Exam
	plans                []domain.Plan
	attemptID            atomic.Uint64
	accounts             map[string]account
	playReviewEmail      string
	playReviewUserID     string
	sessions             map[string]string
	vocabulary           []domain.VocabularyLevelContent
	vocabAttempts        map[string]*vocabularyAttemptState
	assetBaseURL         string
	grammar              []grammarLessonContent
	grammarAttempts      map[string]*grammarAttemptState
	kanji                []kanjiLessonContent
	kanjiAttempts        map[string]*kanjiAttemptState
	reading              []readingLessonContent
	readingAttempts      map[string]*readingAttemptState
	diagnosticAttempts   map[string]*diagnosticAttemptState
	diagnosticOrder      []string
	simulationAttempts   map[string]*simulationAttemptState
	simulationOrder      []string
	simulationNow        func() time.Time
}

func (s *MemoryStore) SetAssetBaseURL(value string) {
	s.mu.Lock()
	s.assetBaseURL = strings.TrimRight(strings.TrimSpace(value), "/")
	s.mu.Unlock()
}

func NewMemoryStore() *MemoryStore {
	result := &MemoryStore{
		purchases:            map[string]domain.Purchase{},
		billingTokens:        map[string]string{},
		profiles:             map[string]domain.Profile{},
		gamification:         map[string]*gamificationState{},
		userAnkiDecks:        map[string]*userAnkiDeckRecord{},
		userAnkiDeckRequests: map[string]string{},
		gamificationNow:      time.Now,
		progress:             map[string]map[string]domain.LessonProgress{},
		categories:           seedCategories(),
		tracks:               seedTracks(),
		exams:                seedExams(),
		plans:                seedPlans(),
		accounts:             map[string]account{},
		sessions:             map[string]string{},
		vocabulary:           seedVocabulary(),
		vocabAttempts:        map[string]*vocabularyAttemptState{},
		grammar:              append(append(append(append(seedGrammar(), seedGrammarN4()...), seedGrammarN3()...), seedGrammarN2()...), seedGrammarN1()...),
		grammarAttempts:      map[string]*grammarAttemptState{},
		kanji:                seedKanji(),
		kanjiAttempts:        map[string]*kanjiAttemptState{},
		reading:              seedReading(),
		readingAttempts:      map[string]*readingAttemptState{},
		diagnosticAttempts:   map[string]*diagnosticAttemptState{},
		simulationAttempts:   map[string]*simulationAttemptState{},
		simulationNow:        time.Now,
	}
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		panic(err)
	}
	result.attemptID.Store(binary.LittleEndian.Uint64(nonce[:]) >> 1)
	result.seedDemoAccount()
	for i := range result.categories {
		if result.categories[i].ID == "grammar" {
			result.categories[i].LessonCount = len(result.grammar)
		}
		if result.categories[i].ID == "reading" {
			result.categories[i].LessonCount = len(result.reading)
		}
		if result.categories[i].ID == "kanji" {
			result.categories[i].LessonCount = len(result.kanji)
		}
		if result.categories[i].ID == "vocabulary" {
			result.categories[i].LessonCount = 0
			for _, level := range result.vocabulary {
				result.categories[i].LessonCount += len(level.Lessons)
			}
		}
	}
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
	dashboard.DailyGoal.MinutesLearned = s.Gamification(userID).MinutesToday
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
	result := append([]domain.LearningCategory(nil), s.categories...)
	if openLearningAccess {
		for i := range result {
			result[i].RequiredTier = domain.TierFree
		}
	}
	return result
}

func (s *MemoryStore) Tracks(system string) []domain.LearningTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := filterTracks(s.tracks, system)
	if openLearningAccess {
		for i := range result {
			result[i].RequiredTier = domain.TierFree
		}
	}
	return result
}

func (s *MemoryStore) Exams(system string) []domain.Exam {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []domain.Exam
	for _, exam := range s.exams {
		if system != "" && !strings.EqualFold(exam.System, system) {
			continue
		}
		if openLearningAccess {
			exam.RequiredTier = domain.TierFree
		}
		result = append(result, exam)
	}
	if system == "" && result == nil {
		return []domain.Exam{}
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
	profile := s.ensureProfileLocked(userID)
	profile.Entitlement = s.entitlementLocked(userID)
	if profile.Timezone == "" {
		profile.Timezone = defaultStudyTimezone
	}
	if state := s.gamification[userID]; state != nil {
		profile.XP = state.XP
		profile.Level = 1 + state.XP/500
		profile.StreakDays, _ = studyStreaks(state, s.gamificationNow())
	}
	return profile
}

func (s *MemoryStore) UpdateStudyTimezone(userID, timezone string) (domain.Profile, error) {
	if !isValidStudyTimezone(timezone) {
		return domain.Profile{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	profile := s.ensureProfileLocked(userID)
	profile.Timezone = timezone
	profile.Entitlement = s.entitlementLocked(userID)
	state := s.gamification[userID]
	if state == nil {
		state = newGamificationState()
	}
	nextState := *state
	nextState.Timezone = timezone
	if s.gamificationDB != nil {
		raw, err := json.Marshal(&nextState)
		if err != nil {
			return domain.Profile{}, err
		}
		if err = s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state"})}).Create(&database.GamificationUser{UserID: userID, State: raw}).Error; err != nil {
			return domain.Profile{}, fmt.Errorf("persist study timezone: %w", err)
		}
	}
	updatedAccounts := make(map[string]account)
	for email, account := range s.accounts {
		if account.ID != userID {
			continue
		}
		account.Timezone = timezone
		if s.billingEnabled {
			raw, err := json.Marshal(account)
			if err != nil {
				return domain.Profile{}, err
			}
			if err = s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state"})}).Create(&database.Account{ID: account.ID, Email: email, State: raw}).Error; err != nil {
				return domain.Profile{}, fmt.Errorf("persist account timezone: %w", err)
			}
		}
		updatedAccounts[email] = account
	}
	s.gamification[userID] = &nextState
	for email, account := range updatedAccounts {
		s.accounts[email] = account
	}
	s.profiles[userID] = profile
	return profile, nil
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
	profile.Entitlement = s.entitlementLocked(userID)
	s.profiles[userID] = profile
	return profile, nil
}

// UpdateProfilePhoto replaces an account's uploaded photo URL. The Google
// account photo remains available as a fallback if the user later removes it.
func (s *MemoryStore) UpdateProfilePhoto(userID, photoURL string) (domain.Profile, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if userID == "" || userID == s.playReviewUserID || strings.TrimSpace(photoURL) == "" {
		return domain.Profile{}, "", ErrInvalid
	}
	var email string
	var account account
	for candidateEmail, candidate := range s.accounts {
		if candidate.ID == userID {
			email, account = candidateEmail, candidate
			break
		}
	}
	if email == "" {
		return domain.Profile{}, "", ErrNotFound
	}
	previousPhotoURL := account.CustomPhotoURL
	account.CustomPhotoURL = strings.TrimSpace(photoURL)
	if s.billingEnabled {
		raw, err := json.Marshal(account)
		if err != nil {
			return domain.Profile{}, "", err
		}
		if err = s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state"})}).Create(&database.Account{ID: account.ID, Email: email, State: raw}).Error; err != nil {
			return domain.Profile{}, "", err
		}
	}
	s.accounts[email] = account
	profile := s.ensureProfileForAccountLocked(account)
	profile.Entitlement = s.entitlementLocked(userID)
	s.profiles[userID] = profile
	return profile, previousPhotoURL, nil
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
	if input.LessonID == "" || (strings.HasPrefix(input.LessonID, "grammar-") || strings.HasPrefix(input.LessonID, "kanji-") || strings.HasPrefix(input.LessonID, "reading-")) || input.Completion < 0 || input.Completion > 100 || input.LastPosition < 0 {
		return domain.LessonProgress{}, ErrInvalid
	}
	input.Completed = input.Completion == 100
	input.UpdatedAt = time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, lesson := s.findVocabularyLessonLocked(input.LessonID); lesson != nil {
		return domain.LessonProgress{}, ErrInvalid
	}
	if s.progress[userID] == nil {
		s.progress[userID] = map[string]domain.LessonProgress{}
	}
	s.progress[userID][input.LessonID] = input

	return input, nil
}

func (s *MemoryStore) Achievements(userID string) []domain.Achievement {
	return s.Gamification(userID).Achievements
}

func (s *MemoryStore) StartExam(userID, examID string) (domain.ExamAttempt, error) {
	for _, level := range SSWLevels() {
		if examID == strings.ToLower(level)+"-simulation-1" {
			return s.StartSimulation(userID, level)
		}
	}
	if rest, ok := strings.CutPrefix(examID, "nat-5q-full-"); ok {
		paperNo, err := strconv.Atoi(rest)
		if err != nil || strconv.Itoa(paperNo) != rest || paperNo < 1 || paperNo > natLastPaper {
			return domain.ExamAttempt{}, ErrNotFound
		}
		if paperNo == 1 {
			return s.StartSimulation(userID, "NAT")
		}
		return s.StartSimulation(userID, fmt.Sprintf("NAT-%d", paperNo))
	}
	if examID == "jft-a2-full-1" {
		return s.StartSimulation(userID, "JFT")
	}
	for paperNo := 2; paperNo <= jftLastPaper; paperNo++ {
		if examID == fmt.Sprintf("jft-a2-full-%d", paperNo) {
			return s.StartSimulation(userID, fmt.Sprintf("JFT-%d", paperNo))
		}
	}
	// Older clients used this ID for the premium N5 catalog placeholder.
	if examID == "jlpt-n5-full-1" {
		return s.StartSimulation(userID)
	}
	if examID == "jlpt-n3-simulation-1" {
		return s.StartSimulation(userID, "N3")
	}
	if examID == "jlpt-n2-simulation-1" {
		return s.StartSimulation(userID, "N2")
	}
	if examID == "jlpt-n1-simulation-1" {
		return s.StartSimulation(userID, "N1")
	}
	if examID == "jlpt-n4-simulation-1" {
		return s.StartSimulation(userID, "N4")
	}
	if base, paperNo, ok := jlptExtraExam(examID); ok {
		return s.StartSimulation(userID, fmt.Sprintf("%s-%d", base, paperNo))
	}
	if examID == "jlpt-n5-simulation-1" {
		return s.StartSimulation(userID)
	}
	if examID == "jlpt-n5-diagnostic" {
		return s.startDiagnostic(userID), nil
	}
	s.mu.RLock()
	var selected *domain.Exam
	for i := range s.exams {
		if s.exams[i].ID == examID {
			copy := s.exams[i]
			selected = &copy
			break
		}
	}
	profile := s.entitlementLocked(userID)
	s.mu.RUnlock()
	if selected == nil {
		return domain.ExamAttempt{}, ErrNotFound
	}
	if !openLearningAccess && selected.RequiredTier == domain.TierPremium && profile.Tier != domain.TierPremium {
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
		{ID: "grammar", Name: "Grammar", Description: "Grammar JLPT N5 sampai N1: pola, contoh kalimat, dan kuis", LessonCount: 196, RequiredTier: domain.TierFree},
		{ID: "kanji", Name: "Kanji", Description: "Kanji JLPT N5 sampai N1: bacaan, contoh pemakaian, dan kuis", LessonCount: 452, RequiredTier: domain.TierFree},
		{ID: "reading", Name: "Reading", Description: "Reading JLPT N5 sampai N1: cerita, kosakata, dan kuis", LessonCount: 60, RequiredTier: domain.TierFree},
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
		{ID: "ssw-agriculture-crop-simulation-1", System: "SSW", Level: "Tanaman", Title: "Pertanian · Tanaman · Simulasi Tertulis 1", QuestionCount: 70, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "ssw-agriculture-livestock-simulation-1", System: "SSW", Level: "Peternakan", Title: "Pertanian · Peternakan · Simulasi Tertulis 1", QuestionCount: 70, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "ssw-fishery-capture-simulation-1", System: "SSW", Level: "Tangkap", Title: "Perikanan Tangkap · Simulasi 1", QuestionCount: 55, DurationMin: 70, RequiredTier: domain.TierPremium},
		{ID: "ssw-fishery-aquaculture-simulation-1", System: "SSW", Level: "Budi daya", Title: "Budi Daya · Simulasi 1", QuestionCount: 50, DurationMin: 70, RequiredTier: domain.TierPremium},
		{ID: "ssw-kaigo-simulation-1", System: "SSW", Level: "Kaigo", Title: "Kaigo · Simulasi Keterampilan 1", QuestionCount: 45, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "ssw-kaigo-jp-simulation-1", System: "SSW", Level: "Kaigo", Title: "Kaigo · Simulasi Bahasa Jepang 1", QuestionCount: 15, DurationMin: 30, RequiredTier: domain.TierPremium},
		{ID: "ssw-cleaning-simulation-1", System: "SSW", Level: "Building cleaning", Title: "Pembersihan Gedung · Simulasi 1", QuestionCount: 50, DurationMin: 50, RequiredTier: domain.TierPremium},
		{ID: "jlpt-n3-simulation-1", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierPremium},
		{ID: "jlpt-n2-simulation-1", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierPremium},
		{ID: "jlpt-n1-simulation-1", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierPremium},
		{ID: "jlpt-n3-simulation-2", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 02 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n3-simulation-3", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 03 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-2", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 02 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-3", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 03 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-2", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 02 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-3", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 03 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n3-simulation-4", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 04 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n3-simulation-5", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 05 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n3-simulation-6", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 06 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n3-simulation-7", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 07 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n3-simulation-8", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 08 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n3-simulation-9", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 09 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n3-simulation-10", System: "JLPT", Level: "N3", Title: "JLPT N3 Simulation · Paket 10 · Kosakata, Grammar & Reading", QuestionCount: 74, DurationMin: 100, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-4", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 04 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-5", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 05 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-6", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 06 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-7", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 07 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-8", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 08 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-9", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 09 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n2-simulation-10", System: "JLPT", Level: "N2", Title: "JLPT N2 Simulation · Paket 10 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 105, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-4", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 04 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-5", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 05 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-6", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 06 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-7", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 07 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-8", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 08 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-9", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 09 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n1-simulation-10", System: "JLPT", Level: "N1", Title: "JLPT N1 Simulation · Paket 10 · Kosakata, Grammar & Reading", QuestionCount: 64, DurationMin: 110, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-1", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-diagnostic", System: "JLPT", Level: "N5", Title: "JLPT N5 Diagnostic", QuestionCount: 25, DurationMin: 30, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-1", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-2", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 02 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-3", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 03 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-2", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 02 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-3", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 03 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-4", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 04 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-5", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 05 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-4", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 04 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-5", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 05 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-6", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 06 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-7", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 07 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-6", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 06 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-7", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 07 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-8", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 08 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-9", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 09 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n5-simulation-10", System: "JLPT", Level: "N5", Title: "JLPT N5 Simulation · Paket 10 · Kosakata & Grammar", QuestionCount: 38, DurationMin: 40, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-8", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 08 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-9", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 09 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "jlpt-n4-simulation-10", System: "JLPT", Level: "N4", Title: "JLPT N4 Simulation · Paket 10 · Kosakata, Grammar & Reading", QuestionCount: 57, DurationMin: 80, RequiredTier: domain.TierFree},
		{ID: "nat-5q-full-1", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 1 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-2", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 2 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-3", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 3 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-4", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 4 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-5", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 5 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-6", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 6 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-7", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 7 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-8", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 8 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-9", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 9 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "nat-5q-full-10", System: "NAT", Level: "5Q", Title: "NAT 5Q Exam 10 · Latihan tertulis", QuestionCount: 43, DurationMin: 60, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-1", System: "JFT", Level: "A2", Title: "JFT Exam 1 · Latihan tertulis", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-2", System: "JFT", Level: "A2", Title: "JFT Exam 2 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-3", System: "JFT", Level: "A2", Title: "JFT Exam 3 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-4", System: "JFT", Level: "A2", Title: "JFT Exam 4 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-5", System: "JFT", Level: "A2", Title: "JFT Exam 5 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-6", System: "JFT", Level: "A2", Title: "JFT Exam 6 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-7", System: "JFT", Level: "A2", Title: "JFT Exam 7 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-8", System: "JFT", Level: "A2", Title: "JFT Exam 8 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-9", System: "JFT", Level: "A2", Title: "JFT Exam 9 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
		{ID: "jft-a2-full-10", System: "JFT", Level: "A2", Title: "JFT Exam 10 · Latihan tertulis orisinal", QuestionCount: 36, DurationMin: 45, RequiredTier: domain.TierPremium},
	}
}

func seedPlans() []domain.Plan {
	features := []string{"Materi N3–N1 yang tersedia", "Seluruh tes yang tersedia", "Restore pembelian", "Sekali bayar tanpa perpanjangan"}
	return []domain.Plan{
		{ID: "free", Name: "Free", BillingType: "free", PriceIDR: 0, Features: []string{"JLPT N5–N4 dasar", "Daily practice", "Progress dasar"}},
		{ID: "premium-lifetime", Name: "Lifetime Access", BillingType: "lifetime", PriceIDR: 99000, Features: features, Recommended: true},
	}
}
