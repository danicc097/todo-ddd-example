package rabbitmq

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
)

type Publisher struct {
	publisher *rabbitmq.Publisher
	exchange  string
}

var _ messaging.Broker = (*Publisher)(nil)

func NewPublisher(conn *rabbitmq.Conn, exchange string) (*Publisher, error) {
	pub, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsExchangeName(exchange),
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{publisher: pub, exchange: exchange}, nil
}

func (p *Publisher) Publish(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error {
	ctx, span := otel.Tracer("rabbitmq").Start(ctx, "rabbitmq.publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination.name", p.exchange),
			attribute.String("messaging.rabbitmq.routing_key", fmt.Sprintf("%s.%s", eventType, aggID.String())),
			attribute.String("peer.service", "rabbitmq"),
		),
	)
	defer span.End()

	routingKey := fmt.Sprintf("%s.%s", eventType, aggID.String())

	amqpHeaders := rabbitmq.Table{}
	for k, v := range headers {
		amqpHeaders[k] = v
	}

	err := p.publisher.PublishWithContext(
		ctx,
		payload,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsExchange(p.exchange),
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsHeaders(amqpHeaders),
	)
	if err != nil {
		span.RecordError(err)
	}

	return err
}

func (p *Publisher) Close() { p.publisher.Close() }
