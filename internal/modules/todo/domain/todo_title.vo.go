package domain

import (
	"strings"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrTitleEmpty   = shared.NewDomainError(apperrors.InvalidInput, "title cannot be empty")
	ErrTitleTooLong = shared.NewDomainError(apperrors.InvalidInput, "title is too long")
)

type TodoTitle struct {
	value string
}

func NewTodoTitle(val string) (TodoTitle, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return TodoTitle{}, ErrTitleEmpty
	}

	if len(val) > 100 {
		return TodoTitle{}, ErrTitleTooLong
	}

	return TodoTitle{value: val}, nil
}

func (t TodoTitle) String() string {
	return t.value
}
