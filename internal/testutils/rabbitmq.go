package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcRabbitMQ "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wagslane/go-rabbitmq"
)

type RabbitMQContainer struct {
	container *tcRabbitMQ.RabbitMQContainer
}

func NewRabbitMQContainer(ctx context.Context, t *testing.T) *RabbitMQContainer {
	t.Helper()

	container, err := tcRabbitMQ.Run(ctx,
		"rabbitmq:3-management-alpine",
		testcontainers.WithWaitStrategy(
			wait.NewLogStrategy("Server startup complete"),
		),
	)
	if err != nil {
		t.Fatalf("failed to start rabbitmq container: %v", err)
	}

	return &RabbitMQContainer{container: container}
}

func (r *RabbitMQContainer) Connect(ctx context.Context, t *testing.T) *rabbitmq.Conn {
	t.Helper()

	var (
		conn *rabbitmq.Conn
		err  error
	)

	for range 15 {
		connStr, err := r.container.AmqpURL(ctx)
		if err != nil {
			t.Fatalf("failed to get rabbitmq connection string: %v", err)
		}

		conn, err = rabbitmq.NewConn(
			connStr,
			rabbitmq.WithConnectionOptionsLogging,
		)
		if err == nil {
			return conn
		}

		time.Sleep(1 * time.Second)
	}

	t.Fatalf("failed to connect to rabbitmq after retries: %v", err)

	return nil
}

func (r *RabbitMQContainer) Close(ctx context.Context, t *testing.T) {
	t.Helper()

	if err := r.container.Terminate(ctx); err != nil {
		t.Logf("failed to terminate rabbitmq container: %v", err)
	}
}

// StartTestConsumer creates a consumer bound to the specified exchange with the given key.
func (r *RabbitMQContainer) StartTestConsumer(t *testing.T, conn *rabbitmq.Conn, exchangeName, exchangeKind, bindingKey string) (<-chan rabbitmq.Delivery, *rabbitmq.Consumer) {
	t.Helper()

	deliveries := make(chan rabbitmq.Delivery, 10)

	consumer, err := rabbitmq.NewConsumer(
		conn,
		"", // random temporary queue name
		rabbitmq.WithConsumerOptionsQueueExclusive,
		rabbitmq.WithConsumerOptionsQueueAutoDelete,
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeKind(exchangeKind),
		rabbitmq.WithConsumerOptionsExchangeDurable, // must match publisher config
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsRoutingKey(bindingKey),
	)
	if err != nil {
		t.Fatalf("failed to create test consumer: %v", err)
	}

	go func() {
		err := consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
			deliveries <- d
			return rabbitmq.Ack
		})
		if err != nil {
			t.Logf("consumer stopped: %v", err)
		}
	}()

	return deliveries, consumer
}
