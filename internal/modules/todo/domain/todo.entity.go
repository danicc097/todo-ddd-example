package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrTodoNotFound = errors.New("todo not found")

type Todo struct {
	id        uuid.UUID
	title     TodoTitle
	status    TodoStatus
	createdAt time.Time
}

func NewTodo(id uuid.UUID, title TodoTitle, status TodoStatus, createdAt time.Time) *Todo {
	return &Todo{
		id:        id,
		title:     title,
		status:    status,
		createdAt: createdAt,
	}
}

func CreateTodo(title TodoTitle) *Todo {
	return NewTodo(uuid.New(), title, StatusPending, time.Now())
}

func (t *Todo) Complete() error {
	if t.status == StatusArchived {
		return ErrInvalidStatus
	}
	t.status = StatusCompleted
	return nil
}

func (t *Todo) ID() uuid.UUID        { return t.id }
func (t *Todo) Title() TodoTitle     { return t.title }
func (t *Todo) Status() TodoStatus   { return t.status }
func (t *Todo) CreatedAt() time.Time { return t.createdAt }

func (t *Todo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        uuid.UUID  `json:"id"`
		Title     string     `json:"title"`
		Status    TodoStatus `json:"status"`
		CreatedAt time.Time  `json:"created_at"`
	}{
		ID:        t.id,
		Title:     t.title.String(),
		Status:    t.status,
		CreatedAt: t.createdAt,
	})
}
