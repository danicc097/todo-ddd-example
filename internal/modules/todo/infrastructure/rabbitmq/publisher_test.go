package rabbitmq_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/rabbitmq"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestPublisher_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rmq := testutils.GetGlobalRabbitMQ(t)

	conn := rmq.Connect(ctx, t)
	defer conn.Close()

	exchange := "test-todo-events-" + uuid.New().String() // prevent cross-talk

	publisher, err := rabbitmq.NewPublisher(conn, exchange)
	require.NoError(t, err)

	defer publisher.Close()

	title := "test title"

	todoID := uuid.New()
	evt := domain.TodoCreatedEvent{
		ID:          domain.TodoID{UUID: todoID},
		Status:      "PENDING",
		Title:       title,
		WorkspaceID: wsDomain.WorkspaceID{UUID: uuid.New()},
	}

	deliveries, _ := rmq.StartTestConsumer(t, conn, exchange, "topic", "#")

	require.Eventually(t, func() bool {
		err = publisher.Publish(ctx, evt)
		if err != nil {
			return false
		}

		select {
		case d := <-deliveries:
			var received domain.TodoCreatedEvent
			if err := json.Unmarshal(d.Body, &received); err != nil {
				return false
			}

			ok := received.ID.UUID == todoID && received.Title == title

			return ok

		case <-time.After(50 * time.Millisecond):
			return false
		}
	}, 10*time.Second, 100*time.Millisecond, "expected to receive the published message")
}
