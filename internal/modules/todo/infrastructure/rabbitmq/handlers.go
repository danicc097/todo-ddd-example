package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

func (d TodoEventDTO) ToEntity() *domain.Todo {
	title, _ := domain.NewTodoTitle(d.Title)
	status := domain.TodoStatus(d.Status)

	return domain.ReconstituteTodo(d.ID, title, status, d.CreatedAt, nil)
}

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
	return NewEventHandler(func(ctx context.Context, p TodoEventDTO) error {
		return pub.PublishTodoCreated(ctx, p.ToEntity())
	})
}

func MakeUpdatedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return NewEventHandler(func(ctx context.Context, p TodoEventDTO) error {
		return pub.PublishTodoUpdated(ctx, p.ToEntity())
	})
}

func MakeTagAddedHandler(pub domain.EventPublisher) func(context.Context, []byte) error {
	return NewEventHandler(func(ctx context.Context, p TagAddedEventDTO) error {
		return pub.PublishTagAdded(ctx, p.TodoID, p.TagID)
	})
}
