package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/wagslane/go-rabbitmq"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type Publisher struct {
	publisher *rabbitmq.Publisher
}

func NewPublisher(conn *rabbitmq.Conn) (*Publisher, error) {
	pub, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsExchangeName("todo_events"),
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return &Publisher{publisher: pub}, nil
}

func (p *Publisher) PublishTodoCreated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, todo.ID().String(), "todo.created", ToTodoEventDTO(todo))
}

func (p *Publisher) PublishTodoUpdated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, todo.ID().String(), "todo.updated", ToTodoEventDTO(todo))
}

func (p *Publisher) PublishTagAdded(ctx context.Context, todoID uuid.UUID, tagID uuid.UUID) error {
	payload := TagAddedEventDTO{
		TodoID: todoID,
		TagID:  tagID,
	}

	return p.publish(ctx, todoID.String(), "todo.tagadded", payload)
}

func (p *Publisher) publish(ctx context.Context, routingKey string, eventType string, body any) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return p.publisher.PublishWithContext(
		ctx,
		bytes,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsExchange("todo_events"),
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsType(eventType),
		rabbitmq.WithPublishOptionsPersistentDelivery,
	)
}

func (p *Publisher) Close() {
	p.publisher.Close()
}
