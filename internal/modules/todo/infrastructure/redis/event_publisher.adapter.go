package redis

import (
	"context"
	"encoding/json"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/redis/go-redis/v9"
)

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}

func (p *RedisPublisher) PublishTodoCreated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, "todo.created", todo)
}

func (p *RedisPublisher) PublishTodoUpdated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, "todo.updated", todo)
}

func (p *RedisPublisher) publish(ctx context.Context, eventType string, todo *domain.Todo) error {
	msg, _ := json.Marshal(map[string]any{
		"id":     todo.ID(),
		"status": todo.Status(),
		"title":  todo.Title().String(),
		"event":  eventType,
	})
	return p.client.Publish(ctx, "todo_updates", msg).Err()
}
