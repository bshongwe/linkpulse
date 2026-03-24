package application

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/domain"
	"github.com/bshongwe/linkpulse/backend/shared/config"
	"github.com/bshongwe/linkpulse/backend/shared/errors"
)

type tokenService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewTokenService(cfg config.JWTConfig) domain.TokenService {
	return &tokenService{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (s *tokenService) GenerateTokenPair(user *domain.User) (*domain.TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(s.accessExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(s.accessSecret)
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(s.refreshExpiry).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(s.refreshSecret)
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresAt:    time.Now().Add(s.accessExpiry),
	}, nil
}

func (s *tokenService) ValidateAccessToken(tokenStr string) (*domain.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(errors.ErrUnauthorized, "unexpected signing method")
		}
		return s.accessSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New(errors.ErrUnauthorized, "invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New(errors.ErrUnauthorized, "invalid claims")
	}

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return nil, err
	}

	return &domain.Claims{
		UserID: userID,
		Email:  claims["email"].(string),
	}, nil
}

func (s *tokenService) RefreshTokens(refreshTokenStr string) (*domain.TokenPair, error) {
	// TODO: Add refresh token validation + blacklist check in future
	// For MVP we re-generate using user ID from refresh token
	token, err := jwt.Parse(refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		return s.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New(errors.ErrUnauthorized, "invalid refresh token")
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	// In real implementation we would fetch user again. For now we fake minimal user
	user := &domain.User{
		ID:    uuid.MustParse(userID),
		Email: claims["email"].(string), // if stored
	}

	return s.GenerateTokenPair(user)
}
