package application

import (
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/domain"
)

// tokenValidatorAdapter bridges domain.TokenService to shared/middleware.TokenValidator.
type tokenValidatorAdapter struct {
	svc domain.TokenService
}

// TokenValidatorAdapter wraps a domain.TokenService so it satisfies middleware.TokenValidator.
func TokenValidatorAdapter(svc domain.TokenService) *tokenValidatorAdapter {
	return &tokenValidatorAdapter{svc: svc}
}

func (a *tokenValidatorAdapter) ValidateAccessToken(token string) (string, string, error) {
	claims, err := a.svc.ValidateAccessToken(token)
	if err != nil {
		return "", "", err
	}
	return claims.UserID.String(), claims.Email, nil
}
