package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
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

func MakeCreatedHandler(pub shared.EventPublisher) func(context.Context, []byte) error {
	return NewEventHandler(func(ctx context.Context, p TodoEventDTO) error {
		evt := domain.TodoCreatedEvent{
			ID:        p.ID,
			Title:     p.Title,
			Status:    p.Status,
			CreatedAt: p.CreatedAt,
			Occurred:  time.Now(),
		}

		return pub.Publish(ctx, evt)
	})
}

func MakeUpdatedHandler(pub shared.EventPublisher) func(context.Context, []byte) error {
	return NewEventHandler(func(ctx context.Context, p TodoEventDTO) error {
		evt := domain.TodoCompletedEvent{
			ID:        p.ID,
			Title:     p.Title,
			Status:    p.Status,
			CreatedAt: p.CreatedAt,
			Occurred:  time.Now(),
		}

		return pub.Publish(ctx, evt)
	})
}

func MakeTagAddedHandler(pub shared.EventPublisher) func(context.Context, []byte) error {
	return NewEventHandler(func(ctx context.Context, p TagAddedEventDTO) error {
		evt := domain.TagAddedEvent{
			TodoID:   p.TodoID,
			TagID:    p.TagID,
			Occurred: time.Now(),
		}

		return pub.Publish(ctx, evt)
	})
}
