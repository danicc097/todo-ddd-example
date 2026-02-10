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
	publisher, err := messaging.NewRabbitMQPublisher(conn)
	require.NoError(t, err)

	t.Run("verifies actual message content", func(t *testing.T) {
		title, _ := domain.NewTodoTitle("Test Todo")
		todo := domain.NewTodo(title)

		ch, _ := conn.Channel()
		msgs, _ := ch.Consume("todo_events", "", true, false, false, false, nil)

		err = publisher.PublishTodoCreated(ctx, todo)
		require.NoError(t, err)

		select {
		case msg := <-msgs:
			var body map[string]any

			err := json.Unmarshal(msg.Body, &body)
			require.NoError(t, err)
			assert.Equal(t, "Test Todo", body["title"])
			assert.Equal(t, "todo.created", msg.Type)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for message")
		}
	})
}
