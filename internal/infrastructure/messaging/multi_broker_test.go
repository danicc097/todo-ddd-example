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

type mockBroker struct {
	publishFunc func(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error
}

func (m *mockBroker) Publish(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error {
	return m.publishFunc(ctx, eventType, aggID, payload, headers)
}

var _ messaging.Broker = (*mockBroker)(nil)

func TestMultiBroker_Publish(t *testing.T) {
	t.Parallel()

	t.Run("publishes to all brokers", func(t *testing.T) {
		count1 := 0
		broker1 := &mockBroker{
			publishFunc: func(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error {
				count1++
				return nil
			},
		}

		count2 := 0
		broker2 := &mockBroker{
			publishFunc: func(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error {
				count2++
				return nil
			},
		}

		multi := messaging.NewMultiBroker(broker1, broker2)
		err := multi.Publish(context.Background(), "test", uuid.New(), nil, nil)

		require.NoError(t, err)
		assert.Equal(t, 1, count1)
		assert.Equal(t, 1, count2)
	})

	t.Run("returns error if any broker fails", func(t *testing.T) {
		broker1 := &mockBroker{
			publishFunc: func(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error {
				return errors.New("fail")
			},
		}

		multi := messaging.NewMultiBroker(broker1)
		err := multi.Publish(context.Background(), "test", uuid.New(), nil, nil)

		assert.Error(t, err)
	})
}
