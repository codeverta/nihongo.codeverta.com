package store

import (
	"errors"
	"nihongo-master-backend/internal/domain"
	"testing"
)

func TestAdminLoginAccount(t *testing.T) {
	s := NewMemoryStore()
	if err := s.ConfigureAdminLoginAccount("", ""); err != nil {
		t.Fatalf("empty configuration should be a no-op: %v", err)
	}
	for name, args := range map[string][2]string{
		"missing password": {"admin@example.com", ""},
		"missing email":    {"", "a-long-unique-password"},
		"invalid email":    {"not-an-email", "a-long-unique-password"},
		"demo email":       {"demo@nihongomaster.app", "a-long-unique-password"},
		"short password":   {"admin@example.com", "short"},
		"padded password":  {"admin@example.com", " a-long-unique-password "},
	} {
		if err := s.ConfigureAdminLoginAccount(args[0], args[1]); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
	if err := s.ConfigureAdminLoginAccount("Admin@Example.com", "a-long-unique-password"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AuthenticateEmail("admin@example.com", "a-long-unique-password"); err != nil {
		t.Fatalf("admin login failed: %v", err)
	}
	if _, err := s.AuthenticateEmail("admin@example.com", "wrong-password-value"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password accepted: %v", err)
	}
	// The admin email must not be claimable through Google Sign-In.
	if _, err := s.AuthenticateGoogle(domain.GoogleIdentity{Subject: "x", Email: "admin@example.com"}); !errors.Is(err, ErrReviewEmailLogin) {
		t.Fatalf("admin email accepted through Google: %v", err)
	}
}
