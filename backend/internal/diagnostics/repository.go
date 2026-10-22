package diagnostics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/database"
)

// Event is an aggregated diagnostic: repeated occurrences update the same row
// while retaining the first and most recent times for an error fingerprint.
type Event struct {
	ID          string    `json:"id" gorm:"type:char(36);primaryKey"`
	Fingerprint string    `json:"-" gorm:"type:char(64);uniqueIndex"`
	Source      string    `json:"source" gorm:"size:24;index"`
	Severity    string    `json:"severity" gorm:"size:16;index"`
	ErrorType   string    `json:"errorType" gorm:"size:160;index"`
	Message     string    `json:"message" gorm:"type:text"`
	StackTrace  string    `json:"stackTrace" gorm:"type:longtext"`
	Platform    string    `json:"platform" gorm:"size:32;index"`
	AppVersion  string    `json:"appVersion" gorm:"size:48"`
	Route       string    `json:"route" gorm:"size:512;index"`
	Method      string    `json:"method" gorm:"size:12"`
	StatusCode  int       `json:"statusCode" gorm:"index"`
	Occurrences int64     `json:"occurrences"`
	FirstSeenAt time.Time `json:"firstSeenAt" gorm:"index"`
	LastSeenAt  time.Time `json:"lastSeenAt" gorm:"index"`
}

func (Event) TableName() string { return "error_events" }

type Repository struct{ db *gorm.DB }

func New(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if err := database.Migrate(db, &Event{}); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Record(ctx context.Context, event Event) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	now := time.Now().UTC()
	event.ID = uuid.NewString()
	event.FirstSeenAt = now
	event.LastSeenAt = now
	event.Occurrences = 1
	event.Source = limited(event.Source, 24)
	event.Severity = limited(event.Severity, 16)
	event.ErrorType = limited(event.ErrorType, 160)
	event.Platform = limited(event.Platform, 32)
	event.AppVersion = limited(event.AppVersion, 48)
	event.Route = limited(event.Route, 512)
	event.Method = limited(event.Method, 12)
	event.Message = limited(event.Message, 2000)
	event.StackTrace = limited(event.StackTrace, 12000)

	identity := strings.Join([]string{
		event.Source,
		event.Severity,
		event.ErrorType,
		event.Platform,
		event.Route,
		event.Method,
		event.Message,
	}, "\x00")
	sum := sha256.Sum256([]byte(identity))
	event.Fingerprint = hex.EncodeToString(sum[:])

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "fingerprint"}},
		DoUpdates: clause.Assignments(map[string]any{
			"occurrences":  gorm.Expr("occurrences + 1"),
			"last_seen_at": event.LastSeenAt,
			"stack_trace":  event.StackTrace,
			"status_code":  event.StatusCode,
		}),
	}).Create(&event).Error
}

func (r *Repository) List(ctx context.Context, page int, query, source string) ([]Event, int64, error) {
	db := r.db.WithContext(ctx).Model(&Event{})
	if query != "" {
		pattern := "%" + query + "%"
		db = db.Where("message LIKE ? OR error_type LIKE ? OR route LIKE ?", pattern, pattern, pattern)
	}
	if source != "" {
		db = db.Where("source = ?", source)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var events []Event
	err := db.Order("last_seen_at DESC, id").Limit(50).Offset((page - 1) * 50).Find(&events).Error
	if events == nil {
		events = []Event{}
	}
	return events, total, err
}

func (r *Repository) Stats(ctx context.Context) (map[string]int64, error) {
	var stats struct {
		ErrorGroups int64
		Occurrences int64
		Flutter     int64
		Backend     int64
	}
	err := r.db.WithContext(ctx).Model(&Event{}).Select(
		"COUNT(*) AS error_groups, COALESCE(SUM(occurrences), 0) AS occurrences, " +
			"COALESCE(SUM(CASE WHEN source = 'flutter' THEN occurrences ELSE 0 END), 0) AS flutter, " +
			"COALESCE(SUM(CASE WHEN source = 'backend' THEN occurrences ELSE 0 END), 0) AS backend",
	).Scan(&stats).Error
	return map[string]int64{"errors": stats.ErrorGroups, "occurrences": stats.Occurrences, "flutter": stats.Flutter, "backend": stats.Backend}, err
}

func limited(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
