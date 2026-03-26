package errors

import (
	"errors"
	"fmt"
)

var As = errors.As

type ErrorCode string

const (
	ErrInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrNotFound      ErrorCode = "NOT_FOUND"
	ErrAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrInternal      ErrorCode = "INTERNAL_ERROR"
	ErrUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrForbidden     ErrorCode = "FORBIDDEN"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func New(code ErrorCode, msg string) error {
	return &AppError{Code: code, Message: msg}
}

func Wrap(err error, code ErrorCode, msg string) error {
	return &AppError{Code: code, Message: msg, Err: err}
}

func IsNotFound(err error) bool {
	var e *AppError
	if ok := As(err, &e); ok {
		return e.Code == ErrNotFound
	}
	return false
}

func IsAlreadyExists(err error) bool {
	var e *AppError
	if ok := As(err, &e); ok {
		return e.Code == ErrAlreadyExists
	}
	return false
}
