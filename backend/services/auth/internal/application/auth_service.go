package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/ports"
)

type AuthService struct {
	userRepo ports.UserRepository
	// tokenService, eventPublisher, etc. will be added later
}

func NewAuthService(userRepo ports.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	// TODO: password hashing, validation, duplicate check
	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}