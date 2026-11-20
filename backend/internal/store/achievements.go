package store

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"nihongo-master-backend/internal/domain"
)

const (
	achievementWeeklyTop10 = "weekly-top10"
	achievementComeback    = "comeback"
)

// achievementStats are derived from the saved gamification state only, so every
// achievement below can be computed (and awarded retroactively) without any
// extra tracking.
type achievementStats struct {
	Lessons       int // distinct lessons passed
	Subject       map[string]int
	Level         map[string]int // lessons passed per JLPT level, "N5".."N1"
	Perfect       int            // lessons passed with a 100% score
	ActiveDays    int
	BestStreak    int
	XP            int
	StudySeconds  int
	GoalDays      int // days with at least the daily goal of study time
	LongestDaySec int
	ExamTypes     int // distinct exams attempted
	ExamsPassed   int // distinct exams passed with at least 80%
	AnkiReviews   int
	Comeback      bool // returned after a break of a week or more
}

var lessonLevelPattern = regexp.MustCompile(`(?:^|-)n([1-5])(?:-|$)`)

// lessonLevel reads the JLPT level from lesson IDs such as "grammar-n4-01" or
// "n5-daily". It returns "" for IDs without a level.
func lessonLevel(id string) string {
	if m := lessonLevelPattern.FindStringSubmatch(strings.ToLower(id)); m != nil {
		return "N" + m[1]
	}
	return ""
}

func newAchievementStats(state *gamificationState, total int, subjects map[string]int, bestStreak int) achievementStats {
	s := achievementStats{
		Lessons: total, Subject: subjects, Level: map[string]int{}, ActiveDays: len(state.Days),
		BestStreak: bestStreak, XP: state.XP, ExamTypes: len(state.Exams),
	}
	for id, p := range state.Lessons {
		if !p.Completed {
			continue
		}
		if level := lessonLevel(id); level != "" {
			s.Level[level]++
		}
		if p.Completion >= 100 {
			s.Perfect++
		}
	}
	for key := range state.Rewards {
		if strings.HasPrefix(key, "exam:") {
			s.ExamsPassed++
		}
	}
	days := make([]string, 0, len(state.Days))
	for day, seconds := range state.Days {
		days = append(days, day)
		s.StudySeconds += seconds
		if seconds >= dailyGoalMinutes*60 {
			s.GoalDays++
		}
		if seconds > s.LongestDaySec {
			s.LongestDaySec = seconds
		}
	}
	sort.Strings(days)
	var previous time.Time
	for _, day := range days {
		date, err := time.Parse("2006-01-02", day)
		if err != nil {
			continue
		}
		if !previous.IsZero() && date.Sub(previous) >= 8*24*time.Hour {
			s.Comeback = true
		}
		previous = date
	}
	for _, deck := range state.Anki {
		if deck == nil {
			continue
		}
		for _, card := range deck.Cards {
			s.AnkiReviews += card.Version
		}
	}
	return s
}

type achievementRule struct {
	ID, Name, Description, Icon string
	Target                      int
	Value                       func(achievementStats) int
}

func subjectCount(subject string) func(achievementStats) int {
	return func(s achievementStats) int { return s.Subject[subject] }
}

func levelCount(level string) func(achievementStats) int {
	return func(s achievementStats) int { return s.Level[level] }
}

func boolValue(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

// achievementRules extend the original dozen in gamificationAchievements.
var achievementRules = []achievementRule{
	// Lessons passed
	{"lessons-25", "Konsisten", "Lulus 25 lesson berbeda", "book", 25, func(s achievementStats) int { return s.Lessons }},
	{"lessons-100", "Seratus Lesson", "Lulus 100 lesson berbeda", "book", 100, func(s achievementStats) int { return s.Lessons }},
	{"perfect-1", "Nilai Sempurna", "Dapatkan nilai 100% pada satu lesson", "star", 1, func(s achievementStats) int { return s.Perfect }},
	{"perfect-10", "Teliti", "Dapatkan nilai 100% pada 10 lesson berbeda", "star", 10, func(s achievementStats) int { return s.Perfect }},
	// Subjects
	{"vocabulary-50", "Kamus Berjalan", "Lulus 50 lesson kosakata", "words", 50, subjectCount("vocabulary")},
	{"grammar-50", "Ahli Grammar", "Lulus 50 lesson grammar", "grammar", 50, subjectCount("grammar")},
	{"kanji-50", "Master Kanji", "Lulus 50 lesson kanji", "kanji", 50, subjectCount("kanji")},
	{"reading-20", "Pembaca Sejati", "Lulus 20 lesson reading", "book", 20, subjectCount("reading")},
	// JLPT levels
	{"level-n5", "Fondasi N5", "Lulus 10 lesson level N5", "footprints", 10, levelCount("N5")},
	{"level-n4", "Langkah N4", "Lulus 10 lesson level N4", "footprints", 10, levelCount("N4")},
	{"level-n3", "Menengah N3", "Lulus 10 lesson level N3", "footprints", 10, levelCount("N3")},
	{"level-n2", "Tantangan N2", "Lulus 10 lesson level N2", "footprints", 10, levelCount("N2")},
	{"level-n1", "Puncak N1", "Lulus 10 lesson level N1", "trophy", 10, levelCount("N1")},
	// Streaks
	{"streak-3", "Mulai Terbiasa", "Belajar 3 hari berturut-turut", "flame", 3, func(s achievementStats) int { return s.BestStreak }},
	{"streak-14", "Dua Minggu Penuh", "Belajar 14 hari berturut-turut", "flame", 14, func(s achievementStats) int { return s.BestStreak }},
	{"streak-60", "Dua Bulan Tanpa Putus", "Belajar 60 hari berturut-turut", "flame", 60, func(s achievementStats) int { return s.BestStreak }},
	{"streak-100", "Seratus Hari", "Belajar 100 hari berturut-turut", "flame", 100, func(s achievementStats) int { return s.BestStreak }},
	{"streak-365", "Setahun Penuh", "Belajar 365 hari berturut-turut", "flame", 365, func(s achievementStats) int { return s.BestStreak }},
	// XP
	{"xp-100", "Poin Pertama", "Kumpulkan 100 XP", "bolt", 100, func(s achievementStats) int { return s.XP }},
	{"xp-1000", "Seribu XP", "Kumpulkan 1.000 XP", "bolt", 1000, func(s achievementStats) int { return s.XP }},
	{"xp-2500", "Pengumpul XP", "Kumpulkan 2.500 XP", "bolt", 2500, func(s achievementStats) int { return s.XP }},
	{"xp-5000", "Bintang XP", "Kumpulkan 5.000 XP", "bolt", 5000, func(s achievementStats) int { return s.XP }},
	{"xp-10000", "Legenda XP", "Kumpulkan 10.000 XP", "bolt", 10000, func(s achievementStats) int { return s.XP }},
	// Habit and time
	{"days-10", "10 Hari Belajar", "Belajar di 10 hari berbeda", "flame", 10, func(s achievementStats) int { return s.ActiveDays }},
	{"days-30", "30 Hari Belajar", "Belajar di 30 hari berbeda", "flame", 30, func(s achievementStats) int { return s.ActiveDays }},
	{"days-100", "100 Hari Belajar", "Belajar di 100 hari berbeda", "flame", 100, func(s achievementStats) int { return s.ActiveDays }},
	{"time-1h", "Satu Jam Pertama", "Belajar total 1 jam", "bolt", 60, func(s achievementStats) int { return s.StudySeconds / 60 }},
	{"time-10h", "10 Jam Belajar", "Belajar total 10 jam", "bolt", 600, func(s achievementStats) int { return s.StudySeconds / 60 }},
	{"goal-1", "Target Tercapai", "Capai target belajar harian 20 menit", "trophy", 1, func(s achievementStats) int { return s.GoalDays }},
	{"goal-7", "Target Seminggu", "Capai target belajar harian di 7 hari berbeda", "trophy", 7, func(s achievementStats) int { return s.GoalDays }},
	{"marathon-1h", "Maraton Belajar", "Belajar 60 menit dalam satu hari", "bolt", 60, func(s achievementStats) int { return s.LongestDaySec / 60 }},
	{achievementComeback, "Kembali Lagi", "Belajar lagi setelah istirahat 7 hari atau lebih", "flame", 1, func(s achievementStats) int { return boolValue(s.Comeback) }},
	// Exams
	{"exams-3", "Penjelajah Ujian", "Kerjakan 3 jenis ujian berbeda", "trophy", 3, func(s achievementStats) int { return s.ExamTypes }},
	{"exams-passed-5", "Pemburu Skor 80", "Dapatkan minimal 80% pada 5 ujian berbeda", "trophy", 5, func(s achievementStats) int { return s.ExamsPassed }},
	// Anki
	{"anki-50", "Pengulang Kartu", "Review 50 kartu Anki", "words", 50, func(s achievementStats) int { return s.AnkiReviews }},
	{"anki-500", "Rajin Review", "Review 500 kartu Anki", "words", 500, func(s achievementStats) int { return s.AnkiReviews }},
	{"anki-2000", "Ahli Flashcard", "Review 2.000 kartu Anki", "words", 2000, func(s achievementStats) int { return s.AnkiReviews }},
}

// extraAchievements returns the achievements added on top of the original set.
func extraAchievements(state *gamificationState, total int, subjects map[string]int, bestStreak int) []domain.Achievement {
	stats := newAchievementStats(state, total, subjects, bestStreak)
	items := make([]domain.Achievement, 0, len(achievementRules)+1)
	for _, rule := range achievementRules {
		items = append(items, domain.Achievement{ID: rule.ID, Name: rule.Name, Description: rule.Description, Icon: rule.Icon, Progress: rule.Value(stats), Target: rule.Target})
	}
	// Awarded when a leaderboard week closes, so progress comes from the unlock.
	items = append(items, domain.Achievement{ID: achievementWeeklyTop10, Name: "Sepuluh Besar Mingguan", Description: "Finis di peringkat 1–10 leaderboard mingguan", Icon: "trophy", Progress: unlockedCount(state, achievementWeeklyTop10), Target: 1})
	return items
}
