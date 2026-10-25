package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserAnkiDeck keeps a user's imported deck metadata and card JSON separate
// from its media blobs, so study and review never rewrite large media payloads.
type UserAnkiDeck struct {
	ID          string    `gorm:"column:deck_id;type:varchar(41);primaryKey"`
	UserID      string    `gorm:"column:user_id;type:char(36);not null;uniqueIndex:user_anki_deck_request,priority:1;index:user_anki_decks_user_id"`
	RequestID   string    `gorm:"column:request_id;type:varchar(64);not null;uniqueIndex:user_anki_deck_request,priority:2"`
	Revision    string    `gorm:"type:char(36);not null"`
	Version     int       `gorm:"not null"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text;not null"`
	SourceURL   string    `gorm:"column:source_url;type:text;not null"`
	SourceID    string    `gorm:"column:source_id;type:varchar(128);not null"`
	Level       string    `gorm:"type:varchar(32);not null"`
	CardCount   int       `gorm:"column:card_count;not null"`
	Cards       []byte    `gorm:"type:longtext;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
}

func (UserAnkiDeck) TableName() string { return "user_anki_decks" }

func (m *UserAnkiDeck) BeforeCreate(_ *gorm.DB) error {
	if err := uuid.Validate(m.UserID); err != nil {
		return err
	}
	if !strings.HasPrefix(m.ID, "user-") {
		return fmt.Errorf("invalid user Anki deck ID")
	}
	if err := uuid.Validate(strings.TrimPrefix(m.ID, "user-")); err != nil {
		return err
	}
	return uuid.Validate(m.Revision)
}

type UserAnkiDeckMedia struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	UserID string `gorm:"column:user_id;type:char(36);not null;uniqueIndex:user_anki_media_name,priority:1;index:user_anki_media_owner"`
	DeckID string `gorm:"column:deck_id;type:varchar(41);not null;uniqueIndex:user_anki_media_name,priority:2"`
	Name   string `gorm:"type:varchar(255);not null;uniqueIndex:user_anki_media_name,priority:3"`
	MIME   string `gorm:"type:varchar(255);not null"`
	Data   []byte `gorm:"type:longblob;not null"`
}

func (UserAnkiDeckMedia) TableName() string { return "user_anki_deck_media" }

func (m *UserAnkiDeckMedia) BeforeCreate(_ *gorm.DB) error {
	if err := uuid.Validate(m.UserID); err != nil {
		return err
	}
	if !strings.HasPrefix(m.DeckID, "user-") {
		return fmt.Errorf("invalid user Anki deck ID")
	}
	if err := uuid.Validate(strings.TrimPrefix(m.DeckID, "user-")); err != nil {
		return err
	}
	return setUUID(&m.ID)
}
