package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	WorkspaceID uuid.UUID `json:"workspace_id,omitempty"`
	Role        Role      `json:"role,omitempty"`
}

type TokenService interface {
	GenerateTokenPair(user *User) (*TokenPair, error)
	ValidateAccessToken(token string) (*Claims, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
}
