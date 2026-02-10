package messaging

import (
	"context"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

// MultiPublisher implements domain.EventPublisher.
type MultiPublisher struct {
	publishers []domain.EventPublisher
}

func NewMultiPublisher(pubs ...domain.EventPublisher) *MultiPublisher {
	return &MultiPublisher{
		publishers: pubs,
	}
}

func (m *MultiPublisher) PublishTodoCreated(ctx context.Context, todo *domain.Todo) error {
	for _, p := range m.publishers {
		if err := p.PublishTodoCreated(ctx, todo); err != nil {
			return err
		}
	}

	return nil
}

func (m *MultiPublisher) PublishTodoUpdated(ctx context.Context, todo *domain.Todo) error {
	for _, p := range m.publishers {
		if err := p.PublishTodoUpdated(ctx, todo); err != nil {
			return err
		}
	}

	return nil
}

func (m *MultiPublisher) PublishTagAdded(ctx context.Context, todoID uuid.UUID, tagID uuid.UUID) error {
	for _, p := range m.publishers {
		if err := p.PublishTagAdded(ctx, todoID, tagID); err != nil {
			return err
		}
	}

	return nil
}
