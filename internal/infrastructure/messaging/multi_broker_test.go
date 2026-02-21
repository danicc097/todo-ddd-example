package messaging_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
)

func TestMultiBroker_Publish(t *testing.T) {
	t.Parallel()

	args := messaging.PublishArgs{
		EventType: "test",
		AggID:     uuid.New(),
		Payload:   nil,
		Headers:   nil,
	}

	t.Run("publishes to all brokers", func(t *testing.T) {
		count1, count2 := 0, 0

		broker1 := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
			count1++
			return nil
		})

		broker2 := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
			count2++
			return nil
		})

		multi := messaging.NewMultiBroker(broker1, broker2)
		err := multi.Publish(context.Background(), args)

		require.NoError(t, err)
		assert.Equal(t, 1, count1)
		assert.Equal(t, 1, count2)
	})

	t.Run("returns error if any broker fails", func(t *testing.T) {
		broker1 := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
			return errors.New("fail")
		})

		multi := messaging.NewMultiBroker(broker1)
		err := multi.Publish(context.Background(), args)

		assert.Error(t, err)
	})
}
