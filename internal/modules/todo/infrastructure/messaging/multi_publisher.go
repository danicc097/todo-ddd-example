package messaging

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

// MultiPublisher implements shared.EventPublisher.
type MultiPublisher struct {
	publishers []domain.EventPublisher
}

func NewMultiPublisher(pubs ...domain.EventPublisher) *MultiPublisher {
	return &MultiPublisher{
		publishers: pubs,
	}
}

func (m *MultiPublisher) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	for _, p := range m.publishers {
		if err := p.Publish(ctx, events...); err != nil {
			return err
		}
	}

	return nil
}
