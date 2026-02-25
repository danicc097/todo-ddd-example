package domain

import (
	"errors"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
)

var ErrConcurrentUpdate = NewDomainError(apperrors.Conflict, "concurrent update detected")

// DomainError defines errors originating in the domain layer.
type DomainError interface {
	error
	Code() apperrors.ErrorCode
}

type domainError struct {
	error

	code apperrors.ErrorCode
}

func (e domainError) Code() apperrors.ErrorCode { return e.code }

func (e domainError) Unwrap() error { return e.error }

func NewDomainError(code apperrors.ErrorCode, message string) error {
	return domainError{
		error: errors.New(message),
		code:  code,
	}
}

func WrapDomainError(err error, code apperrors.ErrorCode) error {
	return domainError{
		error: err,
		code:  code,
	}
}
