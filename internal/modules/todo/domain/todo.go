package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyTitle   = errors.New("title cannot be empty")
	ErrTodoNotFound = errors.New("todo not found")
)

type Todo struct {
	ID        uuid.UUID
	Title     string
	Completed bool
	CreatedAt time.Time
}

func NewTodo(title string) (*Todo, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}
	return &Todo{
		ID:        uuid.New(),
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	}, nil
}

func (t *Todo) Complete() {
	t.Completed = true
}

type TodoRepository interface {
	Save(ctx context.Context, todo *Todo) (uuid.UUID, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	FindAll(ctx context.Context) ([]*Todo, error)
	Update(ctx context.Context, todo *Todo) error
}
