package store

import (
	"encoding/json"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"sort"
	"time"
)

// LearningActivity tracks navigation, independently of quiz scores and rewards.
type LearningActivity struct {
	Kind       string    `json:"kind"`
	LessonID   string    `json:"lessonId"`
	Title      string    `json:"title"`
	Level      string    `json:"level"`
	Position   int       `json:"position"`
	Total      int       `json:"total"`
	Completed  bool      `json:"completed"`
	Locked     bool      `json:"locked"`
	Available  bool      `json:"available"`
	AccessedAt time.Time `json:"accessedAt"`
}

func (s *MemoryStore) learningMetadataLocked(user, kind, id string) (LearningActivity, bool) {
	item := LearningActivity{Kind: kind, LessonID: id, Available: true}
	switch kind {
	case "vocabulary":
		level, lesson := s.findVocabularyLessonLocked(id)
		if lesson == nil {
			return item, false
		}
		item.Title, item.Level, item.Total = lesson.Title, level.Level, len(lesson.Words)
	case "grammar":
		for _, lesson := range s.grammar {
			if lesson.ID == id {
				item.Title, item.Level, item.Total = lesson.Title, lesson.Level, len(lesson.Points)
				break
			}
		}
	case "kanji":
		for _, lesson := range s.kanji {
			if lesson.ID == id {
				item.Title, item.Level, item.Total = lesson.Title, lesson.Level, len(lesson.Characters)
				break
			}
		}
	case "reading":
		for _, lesson := range s.reading {
			if lesson.ID == id {
				item.Title, item.Level, item.Total = lesson.Title, lesson.Level, len(lesson.Pages)
				break
			}
		}
	default:
		return item, false
	}
	if item.Total == 0 {
		return item, false
	}
	item.Completed = s.progress[user][id].Completed
	item.Locked = PremiumLevel(item.Level) && s.profiles[user].Entitlement.Tier != domain.TierPremium
	return item, true
}

func (s *MemoryStore) LearningHistory(user string) []LearningActivity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := []LearningActivity{}
	if state := s.gamification[user]; state != nil {
		for _, saved := range state.LearningActivity {
			item, exists := s.learningMetadataLocked(user, saved.Kind, saved.LessonID)
			if !exists {
				item = saved
				item.Available = false
			}
			item.AccessedAt, item.Position = saved.AccessedAt, saved.Position
			if item.Total > 0 && item.Position >= item.Total {
				item.Position = item.Total - 1
			}
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].AccessedAt.Equal(items[j].AccessedAt) {
			return items[i].Kind+items[i].LessonID < items[j].Kind+items[j].LessonID
		}
		return items[i].AccessedAt.After(items[j].AccessedAt)
	})
	return items
}

func (s *MemoryStore) SaveLearningActivity(user, kind, id string, position int, accessedAt time.Time) (LearningActivity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, exists := s.learningMetadataLocked(user, kind, id)
	if !exists {
		return item, ErrNotFound
	}
	if item.Locked {
		return item, ErrPremium
	}
	now := time.Now().UTC()
	if position < 0 || position >= item.Total || accessedAt.IsZero() || accessedAt.After(now.Add(5*time.Minute)) {
		return item, ErrInvalid
	}
	if accessedAt.After(now) {
		accessedAt = now
	}
	base := s.gamification[user]
	if base == nil {
		base = newGamificationState()
	}
	key := kind + ":" + id
	if old, ok := base.LearningActivity[key]; ok && !accessedAt.After(old.AccessedAt) {
		return old, nil
	}
	// Copy before writing so a failed persistence operation leaves memory unchanged.
	raw, err := json.Marshal(base)
	if err != nil {
		return item, err
	}
	next := newGamificationState()
	if err = json.Unmarshal(raw, next); err != nil {
		return item, err
	}
	if next.LearningActivity == nil {
		next.LearningActivity = map[string]LearningActivity{}
	}
	item.Position, item.AccessedAt = position, accessedAt.UTC()
	next.LearningActivity[key] = item
	if s.gamificationDB != nil {
		raw, err = json.Marshal(next)
		if err != nil {
			return item, err
		}
		err = s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state"})}).Create(&database.GamificationUser{UserID: user, State: raw}).Error
		if err != nil {
			return item, err
		}
	}
	s.gamification[user] = next
	return item, nil
}
