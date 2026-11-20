package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const DemoUserID = "a5090412-9ab4-5852-8229-c7d19eb5cb1c"

func UserUUID(id string) string {
	if parsed, err := uuid.Parse(id); err == nil {
		return parsed.String()
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("nihongo-master:user:"+id)).String()
}
func setUUID(id *string) error {
	if *id == "" {
		*id = uuid.NewString()
	}
	return uuid.Validate(*id)
}
func OptionalUserID(id string) *string {
	if id == "" {
		return nil
	}
	return &id
}

type GamificationUser struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	UserID string `gorm:"column:user_id;type:char(36);not null;uniqueIndex:gamification_user_id"`
	State  []byte `gorm:"type:longblob;not null"`
	// XPTotal and ReachedAt mirror the blob's XP so all-time ranking reads an
	// index instead of the JSON state.
	XPTotal   int        `gorm:"column:xp_total;not null;default:0;index:gamification_xp_rank,priority:1,sort:desc"`
	ReachedAt *time.Time `gorm:"column:reached_at;type:datetime(3);index:gamification_xp_rank,priority:2"`
}

func (GamificationUser) TableName() string { return "gamification_users" }
func (m *GamificationUser) BeforeCreate(_ *gorm.DB) error {
	if err := uuid.Validate(m.UserID); err != nil {
		return err
	}
	return setUUID(&m.ID)
}

type Account struct {
	ID    string `gorm:"type:char(36);primaryKey"`
	Email string `gorm:"type:varchar(320);not null;uniqueIndex:accounts_email"`
	State []byte `gorm:"type:longblob;not null"`
}

func (Account) TableName() string                { return "accounts" }
func (m *Account) BeforeCreate(_ *gorm.DB) error { return setUUID(&m.ID) }

// Session keeps API login sessions across server restarts. Only the SHA-256
// hash of the bearer token is stored, never the token itself.
type Session struct {
	TokenHash string    `gorm:"column:token_hash;type:char(64);primaryKey"`
	UserID    string    `gorm:"column:user_id;type:char(36);not null;index:sessions_user_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (Session) TableName() string { return "sessions" }

type BillingPurchase struct {
	ID          string  `gorm:"type:char(36);primaryKey"`
	PurchaseKey string  `gorm:"column:purchase_key;type:varbinary(2048);not null;uniqueIndex:billing_purchase_key"`
	UserID      *string `gorm:"column:user_id;type:char(36);index:billing_purchase_user_id"`
	State       []byte  `gorm:"type:longblob;not null"`
}

func (BillingPurchase) TableName() string { return "billing_purchases" }
func (m *BillingPurchase) BeforeCreate(_ *gorm.DB) error {
	if m.UserID != nil {
		if err := uuid.Validate(*m.UserID); err != nil {
			return err
		}
	}
	return setUUID(&m.ID)
}

type BillingToken struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	UserID string `gorm:"column:user_id;type:char(36);not null;uniqueIndex:billing_token_user_id"`
	Token  string `gorm:"type:char(36);not null;uniqueIndex:token"`
}

func (BillingToken) TableName() string { return "billing_tokens" }
func (m *BillingToken) BeforeCreate(_ *gorm.DB) error {
	if err := uuid.Validate(m.UserID); err != nil {
		return err
	}
	return setUUID(&m.ID)
}

type DeletedDemo struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	UserID string `gorm:"column:user_id;type:char(36);not null;uniqueIndex:deleted_demo_user_id"`
}

func (DeletedDemo) TableName() string { return "deleted_demo" }
func (m *DeletedDemo) BeforeCreate(_ *gorm.DB) error {
	if err := uuid.Validate(m.UserID); err != nil {
		return err
	}
	return setUUID(&m.ID)
}

type DictionaryResponse struct {
	Endpoint  string `gorm:"type:text"`
	QueryText string `gorm:"type:varchar(512)"`
	Mode      string `gorm:"type:varchar(16)"`
	Page      int
	JLPT      string `gorm:"type:varchar(2)"`
	Common    bool
	ID        string `gorm:"type:char(36);primaryKey"`
	CacheKey  string `gorm:"column:cache_key;type:char(64);not null;uniqueIndex:dictionary_cache_key"`
	Source    string `gorm:"type:varchar(64);not null"`
	Body      []byte `gorm:"type:longblob;not null"`
	FetchedAt int64  `gorm:"column:fetched_at;type:bigint;not null;index:dictionary_responses_age"`
}

func (DictionaryResponse) TableName() string                { return "dictionary_responses" }
func (m *DictionaryResponse) BeforeCreate(_ *gorm.DB) error { return setUUID(&m.ID) }

type DictionaryEntry struct {
	ID        string `gorm:"type:char(36);primaryKey"`
	Source    string `gorm:"type:varchar(64);not null;uniqueIndex:dictionary_entry_source,priority:1"`
	EntryID   string `gorm:"column:source_entry_id;type:varchar(512);not null;uniqueIndex:dictionary_entry_source,priority:2"`
	Kind      string `gorm:"type:varchar(32);not null"`
	Body      []byte `gorm:"type:longblob;not null"`
	FetchedAt int64  `gorm:"column:fetched_at;type:bigint;not null"`
}

func (DictionaryEntry) TableName() string                { return "dictionary_entries" }
func (m *DictionaryEntry) BeforeCreate(_ *gorm.DB) error { return setUUID(&m.ID) }

// XPEvent is one XP award. Raw events are kept for 12 weeks; the leaderboard
// never sums them to rank (it reads weekly_scores and the xp_total column).
type XPEvent struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UserID    string    `gorm:"column:user_id;type:char(36);not null;index:xp_events_user,priority:1"`
	Amount    int       `gorm:"not null"`
	Source    string    `gorm:"type:varchar(32);not null"`
	AwardedAt time.Time `gorm:"column:awarded_at;type:datetime(3);not null;index:xp_events_user,priority:2;index:xp_events_awarded_at"`
}

func (XPEvent) TableName() string { return "xp_events" }

// WeeklyScore is a user's XP for one leaderboard week (Monday 00:00 WIB).
type WeeklyScore struct {
	WeekStart string    `gorm:"column:week_start;type:char(10);primaryKey;index:weekly_scores_rank,priority:1"`
	UserID    string    `gorm:"column:user_id;type:char(36);primaryKey"`
	XP        int       `gorm:"column:xp;not null;index:weekly_scores_rank,priority:2,sort:desc"`
	ReachedAt time.Time `gorm:"column:reached_at;type:datetime(3);not null;index:weekly_scores_rank,priority:3"`
}

func (WeeklyScore) TableName() string { return "weekly_scores" }

// WeeklyResult is the final, permanent rank of a user in a closed week.
type WeeklyResult struct {
	WeekStart string `gorm:"column:week_start;type:char(10);primaryKey"`
	UserID    string `gorm:"column:user_id;type:char(36);primaryKey;index:weekly_results_user"`
	Rank      int    `gorm:"column:rank_position;not null"`
	XP        int    `gorm:"column:xp;not null"`
}

func (WeeklyResult) TableName() string { return "weekly_results" }

// LeaderboardFlag stores the user's own "hide me" setting and the admin's
// silent exclusion. Missing rows mean shown.
type LeaderboardFlag struct {
	UserID   string `gorm:"column:user_id;type:char(36);primaryKey"`
	Hidden   bool   `gorm:"not null;default:false"`
	Excluded bool   `gorm:"not null;default:false"`
}

func (LeaderboardFlag) TableName() string { return "leaderboard_flags" }

// LeaderboardJob is the weekly close schedule. The lease makes one instance
// run a close even if several API processes share the database.
type LeaderboardJob struct {
	ID             string     `gorm:"type:varchar(32);primaryKey"`
	NextRunAt      time.Time  `gorm:"column:next_run_at;not null"`
	LastClosedWeek string     `gorm:"column:last_closed_week;type:char(10);not null;default:''"`
	LeaseOwner     string     `gorm:"column:lease_owner;type:varchar(36);not null;default:''"`
	LeaseUntil     *time.Time `gorm:"column:lease_until"`
	LastError      string     `gorm:"column:last_error;type:text"`
}

func (LeaderboardJob) TableName() string { return "leaderboard_jobs" }
