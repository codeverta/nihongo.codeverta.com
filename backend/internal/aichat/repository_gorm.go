package aichat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"nihongo-master-backend/internal/database"
)

type usageRow struct {
	UserID      string `gorm:"column:user_id;type:char(36);primaryKey"`
	Day         string `gorm:"column:day;type:char(10);primaryKey"`
	FreeUsed    int    `gorm:"column:free_used;not null;default:0"`
	AdGranted   int    `gorm:"column:ad_granted;not null;default:0"`
	AdUsed      int    `gorm:"column:ad_used;not null;default:0"`
	PremiumUsed int    `gorm:"column:premium_used;not null;default:0"`
}

func (usageRow) TableName() string { return "ai_usage" }

type messageRow struct {
	ID               string     `gorm:"column:id;type:char(36);primaryKey"`
	UserID           string     `gorm:"column:user_id;type:char(36);not null;index:ai_messages_conversation,priority:1"`
	LessonKey        string     `gorm:"column:lesson_key;type:varchar(96);not null;index:ai_messages_conversation,priority:2"`
	Role             string     `gorm:"column:role;type:varchar(16);not null"`
	Chip             string     `gorm:"column:chip;type:varchar(32);not null;default:''"`
	Content          string     `gorm:"column:content;type:mediumtext;not null"`
	Locale           string     `gorm:"column:locale;type:varchar(8);not null;default:''"`
	Cached           bool       `gorm:"column:cached;not null;default:false"`
	PromptTokens     int        `gorm:"column:prompt_tokens;not null;default:0"`
	CacheHitTokens   int        `gorm:"column:cache_hit_tokens;not null;default:0"`
	CompletionTokens int        `gorm:"column:completion_tokens;not null;default:0"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:datetime(6);not null;index:ai_messages_conversation,priority:3;index:ai_messages_created"`
	HiddenAt         *time.Time `gorm:"column:hidden_at"`
}

func (messageRow) TableName() string { return "ai_messages" }

type answerRow struct {
	CacheKey  string    `gorm:"column:cache_key;type:char(64);primaryKey"`
	Answer    string    `gorm:"column:answer;type:mediumtext;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (answerRow) TableName() string { return "ai_answer_cache" }

// GormRepository stores AI chat data in MariaDB.
type GormRepository struct{ db *gorm.DB }

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if err := database.Migrate(db, &usageRow{}, &messageRow{}, &answerRow{}); err != nil {
		return nil, err
	}
	return &GormRepository{db: db}, nil
}

func (r *GormRepository) ensureUsage(ctx context.Context, userID, day string) error {
	return r.db.WithContext(ctx).Clauses(clause.Insert{Modifier: "IGNORE"}).
		Create(&usageRow{UserID: userID, Day: day}).Error
}

func (r *GormRepository) Usage(ctx context.Context, userID, day string) (Usage, error) {
	var row usageRow
	err := r.db.WithContext(ctx).Where("user_id = ? AND day = ?", userID, day).Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return Usage{}, nil
	}
	if err != nil {
		return Usage{}, err
	}
	return Usage{FreeUsed: row.FreeUsed, AdGranted: row.AdGranted, AdUsed: row.AdUsed, PremiumUsed: row.PremiumUsed}, nil
}

// Reserve uses a single conditional UPDATE so the limit check and the
// increment are atomic across concurrent requests and API replicas.
func (r *GormRepository) Reserve(ctx context.Context, userID, day string, slot Slot, limit int) (bool, error) {
	if err := r.ensureUsage(ctx, userID, day); err != nil {
		return false, err
	}
	q := r.db.WithContext(ctx).Model(&usageRow{}).Where("user_id = ? AND day = ?", userID, day)
	var result *gorm.DB
	switch slot {
	case SlotFree:
		result = q.Where("free_used < ?", limit).UpdateColumn("free_used", gorm.Expr("free_used + 1"))
	case SlotAd:
		result = q.Where("ad_used < ad_granted").UpdateColumn("ad_used", gorm.Expr("ad_used + 1"))
	case SlotPremium:
		result = q.Where("premium_used < ?", limit).UpdateColumn("premium_used", gorm.Expr("premium_used + 1"))
	default:
		return false, nil
	}
	return result.RowsAffected == 1, result.Error
}

func (r *GormRepository) Release(ctx context.Context, userID, day string, slot Slot) error {
	column := map[Slot]string{SlotFree: "free_used", SlotAd: "ad_used", SlotPremium: "premium_used"}[slot]
	if column == "" {
		return nil
	}
	return r.db.WithContext(ctx).Model(&usageRow{}).
		Where("user_id = ? AND day = ? AND "+column+" > 0", userID, day).
		UpdateColumn(column, gorm.Expr(column+" - 1")).Error
}

func (r *GormRepository) GrantAdCredits(ctx context.Context, userID, day string, credits, maxPerDay int) (bool, error) {
	if credits < 1 {
		return false, nil
	}
	if err := r.ensureUsage(ctx, userID, day); err != nil {
		return false, err
	}
	result := r.db.WithContext(ctx).Model(&usageRow{}).
		Where("user_id = ? AND day = ? AND ad_granted + ? <= ?", userID, day, credits, maxPerDay).
		UpdateColumn("ad_granted", gorm.Expr("ad_granted + ?", credits))
	return result.RowsAffected == 1, result.Error
}

func (r *GormRepository) AppendMessages(ctx context.Context, messages ...Message) error {
	if len(messages) == 0 {
		return nil
	}
	rows := make([]messageRow, len(messages))
	now := time.Now().UTC()
	for i, m := range messages {
		if m.ID == "" {
			m.ID = uuid.NewString()
		}
		if m.CreatedAt.IsZero() {
			// Keep user/assistant pairs ordered even within one transaction.
			m.CreatedAt = now.Add(time.Duration(i) * time.Microsecond)
		}
		rows[i] = messageRow{
			ID: m.ID, UserID: m.UserID, LessonKey: m.LessonKey, Role: m.Role, Chip: m.Chip,
			Content: m.Content, Locale: m.Locale, Cached: m.Cached, PromptTokens: m.PromptTokens,
			CacheHitTokens: m.CacheHitTokens, CompletionTokens: m.CompletionTokens, CreatedAt: m.CreatedAt.UTC(),
		}
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *GormRepository) RecentMessages(ctx context.Context, userID, lessonKey string, limit int) ([]Message, error) {
	var rows []messageRow
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND lesson_key = ? AND hidden_at IS NULL", userID, lessonKey).
		Order("created_at DESC").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Message, len(rows))
	for i, row := range rows {
		out[len(rows)-1-i] = Message{
			ID: row.ID, UserID: row.UserID, LessonKey: row.LessonKey, Role: row.Role, Chip: row.Chip,
			Content: row.Content, Locale: row.Locale, Cached: row.Cached, PromptTokens: row.PromptTokens,
			CacheHitTokens: row.CacheHitTokens, CompletionTokens: row.CompletionTokens, CreatedAt: row.CreatedAt,
		}
	}
	return out, nil
}

func (r *GormRepository) HideConversation(ctx context.Context, userID, lessonKey string) error {
	return r.db.WithContext(ctx).Model(&messageRow{}).
		Where("user_id = ? AND lesson_key = ? AND hidden_at IS NULL", userID, lessonKey).
		UpdateColumn("hidden_at", time.Now().UTC()).Error
}

func (r *GormRepository) PurgeBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff.UTC()).Delete(&messageRow{})
	if result.Error != nil {
		return 0, result.Error
	}
	// Usage rows are tiny but still worth pruning on the same schedule.
	if err := r.db.WithContext(ctx).Where("day < ?", QuotaDay(cutoff)).Delete(&usageRow{}).Error; err != nil {
		return result.RowsAffected, err
	}
	return result.RowsAffected, nil
}

func (r *GormRepository) CachedAnswer(ctx context.Context, key string) (string, bool, error) {
	var row answerRow
	err := r.db.WithContext(ctx).Where("cache_key = ?", key).Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return row.Answer, true, nil
}

func (r *GormRepository) StoreAnswer(ctx context.Context, key, answer string) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&answerRow{CacheKey: key, Answer: answer, CreatedAt: time.Now().UTC()}).Error
}

func (r *GormRepository) ClearAnswers(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("1 = 1").Delete(&answerRow{})
	return result.RowsAffected, result.Error
}
