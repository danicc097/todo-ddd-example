package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Handler func(ctx context.Context, payload []byte) error

type Relay struct {
	pool     *pgxpool.Pool
	q        *db.Queries
	handlers map[string]Handler
	tracer   trace.Tracer
}

func NewRelay(pool *pgxpool.Pool) *Relay {
	return &Relay{
		pool:     pool,
		q:        db.New(),
		handlers: make(map[string]Handler),
		tracer:   otel.Tracer("outbox-relay"),
	}
}

func (r *Relay) Register(eventType string, h Handler) {
	r.handlers[eventType] = h
}

func (r *Relay) Start(ctx context.Context) {
	slog.InfoContext(ctx, "Outbox relay worker started")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Outbox relay worker stopped")
			return
		case <-ticker.C:
			r.processEvents(ctx)
		}
	}
}

func (r *Relay) processEvents(ctx context.Context) {
	ctx, span := r.tracer.Start(ctx, "relay.poll_events")
	defer span.End()

	events, err := r.q.GetUnprocessedOutboxEvents(ctx, r.pool)
	if err != nil {
		slog.ErrorContext(ctx, "Relay fetch failed", slog.String("error", err.Error()))
		return
	}

	for _, event := range events {
		r.handleEvent(ctx, event)
	}
}

func (r *Relay) handleEvent(ctx context.Context, event db.Outbox) {
	ctx, span := r.tracer.Start(ctx, "relay.handle_event", trace.WithAttributes(
		attribute.String("event.id", event.ID.String()),
		attribute.String("event.type", event.EventType),
	))
	defer span.End()

	handler, ok := r.handlers[event.EventType]
	if !ok {
		slog.WarnContext(ctx, "Relay missing handler", slog.String("type", event.EventType))
		r.markProcessed(ctx, event.ID)
		return
	}

	if err := handler(ctx, event.Payload); err != nil {
		slog.ErrorContext(ctx, "Relay handler execution failed", slog.String("id", event.ID.String()), slog.String("error", err.Error()))
		return
	}

	slog.InfoContext(ctx, "Relay processed event", slog.String("id", event.ID.String()), slog.String("type", event.EventType))
	r.markProcessed(ctx, event.ID)
}

func (r *Relay) markProcessed(ctx context.Context, id uuid.UUID) {
	if err := r.q.MarkOutboxEventProcessed(ctx, r.pool, id); err != nil {
		slog.ErrorContext(ctx, "Relay mark status failed", slog.String("id", id.String()), slog.String("error", err.Error()))
	}
}
