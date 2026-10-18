// Package dictionary provides bounded, cached access to free dictionary sources.
package dictionary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	ErrInvalid     = errors.New("invalid dictionary query")
	ErrUnavailable = errors.New("dictionary source unavailable")
	ErrNotFound    = errors.New("kanji not found")
)

type Words struct {
	Query   string            `json:"query"`
	Page    int               `json:"page"`
	HasMore bool              `json:"hasMore"`
	Entries []json.RawMessage `json:"entries"`
	Source  string            `json:"source"`
}
type cached struct {
	body []byte
	at   time.Time
}
type flight struct {
	done    chan struct{}
	body    []byte
	err     error
	refresh bool
}
type Client struct {
	HTTP          *http.Client
	Persistence   *SQLCache
	mu            sync.Mutex
	cache         map[string]cached
	flights       map[string]*flight
	slots         chan struct{}
	next          time.Time
	memoryHits    atomic.Uint64
	databaseHits  atomic.Uint64
	upstreamCalls atomic.Uint64
}

func NewClient() *Client {
	return &Client{HTTP: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 3 || req.URL.Host != via[0].URL.Host || req.URL.Scheme != "https" {
			return ErrUnavailable
		}
		return nil
	}}, cache: map[string]cached{}, flights: map[string]*flight{}, slots: make(chan struct{}, 2)}
}
func (c *Client) Search(ctx context.Context, query string, page int, jlpt string, common bool, modes ...string) (Words, error) {
	query = strings.TrimSpace(query)
	jlpt = strings.ToUpper(strings.TrimSpace(jlpt))
	if query == "" || !utf8.ValidString(query) || utf8.RuneCountInString(query) > 100 || page < 1 || page > 100 {
		return Words{}, ErrInvalid
	}
	for _, r := range query {
		if unicode.IsControl(r) {
			return Words{}, ErrInvalid
		}
	}
	if jlpt != "" && (len(jlpt) != 2 || jlpt[0] != 'N' || jlpt[1] < '1' || jlpt[1] > '5') {
		return Words{}, ErrInvalid
	}
	mode := "auto"
	if len(modes) > 0 {
		mode = modes[0]
	}
	endpoint, err := searchEndpoint(query, page, jlpt, common, mode)
	if err != nil {
		return Words{}, err
	}
	body, err := c.get(ctx, endpoint)
	if err != nil {
		return Words{}, err
	}
	var payload struct {
		Meta struct {
			Status int `json:"status"`
		} `json:"meta"`
		Data []json.RawMessage `json:"data"`
	}
	if json.Unmarshal(body, &payload) != nil || payload.Meta.Status != 200 || payload.Data == nil {
		c.evict(endpoint)
		return Words{}, ErrUnavailable
	}
	return Words{Query: query, Page: page, HasMore: len(payload.Data) >= 20 && page < 100, Entries: payload.Data, Source: "Jisho.org"}, nil
}
func (c *Client) Kanji(ctx context.Context, character string) (json.RawMessage, error) {
	runes := []rune(character)
	if len(runes) != 1 || !unicode.Is(unicode.Han, runes[0]) {
		return nil, ErrInvalid
	}
	endpoint := "https://kanjiapi.dev/v1/kanji/" + url.PathEscape(character)
	body, err := c.get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	var data struct {
		Kanji   string `json:"kanji"`
		Strokes int    `json:"stroke_count"`
	}
	if json.Unmarshal(body, &data) != nil || data.Kanji != character || data.Strokes < 1 {
		c.evict(endpoint)
		return nil, ErrUnavailable
	}
	return json.RawMessage(body), nil
}
func (c *Client) evict(key string) { c.mu.Lock(); delete(c.cache, key); c.mu.Unlock() }

const maxMemoryBytes = 16 << 20

// Stats contains process-local counters; it never exposes account information.
func (c *Client) Stats() map[string]uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	var size uint64
	for _, hit := range c.cache {
		size += uint64(len(hit.body))
	}
	return map[string]uint64{"memoryHits": c.memoryHits.Load(), "databaseHits": c.databaseHits.Load(), "upstreamCalls": c.upstreamCalls.Load(), "memoryBytes": size, "memoryLimitBytes": maxMemoryBytes}
}

func (c *Client) get(ctx context.Context, endpoint string) ([]byte, error) {
	return c.getRequest(ctx, endpoint, false)
}

func (c *Client) getRequest(ctx context.Context, endpoint string, refresh bool) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	if hit, ok := c.cache[endpoint]; !refresh && ok && (c.Persistence != nil || time.Since(hit.at) < 6*time.Hour) {
		c.mu.Unlock()
		c.memoryHits.Add(1)
		return hit.body, nil
	}
	if f, ok := c.flights[endpoint]; ok {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-f.done:
		}
		// A DB read already in progress is not an admin refresh. Wait for it before
		// refreshing, so an older result cannot overwrite the refresh afterward.
		if refresh && !f.refresh {
			return c.getRequest(ctx, endpoint, true)
		}
		return f.body, f.err
	}
	f := &flight{done: make(chan struct{}), refresh: refresh}
	c.flights[endpoint] = f
	c.mu.Unlock()
	// The caller may leave while others still need the result. Bound the shared
	// work separately rather than inheriting the first caller's cancellation.
	go func() {
		work, cancel := context.WithTimeout(context.WithoutCancel(ctx), 12*time.Second)
		defer cancel()
		hit, err := c.loadOrFetch(work, endpoint, refresh)
		c.mu.Lock()
		if err == nil {
			delete(c.cache, endpoint)
			size := len(hit.body)
			for _, h := range c.cache {
				size += len(h.body)
			}
			for len(c.cache) > 0 && (len(c.cache) >= 128 || size > maxMemoryBytes) {
				var oldest string
				var at time.Time
				for key, value := range c.cache {
					if oldest == "" || value.at.Before(at) {
						oldest = key
						at = value.at
					}
				}
				size -= len(c.cache[oldest].body)
				delete(c.cache, oldest)
			}
			if len(hit.body) <= maxMemoryBytes {
				c.cache[endpoint] = hit
			}
		}
		f.body, f.err = hit.body, err
		delete(c.flights, endpoint)
		close(f.done)
		c.mu.Unlock()
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.done:
		return f.body, f.err
	}
}

func (c *Client) loadOrFetch(ctx context.Context, endpoint string, refresh bool) (cached, error) {
	if c.Persistence != nil && !refresh {
		hit, ok, err := c.Persistence.load(ctx, endpoint)
		if err != nil {
			return cached{}, fmt.Errorf("%w: database read failed", ErrUnavailable)
		}
		// Existing cache keys can be reused only when the exact upstream query is
		// unchanged. Never invent aliases from arbitrary archived result readings.
		if !ok && strings.Contains(endpoint, "#") {
			legacy, _ := url.Parse(endpoint)
			legacy.Fragment = ""
			hit, ok, err = c.Persistence.load(ctx, legacy.String())
			if err != nil {
				return cached{}, fmt.Errorf("%w: database read failed", ErrUnavailable)
			}
			if ok {
				if err = c.Persistence.save(ctx, endpoint, hit); err != nil {
					return cached{}, fmt.Errorf("%w: database migration failed", ErrUnavailable)
				}
			}
		}
		if ok {
			if _, _, _, err = responseEntries(endpoint, hit.body); err == nil {
				c.databaseHits.Add(1)
				return hit, nil
			}
		}
	}
	body, err := c.fetch(ctx, endpoint)
	if err != nil {
		return cached{}, err
	}
	if _, _, _, err = responseEntries(endpoint, body); err != nil {
		return cached{}, err
	}
	hit := cached{body: body, at: time.Now()}
	if c.Persistence != nil {
		if err = c.Persistence.save(ctx, endpoint, hit); err != nil {
			return cached{}, fmt.Errorf("%w: database write failed", ErrUnavailable)
		}
	}
	return hit, nil
}
func (c *Client) fetch(ctx context.Context, endpoint string) ([]byte, error) {
	// Bound waiting requests as well as active connections; callers can retry busy sources.
	select {
	case c.slots <- struct{}{}:
		defer func() { <-c.slots }()
	default:
		return nil, ErrUnavailable
	}
	c.mu.Lock()
	wait := time.Until(c.next)
	if wait < 0 {
		wait = 0
	}
	c.next = time.Now().Add(wait + 500*time.Millisecond)
	c.mu.Unlock()
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, ErrUnavailable
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "NihongoMaster/1.0 (dictionary lookup)")
	c.upstreamCalls.Add(1)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if res.StatusCode != 200 {
		return nil, ErrUnavailable
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 || !json.Valid(body) {
		return nil, fmt.Errorf("%w: invalid response", ErrUnavailable)
	}
	return body, nil
}
