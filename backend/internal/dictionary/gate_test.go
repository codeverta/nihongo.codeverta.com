package dictionary

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestUpstreamGateOnlyGuardsUncachedLookups(t *testing.T) {
	c := NewClient()
	calls := 0
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		calls++
		return response(200, `{"meta":{"status":200},"data":[]}`), nil
	})
	gateCalls := 0
	deny := WithUpstreamGate(context.Background(), func() bool { gateCalls++; return false })
	if _, err := c.Search(deny, "cat", 1, "", false); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("denied gate should rate limit, got %v", err)
	}
	if calls != 0 || gateCalls != 1 {
		t.Fatalf("denied lookup reached upstream: calls=%d gate=%d", calls, gateCalls)
	}
	allow := WithUpstreamGate(context.Background(), func() bool { gateCalls++; return true })
	if _, err := c.Search(allow, "cat", 1, "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Search(deny, "cat", 1, "", false); err != nil {
		t.Fatalf("cached lookup must ignore the gate: %v", err)
	}
	if calls != 1 || gateCalls != 2 {
		t.Fatalf("unexpected calls=%d gate=%d", calls, gateCalls)
	}
}

func TestCachedLookupsNeverReachUpstream(t *testing.T) {
	c := NewClient()
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		t.Error("cached-only lookup reached upstream")
		return nil, errors.New("offline")
	})
	if _, err := c.CachedWord(context.Background(), "猫"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing word should be not found: %v", err)
	}
	if _, err := c.CachedKanji(context.Background(), "猫"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing kanji should be not found: %v", err)
	}
	if ids, err := c.ArchivedIDs(context.Background(), "word", 10); err != nil || len(ids) != 0 {
		t.Fatalf("no persistence should list nothing: %v %v", ids, err)
	}
}
