package api

import (
	"net/http"
	"time"

	"nihongo-master-backend/internal/aichat"
)

// leaderboardRequestsPerMinute bounds each account (or guest address).
const leaderboardRequestsPerMinute = 30

var leaderboardLimiter = aichat.NewRateLimiter()

func (s *Server) registerLeaderboard(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/leaderboard", s.optionalAuth(s.leaderboard))
	mux.HandleFunc("GET /v1/me/leaderboard", s.requireAuth(s.myLeaderboard))
	mux.HandleFunc("PUT /v1/me/leaderboard/settings", s.requireAuth(s.updateLeaderboardSettings))
}

func leaderboardAllowed(w http.ResponseWriter, r *http.Request) bool {
	key := userID(r)
	if key == "" {
		key = "ip:" + clientIP(r)
	}
	if !leaderboardLimiter.Allow(key, leaderboardRequestsPerMinute, time.Now()) {
		w.Header().Set("Retry-After", "60")
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Terlalu banyak permintaan. Coba lagi sebentar lagi.")
		return false
	}
	return true
}

// leaderboard serves the public top 50. The body is the same for every
// viewer, so it can be cached for the snapshot lifetime.
func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {
	if !leaderboardAllowed(w, r) {
		return
	}
	page, err := s.store.Leaderboard(r.URL.Query().Get("period"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeData(w, http.StatusOK, page)
}

func (s *Server) myLeaderboard(w http.ResponseWriter, r *http.Request) {
	if !leaderboardAllowed(w, r) {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, http.StatusOK, s.store.MyLeaderboard(userID(r)))
}

func (s *Server) updateLeaderboardSettings(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Hidden *bool `json:"hidden"`
	}
	if err := decodeJSON(r, &input); err != nil || input.Hidden == nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "hidden wajib diisi")
		return
	}
	if err := s.store.SetLeaderboardHidden(userID(r), *input.Hidden); err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]bool{"hidden": *input.Hidden})
}
