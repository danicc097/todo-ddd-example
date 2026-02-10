package testutils

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	tcRabbitMQ "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RabbitMQContainer struct {
	container *tcRabbitMQ.RabbitMQContainer
	conn      *amqp.Connection
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

func (r *RabbitMQContainer) Connect(ctx context.Context, t *testing.T) *amqp.Connection {
	t.Helper()

	var conn *amqp.Connection
	var err error

	for range 15 {
		connStr, err := r.container.AmqpURL(ctx)
		if err != nil {
			t.Fatalf("failed to get rabbitmq connection string: %v", err)
		}

		conn, err = amqp.Dial(connStr)
		if err == nil {
			r.conn = conn
			return conn
		}

		time.Sleep(1 * time.Second)
	}

	t.Fatalf("failed to connect to rabbitmq after retries: %v", err)
	return nil
}

func (r *RabbitMQContainer) Close(ctx context.Context, t *testing.T) {
	t.Helper()

	if r.conn != nil {
		r.conn.Close()
	}

	if err := r.container.Terminate(ctx); err != nil {
		t.Logf("failed to terminate rabbitmq container: %v", err)
	}
}

func (r *RabbitMQContainer) Connection() *amqp.Connection {
	return r.conn
}

func (r *RabbitMQContainer) AmqpURL(ctx context.Context) string {
	url, err := r.container.AmqpURL(ctx)
	if err != nil {
		panic(err)
	}
	return url
}
