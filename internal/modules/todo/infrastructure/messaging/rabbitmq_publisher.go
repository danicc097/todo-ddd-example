package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
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

func (p *RabbitMQPublisher) publish(ctx context.Context, routingKey string, todo *domain.Todo) error {
	body, err := json.Marshal(map[string]any{
		"id":         todo.ID(),
		"title":      todo.Title().String(),
		"status":     todo.Status(),
		"created_at": todo.CreatedAt(),
	})
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
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Type:         routingKey,
		},
	)
}
