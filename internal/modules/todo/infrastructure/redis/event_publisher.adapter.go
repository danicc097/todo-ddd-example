package redis

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
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
	payload := map[string]any{
		"id":     todo.ID(),
		"status": todo.Status(),
		"title":  todo.Title().String(),
		"event":  "todo.created",
	}
	return p.publish(ctx, "todo.created", payload, attribute.String("todo.id", todo.ID().String()))
}

func (p *RedisPublisher) PublishTodoUpdated(ctx context.Context, todo *domain.Todo) error {
	payload := map[string]any{
		"id":     todo.ID(),
		"status": todo.Status(),
		"title":  todo.Title().String(),
		"event":  "todo.updated",
	}
	return p.publish(ctx, "todo.updated", payload, attribute.String("todo.id", todo.ID().String()))
}

func (p *RedisPublisher) PublishTagAdded(ctx context.Context, todoID uuid.UUID, tagID uuid.UUID) error {
	payload := map[string]any{
		"todo_id": todoID,
		"tag_id":  tagID,
		"event":   "todo.tagadded",
	}
	return p.publish(ctx, "todo.tagadded", payload, attribute.String("todo.id", todoID.String()))
}

func (p *RedisPublisher) publish(ctx context.Context, eventType string, payload any, attrs ...attribute.KeyValue) error {
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
