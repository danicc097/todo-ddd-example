package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	infraRedis "github.com/danicc097/todo-ddd-example/internal/infrastructure/redis"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCacheStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rdb := testutils.GetGlobalRedis(t).Connect(ctx, t)
	store := infraRedis.NewCacheStore(rdb)

	t.Run("Get and Set", func(t *testing.T) {
		key := "test-key"
		val := []byte("test-value")
		ttl := 1 * time.Minute

		err := store.Set(ctx, key, val, ttl)
		require.NoError(t, err)

		got, err := store.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, val, got)
	})

	t.Run("Get miss", func(t *testing.T) {
		_, err := store.Get(ctx, "non-existent")
		assert.ErrorIs(t, err, cache.ErrCacheMiss)
	})

	t.Run("Delete", func(t *testing.T) {
		key := "del-key"
		_ = store.Set(ctx, key, []byte("val"), time.Minute)

		err := store.Delete(ctx, key)
		require.NoError(t, err)

		_, err = store.Get(ctx, key)
		assert.ErrorIs(t, err, cache.ErrCacheMiss)
	})

	t.Run("Invalidate with tags", func(t *testing.T) {
		key1 := "key1"
		key2 := "key2"
		tag := "tag1"

		require.NoError(t, store.Set(ctx, key1, []byte("v1"), time.Minute, tag))
		require.NoError(t, store.Set(ctx, key2, []byte("v2"), time.Minute, tag))

		err := store.Invalidate(ctx, tag)
		require.NoError(t, err)

		_, err = store.Get(ctx, key1)
		assert.ErrorIs(t, err, cache.ErrCacheMiss)

		_, err = store.Get(ctx, key2)
		assert.ErrorIs(t, err, cache.ErrCacheMiss)
	})

	t.Run("Incr", func(t *testing.T) {
		key := "incr-key"
		_ = store.Delete(ctx, key)

		val, err := store.Incr(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, int64(1), val)

		val, err = store.Incr(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, int64(2), val)
	})
}
