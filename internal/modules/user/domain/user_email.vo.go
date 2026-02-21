package domain

import (
	"encoding/json"
	"strings"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrInvalidEmail = shared.NewDomainError(apperrors.InvalidInput, "invalid email address")

type UserEmail struct {
	value string
}

func NewUserEmail(val string) (UserEmail, error) {
	val = strings.ToLower(strings.TrimSpace(val))
	if !strings.Contains(val, "@") {
		return UserEmail{}, ErrInvalidEmail
	}

	return UserEmail{value: val}, nil
}

func (e UserEmail) String() string {
	return e.value
}

func (e UserEmail) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.value)
}

func (e UserEmail) MarshalText() ([]byte, error) {
	return []byte(e.value), nil
}

func (e *UserEmail) UnmarshalText(text []byte) error {
	vo, err := NewUserEmail(string(text))
	if err != nil {
		return err
	}

	*e = vo

	return nil
}
