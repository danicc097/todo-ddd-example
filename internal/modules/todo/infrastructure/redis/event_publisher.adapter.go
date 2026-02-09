package redis

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

func (p *RedisPublisher) PublishTodoCreated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, "todo.created", todo)
}

func (p *RedisPublisher) PublishTodoUpdated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, "todo.updated", todo)
}

func (p *RedisPublisher) publish(ctx context.Context, eventType string, todo *domain.Todo) error {
	ctx, span := p.tracer.Start(ctx, "redis.publish", trace.WithAttributes(
		attribute.String("event.type", eventType),
		attribute.String("todo.id", todo.ID().String()),
		attribute.String("redis.channel", "todo_updates"),
	))
	defer span.End()

	payload := map[string]any{
		"id":     todo.ID(),
		"status": todo.Status(),
		"title":  todo.Title().String(),
		"event":  eventType,
	}

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
