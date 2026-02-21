package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type EventEnvelope struct {
	Event     string          `json:"event"`
	Version   int             `json:"version"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// SaveDomainEvents persists events from an aggregate to the outbox.
func SaveDomainEvents(
	ctx context.Context,
	q *db.Queries,
	dbtx db.DBTX,
	mapper domain.EventMapper,
	agg domain.EventsAggregate,
) error {
	ctx, span := otel.Tracer("outbox").Start(ctx, "SaveDomainEvents", trace.WithAttributes(
		semconv.DBSystemNamePostgreSQL,
		semconv.PeerServiceKey.String("postgres"),
	))
	defer span.End()

	for _, e := range agg.Events() {
		eventName, rawPayload, err := mapper.MapEvent(e)
		if err != nil {
			return err
		}

		if rawPayload == nil {
			continue
		}

		envelope := EventEnvelope{
			Event:     string(eventName),
			Version:   1,
			Timestamp: e.OccurredAt(),
		}

		// handle both cases where mapper might already return bytes or a struct
		if b, ok := rawPayload.([]byte); ok {
			envelope.Data = b
		} else {
			b, err := json.Marshal(rawPayload)
			if err != nil {
				return err
			}

			envelope.Data = b
		}

		payload, err := json.Marshal(envelope)
		if err != nil {
			return err
		}

		headers := make(map[string]string)
		otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))

		if we, ok := e.(domain.WorkspacedEvent); ok {
			headers[string(messaging.RoutingWorkspaceID)] = we.WorkspaceID().String()
		}

		headersJSON, err := json.Marshal(headers)
		if err != nil {
			return err
		}

		if err := q.SaveOutboxEvent(ctx, dbtx, db.SaveOutboxEventParams{
			ID:            uuid.New(),
			EventType:     eventName,
			AggregateType: e.AggregateType().String(),
			AggregateID:   e.AggregateID(),
			Payload:       payload,
			Headers:       headersJSON,
		}); err != nil {
			return err
		}
	}

	agg.ClearEvents()

	return nil
}
