package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
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

// NewEventHandler creates a handler that unmarshals JSON payloads into type T.
func NewEventHandler[T any](fn func(context.Context, T) error) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var event T
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}

		return fn(ctx, event)
	}
}

func MakeCreatedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return NewEventHandler(func(ctx context.Context, p TodoEventPayload) error {
		return pub.PublishTodoCreated(ctx, p.ToEntity())
	})
}

func MakeUpdatedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return NewEventHandler(func(ctx context.Context, p TodoEventPayload) error {
		return pub.PublishTodoUpdated(ctx, p.ToEntity())
	})
}

func MakeTagAddedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return NewEventHandler(func(ctx context.Context, p TagAddedPayload) error {
		return pub.PublishTagAdded(ctx, p.TodoID, p.TagID)
	})
}
