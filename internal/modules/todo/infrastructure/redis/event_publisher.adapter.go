package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type RedisPublisher struct {
	client *redis.Client
	tracer trace.Tracer
}

func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{
		client: client,
		tracer: otel.Tracer("redis-publisher"),
	}
}

// Publish implements shared.EventPublisher.
func (p *RedisPublisher) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	for _, event := range events {
		if err := p.publishOne(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (p *RedisPublisher) publishOne(ctx context.Context, event domain.DomainEvent) error {
	var payload map[string]any

	switch e := event.(type) {
	case todoDomain.TodoCreatedEvent:
		payload = map[string]any{
			"id":           e.ID,
			"status":       e.Status,
			"title":        e.Title,
			"event":        e.EventName(),
			"workspace_id": e.WorkspaceID,
		}
	case todoDomain.TodoCompletedEvent:
		payload = map[string]any{
			"id":           e.ID,
			"status":       e.Status,
			"title":        e.Title,
			"event":        e.EventName(),
			"workspace_id": e.WorkspaceID,
		}
	case todoDomain.TagAddedEvent:
		payload = map[string]any{
			"todo_id": e.TodoID,
			"tag_id":  e.TagID,
			"event":   e.EventName(),
		}
	default:
		slog.WarnContext(ctx, "redis publisher: unknown event type", slog.String("type", fmt.Sprintf("%T", event)))
		return nil
	}

	return p.publishToRedis(ctx, event.EventName(), payload, attribute.String("aggregate.id", event.AggregateID().String()))
}

func (p *RedisPublisher) publishToRedis(ctx context.Context, eventType string, payload any, attrs ...attribute.KeyValue) error {
	traceAttrs := append([]attribute.KeyValue{
		attribute.String("event.type", eventType),
		attribute.String("redis.channel", "todo_updates"),
	}, attrs...)

	ctx, span := p.tracer.Start(ctx, "redis.publish", trace.WithAttributes(traceAttrs...))
	defer span.End()

	msg, err := json.Marshal(payload)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal redis payload", slog.String("error", err.Error()))
		return err
	}

	err = p.client.Publish(ctx, "todo_updates", msg).Err()
	if err != nil {
		slog.ErrorContext(ctx, "redis publish failed", slog.String("channel", "todo_updates"), slog.String("error", err.Error()))
		return err
	}

	slog.InfoContext(ctx, "published to redis", slog.String("channel", "todo_updates"), slog.String("event", eventType))

	return nil
}
