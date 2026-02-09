package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type TodoEventPayload struct {
	ID        uuid.UUID         `json:"id"`
	Title     string            `json:"title"`
	Status    domain.TodoStatus `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
}

func (p TodoEventPayload) ToEntity() *domain.Todo {
	title, _ := domain.NewTodoTitle(p.Title)
	return domain.NewTodo(p.ID, title, p.Status, p.CreatedAt)
}

func MakeCreatedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var data TodoEventPayload
		if err := json.Unmarshal(payload, &data); err != nil {
			return err
		}
		return pub.PublishTodoCreated(ctx, data.ToEntity())
	}
}

func MakeUpdatedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var data TodoEventPayload
		if err := json.Unmarshal(payload, &data); err != nil {
			return err
		}
		return pub.PublishTodoUpdated(ctx, data.ToEntity())
	}
}
