package domain

import "github.com/bshongwe/linkpulse/backend/shared/errors"

var (
	ErrInvalidCredentials = errors.New(errors.ErrUnauthorized, "invalid email or password")
	ErrUserAlreadyExists  = errors.New(errors.ErrAlreadyExists, "user with this email already exists")
	ErrInvalidEmail       = errors.New(errors.ErrInvalidInput, "invalid email format")
)
