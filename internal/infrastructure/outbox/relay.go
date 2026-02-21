package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/utils/pointers"
)

// Relay polls the database for outbox events and publishes them to the message broker.
// It provides at-least-once deliveries with infinite backoff.
type Relay struct {
	pool         *pgxpool.Pool
	q            *db.Queries
	broker       messaging.Broker
	tracer       trace.Tracer
	metricLag    metric.Int64Gauge
	metricMaxAge metric.Float64Gauge
	wg           sync.WaitGroup
}

func NewRelay(pool *pgxpool.Pool, broker messaging.Broker) *Relay {
	meter := otel.Meter("outbox-relay")

	lag, _ := meter.Int64Gauge("outbox.lag_count", metric.WithDescription("Number of unprocessed events"))
	age, _ := meter.Float64Gauge("outbox.max_age_seconds", metric.WithDescription("Age of oldest unprocessed event"))

	return &Relay{
		pool:         pool,
		q:            db.New(),
		broker:       broker,
		tracer:       otel.Tracer("outbox-relay"),
		metricLag:    lag,
		metricMaxAge: age,
	}
}

func (r *Relay) Start(ctx context.Context) {
	slog.InfoContext(ctx, "Outbox relay worker started")

	ticker := time.NewTicker(1 * time.Second)
	metricsTicker := time.NewTicker(5 * time.Second)
	cleanupTicker := time.NewTicker(1 * time.Hour)

	defer ticker.Stop()
	defer metricsTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Outbox relay shutting down, waiting for active batch to finish...")
			r.wg.Wait()
			slog.InfoContext(ctx, "Outbox relay stopped")

			return
		case <-metricsTicker.C:
			r.updateMetrics(ctx)
		case <-cleanupTicker.C:
			if err := r.q.DeleteProcessedOutboxEvents(ctx, r.pool); err != nil {
				slog.ErrorContext(ctx, "failed to cleanup outbox", slog.String("error", err.Error()))
			}
		case <-ticker.C:
			r.wg.Add(1)
			r.processEvents(context.WithoutCancel(ctx))
			r.wg.Done()
		}
	}
}

func (r *Relay) updateMetrics(ctx context.Context) {
	stats, err := r.q.GetOutboxLag(ctx, r.pool)
	if err == nil {
		r.metricLag.Record(ctx, stats.TotalLag)
		r.metricMaxAge.Record(ctx, stats.MaxAgeSeconds)
	}
}

func (r *Relay) processEvents(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to begin outbox transaction", slog.String("error", err.Error()))
		return
	}
	defer tx.Rollback(ctx)

	events, err := r.q.GetUnprocessedOutboxEvents(ctx, tx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "failed to fetch outbox events", slog.String("error", err.Error()))
		}

		return
	}

	if len(events) == 0 {
		return
	}

	ctx, span := r.tracer.Start(ctx, "outbox.relay_batch", trace.WithAttributes(
		attribute.Int("batch.size", len(events)),
	))
	defer span.End()

	slog.DebugContext(ctx, "processing outbox batch", slog.Int("count", len(events)))

	for _, event := range events {
		r.processSingleEvent(ctx, tx, event)
	}

	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to commit outbox batch locks", slog.String("error", err.Error()))
	}
}

func (r *Relay) processSingleEvent(ctx context.Context, tx db.DBTX, event db.Outbox) {
	var rawHeaders map[string]string

	idAttr := slog.String("event_id", event.ID.String())

	unmarshalErr := json.Unmarshal(event.Headers, &rawHeaders)

	spanCtx := ctx

	if unmarshalErr == nil {
		carrier := propagation.MapCarrier(rawHeaders)
		spanCtx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	pubCtx, span := r.tracer.Start(spanCtx, "outbox.publish_event",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithLinks(trace.LinkFromContext(ctx)),
		trace.WithAttributes(
			attribute.String("event.type", string(event.EventType)),
			attribute.String("aggregate.id", event.AggregateID.String()),
			attribute.String("aggregate.type", event.AggregateType),
		),
	)
	defer span.End()

	if unmarshalErr != nil {
		span.RecordError(unmarshalErr)
		span.SetStatus(codes.Error, "failed to unmarshal headers")

		slog.ErrorContext(ctx, "failed to unmarshal headers", idAttr, slog.String("error", unmarshalErr.Error()))
		_ = r.q.UpdateOutboxRetries(ctx, tx, db.UpdateOutboxRetriesParams{
			ID:        event.ID,
			LastError: pointers.New("fatal: invalid header JSON"),
		})

		return
	}

	headers := make(map[messaging.Header]string, len(rawHeaders))
	for k, v := range rawHeaders {
		headers[messaging.Header(k)] = v
	}

	pubCtx, cancel := context.WithTimeout(pubCtx, 2*time.Second)
	defer cancel()

	args := messaging.PublishArgs{
		EventType: string(event.EventType),
		AggID:     event.AggregateID,
		Payload:   event.Payload,
		Headers:   headers,
	}

	err := r.broker.Publish(pubCtx, args)
	if err == nil {
		if dbErr := r.q.MarkOutboxEventProcessed(ctx, tx, event.ID); dbErr != nil {
			slog.ErrorContext(ctx, "failed to mark event processed", idAttr, slog.String("error", dbErr.Error()))
		}

		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, "broker publish failed")
	slog.WarnContext(ctx, "event publish failed", idAttr, slog.String("error", err.Error()))

	if dbErr := r.q.UpdateOutboxRetries(ctx, tx, db.UpdateOutboxRetriesParams{
		ID:        event.ID,
		LastError: pointers.New(err.Error()),
	}); dbErr != nil {
		slog.ErrorContext(ctx, "failed to update retries", idAttr, slog.String("error", dbErr.Error()))
	}
}
