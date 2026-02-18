package testutils

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wagslane/go-rabbitmq"
)

var (
	globalRabbitOnce      sync.Once
	globalRabbitContainer *RabbitMQContainer
	globalRabbitErr       error
)

type RabbitMQContainer struct {
	container testcontainers.Container
	amqpURI   string
}

func GetGlobalRabbitMQ(t *testing.T) *RabbitMQContainer {
	ctx := context.Background()

	globalRabbitOnce.Do(func() {
		globalRabbitContainer, globalRabbitErr = newRabbitMQContainer(ctx)
	})

	if globalRabbitErr != nil {
		t.Fatalf("Failed to initialize global rabbitmq container: %v", globalRabbitErr)
	}

	return globalRabbitContainer
}

func newRabbitMQContainer(ctx context.Context) (*RabbitMQContainer, error) {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	_ = os.Setenv("TESTCONTAINERS_REUSE_ENABLE", "true")

	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-management-alpine",
		Name:         "todo-ddd-test-rmq",
		ExposedPorts: []string{"5672/tcp", "15672/tcp"},
		WaitingFor: wait.ForLog("Server startup complete").
			WithStartupTimeout(60 * time.Second),
		SkipReaper: true,
		Labels: map[string]string{
			"todo-ddd-test": "true", // cleanup watchdog
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start rabbitmq container: %w", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5672")
	amqpURI := fmt.Sprintf("amqp://guest:guest@%s:%s/", host, port.Port())

	return &RabbitMQContainer{
		container: container,
		amqpURI:   amqpURI,
	}, nil
}

func (r *RabbitMQContainer) Connect(ctx context.Context, t *testing.T) *rabbitmq.Conn {
	t.Helper()

	var (
		conn *rabbitmq.Conn
		err  error
	)

	for range 50 {
		conn, err = rabbitmq.NewConn(r.amqpURI, rabbitmq.WithConnectionOptionsLogging)
		if err == nil {
			return conn
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("failed to connect to rabbitmq after retries: %v", err)

	return nil
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
		rabbitmq.WithConsumerOptionsExchangeDurable,
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
