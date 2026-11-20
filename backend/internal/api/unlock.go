package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"nihongo-master-backend/internal/aichat"
	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/jobs"
	"nihongo-master-backend/internal/store"
	"nihongo-master-backend/internal/unlock"
)

// Opening a locked item by watching a rewarded ad. See
// docs/rewarded-unlock-spec.md. The ad is only ever proven by Google's signed
// server-side verification callback; the client just reports that it finished.

type unlockAPI struct {
	svc     *unlock.Service
	limiter *aichat.RateLimiter
	// claimWait is how long POST /v1/unlocks/claim waits for Google's callback
	// before telling the app to ask again.
	claimWait time.Duration
	claimPoll time.Duration
	// Callback counters since the server started, for the admin dashboard.
	callbacksOK, callbacksIgnored, callbacksBadSignature atomic.Int64
	startedAt                                            time.Time
}

const (
	unlockRequestsPerMinute = 30
	unlockClaimWait         = 8 * time.Second
	unlockClaimPoll         = 400 * time.Millisecond
	unlockRetention         = 24 * time.Hour
)

var unlockNoncePattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// unlockLessonRoute matches the routes whose content an ad can open: lesson
// pages and the quiz attempts that belong to them. Offline packages, PDFs and
// everything else stay Lifetime-only.
var unlockLessonRoute = regexp.MustCompile(`^/v1/(vocabulary|grammar|kanji|reading)/(lessons|attempts)/`)

// unlockAIRoute matches the AI tutor's conversation routes for a lesson.
var unlockAIRoute = regexp.MustCompile(`^/v1/ai/conversations/(vocabulary|grammar|kanji|reading)/`)

// unlockAnkiRoute matches studying and reviewing a built-in deck.
var unlockAnkiRoute = regexp.MustCompile(`^/v1/anki/decks/[^/]+/(study|review)$`)

// unlockSimulationAttemptRoute matches the routes of one simulation attempt.
var unlockSimulationAttemptRoute = regexp.MustCompile(`^/v1/simulation/attempts/`)

// unlockSimulationOverviewRoute matches a simulation paper's overview.
var unlockSimulationOverviewRoute = regexp.MustCompile(`^/v1/simulation/([^/]+)$`)

// adUnlockableExams are the premium simulation papers an ad can open, one
// attempt per ad. N3–N1 papers 02–10 are already free; only paper 01 is
// premium, and the product decision covers N3.
var adUnlockableExams = map[string]string{"jlpt-n3-simulation-1": "N3"}

func adUnlockableSimulationLevel(level string) bool {
	for _, unlockable := range adUnlockableExams {
		if strings.EqualFold(level, unlockable) {
			return true
		}
	}
	return false
}

// adsSnapshot returns the current ads configuration.
func (s *Server) adsSnapshot() AdsConfig {
	if s.ads == nil {
		return defaultAdsConfig()
	}
	s.ads.mu.RLock()
	defer s.ads.mu.RUnlock()
	return s.ads.value
}

// ConfigureUnlocks wires ad-based unlocking. keys may be nil, which keeps
// Google's callbacks unverifiable (and so every ad unlock impossible) until it
// is set; the fallback path still works.
func (s *Server) ConfigureUnlocks(repo unlock.Repository, keys unlock.KeyProvider) {
	if repo == nil {
		repo = unlock.NewMemoryRepository()
	}
	s.unlocks = &unlockAPI{
		svc: &unlock.Service{
			Repo:   repo,
			Limits: func() unlock.Limits { return s.adsSnapshot().unlockLimits() },
			Verify: &unlock.Verifier{Keys: keys},
			Units: func() []string {
				cfg := s.adsSnapshot()
				return []string{cfg.UnlockRewarded.UnitID, cfg.XPBoostRewarded.UnitID}
			},
		},
		limiter:   aichat.NewRateLimiter(),
		claimWait: unlockClaimWait,
		claimPoll: unlockClaimPoll,
		startedAt: time.Now(),
	}
}

// ConfigureUnlockTiming changes how long a claim waits for Google's callback
// and how often it checks. It exists for tests, which cannot wait seconds.
func (s *Server) ConfigureUnlockTiming(wait, poll time.Duration) {
	if s.unlocks == nil {
		s.ConfigureUnlocks(nil, nil)
	}
	s.unlocks.claimWait, s.unlocks.claimPoll = wait, poll
}

// RunUnlockMaintenance deletes old intents once a day.
func (s *Server) RunUnlockMaintenance(ctx context.Context) {
	if s.unlocks == nil {
		return
	}
	ticker := time.NewTicker(unlockRetention)
	defer ticker.Stop()
	for {
		if n, err := s.unlocks.svc.Repo.PurgeIntentsBefore(ctx, time.Now().Add(-unlock.IntentRetention)); err != nil {
			log.Printf("unlock purge: %v", err)
			s.ReportBackendWorkerError("unlock-purge", err)
		} else if n > 0 {
			log.Printf("unlock purge: removed %d intents", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Server) registerUnlocks(mux *http.ServeMux) {
	if s.unlocks == nil {
		s.ConfigureUnlocks(nil, nil)
	}
	mux.HandleFunc("GET /v1/unlocks", s.requireAuth(s.unlockList))
	mux.HandleFunc("POST /v1/unlocks/intents", s.requireAuth(s.unlockIntent))
	mux.HandleFunc("POST /v1/unlocks/claim", s.requireAuth(s.unlockClaim))
	mux.HandleFunc("POST /v1/unlocks/fallback", s.requireAuth(s.unlockFallback))
	mux.HandleFunc("GET /v1/ads/ssv", s.unlockSSV)
	mux.HandleFunc("POST /v1/jobs/{jobID}/apply-link", s.requireAuth(s.jobApplyLink))
	mux.HandleFunc("GET /v1/admin/unlock-stats", s.unlockStats)
	mux.HandleFunc("POST /v1/me/practice-lessons/{sectorID}/{lessonID}/complete", s.requireAuth(s.completePracticeLesson))
	mux.HandleFunc("POST /v1/xp-boosts/{completionID}/free", s.requireAuth(s.freeXPBoost))
}

// completePracticeLesson records a finished Kaigo or SSW curriculum lesson
// for XP and the streak (docs/ads-xp-spec.md).
func (s *Server) completePracticeLesson(w http.ResponseWriter, r *http.Request) {
	gain, err := s.store.CompletePracticeLesson(userID(r), r.PathValue("sectorID"), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, http.StatusOK, map[string]any{"xp": gain})
}

// unlockStats is the admin view of ad unlocks. "accountsNowLifetime" counts
// accounts that opened something in the window and hold Lifetime Access now;
// it shows overlap, not that the ad caused the purchase.
func (s *Server) unlockStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !s.adsAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum diatur di server.")
		return
	}
	days := 14
	if raw := r.URL.Query().Get("days"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > unlock.MaxStatsDays {
			writeError(w, http.StatusUnprocessableEntity, "invalid_days", "Jumlah hari harus 1 sampai 30.")
			return
		}
		days = n
	}
	report, err := s.unlocks.svc.Stats(r.Context(), days)
	if err != nil {
		log.Printf("unlock stats: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}
	lifetime := 0
	for _, account := range report.Accounts {
		if s.store.Profile(account).Entitlement.Tier == domain.TierPremium {
			lifetime++
		}
	}
	writeData(w, http.StatusOK, map[string]any{
		"since": report.Since, "days": report.Days, "totals": report.Totals,
		"accounts": len(report.Accounts), "accountsNowLifetime": lifetime,
		"intentRetentionDays": int(unlock.IntentRetention / (24 * time.Hour)),
		"callbacks": map[string]int64{
			"ok": s.unlocks.callbacksOK.Load(), "ignored": s.unlocks.callbacksIgnored.Load(),
			"badSignature": s.unlocks.callbacksBadSignature.Load(),
		},
		"callbacksSince": s.unlocks.startedAt.UTC().Format(time.RFC3339),
	})
}

// unlockedByAd reports whether the account opened this lesson with an ad. Only
// lesson and attempt routes qualify, so offline packages stay Lifetime-only.
func (s *Server) unlockedByAd(r *http.Request, id string) bool {
	if s.unlocks == nil {
		return false
	}
	// A paper's overview lists its sections and history, never questions, so
	// it opens while an ad can open the paper: the app needs it to offer one.
	if match := unlockSimulationOverviewRoute.FindStringSubmatch(r.URL.Path); match != nil {
		return adUnlockableSimulationLevel(match[1]) && s.adsSnapshot().unlockOffered(unlock.KindSimulation)
	}
	kind, itemID, ok := s.routeItem(r)
	if !ok {
		return false
	}
	key, ok := unlock.Key(kind, itemID)
	return ok && s.unlocks.svc.Allowed(r.Context(), id, key)
}

// routeItem names the unlockable item a gated route serves. Offline packages,
// PDFs, exports and everything else name none, so they stay Lifetime-only.
func (s *Server) routeItem(r *http.Request) (unlock.Kind, string, bool) {
	switch path := r.URL.Path; {
	case unlockLessonRoute.MatchString(path):
		kind, itemID, ok := s.store.LessonRef(r.PathValue("lessonID"), r.PathValue("attemptID"))
		return unlock.Kind(kind), itemID, ok
	case unlockAIRoute.MatchString(path):
		// The AI tutor of a lesson opens together with the lesson.
		return unlock.Kind(r.PathValue("lessonType")), r.PathValue("lessonID"), true
	case unlockAnkiRoute.MatchString(path):
		return unlock.KindAnki, r.PathValue("deckID"), true
	case unlockSimulationAttemptRoute.MatchString(path):
		return unlock.KindSimulationAttempt, r.PathValue("attemptID"), true
	}
	return "", "", false
}

// writePremiumRequired is the 402 for gated routes.
func (s *Server) writePremiumRequired(w http.ResponseWriter, r *http.Request) {
	kind, itemID, _ := s.routeItem(r)
	s.writeLockedItem(w, kind, itemID, "Materi N3–N1 dan tes selain N5 memerlukan Lifetime Access.")
}

// writeLockedItem is the 402 for locked content. When an ad could open the
// item it also says which item, so the app can offer to watch one.
func (s *Server) writeLockedItem(w http.ResponseWriter, kind unlock.Kind, itemID, message string) {
	body := map[string]any{"code": "premium_required", "message": message}
	if _, clientKind := unlock.ParseKind(string(kind)); clientKind && itemID != "" && s.adsSnapshot().unlockOffered(kind) {
		body["unlockable"] = true
		body["unlock"] = map[string]string{"kind": string(kind), "itemId": itemID}
	}
	writeJSON(w, http.StatusPaymentRequired, map[string]any{"error": body})
}

// unlockLocked reports whether item is currently locked for free accounts,
// which is the only case where unlocking it means anything.
func (s *Server) unlockLocked(r *http.Request, kind unlock.Kind, itemID string) bool {
	cfg := s.adsSnapshot()
	switch {
	case kind.IsLesson():
		return s.store.PremiumRequest("/v1/"+string(kind)+"/lessons/"+itemID, "", itemID, "")
	case kind == unlock.KindAnki:
		deck, err := anki.Get(itemID)
		return err == nil && store.PremiumLevel(deck.Level)
	case kind == unlock.KindSimulation:
		_, unlockable := adUnlockableExams[itemID]
		return unlockable && s.store.ExamRequiresPremium(itemID)
	case kind == unlock.KindNews:
		if !cfg.ContentGates.News || s.news == nil {
			return false
		}
		_, err := s.news.Detail(r.Context(), itemID, false)
		return err == nil
	case kind == unlock.KindJob:
		if !cfg.ContentGates.Jobs || s.jobs == nil {
			return false
		}
		_, err := s.jobs.Get(r.Context(), itemID)
		return err == nil
	}
	return false
}

// itemUnlocked reports whether account id opened item with an ad.
func (s *Server) itemUnlocked(r *http.Request, id string, kind unlock.Kind, itemID string) bool {
	key, ok := unlock.Key(kind, itemID)
	return ok && id != "" && s.unlocks != nil && s.unlocks.svc.Allowed(r.Context(), id, key)
}

// gateNews decides whether the caller may read a news article. It writes the
// refusal and returns true when the article stays locked.
func (s *Server) gateNews(w http.ResponseWriter, r *http.Request, articleID string) bool {
	if !s.adsSnapshot().ContentGates.News {
		return false
	}
	id, ok := s.store.ResolveSession(bearerToken(r))
	if !ok {
		writeError(w, http.StatusUnauthorized, "login_required", "Masuk untuk membaca berita ini.")
		return true
	}
	if s.aiIsPremium(r, id) || s.itemUnlocked(r, id, unlock.KindNews, articleID) {
		return false
	}
	s.writeLockedItem(w, unlock.KindNews, articleID, "Berita ini memerlukan Lifetime Access atau dibuka dengan iklan.")
	return true
}

// jobLinksLocked reports whether job application links are hidden from the
// public catalog.
func (s *Server) jobLinksLocked() bool { return s.adsSnapshot().ContentGates.Jobs }

// publicJob is a job as the public catalog shows it while application links
// are locked: both links are removed, because many sources use the same URL
// for the listing and the application.
type publicJob struct {
	jobs.Job
	ApplyLocked bool `json:"applyLocked"`
}

func lockedJob(job jobs.Job) publicJob {
	job.ApplyURL, job.SourceURL = "", ""
	return publicJob{Job: job, ApplyLocked: true}
}

type publicJobPage struct {
	Items      []publicJob `json:"items"`
	Total      int64       `json:"total"`
	NextCursor string      `json:"nextCursor"`
	HasMore    bool        `json:"hasMore"`
}

func lockedJobPage(page jobs.Page) publicJobPage {
	items := make([]publicJob, 0, len(page.Items))
	for _, job := range page.Items {
		items = append(items, lockedJob(job))
	}
	return publicJobPage{Items: items, Total: page.Total, NextCursor: page.NextCursor, HasMore: page.HasMore}
}

// jobApplyLink returns a job's application link to an account allowed to
// apply: Lifetime, or one that opened this job with an ad. While the jobs gate
// is off the link is public anyway, so every signed-in account gets it.
func (s *Server) jobApplyLink(w http.ResponseWriter, r *http.Request) {
	if s.jobs == nil {
		writeError(w, 503, "jobs_unavailable", "Katalog lowongan belum terhubung. Coba lagi nanti.")
		return
	}
	jobID := r.PathValue("jobID")
	job, err := s.jobs.Get(r.Context(), jobID)
	if err != nil {
		writeJobError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	id := userID(r)
	if s.jobLinksLocked() && !s.aiIsPremium(r, id) && !s.itemUnlocked(r, id, unlock.KindJob, job.ID) {
		s.writeLockedItem(w, unlock.KindJob, job.ID, "Melamar lowongan memerlukan Lifetime Access atau dibuka dengan iklan.")
		return
	}
	writeData(w, http.StatusOK, map[string]string{"applyUrl": job.ApplyURL, "sourceUrl": job.SourceURL})
}

// startSimulationWithCredit handles starting a premium simulation paper that
// an ad can open. One ad buys one attempt: the credit is converted into an
// unlock of the attempt it starts, and resuming that attempt costs nothing.
// It returns false when the request is not such a start, leaving it to the
// normal path.
func (s *Server) startSimulationWithCredit(w http.ResponseWriter, r *http.Request, examID string) bool {
	if _, unlockable := adUnlockableExams[examID]; !unlockable || s.unlocks == nil || !s.store.ExamRequiresPremium(examID) {
		return false
	}
	id := userID(r)
	if s.aiIsPremium(r, id) {
		return false
	}
	ctx := r.Context()
	credit, _ := unlock.Key(unlock.KindSimulation, examID)
	attemptKey := func(attemptID string) string {
		key, _ := unlock.Key(unlock.KindSimulationAttempt, attemptID)
		return key
	}
	active, hasActive := s.store.ActiveSimulationAttempt(id, examID)
	resuming := hasActive && s.unlocks.svc.Allowed(ctx, id, attemptKey(active))
	if !resuming && !s.unlocks.svc.Allowed(ctx, id, credit) {
		s.writeLockedItem(w, unlock.KindSimulation, examID, "Simulasi ini memerlukan Lifetime Access atau dibuka dengan iklan.")
		return true
	}
	// Starting is safe before the credit is spent: an attempt nobody has
	// unlocked cannot be opened, so a lost race leaves nothing usable.
	attempt, err := s.store.StartExamFor(id, examID, true)
	if err != nil {
		writeStoreError(w, err)
		return true
	}
	key := attemptKey(attempt.ID)
	if key == "" {
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return true
	}
	if !s.unlocks.svc.Allowed(ctx, id, key) {
		converted, err := s.unlocks.svc.Convert(ctx, id, credit, key)
		if err != nil {
			log.Printf("unlock convert simulation credit: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
			return true
		}
		if !converted && !s.unlocks.svc.Allowed(ctx, id, key) {
			s.writeLockedItem(w, unlock.KindSimulation, examID, "Simulasi ini memerlukan Lifetime Access atau dibuka dengan iklan.")
			return true
		}
	}
	writeData(w, http.StatusCreated, attempt)
	return true
}

type unlockTarget struct {
	Kind   string `json:"kind"`
	ItemID string `json:"itemId"`
}

// unlockItem validates the requested item and reports whether the account
// should be offered an unlock for it. done is true when a response was written.
func (s *Server) unlockItem(w http.ResponseWriter, r *http.Request, id string) (key string, done bool) {
	if !s.unlocks.limiter.Allow(id, unlockRequestsPerMinute, time.Now()) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Terlalu banyak permintaan. Coba lagi sebentar.")
		return "", true
	}
	var input unlockTarget
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "format permintaan tidak valid")
		return "", true
	}
	kind, ok := unlock.ParseKind(input.Kind)
	key, keyOK := unlock.Key(kind, input.ItemID)
	if !ok || !keyOK {
		writeError(w, http.StatusUnprocessableEntity, "invalid_item", "Jenis atau ID item tidak valid.")
		return "", true
	}
	if kind == unlock.KindXPBoost {
		return key, !s.boostOffered(w, r, id, input.ItemID)
	}
	// An item that is free, or does not exist, has nothing to unlock.
	if !s.unlockLocked(r, kind, input.ItemID) {
		writeError(w, http.StatusUnprocessableEntity, "not_locked", "Item ini tidak terkunci atau tidak ditemukan.")
		return "", true
	}
	if s.aiIsPremium(r, id) {
		writeData(w, http.StatusOK, map[string]any{"alreadyUnlocked": true, "reason": "lifetime"})
		return "", true
	}
	if !s.adsSnapshot().unlockOffered(kind) {
		writeError(w, http.StatusForbidden, "unlock_unavailable", "Membuka dengan iklan sedang tidak tersedia.")
		return "", true
	}
	return key, false
}

func (s *Server) unlockLimitError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, unlock.ErrCapReached):
		writeError(w, http.StatusTooManyRequests, "unlock_daily_limit", "Batas membuka materi gratis hari ini sudah tercapai. Coba lagi besok atau beli Lifetime.")
	case errors.Is(err, unlock.ErrFallbackExhausted):
		writeError(w, http.StatusTooManyRequests, "unlock_fallback_exhausted", "Iklan belum tersedia. Coba lagi nanti atau beli Lifetime.")
	default:
		return false
	}
	return true
}

// boostOffered checks that the completion can be doubled with an ad right now
// and writes the refusal when it cannot.
func (s *Server) boostOffered(w http.ResponseWriter, r *http.Request, id, completionID string) bool {
	cfg := s.adsSnapshot()
	group, err := s.store.XPBoostable(id, completionID, cfg.xpBoostDailyCap())
	if s.writeBoostError(w, err) {
		return false
	}
	if err != nil {
		log.Printf("xp boost check: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return false
	}
	if !cfg.unlockOffered(unlock.KindXPBoost) || !cfg.boostAllows(group) {
		writeError(w, http.StatusForbidden, "unlock_unavailable", "2× XP dengan iklan sedang tidak tersedia.")
		return false
	}
	return true
}

// writeBoostError maps the double XP refusals to responses.
func (s *Server) writeBoostError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusUnprocessableEntity, "not_locked", "Aktivitas ini tidak memberi XP yang bisa digandakan.")
	case errors.Is(err, store.ErrBoostUsed):
		writeError(w, http.StatusConflict, "boost_used", "2× XP untuk aktivitas ini sudah diklaim.")
	case errors.Is(err, store.ErrBoostExpired):
		writeError(w, http.StatusGone, "boost_expired", "2× XP hanya bisa diklaim dalam 30 menit setelah selesai.")
	case errors.Is(err, store.ErrBoostLimit):
		writeError(w, http.StatusTooManyRequests, "boost_limit", "Batas 2× XP hari ini sudah tercapai. Coba lagi besok.")
	default:
		return false
	}
	return true
}

// claimBoost finishes a double XP ad: the verified intent is consumed (never
// an unlock), then the XP of that completion is doubled.
func (s *Server) claimBoost(w http.ResponseWriter, r *http.Request, id, nonce, contentKey string) {
	deadline := time.Now().Add(s.unlocks.claimWait)
	for {
		_, err := s.unlocks.svc.ConsumeVerified(r.Context(), id, nonce)
		switch {
		case err == nil, errors.Is(err, unlock.ErrConsumed):
			completionID := strings.TrimPrefix(contentKey, string(unlock.KindXPBoost)+":")
			gain, boostErr := s.store.BoostXP(id, completionID, s.adsSnapshot().xpBoostDailyCap())
			if errors.Is(boostErr, store.ErrBoostUsed) && errors.Is(err, unlock.ErrConsumed) {
				// A repeated claim: the first one already doubled the XP.
				writeData(w, http.StatusOK, map[string]any{"unlocked": true, "contentKey": contentKey, "alreadyBoosted": true})
				return
			}
			if s.writeBoostError(w, boostErr) {
				return
			}
			if boostErr != nil {
				log.Printf("xp boost: %v", boostErr)
				writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
				return
			}
			writeData(w, http.StatusOK, map[string]any{"unlocked": true, "contentKey": contentKey, "xp": gain})
			return
		case errors.Is(err, unlock.ErrNotVerified):
			if time.Now().After(deadline) {
				writeData(w, http.StatusAccepted, map[string]any{"pending": true})
				return
			}
			select {
			case <-r.Context().Done():
				return
			case <-time.After(s.unlocks.claimPoll):
			}
		case errors.Is(err, unlock.ErrNotFound):
			writeError(w, http.StatusNotFound, "intent_not_found", "Permintaan pembukaan tidak ditemukan.")
			return
		case errors.Is(err, unlock.ErrExpired):
			writeError(w, http.StatusGone, "intent_expired", "Permintaan pembukaan sudah kedaluwarsa. Mulai lagi.")
			return
		default:
			log.Printf("xp boost consume: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
			return
		}
	}
}

// freeXPBoost doubles the XP of an activity without an ad, for Lifetime
// accounts, within the same daily limit.
func (s *Server) freeXPBoost(w http.ResponseWriter, r *http.Request) {
	id := userID(r)
	w.Header().Set("Cache-Control", "no-store")
	if !s.aiIsPremium(r, id) {
		writeError(w, http.StatusPaymentRequired, "premium_required", "Klaim 2× XP tanpa iklan khusus pengguna Lifetime.")
		return
	}
	cfg := s.adsSnapshot()
	completionID := r.PathValue("completionID")
	group, err := s.store.XPBoostable(id, completionID, cfg.xpBoostDailyCap())
	if s.writeBoostError(w, err) {
		return
	}
	if err == nil && !cfg.boostAllows(group) {
		writeError(w, http.StatusForbidden, "unlock_unavailable", "2× XP sedang tidak tersedia untuk aktivitas ini.")
		return
	}
	gain, err := s.store.BoostXP(id, completionID, cfg.xpBoostDailyCap())
	if s.writeBoostError(w, err) {
		return
	}
	if err != nil {
		log.Printf("free xp boost: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}
	writeData(w, http.StatusOK, map[string]any{"xp": gain})
}

func (s *Server) unlockIntent(w http.ResponseWriter, r *http.Request) {
	id := userID(r)
	key, done := s.unlockItem(w, r, id)
	if done {
		return
	}
	intent, err := s.unlocks.svc.CreateIntent(r.Context(), id, key)
	switch {
	case errors.Is(err, unlock.ErrAlreadyUnlocked):
		writeData(w, http.StatusOK, map[string]any{"alreadyUnlocked": true, "reason": "unlocked"})
	case s.unlockLimitError(w, err):
	case err != nil:
		log.Printf("unlock intent: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	default:
		// userId is what the app passes to the ad as its SSV user ID, so that
		// Google's callback names this account. It is the opaque account ID,
		// never the email address.
		writeData(w, http.StatusOK, map[string]any{
			"alreadyUnlocked": false, "nonce": intent.Nonce, "userId": id,
			"expiresAt": intent.ExpiresAt.UTC().Format(time.RFC3339),
		})
	}
}

func (s *Server) unlockClaim(w http.ResponseWriter, r *http.Request) {
	id := userID(r)
	if !s.unlocks.limiter.Allow(id, unlockRequestsPerMinute*4, time.Now()) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Terlalu banyak permintaan. Coba lagi sebentar.")
		return
	}
	var input struct {
		Nonce string `json:"nonce"`
	}
	if err := decodeJSON(r, &input); err != nil || !unlockNoncePattern.MatchString(input.Nonce) {
		writeError(w, http.StatusUnprocessableEntity, "invalid_nonce", "Kode pembukaan tidak valid.")
		return
	}
	if intent, ok, err := s.unlocks.svc.Repo.Intent(r.Context(), input.Nonce); err == nil && ok && intent.UserID == id &&
		strings.HasPrefix(intent.ContentKey, string(unlock.KindXPBoost)+":") {
		s.claimBoost(w, r, id, input.Nonce, intent.ContentKey)
		return
	}
	// Google's callback usually lands a moment after the ad closes, so wait
	// briefly before asking the app to try again.
	deadline := time.Now().Add(s.unlocks.claimWait)
	for {
		granted, err := s.unlocks.svc.Claim(r.Context(), id, input.Nonce)
		switch {
		case err == nil:
			remaining, _, _ := s.unlocks.svc.Remaining(r.Context(), id)
			writeData(w, http.StatusOK, map[string]any{"unlocked": true, "contentKey": granted.ContentKey, "remaining": remaining})
			return
		case errors.Is(err, unlock.ErrNotVerified):
			if time.Now().After(deadline) {
				writeData(w, http.StatusAccepted, map[string]any{"pending": true})
				return
			}
			select {
			case <-r.Context().Done():
				return
			case <-time.After(s.unlocks.claimPoll):
			}
			continue
		case errors.Is(err, unlock.ErrNotFound):
			writeError(w, http.StatusNotFound, "intent_not_found", "Permintaan pembukaan tidak ditemukan.")
		case errors.Is(err, unlock.ErrExpired):
			writeError(w, http.StatusGone, "intent_expired", "Permintaan pembukaan sudah kedaluwarsa. Mulai lagi.")
		case s.unlockLimitError(w, err):
		default:
			log.Printf("unlock claim: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		}
		return
	}
}

func (s *Server) unlockFallback(w http.ResponseWriter, r *http.Request) {
	id := userID(r)
	key, done := s.unlockItem(w, r, id)
	if done {
		return
	}
	// Double XP is a bonus, not access: no ad, no bonus.
	if strings.HasPrefix(key, string(unlock.KindXPBoost)+":") {
		writeError(w, http.StatusForbidden, "boost_no_fallback", "2× XP hanya bisa didapat dengan menonton iklan.")
		return
	}
	granted, err := s.unlocks.svc.Fallback(r.Context(), id, key)
	switch {
	case s.unlockLimitError(w, err):
	case err != nil:
		log.Printf("unlock fallback: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	default:
		remaining, _, _ := s.unlocks.svc.Remaining(r.Context(), id)
		writeData(w, http.StatusOK, map[string]any{"unlocked": true, "contentKey": granted.ContentKey, "source": string(granted.Source), "remaining": remaining})
	}
}

func (s *Server) unlockList(w http.ResponseWriter, r *http.Request) {
	id := userID(r)
	items, err := s.unlocks.svc.Repo.List(r.Context(), id)
	if err != nil {
		log.Printf("unlock list: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.ContentKey)
	}
	remaining, fallbackRemaining, err := s.unlocks.svc.Remaining(r.Context(), id)
	if err != nil {
		log.Printf("unlock usage: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}
	cfg := s.adsSnapshot()
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, http.StatusOK, map[string]any{
		"items": keys, "available": cfg.unlockAvailable(),
		"dailyCap": cfg.unlockLimits().DailyCap, "remaining": remaining, "fallbackRemaining": fallbackRemaining,
		// gates says which content is locked for free accounts right now, so
		// the app only draws locks that are real.
		"gates": map[string]bool{
			"lessons": !store.OpenLearningAccess(), "news": cfg.ContentGates.News, "jobs": cfg.ContentGates.Jobs,
		},
		// offered says, per kind, whether an ad may open it.
		"offered": map[string]bool{
			"lessons": cfg.unlockOffered(unlock.KindGrammar), "anki": cfg.unlockOffered(unlock.KindAnki),
			"simulation": cfg.unlockOffered(unlock.KindSimulation), "news": cfg.unlockOffered(unlock.KindNews),
			"jobs": cfg.unlockOffered(unlock.KindJob),
		},
	})
}

// unlockSSV is the server-to-server callback Google calls after a rewarded ad.
// Only a request signed by Google can verify an intent. A genuine callback that
// cannot be applied still gets 200, because retrying it will not change the
// outcome; only a bad signature or a storage failure is an error.
func (s *Server) unlockSSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	err := s.unlocks.svc.HandleCallback(r.Context(), r.URL.RawQuery)
	switch {
	case err == nil:
		s.unlocks.callbacksOK.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	case errors.Is(err, unlock.ErrBadSignature):
		s.unlocks.callbacksBadSignature.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid signature"))
	case errors.Is(err, unlock.ErrUnavailable):
		w.WriteHeader(http.StatusServiceUnavailable)
	case errors.Is(err, unlock.ErrInvalid), errors.Is(err, unlock.ErrNotFound),
		errors.Is(err, unlock.ErrExpired), errors.Is(err, unlock.ErrReplay):
		log.Printf("unlock ssv ignored: %v", err)
		s.unlocks.callbacksIgnored.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ignored"))
	default:
		log.Printf("unlock ssv: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
