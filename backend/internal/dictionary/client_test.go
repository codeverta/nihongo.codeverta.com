package dictionary

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLiveDictionarySources(t *testing.T) {
	if os.Getenv("DICTIONARY_LIVE_TEST") != "1" {
		t.Skip("set DICTIONARY_LIVE_TEST=1 for real external requests")
	}
	c := NewClient()
	page, err := c.Search(context.Background(), "neko", 1, "", false)
	if err != nil || len(page.Entries) == 0 {
		t.Fatalf("live Jisho search failed: %v", err)
	}
	if _, err = c.Kanji(context.Background(), "猫"); err != nil {
		t.Fatalf("live kanji request failed: %v", err)
	}
	t.Logf("live sources returned %d Jisho entries and verified 猫 kanji details", len(page.Entries))
}

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}
func TestSearchEncodingFiltersPaginationAndCache(t *testing.T) {
	c := NewClient()
	var calls atomic.Int32
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.URL.Host != "jisho.org" || r.URL.Query().Get("keyword") != "猫 & dog #jlpt-n5 #common" || r.URL.Query().Get("page") != "2" {
			t.Errorf("incorrect request: %s", r.URL)
		}
		return response(200, `{"meta":{"status":200},"data":[{"slug":"猫","japanese":[{"word":"猫","reading":"ねこ"}],"senses":[{"english_definitions":["cat"]}]}]}`), nil
	})
	for i := 0; i < 2; i++ {
		page, err := c.Search(context.Background(), " 猫 & dog ", 2, "n5", true)
		if err != nil || len(page.Entries) != 1 || page.HasMore || page.Page != 2 {
			t.Fatalf("unexpected result %+v %v", page, err)
		}
	}
	if calls.Load() != 1 {
		t.Fatal("cache did not prevent duplicate upstream requests")
	}
}
func TestInvalidQueriesNeverReachExternalHost(t *testing.T) {
	c := NewClient()
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid query made a network request")
		return nil, nil
	})
	for _, q := range []string{"", strings.Repeat("猫", 101), "cat\x00"} {
		if _, err := c.Search(context.Background(), q, 1, "", false); !errors.Is(err, ErrInvalid) {
			t.Fatal(err)
		}
	}
	for _, page := range []int{0, 101} {
		if _, err := c.Search(context.Background(), "cat", page, "", false); !errors.Is(err, ErrInvalid) {
			t.Fatal(err)
		}
	}
	if _, err := c.Search(context.Background(), "cat", 1, "N6", false); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	for _, q := range []string{"猫犬", "https://evil.test", "a"} {
		if _, err := c.Kanji(context.Background(), q); !errors.Is(err, ErrInvalid) {
			t.Fatal(err)
		}
	}
}
func TestUpstreamErrorsAreNotCachedAndKanjiIsValidated(t *testing.T) {
	c := NewClient()
	var calls int
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return response(429, "rate limited"), nil
		}
		if calls == 2 {
			return response(200, `{"error":"bad data"}`), nil
		}
		return response(200, `{"kanji":"猫","stroke_count":11,"on_readings":["ビョウ"]}`), nil
	})
	for i := 0; i < 2; i++ {
		if _, err := c.Kanji(context.Background(), "猫"); !errors.Is(err, ErrUnavailable) {
			t.Fatal(err)
		}
		c.next = time.Time{}
	}
	if _, err := c.Kanji(context.Background(), "猫"); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatal("upstream failure was cached")
	}
	c = NewClient()
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return response(404, "{}"), nil })
	if _, err := c.Kanji(context.Background(), "猫"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}
func TestConcurrentIdenticalSearchUsesOneRequest(t *testing.T) {
	c := NewClient()
	entered := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			close(entered)
		}
		<-release
		return response(200, `{"meta":{"status":200},"data":[]}`), nil
	})
	first := make(chan error)
	go func() { _, err := c.Search(context.Background(), "cat", 1, "", false); first <- err }()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Search(ctx, "cat", 1, "", false); !errors.Is(err, context.Canceled) {
		t.Fatal("waiting caller cannot cancel")
	}
	close(release)
	if err := <-first; err != nil {
		t.Fatal(err)
	}
	if _, err := c.Search(context.Background(), "cat", 1, "", false); err != nil || calls.Load() != 1 {
		t.Fatal("duplicate upstream request")
	}
}
