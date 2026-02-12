package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wagslane/go-rabbitmq"

	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type Publisher struct {
	publisher *rabbitmq.Publisher
	exchange  string
}

func NewPublisher(conn *rabbitmq.Conn, exchangeName string) (*Publisher, error) {
	pub, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsExchangeName(exchangeName),
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return &Publisher{
		publisher: pub,
		exchange:  exchangeName,
	}, nil
}

// Publish implements shared.EventPublisher.
func (p *Publisher) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	for _, event := range events {
		if err := p.publishOne(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (p *Publisher) publishOne(ctx context.Context, event domain.DomainEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event %s: %w", event.EventName(), err)
	}

	// Routing key convention: domain.event_type.id (e.g., "todo.created.123-abc")
	// allows binding to e.g. "todo.#" or "#.created.#"
	routingKey := fmt.Sprintf("%s.%s", event.EventName(), event.AggregateID().String())

	return p.publisher.PublishWithContext(
		ctx,
		body,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsExchange(p.exchange),
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsType(event.EventName()),
		rabbitmq.WithPublishOptionsPersistentDelivery,
	)
}

func (p *Publisher) Close() {
	p.publisher.Close()
}
