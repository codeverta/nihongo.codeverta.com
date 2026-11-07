package store

import (
	"sort"
	"sync"
	"time"
)

// lastSeenTracker records the last authenticated API request per user. It is
// kept in memory (reset on restart); persisted learning activity and study
// days cover users who have not called the API since the last restart.
type lastSeenTracker struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

var userLastSeen = &lastSeenTracker{seen: map[string]time.Time{}}

// TouchUser marks userID as active. Updates are throttled to once a minute.
func (s *MemoryStore) TouchUser(userID string, at time.Time) {
	if userID == "" {
		return
	}
	userLastSeen.mu.Lock()
	defer userLastSeen.mu.Unlock()
	if previous, ok := userLastSeen.seen[userID]; ok && at.Sub(previous) < time.Minute {
		return
	}
	userLastSeen.seen[userID] = at
}

func lastSeen(userID string) (time.Time, bool) {
	userLastSeen.mu.Lock()
	defer userLastSeen.mu.Unlock()
	at, ok := userLastSeen.seen[userID]
	return at, ok
}

// AdminUser is one row of the admin dashboard's user list.
type AdminUser struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	DisplayName      string     `json:"displayName"`
	PhotoURL         string     `json:"photoUrl,omitempty"`
	Provider         string     `json:"provider"`
	Plan             string     `json:"plan"`
	Demo             bool       `json:"demo"`
	XP               int        `json:"xp"`
	Level            int        `json:"level"`
	StreakDays       int        `json:"streakDays"`
	LessonsCompleted int        `json:"lessonsCompleted"`
	ExamsCompleted   int        `json:"examsCompleted"`
	ActiveDays       int        `json:"activeDays"`
	Timezone         string     `json:"timezone"`
	LastActiveAt     *time.Time `json:"lastActiveAt"`
	LastActivity     string     `json:"lastActivity"`
}

// AdminUsers lists every known account with progress and last activity,
// most recently active first.
func (s *MemoryStore) AdminUsers() []AdminUser {
	type base struct {
		id, email, name, photo, provider string
	}
	s.mu.RLock()
	accounts := make([]base, 0, len(s.accounts))
	for _, a := range s.accounts {
		provider := "email"
		if a.GoogleSub != "" {
			provider = "google"
		}
		accounts = append(accounts, base{a.ID, a.Email, a.DisplayName, a.profilePhotoURL(), provider})
	}
	s.mu.RUnlock()

	result := make([]AdminUser, 0, len(accounts))
	for _, a := range accounts {
		summary := s.Gamification(a.id)
		row := AdminUser{
			ID: a.id, Email: a.email, DisplayName: a.name, PhotoURL: a.photo,
			Provider: a.provider, Plan: "free", Demo: s.IsDemoUser(a.id),
			XP: summary.XP, Level: summary.Level, StreakDays: summary.StreakDays,
			LessonsCompleted: summary.LessonsCompleted, ExamsCompleted: summary.ExamsCompleted,
			ActiveDays: summary.ActiveDays, Timezone: summary.Timezone,
		}
		var latest time.Time
		s.mu.RLock()
		if a.id == s.playReviewUserID && a.id != "" {
			row.Plan = "review"
		}
		for _, p := range s.purchases {
			if p.UserID == a.id && p.State == "purchased" {
				row.Plan = "lifetime"
				break
			}
		}
		if state := s.gamification[a.id]; state != nil {
			for _, activity := range state.LearningActivity {
				if activity.AccessedAt.After(latest) {
					latest = activity.AccessedAt
					row.LastActivity = activityLabel(activity)
				}
			}
			if latest.IsZero() {
				// Study days have no time of day; use them only when nothing
				// more precise is known.
				for day := range state.Days {
					if parsed, err := time.Parse("2006-01-02", day); err == nil && parsed.After(latest) {
						latest = parsed
						row.LastActivity = "Belajar"
					}
				}
			}
		}
		s.mu.RUnlock()
		if seen, ok := lastSeen(a.id); ok && seen.After(latest) {
			latest = seen
			if row.LastActivity == "" {
				row.LastActivity = "Membuka aplikasi"
			}
		}
		if !latest.IsZero() {
			at := latest.UTC()
			row.LastActiveAt = &at
		}
		result = append(result, row)
	}
	sort.SliceStable(result, func(i, j int) bool {
		a, b := result[i].LastActiveAt, result[j].LastActiveAt
		switch {
		case a == nil && b == nil:
			return result[i].Email < result[j].Email
		case a == nil:
			return false
		case b == nil:
			return true
		default:
			return a.After(*b)
		}
	})
	return result
}

func activityLabel(a LearningActivity) string {
	kind := map[string]string{
		"vocabulary": "Vocabulary", "grammar": "Grammar", "kanji": "Kanji",
		"reading": "Reading", "anki": "Anki", "ssw": "SSW",
	}[a.Kind]
	if kind == "" {
		kind = a.Kind
	}
	label := kind
	if a.Level != "" {
		label += " " + a.Level
	}
	if a.Title != "" {
		label += " · " + a.Title
	}
	return label
}
