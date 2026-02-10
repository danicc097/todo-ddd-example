package messaging_test

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcRabbitMQ "github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain/domainfakes"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/messaging"
)

var _ domain.EventPublisher = (*domainfakes.FakeEventPublisher)(nil)

func TestRabbitMQPublisher_PublishTodoCreated(t *testing.T) {
	ctx := context.Background()

	rabbitmqContainer, err := tcRabbitMQ.Run(ctx,
		"rabbitmq:3-management-alpine",
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate rabbitmq container: %s", err)
		}
	})

	connStr, err := rabbitmqContainer.AmqpURL(ctx)
	require.NoError(t, err)

	conn, err := amqp.Dial(connStr)
	require.NoError(t, err)

	defer conn.Close()

	publisher, err := messaging.NewRabbitMQPublisher(conn)
	require.NoError(t, err)

	title, err := domain.NewTodoTitle("Test Todo")
	require.NoError(t, err)

	todo := domain.NewTodo(title)

	err = publisher.PublishTodoCreated(ctx, todo)
	assert.NoError(t, err)
}

func TestRabbitMQPublisher_PublishTodoUpdated(t *testing.T) {
	ctx := context.Background()

	rabbitmqContainer, err := tcRabbitMQ.Run(ctx,
		"rabbitmq:3-management-alpine",
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate rabbitmq container: %s", err)
		}
	})

	connStr, err := rabbitmqContainer.AmqpURL(ctx)
	require.NoError(t, err)

	conn, err := amqp.Dial(connStr)
	require.NoError(t, err)

	defer conn.Close()

	publisher, err := messaging.NewRabbitMQPublisher(conn)
	require.NoError(t, err)

	title, err := domain.NewTodoTitle("Test Todo")
	require.NoError(t, err)

	todo := domain.NewTodo(title)

	err = publisher.PublishTodoUpdated(ctx, todo)
	assert.NoError(t, err)
}
