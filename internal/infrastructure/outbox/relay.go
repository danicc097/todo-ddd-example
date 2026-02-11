package outbox

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/utils/pointers"
)

type Handler func(ctx context.Context, payload []byte) error

type Relay struct {
	pool         *pgxpool.Pool
	q            *db.Queries
	handlers     map[string]Handler
	tracer       trace.Tracer
	metricLag    metric.Int64Gauge
	metricMaxAge metric.Float64Gauge
	wg           sync.WaitGroup
}

func NewRelay(pool *pgxpool.Pool) *Relay {
	meter := otel.Meter("outbox-relay")

	lag, _ := meter.Int64Gauge("outbox.lag_count", metric.WithDescription("Number of unprocessed events"))
	age, _ := meter.Float64Gauge("outbox.max_age_seconds", metric.WithDescription("Age of oldest unprocessed event"))

	return &Relay{
		pool:         pool,
		q:            db.New(),
		handlers:     make(map[string]Handler),
		tracer:       otel.Tracer("outbox-relay"),
		metricLag:    lag,
		metricMaxAge: age,
	}
}

func (r *Relay) Register(eventType string, h Handler) {
	r.handlers[eventType] = h
}

func (r *Relay) Start(ctx context.Context) {
	slog.InfoContext(ctx, "Outbox relay worker started")

	ticker := time.NewTicker(1 * time.Second)
	metricsTicker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()
	defer metricsTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Outbox relay shutting down, waiting for active batch to finish...")
			r.wg.Wait()
			slog.InfoContext(ctx, "Outbox relay stopped")

			return
		case <-metricsTicker.C:
			r.updateMetrics(ctx)
		case <-ticker.C:
			r.wg.Add(1)
			// Ensure the DB tx finishes even if main ctx is cancelled
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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

	if len(events) > 0 {
		slog.DebugContext(ctx, "processing outbox batch", slog.Int("count", len(events)))
	}

	for _, event := range events {
		h, ok := r.handlers[event.EventType]
		if !ok { // treat as handled
			if err := r.q.MarkOutboxEventProcessed(ctx, tx, event.ID); err != nil {
				slog.ErrorContext(ctx, "failed to mark orphan event processed", slog.String("error", err.Error()))
			}

			continue
		}

		if err := h(ctx, event.Payload); err != nil {
			slog.WarnContext(ctx, "event handler failed, updating retries", slog.String("id", event.ID.String()), slog.String("error", err.Error()))

			if dbErr := r.q.UpdateOutboxRetries(ctx, tx, db.UpdateOutboxRetriesParams{
				ID:        event.ID,
				LastError: pointers.New(err.Error()),
			}); dbErr != nil {
				slog.ErrorContext(ctx, "failed to update retries", slog.String("error", dbErr.Error()))
			}

			continue
		}

		if err := r.q.MarkOutboxEventProcessed(ctx, tx, event.ID); err != nil {
			slog.ErrorContext(ctx, "failed to mark event processed", slog.String("error", err.Error()))
			return // abort the tx to avoid reprocessing
		}
	}

	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to commit outbox transaction", slog.String("error", err.Error()))
	}
}
