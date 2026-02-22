package rabbitmq

import (
	"context"

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

func (p *Publisher) Publish(ctx context.Context, args messaging.PublishArgs) error {
	routingKey := messaging.Keys.EventRoutingKey(args.EventType, args.AggID)

	ctx, span := otel.Tracer("rabbitmq").Start(ctx, "rabbitmq.publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination.name", p.exchange),
			attribute.String("messaging.rabbitmq.routing_key", routingKey),
			attribute.String("peer.service", "rabbitmq"),
		),
	)
	defer span.End()

	amqpHeaders := rabbitmq.Table{}
	for k, v := range args.Headers {
		amqpHeaders[string(k)] = v
	}

	err := p.publisher.PublishWithContext(
		ctx,
		args.Payload,
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
