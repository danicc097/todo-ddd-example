package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	ch *amqp.Channel
}

func NewRabbitMQPublisher(conn *amqp.Connection) (*RabbitMQPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"todo_events",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}
	_ = q

	return &RabbitMQPublisher{ch: ch}, nil
}

func (p *RabbitMQPublisher) PublishTodoCreated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, "todo.created", todo)
}

func (p *RabbitMQPublisher) PublishTodoUpdated(ctx context.Context, todo *domain.Todo) error {
	return p.publish(ctx, "todo.updated", todo)
}

func (p *RabbitMQPublisher) PublishTagAdded(ctx context.Context, todoID uuid.UUID, tagID uuid.UUID) error {
	payload := TagAddedPayload{
		TodoID: todoID,
		TagID:  tagID,
	}
	return p.publish(ctx, "todo.tagadded", payload)
}

func (p *RabbitMQPublisher) publish(ctx context.Context, routingKey string, body any) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx,
		"",            // exchange
		"todo_events", // key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         bytes,
			DeliveryMode: amqp.Persistent,
			Type:         routingKey,
		},
	)
}
