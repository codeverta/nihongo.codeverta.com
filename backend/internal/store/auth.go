package store

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"strings"

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
		existingAccount.PhotoURL = identity.PhotoURL
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
		profile.PhotoURL = account.PhotoURL
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
		PhotoURL:    account.PhotoURL,
		Target:      "JLPT N5",
		Level:       1,
		Entitlement: domain.Entitlement{Tier: domain.TierFree},
	}
}

func (s *MemoryStore) IsDemoUser(id string) bool { return id == database.DemoUserID }
