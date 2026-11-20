package unlock

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultKeysURL is where Google publishes the AdMob SSV verifier keys.
const DefaultKeysURL = "https://www.gstatic.com/admob/reward/verifier-keys.json"

// KeyProvider returns the public key Google signed a callback with.
type KeyProvider interface {
	PublicKey(ctx context.Context, keyID string) (*ecdsa.PublicKey, error)
}

// Callback is the verified content of one SSV request from Google.
type Callback struct {
	AdUnit        string // numeric ad unit ID, the part after "/" in the full ID
	TransactionID string
	UserID        string
	CustomData    string
	RewardItem    string
	Timestamp     time.Time
}

// Verifier checks the signature of SSV callbacks.
type Verifier struct{ Keys KeyProvider }

const (
	signatureParam = "&signature="
	keyIDParam     = "&key_id="
)

// Parse verifies rawQuery (the request's raw query string) and returns its
// fields. Google signs everything before "&signature=", exactly as sent, so
// the query must not be decoded or reordered before checking it.
func (v *Verifier) Parse(ctx context.Context, rawQuery string) (Callback, error) {
	if v == nil || v.Keys == nil {
		return Callback{}, ErrUnavailable
	}
	sigAt := strings.Index(rawQuery, signatureParam)
	if sigAt < 0 {
		return Callback{}, ErrBadSignature
	}
	signed := rawQuery[:sigAt]
	rest := rawQuery[sigAt+len(signatureParam):]
	keyAt := strings.Index(rest, keyIDParam)
	if keyAt < 0 {
		return Callback{}, ErrBadSignature
	}
	encodedSig, keyID := rest[:keyAt], rest[keyAt+len(keyIDParam):]
	if keyID == "" || strings.Contains(keyID, "&") {
		return Callback{}, ErrBadSignature
	}
	signature, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(encodedSig, "="))
	if err != nil {
		return Callback{}, ErrBadSignature
	}
	key, err := v.Keys.PublicKey(ctx, keyID)
	if err != nil {
		return Callback{}, fmt.Errorf("%w: %v", ErrBadSignature, err)
	}
	digest := sha256.Sum256([]byte(signed))
	if !ecdsa.VerifyASN1(key, digest[:], signature) {
		return Callback{}, ErrBadSignature
	}
	values, err := url.ParseQuery(signed)
	if err != nil {
		return Callback{}, ErrBadSignature
	}
	callback := Callback{
		AdUnit:        values.Get("ad_unit"),
		TransactionID: values.Get("transaction_id"),
		UserID:        values.Get("user_id"),
		CustomData:    values.Get("custom_data"),
		RewardItem:    values.Get("reward_item"),
	}
	if millis, err := strconv.ParseInt(values.Get("timestamp"), 10, 64); err == nil {
		callback.Timestamp = time.UnixMilli(millis)
	}
	return callback, nil
}

// GoogleKeys downloads and caches Google's verifier keys.
type GoogleKeys struct {
	URL    string
	Client *http.Client
	// Now is replaceable in tests.
	Now func() time.Time

	mu          sync.Mutex
	keys        map[string]*ecdsa.PublicKey
	fetchedAt   time.Time
	attemptedAt time.Time
}

const (
	keysTTL         = 24 * time.Hour
	keysMinInterval = time.Minute
)

func NewGoogleKeys() *GoogleKeys {
	return &GoogleKeys{URL: DefaultKeysURL, Client: &http.Client{Timeout: 10 * time.Second}}
}

func (g *GoogleKeys) now() time.Time {
	if g.Now != nil {
		return g.Now()
	}
	return time.Now()
}

func (g *GoogleKeys) PublicKey(ctx context.Context, keyID string) (*ecdsa.PublicKey, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	if key, ok := g.keys[keyID]; ok && now.Sub(g.fetchedAt) < keysTTL {
		return key, nil
	}
	// An unknown key ID may mean Google rotated keys, so refresh, but never
	// more than once a minute so bad callbacks cannot hammer the endpoint.
	if now.Sub(g.attemptedAt) >= keysMinInterval {
		g.attemptedAt = now
		keys, err := g.fetch(ctx)
		if err != nil {
			return nil, err
		}
		g.keys, g.fetchedAt = keys, now
	}
	key, ok := g.keys[keyID]
	if !ok {
		return nil, errors.New("unknown key id")
	}
	return key, nil
}

func (g *GoogleKeys) fetch(ctx context.Context) (map[string]*ecdsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, g.URL, nil)
	if err != nil {
		return nil, err
	}
	client := g.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("verifier keys: status %d", response.StatusCode)
	}
	var payload struct {
		Keys []struct {
			KeyID json.Number `json:"keyId"`
			PEM   string      `json:"pem"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, err
	}
	keys := map[string]*ecdsa.PublicKey{}
	for _, entry := range payload.Keys {
		block, _ := pem.Decode([]byte(entry.PEM))
		if block == nil {
			continue
		}
		parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			continue
		}
		if key, ok := parsed.(*ecdsa.PublicKey); ok {
			keys[entry.KeyID.String()] = key
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("verifier keys: none usable")
	}
	return keys, nil
}
