package outbox

import (
	"context"
	"log"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler func(ctx context.Context, payload []byte) error

type Relay struct {
	pool     *pgxpool.Pool
	q        *db.Queries
	handlers map[string]Handler
}

func NewRelay(pool *pgxpool.Pool) *Relay {
	return &Relay{
		pool:     pool,
		q:        db.New(),
		handlers: make(map[string]Handler),
	}
}

func (r *Relay) Register(eventType string, h Handler) {
	r.handlers[eventType] = h
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
		log.Printf("relay: failed to fetch events: %v", err)
		return
	}

	for _, event := range events {
		handler, ok := r.handlers[event.EventType]
		if !ok {
			log.Printf("relay: no handler for event type: %s", event.EventType)
			r.markProcessed(ctx, event.ID)
			continue
		}

		if err := handler(ctx, event.Payload); err != nil {
			log.Printf("relay: handler failed for event %s: %v", event.ID, err)
			continue
		}

		r.markProcessed(ctx, event.ID)
	}
}

func (r *Relay) markProcessed(ctx context.Context, id uuid.UUID) {
	if err := r.q.MarkOutboxEventProcessed(ctx, r.pool, id); err != nil {
		log.Printf("relay: failed to mark event %s processed: %v", id, err)
	}
}
