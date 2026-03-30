package errors_test

import (
	"errors"
	"testing"

	sharedErrors "github.com/bshongwe/linkpulse/backend/shared/errors"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"not found AppError", sharedErrors.New(sharedErrors.ErrNotFound, "not found"), true},
		{"already exists AppError", sharedErrors.New(sharedErrors.ErrAlreadyExists, "exists"), false},
		{"plain error", errors.New("some error"), false},
		{"wrapped not found", sharedErrors.Wrap(errors.New("cause"), sharedErrors.ErrNotFound, "wrapped"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sharedErrors.IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"already exists AppError", sharedErrors.New(sharedErrors.ErrAlreadyExists, "exists"), true},
		{"not found AppError", sharedErrors.New(sharedErrors.ErrNotFound, "not found"), false},
		{"plain error", errors.New("some error"), false},
		{"wrapped already exists", sharedErrors.Wrap(errors.New("cause"), sharedErrors.ErrAlreadyExists, "wrapped"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sharedErrors.IsAlreadyExists(tt.err); got != tt.want {
				t.Errorf("IsAlreadyExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Error(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := sharedErrors.New(sharedErrors.ErrNotFound, "link not found")
		if err.Error() != "link not found" {
			t.Errorf("got %q, want %q", err.Error(), "link not found")
		}
	})
	t.Run("with wrapped error", func(t *testing.T) {
		cause := errors.New("db timeout")
		err := sharedErrors.Wrap(cause, sharedErrors.ErrInternal, "query failed")
		if err.Error() != "query failed: db timeout" {
			t.Errorf("got %q, want %q", err.Error(), "query failed: db timeout")
		}
	})
}
