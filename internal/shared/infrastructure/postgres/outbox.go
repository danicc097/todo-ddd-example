package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

// SaveDomainEvents persists events from an aggregate to the outbox.
func SaveDomainEvents(
	ctx context.Context,
	q *db.Queries,
	dbtx db.DBTX,
	mapper domain.EventMapper,
	agg domain.EventsAggregate,
) error {
	for _, e := range agg.Events() {
		eventName, payload, err := mapper.MapEvent(e)
		if err != nil {
			return err
		}

		if payload == nil {
			continue
		}

		if err := q.SaveOutboxEvent(ctx, dbtx, db.SaveOutboxEventParams{
			ID:        uuid.New(),
			EventType: eventName,
			Payload:   payload,
		}); err != nil {
			return err
		}
	}

	agg.ClearEvents()

	return nil
}
