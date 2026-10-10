package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	backendAuth "nihongo-master-backend/internal/auth"
	"nihongo-master-backend/internal/billing"
	"nihongo-master-backend/internal/dictionary"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

type Server struct {
	dictionary          *dictionary.Client
	billing             *billing.Service
	billingPushAudience string
	billingPushEmail    string
	assetDirectory      string
	store               *store.MemoryStore
	googleVerifier      backendAuth.GoogleVerifier
	branding            *BrandingStore
}

func NewServer(store *store.MemoryStore, googleVerifier backendAuth.GoogleVerifier) *Server {
	return &Server{store: store, googleVerifier: googleVerifier, dictionary: dictionary.NewClient()}
}

type userContextKey struct{}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/dictionary/search", s.dictionarySearch)
	mux.HandleFunc("GET /v1/dictionary/kanji/{character}", s.dictionaryKanji)
	mux.HandleFunc("GET /v1/billing", s.requireAuth(s.billingStatus))
	mux.HandleFunc("POST /v1/billing/verify", s.requireAuth(s.verifyPurchase))
	mux.HandleFunc("POST /v1/billing/notifications/apple", s.appleBillingNotification)
	mux.HandleFunc("POST /v1/billing/notifications/google", s.googleBillingNotification)
	mux.HandleFunc("DELETE /v1/me/account", s.requireAuth(s.deleteAccount))
	s.registerAssets(mux)
	s.registerBranding(mux)
	mux.HandleFunc("GET /v1/simulation/{level}", s.requireAuth(s.simulationOverview))
	mux.HandleFunc("GET /v1/simulation/attempts/{attemptID}", s.requireAuth(s.simulationSession))
	mux.HandleFunc("PUT /v1/simulation/attempts/{attemptID}/answers", s.requireAuth(s.simulationAnswer))
	mux.HandleFunc("POST /v1/simulation/attempts/{attemptID}/submit-section", s.requireAuth(s.simulationSubmit))
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /v1/auth/login", s.login)
	mux.HandleFunc("POST /v1/auth/google", s.googleLogin)
	mux.HandleFunc("POST /v1/auth/logout", s.requireAuth(s.logout))
	mux.HandleFunc("GET /v1/auth/me", s.requireAuth(s.profile))
	mux.HandleFunc("GET /v1/home", s.optionalAuth(s.home))
	mux.HandleFunc("GET /v1/learning/categories", s.categories)
	mux.HandleFunc("GET /v1/learning/tracks", s.tracks)
	mux.HandleFunc("GET /v1/exams", s.exams)
	mux.HandleFunc("POST /v1/exam-attempts", s.requireAuth(s.startExam))
	mux.HandleFunc("GET /v1/diagnostic/n5", s.requireAuth(s.diagnosticOverview))
	mux.HandleFunc("GET /v1/exam-attempts/{attemptID}", s.requireAuth(s.diagnosticSession))
	mux.HandleFunc("PUT /v1/exam-attempts/{attemptID}/answers", s.requireAuth(s.diagnosticAnswer))
	mux.HandleFunc("POST /v1/exam-attempts/{attemptID}/complete", s.requireAuth(s.diagnosticComplete))
	mux.HandleFunc("GET /v1/plans", s.plans)
	mux.HandleFunc("GET /v1/me/profile", s.requireAuth(s.profile))
	mux.HandleFunc("PATCH /v1/me/profile", s.requireAuth(s.updateProfile))
	mux.HandleFunc("GET /v1/me/progress", s.requireAuth(s.progress))
	mux.HandleFunc("PUT /v1/me/progress/{lessonID}", s.requireAuth(s.saveProgress))
	mux.HandleFunc("GET /v1/me/gamification", s.requireAuth(s.gamification))
	mux.HandleFunc("GET /v1/me/achievements", s.requireAuth(s.achievements))
	mux.HandleFunc("GET /v1/vocabulary/levels", s.requireAuth(s.vocabularyLevels))
	mux.HandleFunc("GET /v1/vocabulary/lessons/{lessonID}", s.requireAuth(s.vocabularyLesson))
	mux.HandleFunc("POST /v1/vocabulary/lessons/{lessonID}/attempts", s.requireAuth(s.startVocabularyAttempt))
	mux.HandleFunc("POST /v1/vocabulary/attempts/{attemptID}/answers", s.requireAuth(s.answerVocabularyQuestion))
	mux.HandleFunc("POST /v1/vocabulary/attempts/{attemptID}/complete", s.requireAuth(s.completeVocabularyAttempt))
	mux.HandleFunc("GET /v1/grammar/course", s.requireAuth(s.grammarCourse))
	mux.HandleFunc("GET /v1/grammar/lessons/{lessonID}", s.requireAuth(s.grammarLesson))
	mux.HandleFunc("POST /v1/grammar/lessons/{lessonID}/attempts", s.requireAuth(s.startGrammarAttempt))
	mux.HandleFunc("POST /v1/grammar/attempts/{attemptID}/answers", s.requireAuth(s.answerGrammarQuestion))
	mux.HandleFunc("POST /v1/grammar/attempts/{attemptID}/complete", s.requireAuth(s.completeGrammarAttempt))
	mux.HandleFunc("GET /v1/kanji/course", s.requireAuth(s.kanjiCourse))
	mux.HandleFunc("GET /v1/kanji/lessons/{lessonID}", s.requireAuth(s.kanjiLesson))
	mux.HandleFunc("POST /v1/kanji/lessons/{lessonID}/attempts", s.requireAuth(s.startKanjiAttempt))
	mux.HandleFunc("POST /v1/kanji/attempts/{attemptID}/answers", s.requireAuth(s.answerKanjiQuestion))
	mux.HandleFunc("POST /v1/kanji/attempts/{attemptID}/complete", s.requireAuth(s.completeKanjiAttempt))
	mux.HandleFunc("GET /v1/reading/course", s.requireAuth(s.readingCourse))
	mux.HandleFunc("GET /v1/reading/lessons/{lessonID}", s.requireAuth(s.readingLesson))
	mux.HandleFunc("POST /v1/reading/lessons/{lessonID}/attempts", s.requireAuth(s.startReadingAttempt))
	mux.HandleFunc("POST /v1/reading/attempts/{attemptID}/answers", s.requireAuth(s.answerReadingQuestion))
	mux.HandleFunc("POST /v1/reading/attempts/{attemptID}/complete", s.requireAuth(s.completeReadingAttempt))
	return s.middleware(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.Dashboard(userID(r)))
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	profile, err := s.store.AuthenticateEmail(input.Email, input.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "email atau password salah")
		return
	}
	s.writeSession(w, profile)
}

func (s *Server) googleLogin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDToken string `json:"idToken"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	identity, err := s.googleVerifier.Verify(r.Context(), input.IDToken)
	if errors.Is(err, backendAuth.ErrGoogleNotConfigured) {
		writeError(w, http.StatusServiceUnavailable, "google_not_configured", "GOOGLE_CLIENT_ID belum dikonfigurasi")
		return
	}
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_google_token", "Google ID token tidak valid")
		return
	}
	profile, err := s.store.AuthenticateGoogle(identity)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	s.writeSession(w, profile)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	s.store.RevokeSession(bearerToken(r))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeSession(w http.ResponseWriter, profile domain.Profile) {
	token, err := s.store.CreateSession(profile.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session_error", "gagal membuat sesi")
		return
	}
	writeData(w, http.StatusOK, domain.AuthSession{Token: token, User: s.store.Profile(profile.ID)})
}

func (s *Server) categories(w http.ResponseWriter, _ *http.Request) {
	writeData(w, http.StatusOK, s.store.Categories())
}

func (s *Server) tracks(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.Tracks(r.URL.Query().Get("system")))
}

func (s *Server) exams(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.Exams(r.URL.Query().Get("system")))
}

func (s *Server) plans(w http.ResponseWriter, _ *http.Request) {
	writeData(w, http.StatusOK, s.store.Plans())
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.Profile(userID(r)))
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DisplayName string `json:"displayName"`
		Target      string `json:"target"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	profile, err := s.store.UpdateProfile(userID(r), input.DisplayName, input.Target)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, profile)
}

func (s *Server) progress(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.Progress(userID(r)))
}

func (s *Server) saveProgress(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Completion   int `json:"completion"`
		LastPosition int `json:"lastPosition"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	result, err := s.store.SaveProgress(userID(r), domain.LessonProgress{
		LessonID:     r.PathValue("lessonID"),
		Completion:   input.Completion,
		LastPosition: input.LastPosition,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}

func (s *Server) gamification(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.Gamification(userID(r)))
}

func (s *Server) achievements(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.Achievements(userID(r)))
}

func (s *Server) vocabularyLevels(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, s.store.VocabularyLevels(userID(r)))
}

func (s *Server) vocabularyLesson(w http.ResponseWriter, r *http.Request) {
	lesson, err := s.store.VocabularyLesson(userID(r), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, lesson)
}

func (s *Server) startVocabularyAttempt(w http.ResponseWriter, r *http.Request) {
	attempt, err := s.store.StartVocabularyAttempt(userID(r), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusCreated, attempt)
}

func (s *Server) answerVocabularyQuestion(w http.ResponseWriter, r *http.Request) {
	var input struct {
		QuestionID string `json:"questionId"`
		OptionID   string `json:"optionId"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	feedback, err := s.store.AnswerVocabularyQuestion(
		userID(r), r.PathValue("attemptID"), input.QuestionID, input.OptionID,
	)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, feedback)
}

func (s *Server) completeVocabularyAttempt(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.CompleteVocabularyAttempt(userID(r), r.PathValue("attemptID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusOK, result)
}

func (s *Server) startExam(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ExamID string `json:"examId"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if s.billing != nil && s.store.ExamRequiresPremium(input.ExamID) {
		if err := s.billing.RefreshUser(r.Context(), userID(r)); err != nil {
			writeBillingError(w, err)
			return
		}
	}
	attempt, err := s.store.StartExam(userID(r), input.ExamID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, http.StatusCreated, attempt)
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		id, ok := s.store.ResolveSession(token)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Bearer token tidak valid atau sudah dicabut")
			return
		}
		if !s.authorizePremium(w, r, id) {
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, id)))
	}
}

func (s *Server) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := s.store.ResolveSession(bearerToken(r))
		if ok {
			r = r.WithContext(context.WithValue(r.Context(), userContextKey{}, id))
		}
		next(w, r)
	}
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v", recovered)
				writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
			}
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started))
		}()
		next.ServeHTTP(w, r)
	})
}

func userID(r *http.Request) string {
	id, _ := r.Context().Value(userContextKey{}).(string)
	return id
}

func bearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrInvalid):
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", err.Error())
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, store.ErrPremium):
		writeError(w, http.StatusPaymentRequired, "premium_required", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	}
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}
