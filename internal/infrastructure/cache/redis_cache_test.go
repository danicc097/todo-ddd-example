package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

type TestStub struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestGenericCache_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rc := testutils.NewRedisContainer(ctx, t)
	defer rc.Close(ctx, t)

	client := rc.Connect(ctx, t)

	ttl := 1 * time.Hour
	cache := cache.NewGenericCache[TestStub](client, ttl)

	t.Run("miss", func(t *testing.T) {
		val, err := cache.Get(ctx, "non-existent")
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("hit", func(t *testing.T) {
		key := "test:hit"
		want := TestStub{Name: "Alice", Age: 30}

		err := cache.Set(ctx, key, &want)
		require.NoError(t, err)

		got, err := cache.Get(ctx, key)

		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, want, *got)
	})

	t.Run("invalidate", func(t *testing.T) {
		key := "test:del"
		val := TestStub{Name: "Bob", Age: 40}

		require.NoError(t, cache.Set(ctx, key, &val))

		err := cache.Del(ctx, key)
		require.NoError(t, err)

		got, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("corrupted", func(t *testing.T) {
		key := "test:corrupt"

		err := client.Set(ctx, key, "{not-valid-json", ttl).Err()
		require.NoError(t, err)

		got, err := cache.Get(ctx, key)

		require.Error(t, err)
		assert.Nil(t, got)
	})
}
