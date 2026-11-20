package unlock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// Limits are the per-account daily limits, shared by every kind of item.
type Limits struct {
	DailyCap       int // unlocks per account per day
	FallbackPerDay int // of those, how many may be granted without a verified ad
}

// IntentTTL is how long an intent can be verified and claimed.
const IntentTTL = 10 * time.Minute

// IntentRetention is how long used and expired intents are kept for audit.
const IntentRetention = 7 * 24 * time.Hour

// Service holds the unlock rules. All inputs that decide access (the ad
// callback, the daily limits) are checked here, never trusted from the client.
type Service struct {
	Repo   Repository
	Limits func() Limits
	Verify *Verifier
	// Units returns the ad units whose callbacks are accepted (the unlock
	// unit and the double XP unit). Empty accepts any unit.
	Units func() []string
	Now   func() time.Time
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Service) limits() Limits { return s.Limits() }

var jakarta = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*60*60)
}()

// Day is the quota bucket for t. Limits reset at 00:00 WIB for everyone.
func Day(t time.Time) string { return t.In(jakarta).Format("2006-01-02") }

// CreateIntent starts an unlock. It refuses when the item is already open or
// the account has no unlocks left today, so no ad is shown for nothing.
func (s *Service) CreateIntent(ctx context.Context, userID, contentKey string) (Intent, error) {
	open, err := s.Repo.Has(ctx, userID, contentKey)
	if err != nil {
		return Intent{}, err
	}
	if open {
		return Intent{}, ErrAlreadyUnlocked
	}
	now := s.now()
	// Double XP has its own daily limit, checked by the caller.
	if !strings.HasPrefix(contentKey, string(KindXPBoost)+":") {
		usage, err := s.Repo.Usage(ctx, userID, Day(now))
		if err != nil {
			return Intent{}, err
		}
		if usage.Used >= s.limits().DailyCap {
			return Intent{}, ErrCapReached
		}
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return Intent{}, err
	}
	intent := Intent{
		Nonce: hex.EncodeToString(raw), UserID: userID, ContentKey: contentKey,
		CreatedAt: now, ExpiresAt: now.Add(IntentTTL),
	}
	if err := s.Repo.CreateIntent(ctx, intent); err != nil {
		return Intent{}, err
	}
	return intent, nil
}

// HandleCallback applies one SSV request from Google. ErrBadSignature means
// the request was not signed by Google; any other error means a genuine
// callback that could not be applied (unknown nonce, wrong unit, replay).
func (s *Service) HandleCallback(ctx context.Context, rawQuery string) error {
	callback, err := s.Verify.Parse(ctx, rawQuery)
	if err != nil {
		return err
	}
	if !s.acceptsUnit(callback.AdUnit) {
		return ErrInvalid
	}
	if callback.CustomData == "" || callback.TransactionID == "" {
		return ErrInvalid
	}
	intent, ok, err := s.Repo.Intent(ctx, callback.CustomData)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	if callback.UserID != intent.UserID {
		return ErrInvalid
	}
	verified, err := s.Repo.MarkVerified(ctx, intent.Nonce, callback.TransactionID, s.now())
	if err != nil {
		return err
	}
	if !verified {
		return ErrExpired
	}
	return nil
}

// acceptsUnit reports whether a callback for adUnit (the numeric part after
// "/") belongs to one of the configured units.
func (s *Service) acceptsUnit(adUnit string) bool {
	if s.Units == nil {
		return true
	}
	configured := false
	for _, unit := range s.Units() {
		unit = strings.TrimSpace(unit)
		if slash := strings.LastIndex(unit, "/"); slash >= 0 {
			unit = unit[slash+1:]
		}
		if unit == "" {
			continue
		}
		configured = true
		if unit == adUnit {
			return true
		}
	}
	return !configured
}

// Claim opens the item once its intent has been verified by Google. It
// returns ErrNotVerified while the callback has not arrived yet.
func (s *Service) Claim(ctx context.Context, userID, nonce string) (Unlock, error) {
	now := s.now()
	return s.Repo.Claim(ctx, userID, nonce, Day(now), s.limits().DailyCap, now)
}

// Fallback opens an item without a verified ad, for when no ad could be
// loaded. It is the only path without SSV, so its daily count is small.
func (s *Service) Fallback(ctx context.Context, userID, contentKey string) (Unlock, error) {
	now := s.now()
	limits := s.limits()
	return s.Repo.Fallback(ctx, userID, contentKey, Day(now), limits.DailyCap, limits.FallbackPerDay, now)
}

// Convert consumes the one-time credit fromKey and grants toKey instead.
func (s *Service) Convert(ctx context.Context, userID, fromKey, toKey string) (bool, error) {
	return s.Repo.Convert(ctx, userID, fromKey, toKey)
}

// ConsumeVerified uses up a verified intent without recording an unlock.
func (s *Service) ConsumeVerified(ctx context.Context, userID, nonce string) (Intent, error) {
	return s.Repo.ConsumeVerified(ctx, userID, nonce, s.now())
}

// Allowed reports whether userID has unlocked contentKey.
func (s *Service) Allowed(ctx context.Context, userID, contentKey string) bool {
	open, err := s.Repo.Has(ctx, userID, contentKey)
	return err == nil && open
}

// Remaining reports how many unlocks the account has left today.
func (s *Service) Remaining(ctx context.Context, userID string) (remaining, fallbackRemaining int, err error) {
	usage, err := s.Repo.Usage(ctx, userID, Day(s.now()))
	if err != nil {
		return 0, 0, err
	}
	limits := s.limits()
	remaining = limits.DailyCap - usage.Used
	fallbackRemaining = limits.FallbackPerDay - usage.FallbackUsed
	if remaining < 0 {
		remaining = 0
	}
	if fallbackRemaining < 0 {
		fallbackRemaining = 0
	}
	if fallbackRemaining > remaining {
		fallbackRemaining = remaining
	}
	return remaining, fallbackRemaining, nil
}
