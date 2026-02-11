package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/wagslane/go-rabbitmq"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type RabbitMQPublisher struct {
	publisher *rabbitmq.Publisher
}

func NewRabbitMQPublisher(conn *rabbitmq.Conn) (*RabbitMQPublisher, error) {
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

	return &RabbitMQPublisher{publisher: pub}, nil
}

func (p *RabbitMQPublisher) todoToDTO(todo *domain.Todo) TodoEventPayload {
	return TodoEventPayload{
		ID:        todo.ID(),
		Title:     todo.Title().String(),
		Status:    todo.Status().String(),
		CreatedAt: todo.CreatedAt(),
	}
}

func (p *RabbitMQPublisher) PublishTodoCreated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, todo.ID().String(), "todo.created", p.todoToDTO(todo))
}

func (p *RabbitMQPublisher) PublishTodoUpdated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, todo.ID().String(), "todo.updated", p.todoToDTO(todo))
}

func (p *RabbitMQPublisher) PublishTagAdded(ctx context.Context, todoID uuid.UUID, tagID uuid.UUID) error {
	payload := TagAddedPayload{
		TodoID: todoID,
		TagID:  tagID,
	}

	return p.publish(ctx, todoID.String(), "todo.tagadded", payload)
}

func (p *RabbitMQPublisher) publish(ctx context.Context, routingKey string, eventType string, body any) error {
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

func (p *RabbitMQPublisher) Close() {
	p.publisher.Close()
}
