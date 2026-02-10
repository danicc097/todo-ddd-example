package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/utils/pointers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	metricOutboxLag = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_unprocessed_events_count",
		Help: "Number of events waiting in the outbox",
	})
	metricOutboxMaxAge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_max_age_seconds",
		Help: "Age of the oldest unprocessed event",
	})
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
	metricsTicker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer metricsTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-metricsTicker.C:
			r.updateMetrics(ctx)
		case <-ticker.C:
			r.processEvents(ctx)
		}
	}
}

func (r *Relay) updateMetrics(ctx context.Context) {
	stats, err := r.q.GetOutboxLag(ctx, r.pool)
	if err == nil {
		metricOutboxLag.Set(float64(stats.TotalLag))
		metricOutboxMaxAge.Set(stats.MaxAgeSeconds)
	}
}

func (r *Relay) processEvents(ctx context.Context) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	events, err := r.q.GetUnprocessedOutboxEvents(ctx, tx)
	if err != nil {
		return
	}

	for _, event := range events {
		h, ok := r.handlers[event.EventType]
		if !ok { // treat as handled
			r.q.MarkOutboxEventProcessed(ctx, tx, event.ID)
			continue
		}

		if err := h(ctx, event.Payload); err != nil {
			r.q.UpdateOutboxRetries(ctx, tx, db.UpdateOutboxRetriesParams{
				ID:        event.ID,
				LastError: pointers.New(err.Error()),
			})
			continue
		}
		r.q.MarkOutboxEventProcessed(ctx, tx, event.ID)
	}
	tx.Commit(ctx)
}
