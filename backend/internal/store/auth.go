package store

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
)

func (s *MemoryStore) AuthenticateEmail(email, password string) (domain.Profile, error) {
	s.mu.RLock()
	account, ok := s.accounts[strings.ToLower(strings.TrimSpace(email))]
	s.mu.RUnlock()
	if !ok || bcrypt.CompareHashAndPassword(account.PasswordHash, []byte(password)) != nil {
		return domain.Profile{}, ErrInvalidCredentials
	}
	return s.Profile(account.ID), nil
}

func (s *MemoryStore) AuthenticateGoogle(identity domain.GoogleIdentity) (domain.Profile, error) {
	email := strings.ToLower(strings.TrimSpace(identity.Email))
	s.mu.Lock()
	defer s.mu.Unlock()
	if email == s.playReviewEmail {
		return domain.Profile{}, ErrReviewEmailLogin
	}
	existingAccount, ok := s.accounts[email]
	if !ok {
		id := uuid.NewString()
		// Preserve progress from accounts created before durable account storage.
		if s.gamification[database.UserUUID("google-"+identity.Subject)] != nil {
			id = database.UserUUID("google-" + identity.Subject)
		}
		existingAccount = account{
			ID:          id,
			Email:       email,
			DisplayName: identity.DisplayName,
			PhotoURL:    identity.PhotoURL,
			GoogleSub:   identity.Subject,
		}
	} else {
		existingAccount.GoogleSub = identity.Subject
		if existingAccount.CustomPhotoURL == "" {
			existingAccount.PhotoURL = identity.PhotoURL
		}
		if identity.DisplayName != "" {
			existingAccount.DisplayName = identity.DisplayName
		}
	}
	if s.billingEnabled {
		raw, err := json.Marshal(existingAccount)
		if err != nil {
			return domain.Profile{}, err
		}
		if err = s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state"})}).Create(&database.Account{ID: existingAccount.ID, Email: email, State: raw}).Error; err != nil {
			return domain.Profile{}, err
		}
	}
	s.accounts[email] = existingAccount
	return s.ensureProfileForAccountLocked(existingAccount), nil
}

// ConfigurePlayReviewAccount adds an email/password login for Play's app review.
// Its credentials are supplied by the backend environment and its premium access
// exists only in this API process. Never store the password in the database.
func (s *MemoryStore) ConfigurePlayReviewAccount(email, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" && password == "" {
		return nil
	}
	if email == "" || password == "" {
		return fmt.Errorf("PLAY_REVIEW_EMAIL and PLAY_REVIEW_PASSWORD must both be set")
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return fmt.Errorf("PLAY_REVIEW_EMAIL must be a valid email address")
	}
	if email == "demo@nihongomaster.app" {
		return fmt.Errorf("PLAY_REVIEW_EMAIL must be different from the development demo account")
	}
	if strings.TrimSpace(password) != password || len([]byte(password)) > 72 {
		return fmt.Errorf("PLAY_REVIEW_PASSWORD must not have surrounding spaces and must be at most 72 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash Play review password: %w", err)
	}
	userID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("nihongo-master:play-review:"+email)).String()
	configuredAccount := account{
		ID:           userID,
		Email:        email,
		DisplayName:  "Google Play Reviewer",
		PasswordHash: hash,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.accounts[email]; ok && existing.ID != userID {
		return fmt.Errorf("PLAY_REVIEW_EMAIL is already assigned to another account")
	}
	s.accounts[email] = configuredAccount
	s.profiles[userID] = profileForAccount(configuredAccount)
	s.playReviewEmail = email
	s.playReviewUserID = userID
	return nil
}

func (s *MemoryStore) CreateSession(userID string) (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	s.mu.Lock()
	s.sessions[token] = userID
	s.mu.Unlock()
	return token, nil
}

func (s *MemoryStore) ResolveSession(token string) (string, bool) {
	s.mu.RLock()
	userID, ok := s.sessions[token]
	s.mu.RUnlock()
	return userID, ok
}

func (s *MemoryStore) RevokeSession(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *MemoryStore) seedDemoAccount() {
	hash, err := bcrypt.GenerateFromPassword([]byte("Demo123!"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	account := account{
		ID:           database.DemoUserID,
		Email:        "demo@nihongomaster.app",
		DisplayName:  "Demo Learner",
		PasswordHash: hash,
	}
	s.accounts[account.Email] = account
	s.profiles[account.ID] = profileForAccount(account)
}

func (s *MemoryStore) ensureProfileForAccountLocked(account account) domain.Profile {
	if profile, ok := s.profiles[account.ID]; ok {
		profile.DisplayName = account.DisplayName
		profile.Email = account.Email
		profile.PhotoURL = account.profilePhotoURL()
		s.profiles[account.ID] = profile
		return profile
	}
	profile := profileForAccount(account)
	s.profiles[account.ID] = profile
	return profile
}

func profileForAccount(account account) domain.Profile {
	return domain.Profile{
		ID:          account.ID,
		DisplayName: account.DisplayName,
		Email:       account.Email,
		PhotoURL:    account.profilePhotoURL(),
		Target:      "JLPT N5",
		Level:       1,
		Entitlement: domain.Entitlement{Tier: domain.TierFree},
	}
}

func (a account) profilePhotoURL() string {
	if a.CustomPhotoURL != "" {
		return a.CustomPhotoURL
	}
	return a.PhotoURL
}

func (s *MemoryStore) IsDemoUser(id string) bool { return id == database.DemoUserID }

func (s *MemoryStore) IsPlayReviewUser(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return id != "" && id == s.playReviewUserID
}
