package redis

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
)

type Publisher struct {
	client *redis.Client
}

var _ messaging.Broker = (*Publisher)(nil)

func NewPublisher(client *redis.Client) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) Publish(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error {
	ctx, span := otel.Tracer("redis-pub").Start(ctx, "redis.publish_event",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKey.String("redis"),
			semconv.MessagingDestinationName(cache.Keys.TodoAPIUpdatesChannel()),
			semconv.PeerServiceKey.String("redis"),
		),
	)
	defer span.End()

	channel := cache.Keys.TodoAPIUpdatesChannel()

	if err := p.client.Publish(ctx, channel, payload).Err(); err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "redis publish failed", slog.String("error", err.Error()))

		return err
	}

	return nil
}
