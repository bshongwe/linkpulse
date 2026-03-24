package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/ports"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) ports.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, password_hash, name, email_verified, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.Name, user.EmailVerified, user.CreatedAt)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password_hash, name, email_verified, created_at FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	return &user, err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password_hash, name, email_verified, created_at FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	return &user, err
}
