package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

type listEnv struct {
	handler http.Handler
	memory  *store.MemoryStore
	session domain.AuthSession
}

func newListEnv(t *testing.T) *listEnv {
	t.Helper()
	memory := store.NewMemoryStore()
	t.Cleanup(func() { _ = memory.Close() })
	handler := api.NewServer(memory, nil).Handler()
	return &listEnv{handler: handler, memory: memory, session: aiLogin(t, handler)}
}

func (e *listEnv) get(path, language string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("Authorization", "Bearer "+e.session.Token)
	if language != "" {
		request.Header.Set("Accept-Language", language)
	}
	response := httptest.NewRecorder()
	e.handler.ServeHTTP(response, request)
	return response
}

func (e *listEnv) makeLifetime(t *testing.T) {
	t.Helper()
	token, err := e.memory.BillingAccountToken(e.session.User.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.memory.SaveVerifiedPurchase(e.session.User.ID, domain.Purchase{
		Platform: "android", ID: "list-" + e.session.User.ID, ProductID: "nihongo_master_lifetime",
		AccountToken: token, Proof: "list-test", State: "purchased", CheckedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGrammarLevelListReturnsEveryPointInLessonOrder(t *testing.T) {
	env := newListEnv(t)
	response := env.get("/v1/grammar/levels/n5/points", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data struct {
			Level   string
			Lessons []domain.GrammarLesson
		}
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	course := env.memory.GrammarCourse("", "N5")
	if payload.Data.Level != "N5" || len(payload.Data.Lessons) != len(course.Lessons) {
		t.Fatalf("lessons = %d, course = %d", len(payload.Data.Lessons), len(course.Lessons))
	}
	for i, lesson := range payload.Data.Lessons {
		if lesson.ID != course.Lessons[i].ID || len(lesson.Points) != course.Lessons[i].PointCount || len(lesson.Points) == 0 {
			t.Fatalf("lesson %d mismatch: %s vs %s", i, lesson.ID, course.Lessons[i].ID)
		}
	}
}

func TestKanjiLevelListMatchesEachLessonIncludingEnglish(t *testing.T) {
	t.Setenv("ENGLISH_QA_PREVIEW_ENABLED", "true")
	env := newListEnv(t)
	for _, language := range []string{"", "en"} {
		response := env.get("/v1/kanji/levels/N5/characters", language)
		if response.Code != http.StatusOK {
			t.Fatalf("status %d", response.Code)
		}
		var payload struct {
			Data struct{ Lessons []domain.KanjiLesson }
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Data.Lessons) == 0 {
			t.Fatal("no kanji lessons")
		}
		// The list must show exactly what the lesson itself shows.
		first := payload.Data.Lessons[0]
		lesson := env.get("/v1/kanji/lessons/"+first.ID, language)
		var single struct{ Data domain.KanjiLesson }
		if err := json.Unmarshal(lesson.Body.Bytes(), &single); err != nil {
			t.Fatal(err)
		}
		if len(single.Data.Characters) != len(first.Characters) || single.Data.Characters[0].Meaning != first.Characters[0].Meaning {
			t.Fatalf("%q: list %q vs lesson %q", language, first.Characters[0].Meaning, single.Data.Characters[0].Meaning)
		}
	}
}

func TestLearningListRejectsUnknownLevels(t *testing.T) {
	env := newListEnv(t)
	for _, path := range []string{"/v1/grammar/levels/N6/points", "/v1/kanji/levels/x/characters", "/v1/kanji/levels/N0/pdf"} {
		if response := env.get(path, ""); response.Code != http.StatusBadRequest {
			t.Fatalf("%s: %d", path, response.Code)
		}
	}
}

func TestLearningPDFIsLifetimeOnly(t *testing.T) {
	env := newListEnv(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/grammar/levels/N5/pdf", nil)
	response := httptest.NewRecorder()
	env.handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: %d", response.Code)
	}
	if response := env.get("/v1/grammar/levels/N5/pdf", ""); response.Code != http.StatusPaymentRequired || !strings.Contains(response.Body.String(), "premium_required") {
		t.Fatalf("free: %d %s", response.Code, response.Body.String())
	}
	env.makeLifetime(t)
	for _, path := range []string{"/v1/grammar/levels/N5/pdf", "/v1/kanji/levels/N4/pdf"} {
		response := env.get(path, "")
		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/pdf" || !bytes.HasPrefix(response.Body.Bytes(), []byte("%PDF")) {
			t.Fatalf("%s: %d %q", path, response.Code, response.Header().Get("Content-Type"))
		}
		if disposition := response.Header().Get("Content-Disposition"); !strings.Contains(disposition, "NihongoMaster-") {
			t.Fatalf("disposition = %q", disposition)
		}
	}
}

func TestSimulationPDFIsLifetimeOnly(t *testing.T) {
	env := newListEnv(t)
	if response := env.get("/v1/simulation-pdf/N5", ""); response.Code != http.StatusPaymentRequired {
		t.Fatalf("free: %d", response.Code)
	}
	if response := env.get("/v1/simulation-pdf/N9", ""); response.Code != http.StatusBadRequest {
		t.Fatalf("unknown level: %d", response.Code)
	}
	env.makeLifetime(t)
	for _, level := range []string{"N5", "n3-4", "JFT-2", "NAT", "NAT-3"} {
		response := env.get("/v1/simulation-pdf/"+level, "")
		if response.Code != http.StatusOK || !bytes.HasPrefix(response.Body.Bytes(), []byte("%PDF")) {
			t.Fatalf("%s: %d %s", level, response.Code, response.Body.String())
		}
		if disposition := response.Header().Get("Content-Disposition"); !strings.Contains(disposition, "NihongoMaster-Simulasi-") {
			t.Fatalf("disposition = %q", disposition)
		}
	}
}
