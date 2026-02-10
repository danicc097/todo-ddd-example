package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type TodoEventPayload struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (p TodoEventPayload) ToEntity() *domain.Todo {
	title, _ := domain.NewTodoTitle(p.Title)
	status := domain.TodoStatus(p.Status)
	return domain.ReconstituteTodo(p.ID, title, status, p.CreatedAt, nil)
}

type TagAddedPayload struct {
	TodoID uuid.UUID `json:"todo_id"`
	TagID  uuid.UUID `json:"tag_id"`
}

func MakeCreatedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var p TodoEventPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		return pub.PublishTodoCreated(ctx, p.ToEntity())
	}
}

func MakeUpdatedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var p TodoEventPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		return pub.PublishTodoUpdated(ctx, p.ToEntity())
	}
}

func MakeTagAddedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var p TagAddedPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		return pub.PublishTagAdded(ctx, p.TodoID, p.TagID)
	}
}
