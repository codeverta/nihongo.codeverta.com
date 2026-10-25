package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

const offlinePracticeSchemaVersion = 1

type offlinePracticeContent struct {
	Kind      string                           `json:"kind"`
	LessonID  string                           `json:"lessonId"`
	Level     string                           `json:"level"`
	Questions []domain.OfflinePracticeQuestion `json:"questions"`
}

// ConfigureOfflinePracticeSigningKey allows deployments and API tests to
// inject the dedicated secret. NewServer also reads OFFLINE_PRACTICE_SIGNING_KEY
// and falls back to the required MariaDB password, deriving a separate key.
func (s *Server) ConfigureOfflinePracticeSigningKey(secret string) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		s.offlinePracticeSigningKey = nil
		return
	}
	sum := sha256.Sum256([]byte("nihongo-master:offline-practice:" + secret))
	s.offlinePracticeSigningKey = sum[:]
}

func defaultOfflinePracticeSigningKey(configured, databasePassword string) []byte {
	secret := strings.TrimSpace(configured)
	if secret == "" {
		secret = strings.TrimSpace(databasePassword)
	}
	if secret == "" {
		return nil
	}
	sum := sha256.Sum256([]byte("nihongo-master:offline-practice:" + secret))
	return sum[:]
}

func offlinePracticeSigningKey(configured, keyFile, databasePassword string) []byte {
	if path := strings.TrimSpace(keyFile); path != "" {
		secret, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		return defaultOfflinePracticeSigningKey(string(secret), "")
	}
	return defaultOfflinePracticeSigningKey(configured, databasePassword)
}

func (s *Server) offlinePracticePackage(w http.ResponseWriter, r *http.Request) {
	if len(s.offlinePracticeSigningKey) == 0 {
		writeError(w, http.StatusServiceUnavailable, "offline_practice_unavailable", "Kunci tanda tangan paket latihan offline belum dikonfigurasi.")
		return
	}
	practice, err := s.store.OfflinePracticePackage(userID(r), r.PathValue("kind"), r.PathValue("lessonID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if requestContentLanguage(r.Header.Get("Accept-Language")) == localeEnglish {
		raw, marshalErr := json.Marshal(practice)
		if marshalErr != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "Paket latihan tidak dapat disiapkan.")
			return
		}
		localized, localizationErr := localizePracticePackage(raw)
		if localizationErr != nil || json.Unmarshal(localized, &practice) != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "Paket latihan tidak dapat disiapkan.")
			return
		}
	}
	if len(practice.Questions) == 0 {
		writeError(w, http.StatusNotFound, "practice_unavailable", "Kuis lesson belum tersedia.")
		return
	}
	contentHash, err := offlinePracticeContentHash(practice)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Paket latihan tidak dapat ditandatangani.")
		return
	}
	practice.Manifest.SchemaVersion = offlinePracticeSchemaVersion
	practice.Manifest.ContentHash = contentHash
	practice.Manifest.Version = "v1-" + contentHash[:16]
	practice.Manifest.PackageID = "offline-" + contentHash[:32]
	practice.Manifest.UserID = userID(r)
	practice.Manifest.IssuedAt = time.Now().UTC()
	practice.Manifest.Signature, err = signOfflinePracticeManifest(s.offlinePracticeSigningKey, practice.Manifest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Paket latihan tidak dapat ditandatangani.")
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=86400, stale-if-error=31536000")
	w.Header().Add("Vary", "Authorization")
	w.Header().Set("ETag", `"`+contentHash+`"`)
	writeData(w, http.StatusOK, practice)
}

func (s *Server) syncOfflinePracticeAttempt(w http.ResponseWriter, r *http.Request) {
	if len(s.offlinePracticeSigningKey) == 0 {
		writeError(w, http.StatusServiceUnavailable, "offline_practice_unavailable", "Kunci tanda tangan paket latihan offline belum dikonfigurasi.")
		return
	}
	var input domain.OfflinePracticeSyncInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Data latihan offline tidak valid.")
		return
	}
	if _, err := uuid.Parse(input.AttemptID); err != nil || uuid.MustParse(input.AttemptID) == uuid.Nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_attempt", "ID latihan offline tidak valid.")
		return
	}
	if err := s.verifyOfflinePracticePackage(userID(r), input.Package); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_practice_package", "Paket latihan offline tidak valid atau tidak cocok dengan akun ini.")
		return
	}
	fingerprint, err := offlinePracticeFingerprint(input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_attempt", "Jawaban latihan offline tidak valid.")
		return
	}
	result, err := s.store.SyncOfflinePracticeAttempt(userID(r), input, fingerprint)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeData(w, http.StatusOK, result)
}

func (s *Server) verifyOfflinePracticePackage(userID string, practice domain.OfflinePracticePackage) error {
	manifest := practice.Manifest
	if manifest.SchemaVersion != offlinePracticeSchemaVersion || manifest.UserID != userID || manifest.Kind == "" || manifest.LessonID == "" || manifest.Level == "" || manifest.IssuedAt.IsZero() || manifest.IssuedAt.After(time.Now().UTC().Add(5*time.Minute)) {
		return store.ErrInvalid
	}
	switch manifest.Kind {
	case "vocabulary", "grammar", "kanji", "reading", "ssw":
	default:
		return store.ErrInvalid
	}
	if len(practice.Questions) == 0 || len(practice.Questions) > 300 {
		return store.ErrInvalid
	}
	for _, question := range practice.Questions {
		if question.ID == "" || question.Prompt == "" || len(question.Prompt) > 4096 || question.Explanation == "" || len(question.Explanation) > 8192 || len(question.Options) < 2 || len(question.Options) > 64 {
			return store.ErrInvalid
		}
		foundCorrect := false
		seenOptions := make(map[string]bool, len(question.Options))
		for _, option := range question.Options {
			if option.ID == "" || option.Text == "" || len(option.Text) > 4096 || seenOptions[option.ID] {
				return store.ErrInvalid
			}
			seenOptions[option.ID] = true
			if option.ID == question.CorrectOptionID {
				foundCorrect = true
			}
		}
		if !foundCorrect {
			return store.ErrInvalid
		}
	}
	contentHash, err := offlinePracticeContentHash(practice)
	if err != nil || contentHash != manifest.ContentHash || manifest.Version != "v1-"+contentHash[:16] || manifest.PackageID != "offline-"+contentHash[:32] {
		return store.ErrInvalid
	}
	expected, err := signOfflinePracticeManifest(s.offlinePracticeSigningKey, manifest)
	if err != nil {
		return store.ErrInvalid
	}
	provided, err := base64.RawURLEncoding.DecodeString(manifest.Signature)
	if err != nil {
		return store.ErrInvalid
	}
	want, err := base64.RawURLEncoding.DecodeString(expected)
	if err != nil || !hmac.Equal(provided, want) {
		return store.ErrInvalid
	}
	return nil
}

func offlinePracticeContentHash(practice domain.OfflinePracticePackage) (string, error) {
	content := offlinePracticeContent{
		Kind: practice.Manifest.Kind, LessonID: practice.Manifest.LessonID,
		Level: practice.Manifest.Level, Questions: practice.Questions,
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func signOfflinePracticeManifest(key []byte, manifest domain.OfflinePracticeManifest) (string, error) {
	manifest.Signature = ""
	raw, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	if _, err = mac.Write(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func offlinePracticeFingerprint(input domain.OfflinePracticeSyncInput) (string, error) {
	answers := append([]domain.OfflinePracticeAnswer(nil), input.Answers...)
	sort.Slice(answers, func(i, j int) bool {
		if answers[i].QuestionID == answers[j].QuestionID {
			return answers[i].OptionID < answers[j].OptionID
		}
		return answers[i].QuestionID < answers[j].QuestionID
	})
	canonical, err := json.Marshal(struct {
		AttemptID   string                         `json:"attemptId"`
		PackageID   string                         `json:"packageId"`
		ContentHash string                         `json:"contentHash"`
		Answers     []domain.OfflinePracticeAnswer `json:"answers"`
	}{input.AttemptID, input.Package.Manifest.PackageID, input.Package.Manifest.ContentHash, answers})
	if err != nil {
		return "", fmt.Errorf("marshal offline practice fingerprint: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}
