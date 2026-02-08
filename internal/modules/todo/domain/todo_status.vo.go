package domain

import "errors"

var ErrInvalidStatus = errors.New("invalid todo status")

type TodoStatus string

const (
	StatusPending   TodoStatus = "PENDING"
	StatusCompleted TodoStatus = "COMPLETED"
	StatusArchived  TodoStatus = "ARCHIVED"
)

func NewTodoStatus(val string) (TodoStatus, error) {
	s := TodoStatus(val)
	switch s {
	case StatusPending, StatusCompleted, StatusArchived:
		return s, nil
	default:
		return "", ErrInvalidStatus
	}
}

func (s TodoStatus) String() string {
	return string(s)
}
