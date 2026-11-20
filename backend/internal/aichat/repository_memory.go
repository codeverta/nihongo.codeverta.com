package aichat

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryRepository keeps everything in process memory. It is used by tests
// and as a fallback when no database is configured.
type MemoryRepository struct {
	mu       sync.Mutex
	usage    map[string]*Usage
	messages []memoryMessage
	answers  map[string]string
}

type memoryMessage struct {
	Message
	hidden bool
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{usage: map[string]*Usage{}, answers: map[string]string{}}
}

func (m *MemoryRepository) counters(userID, day string) *Usage {
	key := userID + "|" + day
	u, ok := m.usage[key]
	if !ok {
		u = &Usage{}
		m.usage[key] = u
	}
	return u
}

func (m *MemoryRepository) Usage(_ context.Context, userID, day string) (Usage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return *m.counters(userID, day), nil
}

func (m *MemoryRepository) Reserve(_ context.Context, userID, day string, slot Slot, limit int) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u := m.counters(userID, day)
	switch slot {
	case SlotFree:
		if u.FreeUsed >= limit {
			return false, nil
		}
		u.FreeUsed++
	case SlotAd:
		if u.AdUsed >= u.AdGranted {
			return false, nil
		}
		u.AdUsed++
	case SlotPremium:
		if u.PremiumUsed >= limit {
			return false, nil
		}
		u.PremiumUsed++
	default:
		return false, nil
	}
	return true, nil
}

func (m *MemoryRepository) Release(_ context.Context, userID, day string, slot Slot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u := m.counters(userID, day)
	switch slot {
	case SlotFree:
		u.FreeUsed = max(0, u.FreeUsed-1)
	case SlotAd:
		u.AdUsed = max(0, u.AdUsed-1)
	case SlotPremium:
		u.PremiumUsed = max(0, u.PremiumUsed-1)
	}
	return nil
}

func (m *MemoryRepository) GrantAdCredits(_ context.Context, userID, day string, credits, maxPerDay int) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u := m.counters(userID, day)
	if credits < 1 || u.AdGranted+credits > maxPerDay {
		return false, nil
	}
	u.AdGranted += credits
	return true, nil
}

func (m *MemoryRepository) AppendMessages(_ context.Context, messages ...Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, msg := range messages {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
		if msg.CreatedAt.IsZero() {
			msg.CreatedAt = time.Now().UTC()
		}
		m.messages = append(m.messages, memoryMessage{Message: msg})
	}
	return nil
}

func (m *MemoryRepository) RecentMessages(_ context.Context, userID, lessonKey string, limit int) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Message
	for i := len(m.messages) - 1; i >= 0 && len(out) < limit; i-- {
		msg := m.messages[i]
		if !msg.hidden && msg.UserID == userID && msg.LessonKey == lessonKey {
			out = append(out, msg.Message)
		}
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (m *MemoryRepository) HideConversation(_ context.Context, userID, lessonKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.messages {
		if m.messages[i].UserID == userID && m.messages[i].LessonKey == lessonKey {
			m.messages[i].hidden = true
		}
	}
	return nil
}

func (m *MemoryRepository) PurgeBefore(_ context.Context, cutoff time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.messages[:0]
	var removed int64
	for _, msg := range m.messages {
		if msg.CreatedAt.Before(cutoff) {
			removed++
			continue
		}
		kept = append(kept, msg)
	}
	m.messages = kept
	return removed, nil
}

func (m *MemoryRepository) CachedAnswer(_ context.Context, key string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	answer, ok := m.answers[key]
	return answer, ok, nil
}

func (m *MemoryRepository) StoreAnswer(_ context.Context, key, answer string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.answers[key] = answer
	return nil
}

func (m *MemoryRepository) ClearAnswers(_ context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := int64(len(m.answers))
	m.answers = map[string]string{}
	return n, nil
}
