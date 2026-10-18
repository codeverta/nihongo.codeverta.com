package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const DemoUserID = "a5090412-9ab4-5852-8229-c7d19eb5cb1c"

func UserUUID(id string) string {
	if parsed, err := uuid.Parse(id); err == nil {
		return parsed.String()
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("nihongo-master:user:"+id)).String()
}
func setUUID(id *string) error {
	if *id == "" {
		*id = uuid.NewString()
	}
	return uuid.Validate(*id)
}
func OptionalUserID(id string) *string {
	if id == "" {
		return nil
	}
	return &id
}

type GamificationUser struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	UserID string `gorm:"column:user_id;type:char(36);not null;uniqueIndex:gamification_user_id"`
	State  []byte `gorm:"type:longblob;not null"`
}

func (GamificationUser) TableName() string { return "gamification_users" }
func (m *GamificationUser) BeforeCreate(_ *gorm.DB) error {
	if err := uuid.Validate(m.UserID); err != nil {
		return err
	}
	return setUUID(&m.ID)
}

type Account struct {
	ID    string `gorm:"type:char(36);primaryKey"`
	Email string `gorm:"type:varchar(320);not null;uniqueIndex:accounts_email"`
	State []byte `gorm:"type:longblob;not null"`
}

func (Account) TableName() string                { return "accounts" }
func (m *Account) BeforeCreate(_ *gorm.DB) error { return setUUID(&m.ID) }

type BillingPurchase struct {
	ID          string  `gorm:"type:char(36);primaryKey"`
	PurchaseKey string  `gorm:"column:purchase_key;type:varbinary(2048);not null;uniqueIndex:billing_purchase_key"`
	UserID      *string `gorm:"column:user_id;type:char(36);index:billing_purchase_user_id"`
	State       []byte  `gorm:"type:longblob;not null"`
}

func (BillingPurchase) TableName() string { return "billing_purchases" }
func (m *BillingPurchase) BeforeCreate(_ *gorm.DB) error {
	if m.UserID != nil {
		if err := uuid.Validate(*m.UserID); err != nil {
			return err
		}
	}
	return setUUID(&m.ID)
}

type BillingToken struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	UserID string `gorm:"column:user_id;type:char(36);not null;uniqueIndex:billing_token_user_id"`
	Token  string `gorm:"type:char(36);not null;uniqueIndex:token"`
}

func (BillingToken) TableName() string { return "billing_tokens" }
func (m *BillingToken) BeforeCreate(_ *gorm.DB) error {
	if err := uuid.Validate(m.UserID); err != nil {
		return err
	}
	return setUUID(&m.ID)
}

type DeletedDemo struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	UserID string `gorm:"column:user_id;type:char(36);not null;uniqueIndex:deleted_demo_user_id"`
}

func (DeletedDemo) TableName() string { return "deleted_demo" }
func (m *DeletedDemo) BeforeCreate(_ *gorm.DB) error {
	if err := uuid.Validate(m.UserID); err != nil {
		return err
	}
	return setUUID(&m.ID)
}

type DictionaryResponse struct {
	Endpoint  string `gorm:"type:text"`
	QueryText string `gorm:"type:varchar(512)"`
	Mode      string `gorm:"type:varchar(16)"`
	Page      int
	JLPT      string `gorm:"type:varchar(2)"`
	Common    bool
	ID        string `gorm:"type:char(36);primaryKey"`
	CacheKey  string `gorm:"column:cache_key;type:char(64);not null;uniqueIndex:dictionary_cache_key"`
	Source    string `gorm:"type:varchar(64);not null"`
	Body      []byte `gorm:"type:longblob;not null"`
	FetchedAt int64  `gorm:"column:fetched_at;type:bigint;not null;index:dictionary_responses_age"`
}

func (DictionaryResponse) TableName() string                { return "dictionary_responses" }
func (m *DictionaryResponse) BeforeCreate(_ *gorm.DB) error { return setUUID(&m.ID) }

type DictionaryEntry struct {
	ID        string `gorm:"type:char(36);primaryKey"`
	Source    string `gorm:"type:varchar(64);not null;uniqueIndex:dictionary_entry_source,priority:1"`
	EntryID   string `gorm:"column:source_entry_id;type:varchar(512);not null;uniqueIndex:dictionary_entry_source,priority:2"`
	Kind      string `gorm:"type:varchar(32);not null"`
	Body      []byte `gorm:"type:longblob;not null"`
	FetchedAt int64  `gorm:"column:fetched_at;type:bigint;not null"`
}

func (DictionaryEntry) TableName() string                { return "dictionary_entries" }
func (m *DictionaryEntry) BeforeCreate(_ *gorm.DB) error { return setUUID(&m.ID) }
