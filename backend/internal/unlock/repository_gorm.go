package unlock

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"nihongo-master-backend/internal/database"
)

type intentRow struct {
	Nonce         string     `gorm:"column:nonce;type:char(32);primaryKey"`
	UserID        string     `gorm:"column:user_id;type:char(36);not null;index:unlock_intents_user"`
	ContentKey    string     `gorm:"column:content_key;type:varchar(96);not null"`
	CreatedAt     time.Time  `gorm:"column:created_at;type:datetime(6);not null;index:unlock_intents_created"`
	ExpiresAt     time.Time  `gorm:"column:expires_at;type:datetime(6);not null"`
	VerifiedAt    *time.Time `gorm:"column:verified_at;type:datetime(6)"`
	ConsumedAt    *time.Time `gorm:"column:consumed_at;type:datetime(6)"`
	TransactionID *string    `gorm:"column:transaction_id;type:varchar(128);uniqueIndex:unlock_intents_tx"`
}

func (intentRow) TableName() string { return "unlock_intents" }

type unlockRow struct {
	UserID     string    `gorm:"column:user_id;type:char(36);primaryKey"`
	ContentKey string    `gorm:"column:content_key;type:varchar(96);primaryKey"`
	Source     string    `gorm:"column:source;type:varchar(16);not null"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime(6);not null"`
}

func (unlockRow) TableName() string { return "content_unlocks" }

type usageRow struct {
	UserID       string `gorm:"column:user_id;type:char(36);primaryKey"`
	Day          string `gorm:"column:day;type:char(10);primaryKey"`
	Used         int    `gorm:"column:used;not null;default:0"`
	FallbackUsed int    `gorm:"column:fallback_used;not null;default:0"`
}

func (usageRow) TableName() string { return "unlock_usage" }

// GormRepository stores unlock data in MariaDB.
type GormRepository struct{ db *gorm.DB }

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if err := database.Migrate(db, &intentRow{}, &unlockRow{}, &usageRow{}); err != nil {
		return nil, err
	}
	return &GormRepository{db: db}, nil
}

func isDuplicate(err error) bool {
	var driverErr *mysql.MySQLError
	return errors.As(err, &driverErr) && driverErr.Number == 1062
}

func (r *GormRepository) CreateIntent(ctx context.Context, intent Intent) error {
	row := intentRow{
		Nonce: intent.Nonce, UserID: intent.UserID, ContentKey: intent.ContentKey,
		CreatedAt: intent.CreatedAt.UTC(), ExpiresAt: intent.ExpiresAt.UTC(),
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormRepository) Intent(ctx context.Context, nonce string) (Intent, bool, error) {
	var row intentRow
	err := r.db.WithContext(ctx).Where("nonce = ?", nonce).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Intent{}, false, nil
	}
	if err != nil {
		return Intent{}, false, err
	}
	return row.intent(), true, nil
}

func (row intentRow) intent() Intent {
	intent := Intent{
		Nonce: row.Nonce, UserID: row.UserID, ContentKey: row.ContentKey,
		CreatedAt: row.CreatedAt, ExpiresAt: row.ExpiresAt, VerifiedAt: row.VerifiedAt, ConsumedAt: row.ConsumedAt,
	}
	if row.TransactionID != nil {
		intent.TransactionID = *row.TransactionID
	}
	return intent
}

// MarkVerified uses one conditional UPDATE, and the unique index on the
// transaction ID, so a signed callback can only be applied once.
func (r *GormRepository) MarkVerified(ctx context.Context, nonce, transactionID string, at time.Time) (bool, error) {
	at = at.UTC()
	result := r.db.WithContext(ctx).Model(&intentRow{}).
		Where("nonce = ? AND verified_at IS NULL AND expires_at > ?", nonce, at).
		Updates(map[string]any{"verified_at": at, "transaction_id": transactionID})
	if isDuplicate(result.Error) {
		if r.sameTransaction(ctx, nonce, transactionID) {
			return true, nil
		}
		return false, ErrReplay
	}
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 1 {
		return true, nil
	}
	// Google retries callbacks; the same (nonce, transaction) is a success.
	return r.sameTransaction(ctx, nonce, transactionID), nil
}

func (r *GormRepository) sameTransaction(ctx context.Context, nonce, transactionID string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&intentRow{}).
		Where("nonce = ? AND transaction_id = ?", nonce, transactionID).Count(&count)
	return count == 1
}

// ensureUsage creates the day's counter row. It runs before a transaction
// starts: two transactions inserting the same row and then updating it
// deadlock, which is how concurrent claims failed under load.
func (r *GormRepository) ensureUsage(ctx context.Context, userID, day string) error {
	return r.db.WithContext(ctx).Clauses(clause.Insert{Modifier: "IGNORE"}).
		Create(&usageRow{UserID: userID, Day: day}).Error
}

func isRetryable(err error) bool {
	var driverErr *mysql.MySQLError
	// 1213 deadlock victim, 1205 lock wait timeout: MariaDB asks to retry.
	return errors.As(err, &driverErr) && (driverErr.Number == 1213 || driverErr.Number == 1205)
}

// errRaceLost means another transaction recorded the same unlock first. The
// transaction is rolled back (so nothing is counted twice) and the caller
// returns the unlock that won.
var errRaceLost = errors.New("unlock: recorded concurrently")

// Locking rules for every write transaction, which is what makes deadlocks
// impossible rather than merely retried:
//
//  1. The account's counter row for the day is locked first, by primary key.
//     Every write for one account therefore runs one at a time, and they all
//     take their locks in the same order.
//  2. READ COMMITTED isolation: plain reads see rows other transactions have
//     committed, and MariaDB takes no gap locks, so different accounts never
//     block each other.
//  3. A duplicate unlock row (a race across the WIB midnight, when two days'
//     counters are involved) rolls back and returns the existing unlock.
//
// The retry below remains only as a last resort for lock wait timeouts.
var writeTx = &sql.TxOptions{Isolation: sql.LevelReadCommitted}

// maxTxAttempts bounds the last-resort retry. Tests set it to 1 to prove the
// locking rules alone prevent deadlocks.
var maxTxAttempts = 5

// transact runs fn in a transaction and retries it if MariaDB still asks for
// a retry. fn must therefore be safe to run more than once.
func (r *GormRepository) transact(ctx context.Context, fn func(tx *gorm.DB) error) error {
	var err error
	for attempt := 0; attempt < maxTxAttempts; attempt++ {
		if err = r.db.WithContext(ctx).Transaction(fn, writeTx); !isRetryable(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 15 * time.Millisecond):
		}
	}
	return err
}

// lockUsage locks the account's counter row for the day. It must be the first
// statement of every write transaction (see the locking rules above).
func lockUsage(tx *gorm.DB, userID, day string) (usageRow, error) {
	var usage usageRow
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND day = ?", userID, day).Take(&usage).Error
	return usage, err
}

func (r *GormRepository) existing(tx *gorm.DB, userID, contentKey string) (Unlock, bool, error) {
	var row unlockRow
	err := tx.Where("user_id = ? AND content_key = ?", userID, contentKey).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Unlock{}, false, nil
	}
	if err != nil {
		return Unlock{}, false, err
	}
	return row.unlock(), true, nil
}

func (row unlockRow) unlock() Unlock {
	return Unlock{UserID: row.UserID, ContentKey: row.ContentKey, Source: Source(row.Source), CreatedAt: row.CreatedAt}
}

func (r *GormRepository) Claim(ctx context.Context, userID, nonce, day string, dailyCap int, at time.Time) (Unlock, error) {
	at = at.UTC()
	if err := r.ensureUsage(ctx, userID, day); err != nil {
		return Unlock{}, err
	}
	var granted Unlock
	var contentKey string
	err := r.transact(ctx, func(tx *gorm.DB) error {
		usage, err := lockUsage(tx, userID, day)
		if err != nil {
			return err
		}
		var row intentRow
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("nonce = ? AND user_id = ?", nonce, userID).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		contentKey = row.ContentKey
		if existing, done, err := r.existing(tx, userID, row.ContentKey); err != nil {
			return err
		} else if done {
			granted = existing
			return consumeIntent(tx, row, at)
		}
		if row.VerifiedAt == nil {
			return ErrNotVerified
		}
		if !at.Before(row.ExpiresAt) {
			return ErrExpired
		}
		if usage.Used >= dailyCap {
			return ErrCapReached
		}
		if err := tx.Model(&usageRow{}).Where("user_id = ? AND day = ?", userID, day).
			UpdateColumn("used", gorm.Expr("used + 1")).Error; err != nil {
			return err
		}
		unlockRecord := unlockRow{UserID: userID, ContentKey: row.ContentKey, Source: string(SourceRewarded), CreatedAt: at}
		if err := tx.Create(&unlockRecord).Error; err != nil {
			if isDuplicate(err) {
				return errRaceLost
			}
			return err
		}
		granted = unlockRecord.unlock()
		return consumeIntent(tx, row, at)
	})
	if errors.Is(err, errRaceLost) {
		return r.afterRaceLost(ctx, userID, contentKey, nonce, at)
	}
	if err != nil {
		return Unlock{}, err
	}
	return granted, nil
}

func consumeIntent(tx *gorm.DB, row intentRow, at time.Time) error {
	if row.ConsumedAt != nil {
		return nil
	}
	return tx.Model(&intentRow{}).Where("nonce = ?", row.Nonce).UpdateColumn("consumed_at", at).Error
}

// afterRaceLost returns the unlock another transaction recorded first. nonce
// is the claimed intent, marked consumed so it cannot be claimed again; it is
// empty for the fallback path.
func (r *GormRepository) afterRaceLost(ctx context.Context, userID, contentKey, nonce string, at time.Time) (Unlock, error) {
	existing, done, err := r.existing(r.db.WithContext(ctx), userID, contentKey)
	if err != nil {
		return Unlock{}, err
	}
	if !done {
		return Unlock{}, errRaceLost
	}
	if nonce != "" {
		if err := r.db.WithContext(ctx).Model(&intentRow{}).
			Where("nonce = ? AND consumed_at IS NULL", nonce).UpdateColumn("consumed_at", at).Error; err != nil {
			return Unlock{}, err
		}
	}
	return existing, nil
}

func (r *GormRepository) Fallback(ctx context.Context, userID, contentKey, day string, dailyCap, fallbackCap int, at time.Time) (Unlock, error) {
	at = at.UTC()
	if err := r.ensureUsage(ctx, userID, day); err != nil {
		return Unlock{}, err
	}
	var granted Unlock
	err := r.transact(ctx, func(tx *gorm.DB) error {
		usage, err := lockUsage(tx, userID, day)
		if err != nil {
			return err
		}
		if existing, done, err := r.existing(tx, userID, contentKey); err != nil {
			return err
		} else if done {
			granted = existing
			return nil
		}
		if usage.Used >= dailyCap {
			return ErrCapReached
		}
		if usage.FallbackUsed >= fallbackCap {
			return ErrFallbackExhausted
		}
		if err := tx.Model(&usageRow{}).Where("user_id = ? AND day = ?", userID, day).
			UpdateColumns(map[string]any{"used": gorm.Expr("used + 1"), "fallback_used": gorm.Expr("fallback_used + 1")}).Error; err != nil {
			return err
		}
		unlockRecord := unlockRow{UserID: userID, ContentKey: contentKey, Source: string(SourceFallback), CreatedAt: at}
		if err := tx.Create(&unlockRecord).Error; err != nil {
			if isDuplicate(err) {
				return errRaceLost
			}
			return err
		}
		granted = unlockRecord.unlock()
		return nil
	})
	if errors.Is(err, errRaceLost) {
		return r.afterRaceLost(ctx, userID, contentKey, "", at)
	}
	if err != nil {
		return Unlock{}, err
	}
	return granted, nil
}

// Convert is one conditional UPDATE: of two concurrent conversions of the
// same credit, exactly one changes a row. A taken toKey hits the primary key.
func (r *GormRepository) ConsumeVerified(ctx context.Context, userID, nonce string, at time.Time) (Intent, error) {
	at = at.UTC()
	var consumed Intent
	err := r.transact(ctx, func(tx *gorm.DB) error {
		var row intentRow
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("nonce = ? AND user_id = ?", nonce, userID).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		consumed = row.intent()
		if row.ConsumedAt != nil {
			return ErrConsumed
		}
		if row.VerifiedAt == nil {
			return ErrNotVerified
		}
		if !at.Before(row.ExpiresAt) {
			return ErrExpired
		}
		if err := consumeIntent(tx, row, at); err != nil {
			return err
		}
		consumed.ConsumedAt = &at
		return nil
	})
	if errors.Is(err, ErrConsumed) {
		return consumed, ErrConsumed
	}
	if err != nil {
		return Intent{}, err
	}
	return consumed, nil
}

func (r *GormRepository) Convert(ctx context.Context, userID, fromKey, toKey string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&unlockRow{}).
		Where("user_id = ? AND content_key = ?", userID, fromKey).
		UpdateColumn("content_key", toKey)
	if isDuplicate(result.Error) {
		return false, nil
	}
	return result.RowsAffected == 1, result.Error
}

func (r *GormRepository) Has(ctx context.Context, userID, contentKey string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&unlockRow{}).
		Where("user_id = ? AND content_key = ?", userID, contentKey).Count(&count).Error
	return count > 0, err
}

func (r *GormRepository) List(ctx context.Context, userID string) ([]Unlock, error) {
	var rows []unlockRow
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("content_key").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]Unlock, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.unlock())
	}
	return items, nil
}

func (r *GormRepository) Usage(ctx context.Context, userID, day string) (Usage, error) {
	var row usageRow
	err := r.db.WithContext(ctx).Where("user_id = ? AND day = ?", userID, day).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Usage{}, nil
	}
	if err != nil {
		return Usage{}, err
	}
	return Usage{Used: row.Used, FallbackUsed: row.FallbackUsed}, nil
}

// statsRowLimit bounds one statistics query; a month of activity is far below.
const statsRowLimit = 500000

func (r *GormRepository) Activity(ctx context.Context, since time.Time) (Activity, error) {
	since = since.UTC()
	var intents []struct {
		ContentKey string
		CreatedAt  time.Time
		VerifiedAt *time.Time
		ConsumedAt *time.Time
	}
	if err := r.db.WithContext(ctx).Model(&intentRow{}).Select("content_key, created_at, verified_at, consumed_at").
		Where("created_at >= ?", since).Limit(statsRowLimit).Find(&intents).Error; err != nil {
		return Activity{}, err
	}
	var unlocks []unlockRow
	if err := r.db.WithContext(ctx).Where("created_at >= ?", since).Limit(statsRowLimit).Find(&unlocks).Error; err != nil {
		return Activity{}, err
	}
	activity := Activity{Intents: make([]IntentActivity, 0, len(intents)), Unlocks: make([]Unlock, 0, len(unlocks))}
	for _, intent := range intents {
		activity.Intents = append(activity.Intents, IntentActivity{ContentKey: intent.ContentKey, CreatedAt: intent.CreatedAt, Verified: intent.VerifiedAt != nil, Consumed: intent.ConsumedAt != nil})
	}
	for _, row := range unlocks {
		activity.Unlocks = append(activity.Unlocks, row.unlock())
	}
	return activity, nil
}

func (r *GormRepository) PurgeIntentsBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff.UTC()).Delete(&intentRow{})
	return result.RowsAffected, result.Error
}
