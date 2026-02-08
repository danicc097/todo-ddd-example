package domain

import (
	"context"
	"github.com/google/uuid"
)

//counterfeiter:generate . TodoRepository
type TodoRepository interface {
	Save(ctx context.Context, todo *Todo) (uuid.UUID, error)
	AddTag(ctx context.Context, todoID uuid.UUID, tagID uuid.UUID) error
	SaveEvent(ctx context.Context, eventType string, payload any) error
	FindByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	FindAll(ctx context.Context) ([]*Todo, error)
	Update(ctx context.Context, todo *Todo) error
}
