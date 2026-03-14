package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestAfterCommit(t *testing.T) {
	t.Parallel()

	pool := testutils.GetGlobalPostgresPool(t)

	t.Run("Hook runs after successful commit", func(t *testing.T) {
		hookCalled := false
		_, err := db.RunInTx(context.Background(), pool, func(ctx context.Context) (any, error) {
			db.AfterCommit(ctx, func(ctx context.Context) {
				hookCalled = true
			})

			return nil, nil
		})

		require.NoError(t, err)
		assert.True(t, hookCalled)
	})

	t.Run("Hook doesnt run after rollback", func(t *testing.T) {
		hookCalled := false
		_, err := db.RunInTx(context.Background(), pool, func(ctx context.Context) (any, error) {
			db.AfterCommit(ctx, func(ctx context.Context) {
				hookCalled = true
			})

			return nil, errors.New("rollback")
		})

		assert.Error(t, err)
		assert.False(t, hookCalled)
	})

	t.Run("Hook runs immediately if not in tx", func(t *testing.T) {
		hookCalled := false

		db.AfterCommit(context.Background(), func(ctx context.Context) {
			hookCalled = true
		})
		assert.True(t, hookCalled)
	})
}
