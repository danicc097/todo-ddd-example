package domain

import (
	"context"

	"github.com/google/uuid"
)

// EventPublisher is an output port
//
//counterfeiter:generate . EventPublisher
type EventPublisher interface {
	PublishTodoCreated(ctx context.Context, todo *Todo) error
	PublishTodoUpdated(ctx context.Context, todo *Todo) error
	PublishTagAdded(ctx context.Context, todoID uuid.UUID, tagID uuid.UUID) error
}
