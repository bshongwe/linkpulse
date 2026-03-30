package application

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/ports"
	"github.com/bshongwe/linkpulse/backend/shared/config"
	"github.com/bshongwe/linkpulse/backend/shared/errors"
)

const (
	errUnexpectedSigningMethod = "unexpected signing method"
	errInvalidOrExpiredToken   = "invalid or expired token"
	errInvalidClaims           = "invalid claims"
	errInvalidRefreshToken     = "invalid refresh token"
	errMissingTokenID          = "missing token id"
	errTokenRevoked            = "token has been revoked"
	errInvalidUserIDClaim      = "invalid user_id claim"
	errInvalidEmailClaim       = "invalid email claim"
)

type tokenService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	blacklist     ports.TokenBlacklist
}

func NewTokenService(cfg config.JWTConfig, blacklist ports.TokenBlacklist) domain.TokenService {
	return &tokenService{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
		blacklist:     blacklist,
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
		"jti":     uuid.New().String(),
		"user_id": user.ID,
		"email":   user.Email,
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
			return nil, errors.New(errors.ErrUnauthorized, errUnexpectedSigningMethod)
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

func (s *tokenService) RefreshTokens(ctx context.Context, refreshTokenStr string) (*domain.TokenPair, error) {
	token, err := jwt.Parse(refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		// Guard against algorithm confusion attacks (same as ValidateAccessToken)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(errors.ErrUnauthorized, errUnexpectedSigningMethod)
		}
		return s.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New(errors.ErrUnauthorized, "invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New(errors.ErrUnauthorized, "invalid claims")
	}

	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return nil, errors.New(errors.ErrUnauthorized, "missing token id")
	}

	// Blacklist check — reject revoked tokens (e.g. after logout)
	revoked, err := s.blacklist.IsRevoked(ctx, jti)
	if err != nil || revoked {
		return nil, errors.New(errors.ErrUnauthorized, "token has been revoked")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New(errors.ErrUnauthorized, "invalid user_id claim")
	}
	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New(errors.ErrUnauthorized, "invalid email claim")
	}

	user := &domain.User{
		ID:    uuid.MustParse(userID),
		Email: email,
	}

	return s.GenerateTokenPair(user)
}

func (s *tokenService) RevokeRefreshToken(ctx context.Context, refreshTokenStr string) error {
	token, err := jwt.Parse(refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(errors.ErrUnauthorized, errUnexpectedSigningMethod)
		}
		return s.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return errors.New(errors.ErrUnauthorized, "invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New(errors.ErrUnauthorized, "invalid claims")
	}

	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return errors.New(errors.ErrUnauthorized, "missing token id")
	}

	// TTL = remaining lifetime of the token so the blacklist entry auto-expires
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New(errors.ErrUnauthorized, "invalid exp claim")
	}
	ttl := time.Until(time.Unix(int64(exp), 0))
	if ttl <= 0 {
		return nil // already expired, nothing to revoke
	}

	return s.blacklist.Revoke(ctx, jti, ttl)
}
