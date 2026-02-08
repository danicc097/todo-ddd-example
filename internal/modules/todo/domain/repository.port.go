package domain

import (
	"context"

	"github.com/google/uuid"
)

type TodoRepository interface {
	Save(ctx context.Context, todo *Todo) (uuid.UUID, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	FindAll(ctx context.Context) ([]*Todo, error)
	Update(ctx context.Context, todo *Todo) error
}
