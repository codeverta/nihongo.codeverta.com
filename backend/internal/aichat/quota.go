package aichat

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrAdRequired: free quota used; a rewarded ad can unlock one message.
	ErrAdRequired = errors.New("ai: rewarded ad required")
	// ErrDailyLimit: nothing more today (ads unavailable or ad cap reached).
	ErrDailyLimit = errors.New("ai: daily limit reached")
	// ErrPremiumLimit: a Lifetime account reached its daily limit.
	ErrPremiumLimit = errors.New("ai: premium daily limit reached")
	// ErrAdCapReached: no further ad credits can be granted today.
	ErrAdCapReached = errors.New("ai: ad credit cap reached")
)

var jakarta = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*60*60)
}()

// QuotaDay is the quota bucket for t. Quotas reset at 00:00 WIB for everyone.
func QuotaDay(t time.Time) string { return t.In(jakarta).Format("2006-01-02") }

// Reservation is a message slot held while an answer is generated.
type Reservation struct {
	UserID string
	Day    string
	Slot   Slot
}

// Quota decides which bucket pays for a message.
type Quota struct {
	Repo   Repository
	Config func() Config
}

// Reserve holds one message for userID. adsAvailable reports whether the
// rewarded placement is switched on; when it is off a free account that has
// used its quota is told the day is over instead of being offered an ad.
func (q Quota) Reserve(ctx context.Context, userID string, premium, adsAvailable bool, now time.Time) (Reservation, error) {
	cfg := q.Config()
	day := QuotaDay(now)
	if premium {
		ok, err := q.Repo.Reserve(ctx, userID, day, SlotPremium, cfg.PremiumPerDay)
		if err != nil {
			return Reservation{}, err
		}
		if !ok {
			return Reservation{}, ErrPremiumLimit
		}
		return Reservation{UserID: userID, Day: day, Slot: SlotPremium}, nil
	}
	for _, slot := range []Slot{SlotFree, SlotAd} {
		ok, err := q.Repo.Reserve(ctx, userID, day, slot, cfg.FreePerDay)
		if err != nil {
			return Reservation{}, err
		}
		if ok {
			return Reservation{UserID: userID, Day: day, Slot: slot}, nil
		}
	}
	usage, err := q.Repo.Usage(ctx, userID, day)
	if err != nil {
		return Reservation{}, err
	}
	if adsAvailable && usage.AdGranted < cfg.AdPerDay {
		return Reservation{}, ErrAdRequired
	}
	return Reservation{}, ErrDailyLimit
}

// Release returns a reservation after a failed answer, so failures never
// cost the user a message or an ad credit.
func (q Quota) Release(ctx context.Context, r Reservation) error {
	return q.Repo.Release(ctx, r.UserID, r.Day, r.Slot)
}

// GrantAdCredit records one watched (or unavailable) rewarded ad.
func (q Quota) GrantAdCredit(ctx context.Context, userID string, now time.Time) error {
	ok, err := q.Repo.GrantAdCredit(ctx, userID, QuotaDay(now), q.Config().AdPerDay)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAdCapReached
	}
	return nil
}
