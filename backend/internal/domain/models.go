package domain

import "time"

type AccessTier string

const (
	TierFree    AccessTier = "free"
	TierPremium AccessTier = "premium"
)

type LearningCategory struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	LessonCount  int        `json:"lessonCount"`
	RequiredTier AccessTier `json:"requiredTier"`
}

type LearningTrack struct {
	ID           string     `json:"id"`
	System       string     `json:"system"`
	Level        string     `json:"level"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Progress     int        `json:"progress"`
	RequiredTier AccessTier `json:"requiredTier"`
}

type Exam struct {
	ID            string     `json:"id"`
	System        string     `json:"system"`
	Level         string     `json:"level"`
	Title         string     `json:"title"`
	QuestionCount int        `json:"questionCount"`
	DurationMin   int        `json:"durationMinutes"`
	RequiredTier  AccessTier `json:"requiredTier"`
}

type Plan struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	BillingType string   `json:"billingType"`
	PriceIDR    int      `json:"priceIdr"`
	Features    []string `json:"features"`
	Recommended bool     `json:"recommended"`
}

type Entitlement struct {
	Tier       AccessTier `json:"tier"`
	IsLifetime bool       `json:"isLifetime"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
}

type Profile struct {
	ID          string      `json:"id"`
	DisplayName string      `json:"displayName"`
	Email       string      `json:"email"`
	PhotoURL    string      `json:"photoUrl,omitempty"`
	Target      string      `json:"target"`
	Level       int         `json:"level"`
	XP          int         `json:"xp"`
	StreakDays  int         `json:"streakDays"`
	Entitlement Entitlement `json:"entitlement"`
}

type LessonProgress struct {
	LessonID     string    `json:"lessonId"`
	Completion   int       `json:"completion"`
	LastPosition int       `json:"lastPosition"`
	Completed    bool      `json:"completed"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Achievement struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Progress    int        `json:"progress"`
	Target      int        `json:"target"`
	UnlockedAt  *time.Time `json:"unlockedAt,omitempty"`
}

type ExamAttempt struct {
	ID        string    `json:"id"`
	ExamID    string    `json:"examId"`
	UserID    string    `json:"userId"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"startedAt"`
}

type DailyGoal struct {
	MinutesLearned int `json:"minutesLearned"`
	TargetMinutes  int `json:"targetMinutes"`
	StreakDays     int `json:"streakDays"`
}

type Dashboard struct {
	Banners         []HomeBanner       `json:"banners"`
	IsAuthenticated bool               `json:"isAuthenticated"`
	Profile         *Profile           `json:"profile,omitempty"`
	DailyGoal       DailyGoal          `json:"dailyGoal"`
	Categories      []LearningCategory `json:"categories"`
	Tracks          []LearningTrack    `json:"tracks"`
	Exams           []Exam             `json:"exams"`
	Plans           []Plan             `json:"plans"`
	Achievements    []Achievement      `json:"achievements"`
}
