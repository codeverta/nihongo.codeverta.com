package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"nihongo-master-backend/internal/aichat"
	"nihongo-master-backend/internal/unlock"
	"nihongo-master-backend/internal/diagnostics"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

// aiChatRetention is how long chat rows are kept (also for audit after a
// user hides a conversation).
const aiChatRetention = 90 * 24 * time.Hour

// aiStreamTimeout bounds one answer, including the upstream call.
const aiStreamTimeout = 2 * time.Minute

type aiChatService struct {
	repo      aichat.Repository
	config    *aichat.ConfigStore
	completer aichat.Completer // nil when DEEPSEEK_API_KEY is not configured
	limiter   *aichat.RateLimiter
	lessons   sync.Map // lessonType:lessonID|locale -> aichat.Lesson
	token     string
	now       func() time.Time
}

func (a *aiChatService) quota() aichat.Quota {
	return aichat.Quota{Repo: a.repo, Config: a.config.Get}
}

// ConfigureAIChat wires the AI tutor. completer may be nil, which keeps the
// feature switched off until an API key is configured.
func (s *Server) ConfigureAIChat(repo aichat.Repository, config *aichat.ConfigStore, completer aichat.Completer, adminToken string) {
	if repo == nil {
		repo = aichat.NewMemoryRepository()
	}
	if config == nil {
		config, _ = aichat.LoadConfigStore("")
	}
	s.ai = &aiChatService{repo: repo, config: config, completer: completer, limiter: aichat.NewRateLimiter(), token: adminToken, now: time.Now}
}

// RunAIChatMaintenance purges chat rows past the retention window once a day.
func (s *Server) RunAIChatMaintenance(ctx context.Context) {
	if s.ai == nil {
		return
	}
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		purgeCtx, cancel := context.WithTimeout(ctx, time.Minute)
		if n, err := s.ai.repo.PurgeBefore(purgeCtx, time.Now().Add(-aiChatRetention)); err != nil {
			log.Printf("ai chat purge: %v", err)
			s.ReportBackendWorkerError("ai-chat-purge", err)
		} else if n > 0 {
			log.Printf("ai chat purge: removed %d messages", n)
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Server) registerAIChat(mux *http.ServeMux) {
	if s.ai == nil {
		s.ConfigureAIChat(nil, nil, nil, "")
	}
	mux.HandleFunc("POST /v1/ai/chat", s.requireAuth(s.aiChat))
	mux.HandleFunc("POST /v1/ai/ad-credits", s.requireAuth(s.aiAdCredit))
	mux.HandleFunc("GET /v1/ai/conversations/{lessonType}/{lessonID}", s.requireAuth(s.aiConversation))
	mux.HandleFunc("DELETE /v1/ai/conversations/{lessonType}/{lessonID}", s.requireAuth(s.aiDeleteConversation))
	mux.HandleFunc("GET /v1/admin/ai-config", s.aiAdminGetConfig)
	mux.HandleFunc("PUT /v1/admin/ai-config", s.aiAdminPutConfig)
	mux.HandleFunc("POST /v1/admin/ai-cache/clear", s.aiAdminClearCache)
	// The admin UI lives in the dashboard; /admin/ai redirects to it.
}

type aiChatRequest struct {
	LessonType string `json:"lessonType"`
	LessonID   string `json:"lessonId"`
	Chip       string `json:"chip"`
	Message    string `json:"message"`
}

func aiLocale(r *http.Request) string {
	locale := requestContentLanguage(r.Header.Get("Accept-Language"))
	if locale == localeEnglish && !englishLearningContentEnabled() {
		return localeIndonesian
	}
	return locale
}

// aiIsPremium reports Lifetime status. The stored entitlement expires shortly
// after the last store check, so it is refreshed first (a no-op when recent);
// otherwise Lifetime users could be treated as free and shown ads. A failed
// refresh falls back to the stored state rather than failing the request.
func (s *Server) aiIsPremium(r *http.Request, id string) bool {
	if s.billing != nil {
		if err := s.billing.RefreshUser(r.Context(), id); err != nil {
			log.Printf("ai chat billing refresh: %v", err)
		}
	}
	return s.store.Profile(id).Entitlement.Tier == domain.TierPremium
}

// aiLessonAccess validates the lesson reference and applies the same
// Lifetime gate as the lesson itself.
func (s *Server) aiLessonAccess(w http.ResponseWriter, r *http.Request, lessonType, lessonID string) bool {
	if !aiLessonTypes[lessonType] || !aiLessonIDPattern.MatchString(lessonID) {
		writeError(w, http.StatusUnprocessableEntity, "invalid_lesson", "jenis atau ID pelajaran tidak valid")
		return false
	}
	if s.store.PremiumRequest(aiLessonPath(lessonType, lessonID), "", lessonID, "") && !s.aiIsPremium(r, userID(r)) && !s.aiLessonUnlocked(r, lessonType, lessonID) {
		writeError(w, http.StatusPaymentRequired, "premium_required", "Materi ini memerlukan Lifetime Access.")
		return false
	}
	return true
}

// aiLessonUnlocked reports whether the account opened this lesson with an ad,
// so its AI tutor opens with it.
func (s *Server) aiLessonUnlocked(r *http.Request, lessonType, lessonID string) bool {
	key, ok := unlock.Key(unlock.Kind(lessonType), lessonID)
	return ok && s.unlocks != nil && s.unlocks.svc.Allowed(r.Context(), userID(r), key)
}

func (s *Server) aiChat(w http.ResponseWriter, r *http.Request) {
	cfg := s.ai.config.Get()
	if !cfg.Enabled || s.ai.completer == nil {
		writeError(w, http.StatusServiceUnavailable, "ai_disabled", "Fitur Tanya AI sedang tidak tersedia.")
		return
	}
	var input aiChatRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "format permintaan tidak valid")
		return
	}
	chip, ok := aichat.ParseChip(input.Chip)
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "invalid_chip", "aksi cepat tidak dikenal")
		return
	}
	if utf8.RuneCountInString(input.Message) > aichat.MaxMessageRunes {
		writeError(w, http.StatusUnprocessableEntity, "message_too_long", fmt.Sprintf("Pesan maksimal %d karakter.", aichat.MaxMessageRunes))
		return
	}
	message := aichat.SanitizeMessage(input.Message)
	if chip.NeedsMessage() && message == "" {
		writeError(w, http.StatusUnprocessableEntity, "message_required", "Tulis pertanyaan atau kalimatmu dulu.")
		return
	}
	if !chip.NeedsMessage() {
		// Quick actions other than correction take no free text, which keeps
		// their answers shareable through the cache.
		message = ""
	}
	if !s.aiLessonAccess(w, r, input.LessonType, input.LessonID) {
		return
	}
	id := userID(r)
	now := s.ai.now()
	if !s.ai.limiter.Allow(id, cfg.PerMinute, now) {
		writeError(w, http.StatusTooManyRequests, "ai_rate_limited", "Terlalu cepat. Tunggu sebentar lalu coba lagi.")
		return
	}
	locale := aiLocale(r)
	lesson, err := s.aiLessonContext(id, input.LessonType, input.LessonID, locale)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "pelajaran tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}

	premium := s.aiIsPremium(r, id)
	adsAvailable := false
	if s.ads != nil {
		s.ads.mu.RLock()
		adsAvailable = s.ads.value.aiChatRewardedAvailable()
		s.ads.mu.RUnlock()
	}
	quota := s.ai.quota()
	reservation, err := quota.Reserve(r.Context(), id, premium, adsAvailable, now)
	switch {
	case errors.Is(err, aichat.ErrAdRequired):
		writeError(w, http.StatusTooManyRequests, "ai_ad_required", "Kuota gratis hari ini sudah habis. Tonton iklan singkat untuk mengirim pertanyaan ini.")
		return
	case errors.Is(err, aichat.ErrDailyLimit):
		writeError(w, http.StatusTooManyRequests, "ai_daily_limit", "Kuota Tanya AI hari ini sudah habis. Coba lagi besok atau buka Lifetime Access.")
		return
	case errors.Is(err, aichat.ErrPremiumAdOffer):
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": map[string]any{
			"code":          "ai_premium_ad_offer",
			"message":       fmt.Sprintf("Kamu sudah memakai %d pesan Tanya AI hari ini. Tonton iklan singkat untuk %d pesan lagi.", cfg.PremiumPerDay, cfg.LifetimeAdBonus),
			"messagesPerAd": cfg.LifetimeAdBonus,
		}})
		return
	case errors.Is(err, aichat.ErrPremiumLimit):
		writeError(w, http.StatusTooManyRequests, "ai_premium_limit", "Batas harian tercapai, coba lagi besok.")
		return
	case err != nil:
		log.Printf("ai chat reserve: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}

	// From here the reservation must be released unless an answer completes.
	committed := false
	defer func() {
		if !committed {
			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
			defer cancel()
			if err := quota.Release(ctx, reservation); err != nil {
				log.Printf("ai chat release: %v", err)
			}
		}
	}()

	stream, err := newSSEStream(w)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "streaming tidak didukung")
		return
	}
	userMessage := aichat.Message{UserID: id, LessonKey: lesson.Key, Role: aichat.RoleUser, Chip: string(chip), Content: message, Locale: locale}

	cacheKey := ""
	if chip.Cacheable(message) {
		cacheKey = aichat.AnswerCacheKey(lesson, chip, locale)
		if answer, hit, err := s.ai.repo.CachedAnswer(r.Context(), cacheKey); err == nil && hit {
			assistant := aichat.Message{UserID: id, LessonKey: lesson.Key, Role: aichat.RoleAssistant, Chip: string(chip), Content: answer, Locale: locale, Cached: true}
			if err := s.ai.repo.AppendMessages(r.Context(), userMessage, assistant); err != nil {
				log.Printf("ai chat store cached exchange: %v", err)
			}
			committed = true
			_ = stream.send("delta", map[string]string{"text": answer})
			_ = stream.send("done", map[string]any{"cached": true})
			s.aiUsageAlert(r, id, reservation.Day)
			return
		} else if err != nil {
			log.Printf("ai chat cache lookup: %v", err)
		}
	}

	history, err := s.ai.repo.RecentMessages(r.Context(), id, lesson.Key, aichat.HistoryMessages)
	if err != nil {
		log.Printf("ai chat history: %v", err)
		history = nil
	}
	ctx, cancel := context.WithTimeout(r.Context(), aiStreamTimeout)
	defer cancel()
	completion, err := s.ai.completer.Stream(ctx, aichat.CompletionRequest{
		Messages:    aichat.BuildMessages(lesson, locale, history, chip, message),
		MaxTokens:   aichat.MaxTokensFor(chip),
		Temperature: aiTemperature(chip),
	}, func(delta string) error {
		return stream.send("delta", map[string]string{"text": delta})
	})
	if err == nil && strings.TrimSpace(completion.Text) == "" {
		err = errors.New("empty answer")
	}
	if err != nil {
		log.Printf("ai chat upstream (%s): %v", lesson.Key, err)
		_ = stream.send("error", map[string]string{"code": "ai_upstream_error", "message": "Jawaban gagal dibuat. Kuota kamu tidak berkurang, silakan coba lagi."})
		return
	}
	assistant := aichat.Message{
		UserID: id, LessonKey: lesson.Key, Role: aichat.RoleAssistant, Chip: string(chip), Content: completion.Text, Locale: locale,
		PromptTokens: completion.Usage.PromptTokens, CacheHitTokens: completion.Usage.CacheHitTokens, CompletionTokens: completion.Usage.CompletionTokens,
	}
	storeCtx, storeCancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
	defer storeCancel()
	if err := s.ai.repo.AppendMessages(storeCtx, userMessage, assistant); err != nil {
		log.Printf("ai chat store exchange: %v", err)
	}
	// Only complete answers are shared; truncated ones would mislead others.
	if cacheKey != "" && completion.FinishReason == "stop" {
		if err := s.ai.repo.StoreAnswer(storeCtx, cacheKey, completion.Text); err != nil {
			log.Printf("ai chat cache store: %v", err)
		}
	}
	committed = true
	_ = stream.send("done", map[string]any{"cached": false, "truncated": completion.FinishReason == "length"})
	s.aiUsageAlert(r, id, reservation.Day)
}

// aiTemperature keeps factual answers stable and quizzes varied.
func aiTemperature(chip aichat.Chip) float64 {
	switch chip {
	case aichat.ChipQuiz, aichat.ChipExamples:
		return 1.0
	case aichat.ChipCorrect:
		return 0.3
	}
	return 0.7
}

// aiUsageAlert records one admin diagnostic when an account crosses the
// configured daily alert threshold.
func (s *Server) aiUsageAlert(r *http.Request, id, day string) {
	threshold := s.ai.config.Get().AlertPerDay
	if threshold <= 0 {
		return
	}
	usage, err := s.ai.repo.Usage(r.Context(), id, day)
	if err != nil || usage.Total() != threshold {
		return
	}
	s.recordBackendDiagnostic(r, diagnostics.Event{
		Severity: "warning", ErrorType: "ai_usage_alert", StatusCode: http.StatusOK,
		Message: fmt.Sprintf("Akun %s sudah mengirim %d pesan Tanya AI pada %s", id, usage.Total(), day),
	})
}

func (s *Server) aiAdCredit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		// Status is "watched" or "unavailable" (no fill). Both grant a
		// credit and both count toward the daily ad cap.
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &input); err != nil || (input.Status != "watched" && input.Status != "unavailable") {
		writeError(w, http.StatusBadRequest, "invalid_json", "format permintaan tidak valid")
		return
	}
	premium := s.aiIsPremium(r, userID(r))
	if premium && input.Status != "watched" {
		// Lifetime accounts get several messages per ad, so only a watched
		// ad pays; with no fill they simply try again later.
		writeError(w, http.StatusUnprocessableEntity, "ad_not_watched", "Iklan belum ditonton. Coba lagi sebentar lagi.")
		return
	}
	err := s.ai.quota().GrantAdCredit(r.Context(), userID(r), premium, s.ai.now())
	if errors.Is(err, aichat.ErrAdCapReached) {
		writeError(w, http.StatusTooManyRequests, "ai_daily_limit", "Kuota Tanya AI hari ini sudah habis. Coba lagi besok atau buka Lifetime Access.")
		return
	}
	if err != nil {
		log.Printf("ai ad credit: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type aiConversationMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Chip      string    `json:"chip,omitempty"`
	Content   string    `json:"content"`
	Cached    bool      `json:"cached,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *Server) aiConversation(w http.ResponseWriter, r *http.Request) {
	lessonType, lessonID := r.PathValue("lessonType"), r.PathValue("lessonID")
	if !s.aiLessonAccess(w, r, lessonType, lessonID) {
		return
	}
	messages, err := s.ai.repo.RecentMessages(r.Context(), userID(r), lessonType+":"+lessonID, 50)
	if err != nil {
		log.Printf("ai conversation: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}
	out := make([]aiConversationMessage, len(messages))
	for i, m := range messages {
		out[i] = aiConversationMessage{ID: m.ID, Role: m.Role, Chip: m.Chip, Content: m.Content, Cached: m.Cached, CreatedAt: m.CreatedAt}
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeData(w, http.StatusOK, map[string]any{"messages": out})
}

func (s *Server) aiDeleteConversation(w http.ResponseWriter, r *http.Request) {
	lessonType, lessonID := r.PathValue("lessonType"), r.PathValue("lessonID")
	if !aiLessonTypes[lessonType] || !aiLessonIDPattern.MatchString(lessonID) {
		writeError(w, http.StatusUnprocessableEntity, "invalid_lesson", "jenis atau ID pelajaran tidak valid")
		return
	}
	if err := s.ai.repo.HideConversation(r.Context(), userID(r), lessonType+":"+lessonID); err != nil {
		log.Printf("ai delete conversation: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// sseStream writes Server-Sent Events and flushes after each one.
type sseStream struct {
	w          http.ResponseWriter
	controller *http.ResponseController
}

func newSSEStream(w http.ResponseWriter) (*sseStream, error) {
	controller := http.NewResponseController(w)
	// The server's global WriteTimeout is shorter than a long answer.
	_ = controller.SetWriteDeadline(time.Now().Add(aiStreamTimeout + 15*time.Second))
	h := w.Header()
	h.Set("Content-Type", "text/event-stream; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Accel-Buffering", "no") // disable proxy buffering (nginx)
	w.WriteHeader(http.StatusOK)
	if err := controller.Flush(); err != nil {
		return nil, err
	}
	return &sseStream{w: w, controller: controller}, nil
}

func (s *sseStream) send(event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	return s.controller.Flush()
}

func (s *Server) aiAdminAuthorized(r *http.Request) bool {
	return adminRequestAuthorized(r, s.ai.token)
}

func (s *Server) aiAdminGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !s.aiAdminAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum diatur di server.")
		return
	}
	writeData(w, http.StatusOK, map[string]any{"config": s.ai.config.Get(), "apiKeyConfigured": s.ai.completer != nil})
}

func (s *Server) aiAdminPutConfig(w http.ResponseWriter, r *http.Request) {
	if !s.aiAdminAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum diatur di server.")
		return
	}
	// Start from the saved settings so an older dashboard that does not send
	// the newer fields keeps their stored values.
	input := s.ai.config.Get()
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "format pengaturan tidak valid")
		return
	}
	if err := input.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_ai_config", err.Error())
		return
	}
	if err := s.ai.config.Set(input); err != nil {
		writeError(w, http.StatusInternalServerError, "save_failed", "gagal menyimpan pengaturan")
		return
	}
	writeData(w, http.StatusOK, map[string]any{"config": input, "apiKeyConfigured": s.ai.completer != nil})
}

func (s *Server) aiAdminClearCache(w http.ResponseWriter, r *http.Request) {
	if !s.aiAdminAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum diatur di server.")
		return
	}
	n, err := s.ai.repo.ClearAnswers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "clear_failed", "gagal menghapus cache")
		return
	}
	writeData(w, http.StatusOK, map[string]any{"removed": n})
}
