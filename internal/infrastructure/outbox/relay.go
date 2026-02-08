package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Relay struct {
	pool      *pgxpool.Pool
	q         *db.Queries
	publisher domain.EventPublisher
}

func NewRelay(pool *pgxpool.Pool, publisher domain.EventPublisher) *Relay {
	return &Relay{
		pool:      pool,
		q:         db.New(),
		publisher: publisher,
	}
}

func (r *Relay) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.processEvents(ctx)
		}
	}
}

func (r *Relay) processEvents(ctx context.Context) {
	events, err := r.q.GetUnprocessedOutboxEvents(ctx, r.pool)
	if err != nil {
		log.Printf("failed to fetch outbox events: %v", err)
		return
	}

	for _, event := range events {
		var payload map[string]any
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			log.Printf("failed to unmarshal outbox payload %s: %v", event.ID, err)
			continue // retry later
		}

		err = r.publish(ctx, event.EventType, payload)
		if err != nil {
			log.Printf("failed to relay event %s: %v", event.ID, err)
			continue // retry later
		}

		if err := r.q.MarkOutboxEventProcessed(ctx, r.pool, event.ID); err != nil {
			log.Printf("failed to mark event %s as processed: %v", event.ID, err)
		}
	}
}

func (r *Relay) publish(ctx context.Context, eventType string, data map[string]any) error {
	log.Printf("Relaying event: %s", eventType)

	switch eventType {
	case "todo.created", "todo.completed":
		// TODO: reconstruct
		return nil
	default:
		log.Printf("Warning: unknown event type ignored: %s", eventType)
		return nil
	}
}
