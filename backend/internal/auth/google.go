package auth

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/api/idtoken"
	"nihongo-master-backend/internal/domain"
)

var (
	ErrGoogleNotConfigured = errors.New("google oauth is not configured")
	ErrInvalidGoogleToken  = errors.New("invalid google id token")
)

type GoogleVerifier interface {
	Verify(context.Context, string) (domain.GoogleIdentity, error)
}

type IDTokenVerifier struct {
	audience string
}

func NewGoogleVerifier(audience string) *IDTokenVerifier {
	return &IDTokenVerifier{audience: strings.TrimSpace(audience)}
}

func (v *IDTokenVerifier) Verify(ctx context.Context, rawToken string) (domain.GoogleIdentity, error) {
	if v.audience == "" {
		return domain.GoogleIdentity{}, ErrGoogleNotConfigured
	}
	payload, err := idtoken.Validate(ctx, rawToken, v.audience)
	if err != nil {
		return domain.GoogleIdentity{}, ErrInvalidGoogleToken
	}
	identity := domain.GoogleIdentity{
		Subject:     payload.Subject,
		Email:       stringClaim(payload.Claims, "email"),
		DisplayName: stringClaim(payload.Claims, "name"),
		PhotoURL:    stringClaim(payload.Claims, "picture"),
	}
	verified, _ := payload.Claims["email_verified"].(bool)
	if identity.Subject == "" || identity.Email == "" || !verified {
		return domain.GoogleIdentity{}, ErrInvalidGoogleToken
	}
	return identity, nil
}

func stringClaim(claims map[string]any, key string) string {
	value, _ := claims[key].(string)
	return strings.TrimSpace(value)
}
