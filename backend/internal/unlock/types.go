// Package unlock lets a free account open one locked item (a lesson, later
// also decks, news, jobs) by watching a rewarded ad, verified by AdMob
// server-side verification (SSV). The client never decides that an ad was
// watched: only a signed callback from Google marks an intent as verified.
package unlock

import (
	"context"
	"errors"
	"regexp"
	"time"
)

// Kind is the type of item that can be unlocked.
type Kind string

const (
	KindGrammar Kind = "grammar"
	KindKanji   Kind = "kanji"
	KindReading Kind = "reading"
	KindVocab   Kind = "vocabulary"
	// KindAnki opens one built-in JLPT deck, permanently.
	KindAnki Kind = "anki"
	// KindSimulation is a one-time credit for one simulation exam: starting
	// an attempt converts it into KindSimulationAttempt for that attempt.
	KindSimulation Kind = "simulation"
	// KindSimulationAttempt opens one simulation attempt. It is only ever
	// created by converting a KindSimulation credit, never requested directly.
	KindSimulationAttempt Kind = "simulation-attempt"
	// KindNews opens one news article, permanently.
	KindNews Kind = "news"
	// KindJob opens one job's application link, permanently.
	KindJob Kind = "job"
	// KindXPBoost doubles the XP of one finished activity. Its ID is the
	// activity's completion ID (for example "grammar:<attempt>"). It is only
	// ever verified by SSV and then consumed; it never becomes an unlock.
	KindXPBoost Kind = "xp-boost"
)

// clientKinds are the kinds an app may ask to unlock.
var clientKinds = map[Kind]bool{
	KindGrammar: true, KindKanji: true, KindReading: true, KindVocab: true,
	KindAnki: true, KindSimulation: true, KindNews: true, KindJob: true,
	KindXPBoost: true,
}

var kinds = func() map[Kind]bool {
	all := map[Kind]bool{KindSimulationAttempt: true}
	for kind := range clientKinds {
		all[kind] = true
	}
	return all
}()

var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// completionIDPattern allows the colons of completion IDs; content keys are
// at most 96 characters, so "xp-boost:" leaves room for 86.
var completionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_:-]{1,86}$`)

// ParseKind validates a client-supplied kind.
func ParseKind(value string) (Kind, bool) {
	kind := Kind(value)
	return kind, clientKinds[kind]
}

// IsLesson reports whether kind is a study lesson (grammar, kanji, reading,
// vocabulary).
func (k Kind) IsLesson() bool {
	return k == KindGrammar || k == KindKanji || k == KindReading || k == KindVocab
}

// Key builds the stable key stored for an unlocked item, for example
// "grammar:n3-lesson-01". It reports false for unknown kinds or malformed IDs.
func Key(kind Kind, id string) (string, bool) {
	pattern := idPattern
	if kind == KindXPBoost {
		pattern = completionIDPattern
	}
	if !kinds[kind] || !pattern.MatchString(id) {
		return "", false
	}
	return string(kind) + ":" + id, true
}

// Source records how an unlock was granted.
type Source string

const (
	SourceRewarded Source = "rewarded"
	SourceFallback Source = "fallback"
)

var (
	ErrInvalid           = errors.New("unlock: invalid request")
	ErrNotFound          = errors.New("unlock: intent not found")
	ErrNotVerified       = errors.New("unlock: ad not verified yet")
	ErrExpired           = errors.New("unlock: intent expired")
	ErrCapReached        = errors.New("unlock: daily unlock limit reached")
	ErrFallbackExhausted = errors.New("unlock: daily fallback limit reached")
	ErrAlreadyUnlocked   = errors.New("unlock: already unlocked")
	ErrReplay            = errors.New("unlock: ad transaction already used")
	ErrBadSignature      = errors.New("unlock: invalid ssv signature")
	ErrUnavailable       = errors.New("unlock: ssv verification is not configured")
	ErrConsumed          = errors.New("unlock: intent already used")
)

// Intent is a one-time request to unlock ContentKey for UserID. It becomes
// verified when Google's SSV callback arrives carrying its Nonce.
type Intent struct {
	Nonce         string
	UserID        string
	ContentKey    string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	VerifiedAt    *time.Time
	ConsumedAt    *time.Time
	TransactionID string
}

// Unlock is one item an account may open without Lifetime Access.
type Unlock struct {
	UserID     string
	ContentKey string
	Source     Source
	CreatedAt  time.Time
}

// Usage is one account's counters for one quota day.
type Usage struct {
	Used         int // all unlocks granted today
	FallbackUsed int // the subset granted without a verified ad
}

// Repository is the persistence boundary. Claim, Fallback and MarkVerified
// must be atomic so concurrent requests cannot exceed a limit or replay an ad.
type Repository interface {
	CreateIntent(ctx context.Context, intent Intent) error
	Intent(ctx context.Context, nonce string) (Intent, bool, error)
	// MarkVerified records the ad against an unverified, unexpired intent. It
	// returns false when the intent is unknown, expired, or already verified
	// by a different transaction; ErrReplay when the transaction ID belongs to
	// another intent. Repeating the same (nonce, transaction) is a success.
	MarkVerified(ctx context.Context, nonce, transactionID string, at time.Time) (bool, error)
	// Claim consumes a verified intent owned by userID and records the unlock
	// under the daily cap. Claiming twice returns the same unlock.
	Claim(ctx context.Context, userID, nonce, day string, dailyCap int, at time.Time) (Unlock, error)
	// Fallback records an unlock without an ad, bounded by both limits.
	Fallback(ctx context.Context, userID, contentKey, day string, dailyCap, fallbackCap int, at time.Time) (Unlock, error)
	// Convert turns the unlock fromKey into toKey, consuming a one-time
	// credit. It reports false when fromKey is not held or toKey already is.
	Convert(ctx context.Context, userID, fromKey, toKey string) (bool, error)
	// ConsumeVerified marks a verified, unexpired intent owned by userID as
	// used and returns it, without recording an unlock or counting toward the
	// unlock limits (double XP has its own). A second call returns
	// ErrConsumed with the intent.
	ConsumeVerified(ctx context.Context, userID, nonce string, at time.Time) (Intent, error)
	Has(ctx context.Context, userID, contentKey string) (bool, error)
	List(ctx context.Context, userID string) ([]Unlock, error)
	Usage(ctx context.Context, userID, day string) (Usage, error)
	PurgeIntentsBefore(ctx context.Context, cutoff time.Time) (int64, error)
	// Activity returns intents and unlocks created at or after since, for
	// statistics.
	Activity(ctx context.Context, since time.Time) (Activity, error)
}
