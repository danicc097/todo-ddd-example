package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

// UnitOfWork defines an interface for managing atomic operations and domain events.
type UnitOfWork interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error

	// Collect registers an aggregate and its mapper to have its events saved automatically.
	Collect(ctx context.Context, mapper domain.EventMapper, agg domain.EventsAggregate)
}
