package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/learningpdf"
	"nihongo-master-backend/internal/store"
)

// Level-wide grammar and kanji lists (every signed-in account) and their PDF
// cheat sheets (Lifetime Access only).

var jlptLevels = map[string]bool{"N5": true, "N4": true, "N3": true, "N2": true, "N1": true}

// learningPDFCache keeps rendered PDFs keyed by kind, level, locale and a
// hash of the localized content, so edited lessons never serve a stale file.
var learningPDFCache sync.Map

func (s *Server) registerLearningLists(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/grammar/levels/{level}/points", s.requireAuth(s.grammarLevelPoints))
	mux.HandleFunc("GET /v1/kanji/levels/{level}/characters", s.requireAuth(s.kanjiLevelCharacters))
	mux.HandleFunc("GET /v1/grammar/levels/{level}/pdf", s.requireAuth(s.learningLevelPDF("grammar")))
	mux.HandleFunc("GET /v1/kanji/levels/{level}/pdf", s.requireAuth(s.learningLevelPDF("kanji")))
	// Not under /v1/simulation/{level}/… to avoid clashing with the attempt routes.
	mux.HandleFunc("GET /v1/simulation-pdf/{level}", s.requireAuth(s.simulationPDF))
}

// requireLifetimeDownload enforces the manga rule for PDF downloads.
func (s *Server) requireLifetimeDownload(w http.ResponseWriter, r *http.Request) bool {
	id := userID(r)
	if s.billing != nil {
		if err := s.billing.RefreshUser(r.Context(), id); err != nil {
			writeBillingError(w, err)
			return false
		}
	}
	entitlement := s.store.Profile(id).Entitlement
	if entitlement.Tier != domain.TierPremium || !entitlement.IsLifetime {
		writeError(w, http.StatusPaymentRequired, "premium_required", "Download PDF khusus pengguna Lifetime Access.")
		return false
	}
	return true
}

func writePDF(w http.ResponseWriter, name string, pdf []byte) {
	h := w.Header()
	h.Set("Content-Type", "application/pdf")
	h.Set("Content-Disposition", `attachment; filename="`+name+`"`)
	h.Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

// cachedPDF renders once per key; key must include a content hash.
func cachedPDF(kind string, parts []string, content any, render func() ([]byte, error)) ([]byte, error) {
	raw, _ := json.Marshal(content)
	sum := sha256.Sum256(raw)
	key := strings.Join(append(append([]string{kind}, parts...), hex.EncodeToString(sum[:])), "|")
	if cached, ok := learningPDFCache.Load(key); ok {
		return cached.([]byte), nil
	}
	out, err := render()
	if err != nil {
		log.Printf("learning pdf %s: %v", key, err)
		return nil, err
	}
	learningPDFCache.Store(key, out)
	return out, nil
}

func (s *Server) simulationPDF(w http.ResponseWriter, r *http.Request) {
	level := strings.ToUpper(r.PathValue("level"))
	paper, err := s.store.SimulationPaper(level)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_level", "jenis simulasi tidak dikenal")
		return
	}
	if !s.requireLifetimeDownload(w, r) {
		return
	}
	locale := aiLocale(r)
	exam := learningpdf.ExamPaper{Title: paper.Title, Sections: paper.Sections, Omitted: paper.Omitted, Reviews: paper.Reviews}
	exam = localizeLesson(exam, "/v1/simulation/"+strings.ToLower(level), locale)
	pdf, err := cachedPDF("simulation", []string{level, locale}, exam, func() ([]byte, error) {
		return learningpdf.Exam(exam, locale)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "pdf_failed", "PDF belum dapat dibuat. Coba lagi.")
		return
	}
	writePDF(w, "NihongoMaster-Simulasi-"+level+".pdf", pdf)
}

// learningListLevel validates the level and applies the same study gate as
// the course itself (open to everyone unless open access is switched off).
func (s *Server) learningListLevel(w http.ResponseWriter, r *http.Request) (string, bool) {
	level := strings.ToUpper(r.PathValue("level"))
	if !jlptLevels[level] {
		writeError(w, http.StatusBadRequest, "invalid_level", "level harus N5, N4, N3, N2, atau N1")
		return "", false
	}
	if store.PremiumLevel(level) && s.store.Profile(userID(r)).Entitlement.Tier != domain.TierPremium {
		writeError(w, http.StatusPaymentRequired, "premium_required", "Materi N3–N1 memerlukan Lifetime Access.")
		return "", false
	}
	return level, true
}

// localizeLesson returns lesson as it is served by its own lesson route, so
// English users get exactly the translations they see inside the lesson
// (kanji translations depend on the lesson ID).
func localizeLesson[T any](lesson T, path, locale string) T {
	if locale != localeEnglish {
		return lesson
	}
	body, err := json.Marshal(map[string]any{"data": lesson})
	if err != nil {
		return lesson
	}
	translated, err := translateLearningJSONForPath(body, path)
	if err != nil {
		return lesson
	}
	var out struct{ Data T }
	if err := json.Unmarshal(translated, &out); err != nil {
		return lesson
	}
	return out.Data
}

func (s *Server) localizedGrammarLevel(userID, level, locale string) []domain.GrammarLesson {
	lessons := s.store.GrammarLevelLessons(userID, level)
	for i := range lessons {
		lessons[i] = localizeLesson(lessons[i], "/v1/grammar/lessons/"+lessons[i].ID, locale)
	}
	return lessons
}

func (s *Server) localizedKanjiLevel(userID, level, locale string) []domain.KanjiLesson {
	lessons := s.store.KanjiLevelLessons(userID, level)
	for i := range lessons {
		lessons[i] = localizeLesson(lessons[i], "/v1/kanji/lessons/"+lessons[i].ID, locale)
	}
	return lessons
}

func (s *Server) grammarLevelPoints(w http.ResponseWriter, r *http.Request) {
	level, ok := s.learningListLevel(w, r)
	if !ok {
		return
	}
	lessons := s.localizedGrammarLevel(userID(r), level, aiLocale(r))
	writeData(w, http.StatusOK, map[string]any{"level": level, "lessons": lessons})
}

func (s *Server) kanjiLevelCharacters(w http.ResponseWriter, r *http.Request) {
	level, ok := s.learningListLevel(w, r)
	if !ok {
		return
	}
	lessons := s.localizedKanjiLevel(userID(r), level, aiLocale(r))
	writeData(w, http.StatusOK, map[string]any{"level": level, "lessons": lessons})
}

func (s *Server) learningLevelPDF(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		level, ok := s.learningListLevel(w, r)
		if !ok {
			return
		}
		if !s.requireLifetimeDownload(w, r) {
			return
		}
		locale := aiLocale(r)
		var (
			content any
			render  func() ([]byte, error)
		)
		if kind == "grammar" {
			lessons := s.localizedGrammarLevel("", level, locale)
			content = lessons
			render = func() ([]byte, error) { return learningpdf.Grammar(level, locale, lessons) }
		} else {
			lessons := s.localizedKanjiLevel("", level, locale)
			content = lessons
			render = func() ([]byte, error) { return learningpdf.Kanji(level, locale, lessons) }
		}
		pdf, err := cachedPDF(kind, []string{level, locale}, content, render)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "pdf_failed", "PDF belum dapat dibuat. Coba lagi.")
			return
		}
		name := fmt.Sprintf("NihongoMaster-%s-%s.pdf", map[string]string{"grammar": "Grammar", "kanji": "Kanji"}[kind], level)
		writePDF(w, name, pdf)
	}
}
