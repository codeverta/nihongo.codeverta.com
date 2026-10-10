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
	"strconv"
	"strings"
	"sync"
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
	done chan struct{}
	body []byte
	err  error
}
type Client struct {
	HTTP    *http.Client
	mu      sync.Mutex
	cache   map[string]cached
	flights map[string]*flight
	slots   chan struct{}
	next    time.Time
}

func NewClient() *Client {
	return &Client{HTTP: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 3 || req.URL.Host != via[0].URL.Host || req.URL.Scheme != "https" {
			return ErrUnavailable
		}
		return nil
	}}, cache: map[string]cached{}, flights: map[string]*flight{}, slots: make(chan struct{}, 2)}
}
func (c *Client) Search(ctx context.Context, query string, page int, jlpt string, common bool) (Words, error) {
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
	keyword := query
	if jlpt != "" {
		keyword += " #jlpt-" + strings.ToLower(jlpt)
	}
	if common {
		keyword += " #common"
	}
	endpoint := "https://jisho.org/api/v1/search/words?" + url.Values{"keyword": {keyword}, "page": {strconv.Itoa(page)}}.Encode()
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
func (c *Client) get(ctx context.Context, endpoint string) ([]byte, error) {
	c.mu.Lock()
	if hit, ok := c.cache[endpoint]; ok && time.Since(hit.at) < 6*time.Hour {
		c.mu.Unlock()
		return hit.body, nil
	}
	if f, ok := c.flights[endpoint]; ok {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-f.done:
			return f.body, f.err
		}
	}
	f := &flight{done: make(chan struct{})}
	c.flights[endpoint] = f
	c.mu.Unlock()
	body, err := c.fetch(ctx, endpoint)
	c.mu.Lock()
	if err == nil {
		if len(c.cache) >= 128 {
			var oldest string
			var at time.Time
			for key, value := range c.cache {
				if oldest == "" || value.at.Before(at) {
					oldest = key
					at = value.at
				}
			}
			delete(c.cache, oldest)
		}
		c.cache[endpoint] = cached{body: body, at: time.Now()}
	}
	f.body, f.err = body, err
	delete(c.flights, endpoint)
	close(f.done)
	c.mu.Unlock()
	return body, err
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
