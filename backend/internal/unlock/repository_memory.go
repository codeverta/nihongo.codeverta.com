package unlock

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MemoryRepository keeps everything in process memory. It is used by tests
// and as a fallback when no database is configured.
type MemoryRepository struct {
	mu      sync.Mutex
	intents map[string]*Intent
	txs     map[string]string // transaction ID -> nonce
	unlocks map[string]Unlock // userID|contentKey
	usage   map[string]*Usage // userID|day
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		intents: map[string]*Intent{},
		txs:     map[string]string{},
		unlocks: map[string]Unlock{},
		usage:   map[string]*Usage{},
	}
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

func (m *MemoryRepository) CreateIntent(_ context.Context, intent Intent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.intents[intent.Nonce]; exists {
		return ErrInvalid
	}
	copyOf := intent
	m.intents[intent.Nonce] = &copyOf
	return nil
}

func (m *MemoryRepository) Intent(_ context.Context, nonce string) (Intent, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	intent, ok := m.intents[nonce]
	if !ok {
		return Intent{}, false, nil
	}
	return *intent, true, nil
}

func (m *MemoryRepository) MarkVerified(_ context.Context, nonce, transactionID string, at time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if owner, used := m.txs[transactionID]; used {
		if owner == nonce {
			return true, nil
		}
		return false, ErrReplay
	}
	intent, ok := m.intents[nonce]
	if !ok || intent.VerifiedAt != nil || !at.Before(intent.ExpiresAt) {
		return false, nil
	}
	verified := at
	intent.VerifiedAt = &verified
	intent.TransactionID = transactionID
	m.txs[transactionID] = nonce
	return true, nil
}

func (m *MemoryRepository) Claim(_ context.Context, userID, nonce, day string, dailyCap int, at time.Time) (Unlock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	intent, ok := m.intents[nonce]
	if !ok || intent.UserID != userID {
		return Unlock{}, ErrNotFound
	}
	key := userID + "|" + intent.ContentKey
	if existing, done := m.unlocks[key]; done {
		consume(intent, at)
		return existing, nil
	}
	if intent.VerifiedAt == nil {
		return Unlock{}, ErrNotVerified
	}
	if !at.Before(intent.ExpiresAt) {
		return Unlock{}, ErrExpired
	}
	u := m.counters(userID, day)
	if u.Used >= dailyCap {
		return Unlock{}, ErrCapReached
	}
	u.Used++
	granted := Unlock{UserID: userID, ContentKey: intent.ContentKey, Source: SourceRewarded, CreatedAt: at}
	m.unlocks[key] = granted
	consume(intent, at)
	return granted, nil
}

func consume(intent *Intent, at time.Time) {
	if intent.ConsumedAt == nil {
		consumed := at
		intent.ConsumedAt = &consumed
	}
}

func (m *MemoryRepository) Fallback(_ context.Context, userID, contentKey, day string, dailyCap, fallbackCap int, at time.Time) (Unlock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, done := m.unlocks[userID+"|"+contentKey]; done {
		return existing, nil
	}
	u := m.counters(userID, day)
	if u.Used >= dailyCap {
		return Unlock{}, ErrCapReached
	}
	if u.FallbackUsed >= fallbackCap {
		return Unlock{}, ErrFallbackExhausted
	}
	u.Used++
	u.FallbackUsed++
	granted := Unlock{UserID: userID, ContentKey: contentKey, Source: SourceFallback, CreatedAt: at}
	m.unlocks[userID+"|"+contentKey] = granted
	return granted, nil
}

func (m *MemoryRepository) ConsumeVerified(_ context.Context, userID, nonce string, at time.Time) (Intent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	intent, ok := m.intents[nonce]
	if !ok || intent.UserID != userID {
		return Intent{}, ErrNotFound
	}
	if intent.ConsumedAt != nil {
		return *intent, ErrConsumed
	}
	if intent.VerifiedAt == nil {
		return Intent{}, ErrNotVerified
	}
	if !at.Before(intent.ExpiresAt) {
		return Intent{}, ErrExpired
	}
	consume(intent, at)
	return *intent, nil
}

func (m *MemoryRepository) Convert(_ context.Context, userID, fromKey, toKey string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	from, held := m.unlocks[userID+"|"+fromKey]
	if _, taken := m.unlocks[userID+"|"+toKey]; !held || taken {
		return false, nil
	}
	delete(m.unlocks, userID+"|"+fromKey)
	from.ContentKey = toKey
	m.unlocks[userID+"|"+toKey] = from
	return true, nil
}

func (m *MemoryRepository) Has(_ context.Context, userID, contentKey string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.unlocks[userID+"|"+contentKey]
	return ok, nil
}

func (m *MemoryRepository) List(_ context.Context, userID string) ([]Unlock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := []Unlock{}
	for _, u := range m.unlocks {
		if u.UserID == userID {
			items = append(items, u)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ContentKey < items[j].ContentKey })
	return items, nil
}

func (m *MemoryRepository) Usage(_ context.Context, userID, day string) (Usage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return *m.counters(userID, day), nil
}

func (m *MemoryRepository) Activity(_ context.Context, since time.Time) (Activity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var activity Activity
	for _, intent := range m.intents {
		if !intent.CreatedAt.Before(since) {
			activity.Intents = append(activity.Intents, IntentActivity{ContentKey: intent.ContentKey, CreatedAt: intent.CreatedAt, Verified: intent.VerifiedAt != nil, Consumed: intent.ConsumedAt != nil})
		}
	}
	for _, u := range m.unlocks {
		if !u.CreatedAt.Before(since) {
			activity.Unlocks = append(activity.Unlocks, u)
		}
	}
	return activity, nil
}

func (m *MemoryRepository) PurgeIntentsBefore(_ context.Context, cutoff time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var removed int64
	for nonce, intent := range m.intents {
		if intent.CreatedAt.Before(cutoff) {
			if intent.TransactionID != "" {
				delete(m.txs, intent.TransactionID)
			}
			delete(m.intents, nonce)
			removed++
		}
	}
	return removed, nil
}
