package aichat

import (
	"sync"
	"time"
)

// RateLimiter is a per-account sliding one-minute window. It is invisible to
// normal users and stops scripts or shared accounts from bursting.
type RateLimiter struct {
	mu     sync.Mutex
	recent map[string][]time.Time
	calls  int
}

func NewRateLimiter() *RateLimiter { return &RateLimiter{recent: map[string][]time.Time{}} }

// Allow records an attempt and reports whether it is within perMinute.
func (l *RateLimiter) Allow(userID string, perMinute int, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-time.Minute)
	kept := l.recent[userID][:0]
	for _, t := range l.recent[userID] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	allowed := len(kept) < perMinute
	if allowed {
		kept = append(kept, now)
	}
	l.recent[userID] = kept
	// Periodically drop idle accounts so the map does not grow forever.
	if l.calls++; l.calls%1024 == 0 {
		for id, times := range l.recent {
			if len(times) == 0 || !times[len(times)-1].After(cutoff) {
				delete(l.recent, id)
			}
		}
	}
	return allowed
}
