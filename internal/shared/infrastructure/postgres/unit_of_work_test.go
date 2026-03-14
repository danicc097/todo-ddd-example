package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestUnitOfWork_Execute(t *testing.T) {
	t.Parallel()

	pool := testutils.GetGlobalPostgresPool(t)
	uow := postgres.NewUnitOfWork(pool)

	t.Run("Commit persists events", func(t *testing.T) {
		ctx := context.Background()
		eventType := testutils.RandomEventType()
		aggID := uuid.New()

		agg := &mockAggregate{
			events: []domain.DomainEvent{
				mockEvent{id: aggID, name: eventType},
			},
		}

		err := uow.Execute(ctx, func(ctx context.Context) error {
			uow.Collect(ctx, &mockMapper{}, agg)
			return nil
		})

		require.NoError(t, err)

		var count int

		err = pool.QueryRow(ctx, "SELECT count(*) FROM outbox WHERE event_type = $1 AND aggregate_id = $2", eventType, aggID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Empty(t, agg.Events(), "Events should be cleared after save")
	})

	t.Run("Rollback does not persist events", func(t *testing.T) {
		ctx := context.Background()
		eventType := testutils.RandomEventType()
		aggID := uuid.New()

		agg := &mockAggregate{
			events: []domain.DomainEvent{
				mockEvent{id: aggID, name: eventType},
			},
		}

		err := uow.Execute(ctx, func(ctx context.Context) error {
			uow.Collect(ctx, &mockMapper{}, agg)
			return errors.New("simulated error")
		})

		assert.Error(t, err)

		var count int

		err = pool.QueryRow(ctx, "SELECT count(*) FROM outbox WHERE event_type = $1 AND aggregate_id = $2", eventType, aggID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, agg.Events(), "Events should NOT be cleared if transaction rolled back")
	})
}
