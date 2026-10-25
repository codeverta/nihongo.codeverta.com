package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/ankitransfer"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
)

const (
	maxUserAnkiDecks        = 50
	maxUserAnkiCards        = ankitransfer.MaxCards
	maxUserAnkiStorageBytes = 100 << 20
)

var ErrAnkiQuota = errors.New("personal Anki storage limit exceeded")

type userAnkiDeckRecord struct {
	ID           string
	UserID       string
	RequestID    string
	Revision     string
	Version      int
	Title        string
	Description  string
	SourceURL    string
	SourceID     string
	Level        string
	Cards        []ankitransfer.Card
	Media        map[string]userAnkiMedia
	StorageBytes int64
	CreatedAt    time.Time
}

type userAnkiMedia struct {
	MIME string
	Size int64
	Data []byte
}

type userAnkiMediaMetadata struct {
	UserID string `gorm:"column:user_id"`
	DeckID string `gorm:"column:deck_id"`
	Name   string `gorm:"column:name"`
	MIME   string `gorm:"column:mime"`
	Size   int64  `gorm:"column:size"`
}

func isUserAnkiDeckID(id string) bool { return strings.HasPrefix(id, "user-") }

func ankiUserDeckRequestKey(user, requestID string) string { return user + "\x00" + requestID }

func ankiMediaURL(deckID, value string) string {
	if !strings.HasPrefix(value, "media:") {
		return value
	}
	name := strings.TrimPrefix(value, "media:")
	return "/anki/decks/" + url.PathEscape(deckID) + "/media/" + url.PathEscape(name)
}

func ankiUserMediaNameValid(name string) bool {
	if name == "" || len(name) > 255 || name == "." || name == ".." || strings.ContainsAny(name, "/\\:") || strings.TrimSpace(name) != name || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == '\\' || r == '/' || r == ':' {
			return false
		}
	}
	base := strings.TrimRight(strings.SplitN(name, ".", 2)[0], " .")
	reserved := map[string]bool{"CON": true, "PRN": true, "AUX": true, "NUL": true}
	for i := 1; i <= 9; i++ {
		reserved[fmt.Sprintf("COM%d", i)] = true
		reserved[fmt.Sprintf("LPT%d", i)] = true
	}
	if reserved[strings.ToUpper(base)] {
		return false
	}
	return true
}

func validateUserAnkiDocument(doc ankitransfer.Document, requestID string) error {
	if strings.TrimSpace(requestID) == "" {
		return ErrInvalid
	}
	parsed, err := uuid.Parse(requestID)
	if err != nil || parsed == uuid.Nil {
		return ErrInvalid
	}
	if doc.Version != ankitransfer.DocumentVersion || strings.TrimSpace(doc.Title) == "" || utf8.RuneCountInString(doc.Title) > 255 || len(doc.Description) > 60000 || len(doc.SourceURL) > 60000 || utf8.RuneCountInString(doc.SourceID) > 128 || utf8.RuneCountInString(doc.Level) > 32 || len(doc.Cards) < 1 || len(doc.Cards) > maxUserAnkiCards || len(doc.Media) > 10000 {
		return ErrInvalid
	}
	media := make(map[string]struct{}, len(doc.Media))
	var totalMediaBytes int64
	for _, item := range doc.Media {
		if _, err := ankitransfer.DetectMediaMIME(item.Data); err != nil {
			return ErrInvalid
		}
		if !ankiUserMediaNameValid(item.Name) || utf8.RuneCountInString(item.MIME) > 255 {
			return ErrInvalid
		}
		if item.MIME != "" {
			if _, _, err := mime.ParseMediaType(item.MIME); err != nil {
				return ErrInvalid
			}
		}
		key := strings.ToLower(item.Name)
		if _, duplicate := media[key]; duplicate {
			return ErrInvalid
		}
		media[key] = struct{}{}
		totalMediaBytes += int64(len(item.Data))
		if totalMediaBytes > ankitransfer.MaxMediaBytes {
			return ErrInvalid
		}
	}
	for _, card := range doc.Cards {
		if strings.TrimSpace(card.Front) == "" || len(card.Answers) == 0 {
			return ErrInvalid
		}
		for _, answer := range card.Answers {
			if strings.TrimSpace(answer) == "" {
				return ErrInvalid
			}
		}
		for _, ref := range []string{card.ImageURL, card.Audio} {
			if strings.HasPrefix(ref, "media:") {
				if _, ok := media[strings.ToLower(strings.TrimPrefix(ref, "media:"))]; !ok {
					return ErrInvalid
				}
			} else if ref != "" {
				parsedURL, err := url.Parse(ref)
				if err != nil || (!strings.EqualFold(parsedURL.Scheme, "http") && !strings.EqualFold(parsedURL.Scheme, "https")) || parsedURL.Host == "" {
					return ErrInvalid
				}
			}
		}
	}
	return nil
}

func copyAnkiTransferCards(cards []ankitransfer.Card) []ankitransfer.Card {
	result := make([]ankitransfer.Card, len(cards))
	for i, card := range cards {
		card.Answers = append([]string(nil), card.Answers...)
		result[i] = card
	}
	return result
}

func userAnkiDeckPremium(record *userAnkiDeckRecord) bool {
	if record == nil {
		return false
	}
	if record.Level == "N1" || record.Level == "N2" || record.Level == "N3" {
		return true
	}
	if record.SourceID != "" {
		deck, err := anki.Get(record.SourceID)
		return err == nil && ankiPremium(deck)
	}
	return false
}

func userAnkiDeckFromRecord(record *userAnkiDeckRecord) AnkiDeckView {
	previewCount := len(record.Cards)
	if previewCount > 3 {
		previewCount = 3
	}
	preview := make([]anki.Card, 0, previewCount)
	for index, card := range record.Cards[:previewCount] {
		preview = append(preview, anki.Card{
			Index:       index,
			SourceIndex: index,
			Front:       card.Front,
			Answers:     append([]string(nil), card.Answers...),
			Meaning:     card.Meaning,
			ImageURL:    ankiMediaURL(record.ID, card.ImageURL),
		})
	}
	deck := anki.Deck{
		ID:           record.ID,
		Title:        record.Title,
		Category:     "Personal",
		Level:        record.Level,
		Description:  record.Description,
		Instructions: "Ingat jawaban sebelum membalik kartu; cocokkan dengan semua jawaban yang tersedia.",
		OriginalName: record.Title,
		SourceURL:    record.SourceURL,
		CardCount:    len(record.Cards),
		SourceCount:  len(record.Cards),
		ChunkSize:    1,
		Available:    true,
		Preview:      preview,
	}
	requiredTier := "free"
	if userAnkiDeckPremium(record) {
		requiredTier = "lifetime"
	}
	return AnkiDeckView{Deck: deck, RequiredTier: requiredTier, SourceID: record.SourceID, Personal: true}
}

func (s *MemoryStore) userAnkiDeckViewLocked(user string, record *userAnkiDeckRecord) AnkiDeckView {
	view := userAnkiDeckFromRecord(record)
	view.Stats = ankiStats(ankiStateForRevision(s.gamification[user], record.ID, record.Revision), len(record.Cards), s.gamificationNow())
	view.Locked = view.RequiredTier == "lifetime" && s.profiles[user].Entitlement.Tier != domain.TierPremium
	if view.Locked {
		view.Preview = nil
	}
	return view
}

func (s *MemoryStore) userAnkiDeck(user, id string) (AnkiDeckView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record := s.userAnkiDecks[id]
	if record == nil || record.UserID != user {
		return AnkiDeckView{}, ErrNotFound
	}
	return s.userAnkiDeckViewLocked(user, record), nil
}

func (s *MemoryStore) ImportAnkiDeck(user string, doc ankitransfer.Document, requestID string) (AnkiDeckView, error) {
	if user == "" || uuid.Validate(user) != nil || validateUserAnkiDocument(doc, requestID) != nil {
		return AnkiDeckView{}, ErrInvalid
	}
	parsedRequestID, _ := uuid.Parse(requestID)
	requestID = parsedRequestID.String()
	cards := copyAnkiTransferCards(doc.Cards)
	media := make(map[string]userAnkiMedia, len(doc.Media))
	canonicalMediaNames := make(map[string]string, len(doc.Media))
	for _, item := range doc.Media {
		mime, _ := ankitransfer.DetectMediaMIME(item.Data)
		media[item.Name] = userAnkiMedia{MIME: mime, Size: int64(len(item.Data)), Data: append([]byte(nil), item.Data...)}
		canonicalMediaNames[strings.ToLower(item.Name)] = item.Name
	}
	for index := range cards {
		for _, field := range []*string{&cards[index].ImageURL, &cards[index].Audio} {
			if strings.HasPrefix(*field, "media:") {
				name := strings.TrimPrefix(*field, "media:")
				*field = "media:" + canonicalMediaNames[strings.ToLower(name)]
			}
		}
	}
	cardsJSON, err := json.Marshal(cards)
	if err != nil {
		return AnkiDeckView{}, err
	}
	storageBytes := int64(len(cardsJSON))
	for _, item := range media {
		storageBytes += item.Size
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	requestKey := ankiUserDeckRequestKey(user, requestID)
	if id := s.userAnkiDeckRequests[requestKey]; id != "" {
		if record := s.userAnkiDecks[id]; record != nil {
			return s.userAnkiDeckViewLocked(user, record), nil
		}
	}
	userDeckCount := 0
	userStorageBytes := int64(0)
	for _, record := range s.userAnkiDecks {
		if record.UserID == user {
			userDeckCount++
			userStorageBytes += record.StorageBytes
		}
	}
	if userDeckCount >= maxUserAnkiDecks || userStorageBytes+storageBytes > maxUserAnkiStorageBytes {
		return AnkiDeckView{}, ErrAnkiQuota
	}
	deckID := "user-" + uuid.NewString()
	revision := uuid.NewString()
	record := &userAnkiDeckRecord{
		ID: deckID, UserID: user, RequestID: requestID, Revision: revision,
		Version: doc.Version, Title: strings.TrimSpace(doc.Title), Description: doc.Description,
		SourceURL: doc.SourceURL, SourceID: doc.SourceID, Level: doc.Level,
		Cards: cards, Media: media, StorageBytes: storageBytes, CreatedAt: s.gamificationNow().UTC(),
	}
	if s.gamificationDB != nil {
		stored := database.UserAnkiDeck{
			ID: record.ID, UserID: user, RequestID: requestID, Revision: revision, Version: doc.Version,
			Title: record.Title, Description: record.Description, SourceURL: record.SourceURL,
			SourceID: record.SourceID, Level: record.Level, CardCount: len(cards), Cards: cardsJSON, CreatedAt: record.CreatedAt,
		}
		if err := s.ormWriter.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&stored).Error; err != nil {
				return err
			}
			for name, item := range media {
				row := database.UserAnkiDeckMedia{UserID: user, DeckID: deckID, Name: name, MIME: item.MIME, Data: item.Data}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return AnkiDeckView{}, err
		}
		metadata := make(map[string]userAnkiMedia, len(media))
		for name, item := range media {
			metadata[name] = userAnkiMedia{MIME: item.MIME, Size: item.Size}
		}
		record.Media = metadata
	}
	s.userAnkiDecks[deckID] = record
	s.userAnkiDeckRequests[requestKey] = deckID
	return s.userAnkiDeckViewLocked(user, record), nil
}

func (s *MemoryStore) UserAnkiDecks(user string) ([]AnkiDeckView, error) {
	if user == "" {
		return nil, ErrNotFound
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]*userAnkiDeckRecord, 0)
	for _, record := range s.userAnkiDecks {
		if record.UserID == user {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].CreatedAt.Equal(records[j].CreatedAt) {
			return records[i].ID < records[j].ID
		}
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	result := make([]AnkiDeckView, 0, len(records))
	for _, record := range records {
		result = append(result, s.userAnkiDeckViewLocked(user, record))
	}
	return result, nil
}

func (s *MemoryStore) ExportAnkiDeck(user, id string) (ankitransfer.Document, error) {
	if !isUserAnkiDeckID(id) {
		view, err := s.AnkiDeck(user, id)
		if err != nil {
			return ankitransfer.Document{}, err
		}
		if view.Locked {
			return ankitransfer.Document{}, ErrPremium
		}
		if !view.Available || view.CardCount < 1 || view.CardCount > ankitransfer.MaxCards {
			return ankitransfer.Document{}, ErrInvalid
		}
		doc := ankitransfer.Document{Version: ankitransfer.DocumentVersion, Title: view.Title, Description: strings.Join([]string{view.Description, view.OriginalDescription, view.Instructions, "Kotoba source license:\n" + anki.License()}, "\n\n"), SourceURL: view.SourceURL, SourceID: view.ID, Level: view.Level, Cards: make([]ankitransfer.Card, 0, view.CardCount)}
		for index := 0; index < view.CardCount; index++ {
			card, err := anki.ReadCard(id, index)
			if err != nil {
				return ankitransfer.Document{}, err
			}
			doc.Cards = append(doc.Cards, ankitransfer.Card{Front: card.Front, Answers: append([]string(nil), card.Answers...), Meaning: card.Meaning, ImageURL: card.ImageURL})
		}
		return doc, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	record := s.userAnkiDecks[id]
	if record == nil || record.UserID != user {
		return ankitransfer.Document{}, ErrNotFound
	}
	if userAnkiDeckPremium(record) && s.profiles[user].Entitlement.Tier != domain.TierPremium {
		return ankitransfer.Document{}, ErrPremium
	}
	doc := ankitransfer.Document{
		Version: record.Version, Title: record.Title, Description: record.Description,
		SourceURL: record.SourceURL, SourceID: record.SourceID, Level: record.Level,
		Cards: copyAnkiTransferCards(record.Cards), Media: make([]ankitransfer.Media, 0, len(record.Media)),
	}
	if s.gamificationDB != nil {
		var mediaBytes int64
		if err := s.ormDB.Model(&database.UserAnkiDeckMedia{}).
			Where("user_id = ? AND deck_id = ?", user, id).
			Select("COALESCE(SUM(OCTET_LENGTH(data)), 0)").Scan(&mediaBytes).Error; err != nil {
			return ankitransfer.Document{}, err
		}
		if mediaBytes+record.StorageBytes-int64(sumUserAnkiMediaSizes(record.Media)) > maxUserAnkiStorageBytes {
			return ankitransfer.Document{}, ErrInvalid
		}
		var stored []database.UserAnkiDeckMedia
		if err := s.ormDB.Where("user_id = ? AND deck_id = ?", user, id).Order("name").Find(&stored).Error; err != nil {
			return ankitransfer.Document{}, err
		}
		if len(stored) != len(record.Media) || len(stored) > 10000 {
			return ankitransfer.Document{}, ErrInvalid
		}
		for _, item := range stored {
			if _, exists := record.Media[item.Name]; !exists {
				return ankitransfer.Document{}, ErrInvalid
			}
			doc.Media = append(doc.Media, ankitransfer.Media{Name: item.Name, MIME: item.MIME, Data: append([]byte(nil), item.Data...)})
		}
		return doc, nil
	}
	names := make([]string, 0, len(record.Media))
	for name := range record.Media {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		item := record.Media[name]
		doc.Media = append(doc.Media, ankitransfer.Media{Name: name, MIME: item.MIME, Data: append([]byte(nil), item.Data...)})
	}
	return doc, nil
}

func (s *MemoryStore) AnkiDeckMedia(user, id, name string) ([]byte, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record := s.userAnkiDecks[id]
	if record == nil || record.UserID != user {
		return nil, "", ErrNotFound
	}
	if userAnkiDeckPremium(record) && s.profiles[user].Entitlement.Tier != domain.TierPremium {
		return nil, "", ErrPremium
	}
	item, ok := record.Media[name]
	if !ok {
		return nil, "", ErrNotFound
	}
	if s.gamificationDB != nil {
		var row database.UserAnkiDeckMedia
		err := s.ormDB.Where("user_id = ? AND deck_id = ? AND name = ?", user, id, name).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrNotFound
		}
		if err != nil {
			return nil, "", err
		}
		return append([]byte(nil), row.Data...), row.MIME, nil
	}
	return append([]byte(nil), item.Data...), item.MIME, nil
}

func sumUserAnkiMediaSizes(media map[string]userAnkiMedia) int64 {
	var size int64
	for _, item := range media {
		size += item.Size
	}
	return size
}

func cloneGamificationState(state *gamificationState) (*gamificationState, error) {
	if state == nil {
		return nil, nil
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	next := newGamificationState()
	if err := json.Unmarshal(raw, next); err != nil {
		return nil, err
	}
	return next, nil
}

func (s *MemoryStore) DeleteUserAnkiDeck(user, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.userAnkiDecks[id]
	if record == nil || record.UserID != user {
		return ErrNotFound
	}
	nextState, err := cloneGamificationState(s.gamification[user])
	if err != nil {
		return err
	}
	if nextState != nil && nextState.Anki != nil {
		delete(nextState.Anki, id)
	}
	if s.gamificationDB != nil {
		if err := s.ormWriter.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ? AND deck_id = ?", user, id).Delete(&database.UserAnkiDeckMedia{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ? AND deck_id = ?", user, id).Delete(&database.UserAnkiDeck{}).Error; err != nil {
				return err
			}
			if nextState != nil {
				raw, err := json.Marshal(nextState)
				if err != nil {
					return err
				}
				if err := tx.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state"})}).Create(&database.GamificationUser{UserID: user, State: raw}).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	delete(s.userAnkiDecks, id)
	delete(s.userAnkiDeckRequests, ankiUserDeckRequestKey(user, record.RequestID))
	if nextState != nil {
		s.gamification[user] = nextState
	}
	return nil
}

// forgetUserAnkiDecksLocked removes cached imported decks after the durable
// account deletion transaction has committed.
func (s *MemoryStore) forgetUserAnkiDecksLocked(user string) {
	for id, record := range s.userAnkiDecks {
		if record.UserID == user {
			delete(s.userAnkiDecks, id)
			delete(s.userAnkiDeckRequests, ankiUserDeckRequestKey(user, record.RequestID))
		}
	}
}

func persistedUserAnkiDeck(record database.UserAnkiDeck, media []userAnkiMediaMetadata) (*userAnkiDeckRecord, error) {
	if record.ID == "" || !isUserAnkiDeckID(record.ID) || uuid.Validate(record.UserID) != nil || uuid.Validate(record.Revision) != nil || record.CardCount < 1 || record.CardCount > maxUserAnkiCards {
		return nil, fmt.Errorf("invalid imported Anki deck record")
	}
	var cards []ankitransfer.Card
	if err := json.Unmarshal(record.Cards, &cards); err != nil || len(cards) != record.CardCount {
		return nil, fmt.Errorf("invalid imported Anki deck cards: %w", errors.Join(err, ErrInvalid))
	}
	loadedMedia := make(map[string]userAnkiMedia, len(media))
	storageBytes := int64(len(record.Cards))
	for _, item := range media {
		if item.UserID != record.UserID || item.DeckID != record.ID || !ankiUserMediaNameValid(item.Name) {
			return nil, fmt.Errorf("invalid imported Anki deck media")
		}
		if item.Size < 0 {
			return nil, fmt.Errorf("invalid imported Anki deck media size")
		}
		loadedMedia[item.Name] = userAnkiMedia{MIME: item.MIME, Size: item.Size}
		storageBytes += item.Size
	}
	return &userAnkiDeckRecord{
		ID: record.ID, UserID: record.UserID, RequestID: record.RequestID, Revision: record.Revision,
		Version: record.Version, Title: record.Title, Description: record.Description,
		SourceURL: record.SourceURL, SourceID: record.SourceID, Level: record.Level,
		Cards: cards, Media: loadedMedia, StorageBytes: storageBytes, CreatedAt: record.CreatedAt,
	}, nil
}
