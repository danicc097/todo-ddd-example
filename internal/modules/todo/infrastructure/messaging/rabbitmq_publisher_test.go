package messaging_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestRabbitMQPublisher_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mq := testutils.NewRabbitMQContainer(ctx, t)
	defer mq.Close(ctx, t)

	conn := mq.Connect(ctx, t)
	defer conn.Close()

	publisher, err := messaging.NewRabbitMQPublisher(conn)
	require.NoError(t, err)

	defer publisher.Close()

	t.Run("verifies message is routed by ID to the correct exchange", func(t *testing.T) {
		title, _ := domain.NewTodoTitle("Test Todo")
		todo := domain.NewTodo(title)

		msgs, consumer := mq.StartTestConsumer(t, conn, "todo_events", "topic", "#")
		defer consumer.Close()

		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-time.After(5 * time.Second):
					return
				case <-ticker.C: // consumer may not be ready, so spam publish
					_ = publisher.PublishTodoCreated(ctx, todo)
				}
			}
		}()

		select {
		case msg := <-msgs:
			var body map[string]any

			err := json.Unmarshal(msg.Body, &body)
			require.NoError(t, err)
			assert.Equal(t, "Test Todo", body["title"])
			assert.Equal(t, "todo.created", msg.Type)
			assert.Equal(t, todo.ID().String(), msg.RoutingKey)

		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for message")
		}
	})
}
