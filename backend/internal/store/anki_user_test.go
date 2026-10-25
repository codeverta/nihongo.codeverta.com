package store

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/ankitransfer"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/testdb"
)

var sampleAnkiPNG = func() []byte {
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	return b.Bytes()
}()
var sampleAnkiMP3 = append([]byte("ID3\x04\x00\x00\x00\x00\x00\x00"), make([]byte, 30)...)

func sampleUserAnkiDocument(sourceID string) ankitransfer.Document {
	return ankitransfer.Document{
		Version:     ankitransfer.DocumentVersion,
		Title:       "Personal kana deck",
		Description: "A private deck",
		SourceID:    sourceID,
		Level:       "N5",
		Cards: []ankitransfer.Card{
			{Front: "猫", Answers: []string{"ねこ"}, Meaning: "cat", ImageURL: "media:CAT.png", Audio: "media:cat.mp3"},
			{Front: "犬", Answers: []string{"いぬ"}, Meaning: "dog", ImageURL: "https://example.test/dog.png"},
		},
		Media: []ankitransfer.Media{
			{Name: "cat.png", MIME: "image/png", Data: sampleAnkiPNG},
			{Name: "cat.mp3", MIME: "audio/mpeg", Data: sampleAnkiMP3},
		},
	}
}

func TestPrivateAnkiImportSRSOwnershipAndExport(t *testing.T) {
	s := NewMemoryStore()
	now := time.Date(2026, 10, 25, 8, 0, 0, 0, time.UTC)
	s.gamificationNow = func() time.Time { return now }
	user := database.DemoUserID
	other, err := s.AuthenticateGoogle(domain.GoogleIdentity{Subject: "private-anki-other", Email: "private-anki-other@example.test", DisplayName: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	requestID := uuid.NewString()
	doc := sampleUserAnkiDocument("")
	view, err := s.ImportAnkiDeck(user, doc, requestID)
	if err != nil {
		t.Fatal(err)
	}
	if !view.Personal || view.RequiredTier != "free" || len(view.Preview) != 2 || !strings.HasPrefix(view.ID, "user-") {
		t.Fatalf("unexpected imported deck view: %+v", view)
	}
	if view.Preview[0].ImageURL != "/anki/decks/"+view.ID+"/media/cat.png" {
		t.Fatalf("media reference was not canonicalized: %q", view.Preview[0].ImageURL)
	}
	// A retry with the same request ID returns the original deck even if the
	// retried payload differs, so a lost response cannot create a duplicate.
	retryDoc := sampleUserAnkiDocument("")
	retryDoc.Title = "Changed during retry"
	retried, err := s.ImportAnkiDeck(user, retryDoc, requestID)
	if err != nil || retried.ID != view.ID {
		t.Fatalf("import retry created a second deck: %+v, %v", retried, err)
	}
	decks, err := s.UserAnkiDecks(user)
	if err != nil || len(decks) != 1 || decks[0].ID != view.ID {
		t.Fatalf("private deck listing is wrong: %+v, %v", decks, err)
	}
	if _, err := s.AnkiDeck(other.ID, view.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other account read private deck: %v", err)
	}
	if _, err := s.ExportAnkiDeck(other.ID, view.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other account exported private deck: %v", err)
	}
	if _, _, err := s.AnkiDeckMedia(other.ID, view.ID, "cat.png"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other account read private media: %v", err)
	}

	session, err := s.AnkiStudy(user, view.ID, 2)
	if err != nil || len(session.Cards) != 2 {
		t.Fatalf("could not study private deck: %+v, %v", session, err)
	}
	if session.Cards[0].ImageURL != "/anki/decks/"+view.ID+"/media/cat.png" || session.Cards[0].AudioURL != "/anki/decks/"+view.ID+"/media/cat.mp3" {
		t.Fatalf("study card media refs were not rewritten: %+v", session.Cards[0])
	}
	input := AnkiReviewInput{Index: 0, Version: 0, Rating: "good", RequestID: uuid.NewString()}
	first, err := s.AnkiReview(user, view.ID, input)
	if err != nil || first.Version != 1 {
		t.Fatalf("private review failed: %+v, %v", first, err)
	}
	retriedReview, err := s.AnkiReview(user, view.ID, input)
	if err != nil || retriedReview.Version != first.Version || !retriedReview.Due.Equal(first.Due) {
		t.Fatalf("private review retry changed the schedule: %+v, %v", retriedReview, err)
	}
	view, err = s.AnkiDeck(user, view.ID)
	if err != nil || view.Stats.Seen != 1 {
		t.Fatalf("private review state was not saved: %+v, %v", view.Stats, err)
	}

	exported, err := s.ExportAnkiDeck(user, view.ID)
	if err != nil || exported.Title != doc.Title || len(exported.Cards) != len(doc.Cards) || len(exported.Media) != len(doc.Media) || exported.Cards[0].ImageURL != "media:cat.png" {
		t.Fatalf("private deck export did not preserve content: %+v, %v", exported, err)
	}
	media, mime, err := s.AnkiDeckMedia(user, view.ID, "cat.mp3")
	if err != nil || mime != "audio/mpeg" || !bytes.Equal(media, sampleAnkiMP3) {
		t.Fatalf("private deck media was not preserved: %q, %q, %v", media, mime, err)
	}
	if err := s.DeleteUserAnkiDeck(other.ID, view.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other account deleted private deck: %v", err)
	}
	if err := s.DeleteUserAnkiDeck(user, view.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AnkiDeck(user, view.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted private deck still exists: %v", err)
	}
}

func TestPrivateAnkiLimitsAndCatalogPremiumProvenance(t *testing.T) {
	s := NewMemoryStore()
	user := database.DemoUserID
	tooManyCards := sampleUserAnkiDocument("")
	tooManyCards.Cards = make([]ankitransfer.Card, maxUserAnkiCards+1)
	for i := range tooManyCards.Cards {
		tooManyCards.Cards[i] = ankitransfer.Card{Front: "front", Answers: []string{"answer"}}
	}
	if _, err := s.ImportAnkiDeck(user, tooManyCards, uuid.NewString()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("oversized deck was accepted: %v", err)
	}
	// The importer cannot downgrade a recognized premium source by changing its
	// supplied level field to a free level.
	premiumDoc := ankitransfer.Document{Version: ankitransfer.DocumentVersion, Title: "Copied catalog", SourceID: "n3", Level: "N5", Cards: []ankitransfer.Card{{Front: "猫", Answers: []string{"ねこ"}}}}
	premium, err := s.ImportAnkiDeck(user, premiumDoc, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	if premium.RequiredTier != "lifetime" || !premium.Locked || premium.Level != "N5" {
		t.Fatalf("premium provenance was lost or client level was trusted: %+v", premium)
	}
	if _, err := s.AnkiStudy(user, premium.ID, 1); !errors.Is(err, ErrPremium) {
		t.Fatalf("free account studied a premium-derived deck: %v", err)
	}

	limitStore := NewMemoryStore()
	for i := 0; i < maxUserAnkiDecks; i++ {
		doc := ankitransfer.Document{Version: ankitransfer.DocumentVersion, Title: "Personal", Cards: []ankitransfer.Card{{Front: "猫", Answers: []string{"ねこ"}}}}
		if _, err := limitStore.ImportAnkiDeck(user, doc, uuid.NewString()); err != nil {
			t.Fatalf("deck %d was rejected: %v", i+1, err)
		}
	}
	personal := ankitransfer.Document{Version: ankitransfer.DocumentVersion, Title: "Too many", Cards: []ankitransfer.Card{{Front: "犬", Answers: []string{"いぬ"}}}}
	if _, err := limitStore.ImportAnkiDeck(user, personal, uuid.NewString()); !errors.Is(err, ErrAnkiQuota) {
		t.Fatalf("deck limit was not enforced: %v", err)
	}
}

func TestPrivateAnkiPersistenceRetryDeleteAndFailedImport(t *testing.T) {
	dsn := testdb.DSN(t)
	user := database.DemoUserID
	s := NewMemoryStore()
	if err := s.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	requestID := uuid.NewString()
	view, err := s.ImportAnkiDeck(user, sampleUserAnkiDocument(""), requestID)
	if err != nil {
		t.Fatal(err)
	}
	input := AnkiReviewInput{Index: 0, Version: 0, Rating: "good", RequestID: uuid.NewString()}
	first, err := s.AnkiReview(user, view.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	restored := NewMemoryStore()
	if err := restored.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	retried, err := restored.ImportAnkiDeck(user, sampleUserAnkiDocument(""), requestID)
	if err != nil || retried.ID != view.ID {
		t.Fatalf("import retry after restart created a duplicate: %+v, %v", retried, err)
	}
	view, err = restored.AnkiDeck(user, view.ID)
	if err != nil || view.Stats.Seen != 1 {
		t.Fatalf("deck or progress did not survive restart: %+v, %v", view, err)
	}
	media, mime, err := restored.AnkiDeckMedia(user, view.ID, "cat.png")
	if err != nil || mime != "image/png" || !bytes.Equal(media, sampleAnkiPNG) {
		t.Fatalf("media did not survive restart: %q, %q, %v", media, mime, err)
	}
	second, err := restored.AnkiReview(user, view.ID, input)
	if err != nil || second.Version != first.Version || !second.Due.Equal(first.Due) {
		t.Fatalf("review retry after restart changed the schedule: %+v, %v", second, err)
	}

	if err := restored.persistenceConn.Close(); err != nil {
		t.Fatal(err)
	}
	failedRequest := uuid.NewString()
	if _, err := restored.ImportAnkiDeck(user, sampleUserAnkiDocument(""), failedRequest); err == nil {
		t.Fatal("failed database import was hidden")
	}
	decks, err := restored.UserAnkiDecks(user)
	if err != nil || len(decks) != 1 {
		t.Fatalf("failed import changed the in-memory list: %+v, %v", decks, err)
	}
	// Reopen a healthy connection through a separate store so durable account
	// deletion can be checked after a successful transaction.
	_ = restored.Close()
	cleanup := NewMemoryStore()
	if err := cleanup.EnableGamificationPersistence(dsn); err != nil {
		t.Fatal(err)
	}
	defer cleanup.Close()
	if _, err := cleanup.AnkiDeck(user, view.ID); err != nil {
		t.Fatal(err)
	}
	if err := cleanup.DeleteAccount(user); err != nil {
		t.Fatal(err)
	}
	var deckRows, mediaRows int64
	if err := cleanup.ormDB.Model(&database.UserAnkiDeck{}).Where("user_id = ?", user).Count(&deckRows).Error; err != nil {
		t.Fatal(err)
	}
	if err := cleanup.ormDB.Model(&database.UserAnkiDeckMedia{}).Where("user_id = ?", user).Count(&mediaRows).Error; err != nil {
		t.Fatal(err)
	}
	if deckRows != 0 || mediaRows != 0 {
		t.Fatalf("account deletion retained private Anki data: decks=%d media=%d", deckRows, mediaRows)
	}
}
