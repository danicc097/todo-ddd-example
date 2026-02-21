package domain

import (
	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrInvalidStatus = shared.NewDomainError(apperrors.Unprocessable, "invalid status transition")

type TodoStatus string

const (
	StatusPending   TodoStatus = "PENDING"
	StatusCompleted TodoStatus = "COMPLETED"
	StatusArchived  TodoStatus = "ARCHIVED"
)

func (s TodoStatus) String() string {
	return string(s)
}
