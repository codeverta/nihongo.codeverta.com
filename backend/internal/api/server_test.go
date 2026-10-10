package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nihongo-master-backend/internal/api"
	backendAuth "nihongo-master-backend/internal/auth"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

type fakeGoogleVerifier struct {
	identity domain.GoogleIdentity
	err      error
}

func (f fakeGoogleVerifier) Verify(context.Context, string) (domain.GoogleIdentity, error) {
	return f.identity, f.err
}

func newServer(premium ...bool) http.Handler {
	memory := store.NewMemoryStore()
	if len(premium) > 0 && premium[0] {
		token, _ := memory.BillingAccountToken("demo-user")
		if err := memory.SaveVerifiedPurchase("demo-user", domain.Purchase{Platform: "android", ID: "test-purchase", ProductID: "nihongo_master_lifetime", AccountToken: token, State: "purchased", CheckedAt: time.Now()}); err != nil {
			panic(err)
		}
	}
	return api.NewServer(
		memory,
		fakeGoogleVerifier{identity: domain.GoogleIdentity{
			Subject:     "google-123",
			Email:       "google@example.com",
			DisplayName: "Google User",
		}},
	).Handler()
}

func TestGuestCanAccessHome(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/v1/home", nil)
	response := httptest.NewRecorder()

	newServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data domain.Dashboard `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.IsAuthenticated || payload.Data.Profile != nil {
		t.Fatal("guest dashboard must not expose a synthetic profile")
	}
	if len(payload.Data.Categories) == 0 || len(payload.Data.Tracks) == 0 || len(payload.Data.Plans) == 0 {
		t.Fatal("guest dashboard must contain the public catalog")
	}
}

func TestAuthenticatedHomeIncludesPersonalData(t *testing.T) {
	server := newServer()
	token := login(t, server)
	request := httptest.NewRequest(http.MethodGet, "/v1/home", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data domain.Dashboard `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Data.IsAuthenticated || payload.Data.Profile == nil {
		t.Fatal("authenticated dashboard must include its profile")
	}
	if payload.Data.Profile.Email != "demo@nihongomaster.app" {
		t.Fatalf("unexpected profile: %#v", payload.Data.Profile)
	}
}

func TestProtectedFeatureRequiresBearerToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/v1/me/profile", nil)
	response := httptest.NewRecorder()

	newServer().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestEmailLoginCreatesRevocableSession(t *testing.T) {
	server := newServer()
	token := login(t, server)

	profileRequest := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	profileRequest.Header.Set("Authorization", "Bearer "+token)
	profileResponse := httptest.NewRecorder()
	server.ServeHTTP(profileResponse, profileRequest)
	if profileResponse.Code != http.StatusOK {
		t.Fatalf("expected authenticated profile, got %d", profileResponse.Code)
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	logoutRequest.Header.Set("Authorization", "Bearer "+token)
	logoutResponse := httptest.NewRecorder()
	server.ServeHTTP(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", logoutResponse.Code)
	}

	revokedRequest := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	revokedRequest.Header.Set("Authorization", "Bearer "+token)
	revokedResponse := httptest.NewRecorder()
	server.ServeHTTP(revokedResponse, revokedRequest)
	if revokedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked token to return 401, got %d", revokedResponse.Code)
	}
}

func TestGoogleLoginUsesVerifiedIdentity(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/auth/google",
		bytes.NewBufferString(`{"idToken":"verified-by-fake"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
}

func TestInvalidGoogleTokenIsRejected(t *testing.T) {
	server := api.NewServer(
		store.NewMemoryStore(),
		fakeGoogleVerifier{err: backendAuth.ErrInvalidGoogleToken},
	).Handler()
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/auth/google",
		bytes.NewBufferString(`{"idToken":"bad"}`),
	)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestSaveProgressRejectsInvalidPercentage(t *testing.T) {
	server := newServer()
	token := login(t, server)
	body := bytes.NewBufferString(`{"completion":101,"lastPosition":20}`)
	request := httptest.NewRequest(http.MethodPut, "/v1/me/progress/lesson-n5-1", body)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", response.Code, response.Body.String())
	}
}

func TestFreeUserCannotStartPremiumExam(t *testing.T) {
	server := newServer()
	token := login(t, server)
	body := bytes.NewBufferString(`{"examId":"jlpt-n4-simulation-1"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/exam-attempts", body)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d: %s", response.Code, response.Body.String())
	}
}

func TestVocabularyCourseCoversEveryJLPTLevel(t *testing.T) {
	server := newServer()
	token := login(t, server)
	request := httptest.NewRequest(http.MethodGet, "/v1/vocabulary/levels", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data []domain.VocabularyLevel `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Data) != 5 {
		t.Fatalf("expected N5-N1, got %d levels", len(payload.Data))
	}
	for _, level := range payload.Data {
		if len(level.Lessons) == 0 || level.Lessons[0].WordCount < 4 {
			t.Fatalf("level %s has incomplete lesson content", level.Level)
		}
	}
}

func TestVocabularyAttemptEvaluatesAnswersPersistsProgressAndRetries(t *testing.T) {
	server := newServer()
	token := login(t, server)
	attempt := startVocabularyAttempt(t, server, token, "n5-daily")
	requestLesson := httptest.NewRequest(http.MethodGet, "/v1/vocabulary/lessons/n5-daily", nil)
	requestLesson.Header.Set("Authorization", "Bearer "+token)
	lessonResponse := httptest.NewRecorder()
	server.ServeHTTP(lessonResponse, requestLesson)
	var lesson struct {
		Data domain.VocabularyLesson `json:"data"`
	}
	if err := json.Unmarshal(lessonResponse.Body.Bytes(), &lesson); err != nil || lessonResponse.Code != 200 {
		t.Fatal("cannot load vocabulary lesson")
	}
	correctOption := func(index int, question domain.VocabularyQuestion) string {
		word := lesson.Data.Words[index%len(lesson.Data.Words)]
		text := word.Meaning
		switch question.Kind {
		case "Cara baca":
			text = word.Reading
		case "Penulisan kata":
			text = word.Term
		case "Pemakaian kalimat":
			text = word.Example
		}
		for _, option := range question.Options {
			if option.Text == text {
				return option.ID
			}
		}
		t.Fatalf("question has no matching answer for %s", question.ID)
		return ""
	}
	wrongCount := (len(attempt.Questions) + 3) / 4
	for index, question := range attempt.Questions {
		option := correctOption(index, question)
		if index < wrongCount {
			for _, choice := range question.Options {
				if choice.ID != option {
					option = choice.ID
					break
				}
			}
		}
		feedback := answerVocabulary(t, server, token, attempt.ID, question.ID, option)
		if feedback.Correct == (index < wrongCount) {
			t.Fatalf("unexpected correctness for question %d", index)
		}
		if feedback.CorrectOptionID == "" || feedback.Explanation == "" {
			t.Fatal("feedback must teach the correct answer")
		}
	}
	result := completeVocabulary(t, server, token, attempt.ID)
	expected := (len(attempt.Questions) - wrongCount) * 100 / len(attempt.Questions)
	if result.Score != expected || result.Passed || result.BestScore != expected {
		t.Fatalf("unexpected first result: %#v", result)
	}

	retry := startVocabularyAttempt(t, server, token, "n5-daily")
	for index, question := range retry.Questions {
		answerVocabulary(t, server, token, retry.ID, question.ID, correctOption(index, question))
	}
	retryResult := completeVocabulary(t, server, token, retry.ID)
	if retryResult.Score != 100 || !retryResult.Passed || retryResult.BestScore != 100 {
		t.Fatalf("unexpected retry result: %#v", retryResult)
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/vocabulary/levels", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if !strings.Contains(response.Body.String(), `"bestScore":100`) || !strings.Contains(response.Body.String(), `"completed":true`) {
		t.Fatalf("saved progress missing: %s", response.Body.String())
	}
}

func startVocabularyAttempt(t *testing.T, server http.Handler, token, lessonID string) domain.VocabularyAttempt {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/v1/vocabulary/lessons/"+lessonID+"/attempts", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("start attempt failed: %d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data domain.VocabularyAttempt `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Data
}

func answerVocabulary(t *testing.T, server http.Handler, token, attemptID, questionID, optionID string) domain.VocabularyAnswerFeedback {
	t.Helper()
	body, err := json.Marshal(map[string]string{"questionId": questionID, "optionId": optionID})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/vocabulary/attempts/"+attemptID+"/answers", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("answer failed: %d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data domain.VocabularyAnswerFeedback `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Data
}

func completeVocabulary(t *testing.T, server http.Handler, token, attemptID string) domain.VocabularyAttemptResult {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/v1/vocabulary/attempts/"+attemptID+"/complete", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("complete failed: %d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data domain.VocabularyAttemptResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Data
}

func login(t *testing.T, server http.Handler) string {
	t.Helper()
	body := bytes.NewBufferString(`{"email":"demo@nihongomaster.app","password":"Demo123!"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/auth/login", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data domain.AuthSession `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Token == "" {
		t.Fatal(errors.New("expected session token"))
	}
	return payload.Data.Token
}
