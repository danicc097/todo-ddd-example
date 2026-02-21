package domain

import (
	"encoding/json"
	"strings"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrTagNameEmpty   = shared.NewDomainError(apperrors.InvalidInput, "tag name cannot be empty")
	ErrTagNameTooLong = shared.NewDomainError(apperrors.InvalidInput, "tag name is too long")
)

const tagMaxLen = 20

type TagName struct {
	value string
}

func NewTagName(val string) (TagName, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return TagName{}, ErrTagNameEmpty
	}

	if len(val) > tagMaxLen {
		return TagName{}, ErrTagNameTooLong
	}

	return TagName{value: val}, nil
}

func (t TagName) String() string {
	return t.value
}

func (t TagName) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.value)
}

func (t TagName) MarshalText() ([]byte, error) {
	return []byte(t.value), nil
}

func (t *TagName) UnmarshalText(text []byte) error {
	vo, err := NewTagName(string(text))
	if err != nil {
		return err
	}

	*t = vo

	return nil
}
