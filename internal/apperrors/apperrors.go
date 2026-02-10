package apperrors

import (
	"fmt"
)

type AppError struct {
	Code    string
	Message string
	Op      string
	Err     error
	Status  int
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code string, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

func Wrap(err error, code string, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     err,
	}
}

const (
	ErrCodeNotFound      = "RESOURCE_NOT_FOUND"
	ErrCodeInvalidInput  = "INVALID_INPUT"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeConflict      = "RESOURCE_CONFLICT"
	ErrCodeUnprocessable = "UNPROCESSABLE_ENTITY"
)
