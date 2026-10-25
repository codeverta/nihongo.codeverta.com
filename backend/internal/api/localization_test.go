package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

func TestLearningAPISelectsLocaleFromAcceptLanguageWithoutChangingJapaneseSource(t *testing.T) {
	t.Setenv("ENGLISH_QA_PREVIEW_ENABLED", "true")
	server := newServer()
	token := login(t, server)
	request := func(path, acceptLanguage string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.Header.Set("Authorization", "Bearer "+token)
		if acceptLanguage != "" {
			r.Header.Set("Accept-Language", acceptLanguage)
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	type vocabularyPayload struct {
		Data struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Words       []struct {
				Term           string `json:"term"`
				Meaning        string `json:"meaning"`
				Example        string `json:"example"`
				ExampleMeaning string `json:"exampleMeaning"`
			} `json:"words"`
		} `json:"data"`
	}
	decodeVocabulary := func(w *httptest.ResponseRecorder) vocabularyPayload {
		t.Helper()
		var result vocabularyPayload
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if w.Code != http.StatusOK || len(result.Data.Words) == 0 {
			t.Fatalf("invalid learning response: %d %s", w.Code, w.Body.String())
		}
		return result
	}

	indonesian := request("/v1/vocabulary/lessons/n5-daily", "")
	idPayload := decodeVocabulary(indonesian)
	if idPayload.Data.Words[0].Meaning != "saya" {
		t.Fatalf("requests without a locale must retain Indonesian for existing clients; got %q", idPayload.Data.Words[0].Meaning)
	}

	english := request("/v1/vocabulary/lessons/n5-daily", "en-US,en;q=0.9,id;q=0.5")
	enPayload := decodeVocabulary(english)
	if enPayload.Data.Title != "Self-introductions" || enPayload.Data.Description != "Learn how to introduce yourself through images, two examples for each word, and practice with meanings, readings, and sentences." || enPayload.Data.Words[0].Meaning != "I; me" || enPayload.Data.Words[0].ExampleMeaning != "I am a student." {
		t.Fatalf("English learning content was not selected: %#v", enPayload.Data)
	}
	expectedMeanings := []string{"I; me", "student", "teacher", "friend", "child", "name", "Japan", "abroad; a foreign country"}
	if len(enPayload.Data.Words) != len(expectedMeanings) {
		t.Fatalf("expected all eight lesson words, got %d", len(enPayload.Data.Words))
	}
	for index, expected := range expectedMeanings {
		if enPayload.Data.Words[index].Meaning != expected {
			t.Fatalf("word %d meaning was not translated: got %q, want %q", index, enPayload.Data.Words[index].Meaning, expected)
		}
	}
	if enPayload.Data.Words[0].Term != "私" || enPayload.Data.Words[0].Example != "私は学生です。" {
		t.Fatalf("Japanese study text must be preserved: %#v", enPayload.Data.Words[0])
	}
	if got := english.Header().Get("Vary"); got == "" {
		t.Fatal("localized responses must vary by Accept-Language")
	}
	if got := english.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("localized learning response must not be shared across locales, got Cache-Control %q", got)
	}
	readingRequest := request("/v1/reading/lessons/reading-n5-12", "en")
	var readingPayload map[string]any
	if err := json.Unmarshal(readingRequest.Body.Bytes(), &readingPayload); err != nil {
		t.Fatal(err)
	}
	var foundReadingMeaning bool
	var visitReading func(any)
	visitReading = func(value any) {
		switch item := value.(type) {
		case map[string]any:
			if item["meaning"] == "tidy up" {
				foundReadingMeaning = true
			}
			for _, child := range item {
				visitReading(child)
			}
		case []any:
			for _, child := range item {
				visitReading(child)
			}
		}
	}
	visitReading(readingPayload)
	if !foundReadingMeaning {
		t.Fatalf("Reading material should use its context-specific glossary translation, got %s", readingRequest.Body.String())
	}
	// A different request on the same server must still receive the source locale.
	if got := decodeVocabulary(request("/v1/vocabulary/lessons/n5-daily", "id-ID,id;q=0.9")).Data.Words[0].Meaning; got != "saya" {
		t.Fatalf("English content leaked into a later Indonesian request: %q", got)
	}

	grammarRequest := request("/v1/grammar/lessons/grammar-n5-01", "en")
	var grammar struct {
		Data struct {
			Chapter string `json:"chapter"`
			Title   string `json:"title"`
			Points  []struct {
				Meaning  string `json:"meaning"`
				Examples []struct {
					Japanese string `json:"japanese"`
					Meaning  string `json:"meaning"`
				} `json:"examples"`
			} `json:"points"`
		} `json:"data"`
	}
	if err := json.Unmarshal(grammarRequest.Body.Bytes(), &grammar); err != nil {
		t.Fatal(err)
	}
	if grammarRequest.Code != http.StatusOK || grammar.Data.Chapter != "Basic sentences" || grammar.Data.Title != "Identity and negation" || grammar.Data.Points[0].Meaning != "is …" {
		t.Fatalf("English grammar content was not selected: %s", grammarRequest.Body.String())
	}
	if got := grammar.Data.Points[0].Examples[0]; got.Japanese != "私は学生です。" || got.Meaning != "I am a student." {
		t.Fatalf("grammar translation must preserve Japanese source text: %#v", got)
	}
	quizRequest := httptest.NewRequest(http.MethodPost, "/v1/grammar/lessons/grammar-n5-01/attempts", nil)
	quizRequest.Header.Set("Authorization", "Bearer "+token)
	quizRequest.Header.Set("Accept-Language", "en")
	quizResponse := httptest.NewRecorder()
	server.ServeHTTP(quizResponse, quizRequest)
	var quiz struct {
		Data struct {
			Questions []struct {
				Prompt  string `json:"prompt"`
				Options []struct {
					Text string `json:"text"`
				} `json:"options"`
			} `json:"questions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(quizResponse.Body.Bytes(), &quiz); err != nil {
		t.Fatal(err)
	}
	if quizResponse.Code != http.StatusCreated || len(quiz.Data.Questions) == 0 || quiz.Data.Questions[0].Prompt != "私は学生（　）。 Complete the polite sentence." {
		t.Fatalf("quiz instructions were not localized: %d %s", quizResponse.Code, quizResponse.Body.String())
	}
	containsJapaneseAnswer := false
	for _, option := range quiz.Data.Questions[0].Options {
		if option.Text == "です" {
			containsJapaneseAnswer = true
		}
	}
	if !containsJapaneseAnswer {
		t.Fatalf("Japanese quiz choices must stay intact: %s", quizResponse.Body.String())
	}
	vocabularyQuizRequest := httptest.NewRequest(http.MethodPost, "/v1/vocabulary/lessons/n5-daily/attempts", nil)
	vocabularyQuizRequest.Header.Set("Authorization", "Bearer "+token)
	vocabularyQuizRequest.Header.Set("Accept-Language", "en")
	vocabularyQuizResponse := httptest.NewRecorder()
	server.ServeHTTP(vocabularyQuizResponse, vocabularyQuizRequest)
	var vocabularyQuiz struct {
		Data struct {
			ID        string `json:"id"`
			Questions []struct {
				ID               string `json:"id"`
				Kind             string `json:"kind"`
				ImageDescription string `json:"imageDescription"`
				Prompt           string `json:"prompt"`
				Options          []struct {
					ID   string `json:"id"`
					Text string `json:"text"`
				} `json:"options"`
			} `json:"questions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(vocabularyQuizResponse.Body.Bytes(), &vocabularyQuiz); err != nil {
		t.Fatal(err)
	}
	if vocabularyQuizResponse.Code != http.StatusCreated || len(vocabularyQuiz.Data.Questions) == 0 {
		t.Fatalf("English vocabulary quiz could not be started: %d %s", vocabularyQuizResponse.Code, vocabularyQuizResponse.Body.String())
	}
	firstVocabularyQuestion := vocabularyQuiz.Data.Questions[0]
	if firstVocabularyQuestion.Kind != "Word meaning" || firstVocabularyQuestion.Prompt != "What does 私 mean?" || firstVocabularyQuestion.ImageDescription != "Vocabulary context clue illustration" {
		t.Fatalf("vocabulary quiz labels and prompts should use reviewed English while retaining Japanese: %#v", firstVocabularyQuestion)
	}
	foundTranslatedVocabularyChoice := false
	for _, option := range firstVocabularyQuestion.Options {
		if option.Text == "I; me" {
			foundTranslatedVocabularyChoice = true
		}
	}
	if !foundTranslatedVocabularyChoice {
		t.Fatalf("vocabulary quiz meanings should be translated without changing the answer ID: %#v", firstVocabularyQuestion.Options)
	}
	answerBody, err := json.Marshal(map[string]string{"questionId": firstVocabularyQuestion.ID, "optionId": firstVocabularyQuestion.Options[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	vocabularyAnswerRequest := httptest.NewRequest(http.MethodPost, "/v1/vocabulary/attempts/"+vocabularyQuiz.Data.ID+"/answers", bytes.NewReader(answerBody))
	vocabularyAnswerRequest.Header.Set("Authorization", "Bearer "+token)
	vocabularyAnswerRequest.Header.Set("Accept-Language", "en")
	vocabularyAnswerResponse := httptest.NewRecorder()
	server.ServeHTTP(vocabularyAnswerResponse, vocabularyAnswerRequest)
	var vocabularyFeedback struct {
		Data struct {
			Explanation string `json:"explanation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(vocabularyAnswerResponse.Body.Bytes(), &vocabularyFeedback); err != nil {
		t.Fatal(err)
	}
	if vocabularyAnswerResponse.Code != http.StatusOK || !strings.Contains(vocabularyFeedback.Data.Explanation, "means I; me.") || !strings.Contains(vocabularyFeedback.Data.Explanation, "私は学生です。") || strings.Contains(vocabularyFeedback.Data.Explanation, "Saya adalah pelajar.") {
		t.Fatalf("vocabulary quiz feedback should be English and retain Japanese examples: %d %s", vocabularyAnswerResponse.Code, vocabularyAnswerResponse.Body.String())
	}
	kanjiQuizRequest := httptest.NewRequest(http.MethodPost, "/v1/kanji/lessons/kanji-n4-01/attempts", nil)
	kanjiQuizRequest.Header.Set("Authorization", "Bearer "+token)
	kanjiQuizRequest.Header.Set("Accept-Language", "en")
	kanjiQuizResponse := httptest.NewRecorder()
	server.ServeHTTP(kanjiQuizResponse, kanjiQuizRequest)
	var kanjiQuiz struct {
		Data struct {
			Questions []struct {
				Prompt  string `json:"prompt"`
				Options []struct {
					Text string `json:"text"`
				} `json:"options"`
			} `json:"questions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(kanjiQuizResponse.Body.Bytes(), &kanjiQuiz); err != nil {
		t.Fatal(err)
	}
	if kanjiQuizResponse.Code != http.StatusCreated || len(kanjiQuiz.Data.Questions) == 0 {
		t.Fatalf("English Kanji quiz could not be started: %d %s", kanjiQuizResponse.Code, kanjiQuizResponse.Body.String())
	}
	if first := kanjiQuiz.Data.Questions[0]; first.Prompt != "What is the main meaning of the kanji 親 on this card?" {
		t.Fatalf("Kanji prompt should be English while retaining the Japanese character: %q", first.Prompt)
	} else {
		foundEnglishChoice := false
		for _, option := range first.Options {
			if option.Text == "parent; close" {
				foundEnglishChoice = true
			}
		}
		if !foundEnglishChoice {
			t.Fatalf("Kanji answer choices should use their reviewed English translation: %#v", first.Options)
		}
	}
	if unsupported := decodeVocabulary(request("/v1/vocabulary/lessons/n5-daily", "fr-CA")); unsupported.Data.Words[0].Meaning != "saya" {
		t.Fatal("unsupported API languages must keep the Indonesian compatibility fallback")
	}
}

func TestEnglishOfflinePracticePackageIsLocalizedBeforeSigningAndVariesByLocale(t *testing.T) {
	t.Setenv("ENGLISH_QA_PREVIEW_ENABLED", "true")
	server := offlinePracticeServer(store.NewMemoryStore(), fakeGoogleVerifier{})
	token := login(t, server)
	fetch := func(acceptLanguage string) (*httptest.ResponseRecorder, domain.OfflinePracticePackage) {
		r := httptest.NewRequest(http.MethodGet, "/v1/offline/practice/grammar/grammar-n5-01", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Accept-Language", acceptLanguage)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		var payload struct {
			Data domain.OfflinePracticePackage `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode offline package: %v: %s", err, w.Body.String())
		}
		return w, payload.Data
	}
	englishResponse, english := fetch("en")
	if englishResponse.Code != http.StatusOK || len(english.Questions) == 0 || english.Questions[0].Prompt != "私は学生（　）。 Complete the polite sentence." {
		t.Fatalf("offline package was not localized before signing: %d %s", englishResponse.Code, englishResponse.Body.String())
	}
	if englishResponse.Header().Get("ETag") != `"`+english.Manifest.ContentHash+`"` || englishResponse.Header().Get("Content-Language") != "en" {
		t.Fatalf("English cache metadata does not identify the localized package: %#v", englishResponse.Header())
	}
	indonesianResponse, indonesian := fetch("id")
	if indonesianResponse.Code != http.StatusOK || english.Manifest.ContentHash == indonesian.Manifest.ContentHash {
		t.Fatalf("locale-specific offline packages must have distinct content hashes: en=%s id=%s", english.Manifest.ContentHash, indonesian.Manifest.ContentHash)
	}

	input := domain.OfflinePracticeSyncInput{AttemptID: uuid.NewString(), Package: english}
	for _, question := range english.Questions {
		input.Answers = append(input.Answers, domain.OfflinePracticeAnswer{QuestionID: question.ID, OptionID: question.CorrectOptionID})
	}
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/offline/practice-attempts/sync", bytes.NewReader(body))
	syncRequest.Header.Set("Authorization", "Bearer "+token)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusOK {
		t.Fatalf("localized signed package did not validate for offline sync: %d %s", syncResponse.Code, syncResponse.Body.String())
	}
}

func TestLearningAPIKeepsEnglishBehindTheReviewGate(t *testing.T) {
	t.Setenv("PUBLIC_ENGLISH_CONTENT_ENABLED", "false")
	t.Setenv("ENGLISH_QA_PREVIEW_ENABLED", "false")
	server := newServer()
	token := login(t, server)
	r := httptest.NewRequest(http.MethodGet, "/v1/vocabulary/lessons/n5-daily", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Accept-Language", "en")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	var payload struct {
		Data struct {
			Words []struct {
				Meaning string `json:"meaning"`
			} `json:"words"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || len(payload.Data.Words) == 0 || payload.Data.Words[0].Meaning != "saya" || w.Header().Get("Content-Language") != "id" {
		t.Fatalf("English API responses must stay unavailable before review: %d %s (%q)", w.Code, w.Body.String(), w.Header().Get("Content-Language"))
	}
}

func TestEnglishLearningErrorMessageMatchesContentLanguage(t *testing.T) {
	t.Setenv("PUBLIC_ENGLISH_CONTENT_ENABLED", "false")
	t.Setenv("ENGLISH_QA_PREVIEW_ENABLED", "true")
	server := newServer()
	r := httptest.NewRequest(http.MethodGet, "/v1/vocabulary/lessons/n5-daily", nil)
	r.Header.Set("Authorization", "Bearer expired-token")
	r.Header.Set("Accept-Language", "en")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusUnauthorized || payload.Error.Code != "unauthorized" || payload.Error.Message != "Your session is invalid or has expired. Please sign in again." || w.Header().Get("Content-Language") != "en" {
		t.Fatalf("English API errors must match their Content-Language: %d %#v (%q)", w.Code, payload.Error, w.Header().Get("Content-Language"))
	}
}
