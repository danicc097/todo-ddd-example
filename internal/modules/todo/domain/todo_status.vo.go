package domain

import "errors"

var ErrInvalidStatus = errors.New("invalid status transition")

type TodoStatus string

const (
	StatusPending   TodoStatus = "PENDING"
	StatusCompleted TodoStatus = "COMPLETED"
	StatusArchived  TodoStatus = "ARCHIVED"
)

func (s TodoStatus) String() string {
	return string(s)
}
