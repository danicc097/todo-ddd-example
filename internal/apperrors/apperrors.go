package apperrors

import "fmt"

type ErrorCode string

const (
	NotFound      ErrorCode = "RESOURCE_NOT_FOUND"
	InvalidInput  ErrorCode = "INVALID_INPUT"
	Internal      ErrorCode = "INTERNAL_ERROR"
	Conflict      ErrorCode = "RESOURCE_CONFLICT"
	Unprocessable ErrorCode = "UNPROCESSABLE_ENTITY"
	Unauthorized  ErrorCode = "UNAUTHORIZED"
	MFARequired   ErrorCode = "MFA_REQUIRED"
)

type AppError struct {
	Code       ErrorCode
	Message    string
	Op         string
	Err        error
	Validation any
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

func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
