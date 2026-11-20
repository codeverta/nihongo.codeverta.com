package aichat

import (
	"context"
	"time"
)

// Slot is the bucket a message is paid from.
type Slot string

const (
	SlotFree    Slot = "free"
	SlotAd      Slot = "ad"
	SlotPremium Slot = "premium"
)

// Usage is one account's counters for one quota day.
type Usage struct {
	FreeUsed    int
	AdGranted   int
	AdUsed      int
	PremiumUsed int
}

func (u Usage) Total() int { return u.FreeUsed + u.AdUsed + u.PremiumUsed }

// Message is one stored chat turn. Conversations are keyed by account and
// lesson; the server alone writes them, so clients cannot forge history.
type Message struct {
	ID               string
	UserID           string
	LessonKey        string
	Role             string // "user" or "assistant"
	Chip             string
	Content          string
	Locale           string
	Cached           bool
	PromptTokens     int
	CacheHitTokens   int
	CompletionTokens int
	CreatedAt        time.Time
}

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Repository is the persistence boundary. Reserve/Release/GrantAdCredits must
// be atomic so concurrent requests cannot exceed a limit.
type Repository interface {
	Usage(ctx context.Context, userID, day string) (Usage, error)
	// Reserve consumes one message from slot when it is under limit. For
	// SlotAd the limit is the number of credits granted, and limit is ignored.
	Reserve(ctx context.Context, userID, day string, slot Slot, limit int) (bool, error)
	// Release gives back a reservation whose answer failed.
	Release(ctx context.Context, userID, day string, slot Slot) error
	// GrantAdCredits adds credits ad credits unless the day's total would
	// pass maxPerDay.
	GrantAdCredits(ctx context.Context, userID, day string, credits, maxPerDay int) (bool, error)

	AppendMessages(ctx context.Context, messages ...Message) error
	// RecentMessages returns the newest visible messages in chronological order.
	RecentMessages(ctx context.Context, userID, lessonKey string, limit int) ([]Message, error)
	// HideConversation removes a conversation from the user's view. Rows stay
	// for the audit retention window and are purged by PurgeBefore.
	HideConversation(ctx context.Context, userID, lessonKey string) error
	PurgeBefore(ctx context.Context, cutoff time.Time) (int64, error)

	CachedAnswer(ctx context.Context, key string) (string, bool, error)
	StoreAnswer(ctx context.Context, key, answer string) error
	ClearAnswers(ctx context.Context) (int64, error)
}
