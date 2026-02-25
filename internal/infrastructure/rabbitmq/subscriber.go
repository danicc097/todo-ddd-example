package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wagslane/go-rabbitmq"
)

type Handler func(ctx context.Context, delivery rabbitmq.Delivery) error

type Subscriber struct {
	conn *rabbitmq.Conn
}

func NewSubscriber(conn *rabbitmq.Conn) *Subscriber {
	return &Subscriber{conn: conn}
}

// Subscribe starts a consumer for a specific queue and exchange pattern.
func (s *Subscriber) Subscribe(
	queue string,
	exchange string,
	routingKeys []string,
	handler Handler,
	options ...func(*rabbitmq.ConsumerOptions),
) (*rabbitmq.Consumer, error) {
	opts := []func(*rabbitmq.ConsumerOptions){
		rabbitmq.WithConsumerOptionsQueueDurable,
		rabbitmq.WithConsumerOptionsExchangeName(exchange),
		rabbitmq.WithConsumerOptionsExchangeKind("topic"),
		rabbitmq.WithConsumerOptionsExchangeDurable,
	}
	for _, rk := range routingKeys {
		opts = append(opts, rabbitmq.WithConsumerOptionsRoutingKey(rk))
	}

	opts = append(opts, options...)

	consumer, err := rabbitmq.NewConsumer(s.conn, queue, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	go func() {
		err := consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
			if err := handler(context.Background(), d); err != nil {
				slog.Error("subscriber handler failed",
					slog.String("queue", queue),
					slog.String("error", err.Error()),
				)

				return rabbitmq.NackRequeue // basic retry
			}

			return rabbitmq.Ack
		})
		if err != nil {
			slog.Error("consumer stopped", slog.String("queue", queue), slog.String("error", err.Error()))
		}
	}()

	return consumer, nil
}
