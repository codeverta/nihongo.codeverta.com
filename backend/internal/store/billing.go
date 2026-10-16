package store

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"strings"
	"time"
)

// Billing shares the same durable database and process lock as progress.
func (s *MemoryStore) EnableBillingPersistence() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gamificationDB == nil {
		return fmt.Errorf("billing requires a persistent database")
	}
	if err := database.Migrate(s.ormDB, &database.BillingPurchase{}, &database.BillingToken{}, &database.Account{}, &database.DeletedDemo{}); err != nil {
		return err
	}
	var purchases []database.BillingPurchase
	if err := s.ormDB.Find(&purchases).Error; err != nil {
		return err
	}
	for _, record := range purchases {
		var purchase domain.Purchase
		if err := json.Unmarshal(record.State, &purchase); err != nil {
			return err
		}
		s.purchases[purchase.Platform+":"+purchase.ID] = purchase
	}
	var tokens []database.BillingToken
	if err := s.ormDB.Find(&tokens).Error; err != nil {
		return err
	}
	for _, record := range tokens {
		s.billingTokens[record.UserID] = record.Token
	}
	var accounts []database.Account
	if err := s.ormDB.Find(&accounts).Error; err != nil {
		return err
	}
	for _, record := range accounts {
		var a account
		if err := json.Unmarshal(record.State, &a); err != nil {
			return err
		}
		s.accounts[a.Email] = a
		s.profiles[a.ID] = profileForAccount(a)
	}
	var deleted int64
	if err := s.ormDB.Model(&database.DeletedDemo{}).Where("user_id = ?", database.DemoUserID).Count(&deleted).Error; err != nil {
		return err
	}

	if deleted > 0 {
		delete(s.accounts, "demo@nihongomaster.app")
		delete(s.profiles, database.DemoUserID)
	}
	s.billingEnabled = true
	return nil
}

func (s *MemoryStore) BillingAccountToken(user string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if token := s.billingTokens[user]; token != "" {
		return token, nil
	}
	token := uuid.NewString()
	if s.billingEnabled {
		if err := s.ormWriter.Create(&database.BillingToken{UserID: user, Token: token}).Error; err != nil {
			return "", err
		}
	}
	s.billingTokens[user] = token
	return token, nil
}

func (s *MemoryStore) entitlementLocked(user string) domain.Entitlement {
	if user != "" && user == s.playReviewUserID {
		return domain.Entitlement{Tier: domain.TierPremium}
	}
	for _, p := range s.purchases {
		if p.UserID == user && p.State == "purchased" && time.Since(p.CheckedAt) < 15*time.Minute {
			return domain.Entitlement{Tier: domain.TierPremium, IsLifetime: true}
		}
	}
	return domain.Entitlement{Tier: domain.TierFree}
}

func (s *MemoryStore) Purchases(user string) []domain.Purchase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []domain.Purchase{}
	for _, p := range s.purchases {
		if user == "" || p.UserID == user {
			result = append(result, p)
		}
	}
	return result
}

func (s *MemoryStore) Purchase(platform, id string) (domain.Purchase, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.purchases[platform+":"+id]
	return p, ok
}

// A purchase is bound to one application account. After deletion its store proof
// can restore to a new account; it cannot silently transfer between live accounts.
func (s *MemoryStore) SaveVerifiedPurchase(user string, p domain.Purchase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.billingTokens[user] == "" {
		return ErrInvalid
	}
	key := p.Platform + ":" + p.ID
	old, exists := s.purchases[key]
	if exists && old.UserID != "" && old.UserID != user {
		return ErrPremium
	}
	if !exists && (p.AccountToken == "" || p.AccountToken != s.billingTokens[user]) {
		return ErrPremium
	}
	if exists && old.AccountToken != p.AccountToken {
		return ErrPremium
	}
	p.UserID = user
	if s.billingEnabled {
		raw, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err = s.ormWriter.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"state", "user_id"})}).Create(&database.BillingPurchase{PurchaseKey: key, UserID: database.OptionalUserID(user), State: raw}).Error; err != nil {
			return err
		}
	}
	s.purchases[key] = p
	return nil
}

func (s *MemoryStore) UpdateVerifiedPurchase(p domain.Purchase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := p.Platform + ":" + p.ID
	old, ok := s.purchases[key]
	if !ok {
		return ErrNotFound
	}
	if p.AccountToken != old.AccountToken {
		return ErrInvalid
	}
	p.UserID = old.UserID
	if s.billingEnabled {
		raw, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err = s.ormWriter.Model(&database.BillingPurchase{}).Where("purchase_key = ?", key).Updates(map[string]any{"state": raw, "user_id": database.OptionalUserID(p.UserID)}).Error; err != nil {
			return err
		}
	}
	s.purchases[key] = p
	return nil
}

func PremiumLevel(level string) bool {
	switch strings.ToUpper(level) {
	case "N3", "N2", "N1":
		return true
	}
	return false
}

func (s *MemoryStore) ExamRequiresPremium(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.exams {
		if e.ID == id {
			return !(e.System == "JLPT" && e.Level == "N5")
		}
	}
	return false
}

// Authorization covers catalogs, direct lesson URLs, and ongoing quiz IDs.
func (s *MemoryStore) PremiumRequest(path, level, lessonID, attemptID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if strings.Contains(path, "/simulation/") && attemptID == "" {
		return strings.ToUpper(level) == "N4" || strings.ToUpper(level) == "N3" || strings.ToUpper(level) == "JFT" || strings.ToUpper(level) == "NAT"
	}
	if strings.Contains(path, "/course") {
		if strings.Contains(path, "/grammar/") {
			for _, l := range s.grammar {
				if l.Level == level {
					return PremiumLevel(level)
				}
			}
		}
		if strings.Contains(path, "/kanji/") {
			for _, l := range s.kanji {
				if l.Level == level {
					return PremiumLevel(level)
				}
			}
		}
		if strings.Contains(path, "/reading/") {
			for _, l := range s.reading {
				if l.Level == level {
					return PremiumLevel(level)
				}
			}
		}
		return false
	}
	for _, l := range s.grammar {
		if l.ID == lessonID {
			return PremiumLevel(l.Level)
		}
	}
	for _, l := range s.kanji {
		if l.ID == lessonID {
			return PremiumLevel(l.Level)
		}
	}
	for _, l := range s.reading {
		if l.ID == lessonID {
			return PremiumLevel(l.Level)
		}
	}
	for _, l := range s.vocabulary {
		for _, lesson := range l.Lessons {
			if lesson.ID == lessonID {
				return PremiumLevel(l.Level)
			}
		}
	}
	id := ""
	if a := s.grammarAttempts[attemptID]; a != nil {
		id = a.Attempt.LessonID
	}
	if a := s.kanjiAttempts[attemptID]; a != nil {
		id = a.Attempt.LessonID
	}
	if a := s.readingAttempts[attemptID]; a != nil {
		id = a.Attempt.LessonID
	}
	if a := s.vocabAttempts[attemptID]; a != nil {
		id = a.attempt.LessonID
	}
	if a := s.simulationAttempts[attemptID]; a != nil {
		return a.session.Attempt.ExamID != "jlpt-n5-simulation-1"
	}
	for _, level := range s.vocabulary {
		for _, lesson := range level.Lessons {
			if lesson.ID == id {
				return PremiumLevel(level.Level)
			}
		}
	}
	return strings.Contains(id, "-n3-") || strings.Contains(id, "-n2-") || strings.Contains(id, "-n1-")
}

func (s *MemoryStore) DeleteAccount(user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	detached := map[string]domain.Purchase{}
	for key, p := range s.purchases {
		if p.UserID == user {
			p.UserID = ""
			detached[key] = p
		}
	}
	if s.billingEnabled {
		if err := s.ormWriter.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ?", user).Delete(&database.GamificationUser{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", user).Delete(&database.BillingToken{}).Error; err != nil {
				return err
			}
			for email, a := range s.accounts {
				if a.ID == user {
					if err := tx.Where("email = ?", email).Delete(&database.Account{}).Error; err != nil {
						return err
					}
				}
			}
			if user == database.DemoUserID {
				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&database.DeletedDemo{UserID: user}).Error; err != nil {
					return err
				}
			}
			for key, p := range detached {
				raw, err := json.Marshal(p)
				if err != nil {
					return err
				}
				if err := tx.Model(&database.BillingPurchase{}).Where("purchase_key = ?", key).Updates(map[string]any{"state": raw, "user_id": database.OptionalUserID(p.UserID)}).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	} else if s.gamificationDB != nil {
		if err := s.ormWriter.Where("user_id = ?", user).Delete(&database.GamificationUser{}).Error; err != nil {
			return err
		}
	}

	for key, p := range detached {
		s.purchases[key] = p
	}
	delete(s.profiles, user)
	delete(s.progress, user)
	delete(s.gamification, user)
	delete(s.billingTokens, user)
	for email, a := range s.accounts {
		if a.ID == user {
			delete(s.accounts, email)
		}
	}
	for token, id := range s.sessions {
		if id == user {
			delete(s.sessions, token)
		}
	}
	for id, a := range s.vocabAttempts {
		if a.userID == user {
			delete(s.vocabAttempts, id)
		}
	}
	for id, a := range s.grammarAttempts {
		if a.UserID == user {
			delete(s.grammarAttempts, id)
		}
	}
	for id, a := range s.kanjiAttempts {
		if a.UserID == user {
			delete(s.kanjiAttempts, id)
		}
	}
	for id, a := range s.readingAttempts {
		if a.UserID == user {
			delete(s.readingAttempts, id)
		}
	}
	for id, a := range s.diagnosticAttempts {
		if a.session.Attempt.UserID == user {
			delete(s.diagnosticAttempts, id)
		}
	}
	for id, a := range s.simulationAttempts {
		if a.session.Attempt.UserID == user {
			delete(s.simulationAttempts, id)
		}
	}
	s.diagnosticOrder = filterAttemptIDs(s.diagnosticOrder, func(id string) bool { return s.diagnosticAttempts[id] != nil })
	s.simulationOrder = filterAttemptIDs(s.simulationOrder, func(id string) bool { return s.simulationAttempts[id] != nil })
	return nil
}
func filterAttemptIDs(ids []string, keep func(string) bool) []string {
	result := []string{}
	for _, id := range ids {
		if keep(id) {
			result = append(result, id)
		}
	}
	return result
}
