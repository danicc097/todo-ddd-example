package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestRedisPublisher_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rc := testutils.NewRedisContainer(ctx, t)
	defer rc.Close(ctx, t)

	client := rc.Connect(ctx, t)

	pubsub := client.Subscribe(ctx, "todo_updates")
	defer pubsub.Close()

	_, err := pubsub.Receive(ctx) // block until subscription is active
	require.NoError(t, err)

	title := "test title"

	publisher := redis.NewRedisPublisher(client)
	todoID := uuid.New()
	evt := domain.TodoCreatedEvent{
		ID:     domain.TodoID{UUID: todoID},
		Status: "PENDING",
		Title:  title,
	}

	err = publisher.Publish(ctx, evt)
	require.NoError(t, err)

	select {
	case msg := <-pubsub.Channel():
		assert.Contains(t, msg.Payload, todoID.String())
		assert.Contains(t, msg.Payload, title)
		assert.Contains(t, msg.Payload, "todo.created")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for redis message")
	}
}
