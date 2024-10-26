package interr

import (
	"errors"
	"fmt"
)

var (
	ErrInternalError = errors.New("internal error")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type InternalErr interface {
	Error() string
	Unwrap() error
}

type InternalError struct {
	error   error
	wrapped error
	message string
}

func (e InternalError) Error() string {
	return fmt.Sprintf("%s %s: %s", e.message, e.wrapped, e.error)
}

func (e InternalError) Unwrap() error {
	return fmt.Errorf("%w: %w", e.error, e.wrapped)
}

func NewInternalError(wrapped error, message string) InternalErr {
	return InternalError{
		error:   ErrInternalError,
		wrapped: wrapped,
		message: message,
	}
}

func NewNotFoundError(wrapped error, message string) InternalErr {
	return InternalError{
		error:   ErrNotFound,
		wrapped: wrapped,
		message: message,
	}
}

func NewAlreadyExistsError(wrapped error, message string) InternalErr {
	return InternalError{
		error:   ErrAlreadyExists,
		wrapped: wrapped,
		message: message,
	}
}
