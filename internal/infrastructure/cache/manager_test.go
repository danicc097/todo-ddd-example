package cache_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache/mocks"
)

func TestGetOrFetch(t *testing.T) {
	t.Parallel()

	codec := cache.NewMsgpackCodec[string]()
	key := "test-key"
	val := "test-value"
	encoded, _ := codec.Marshal(val)

	t.Run("Cache hit", func(t *testing.T) {
		store := new(mocks.FakeStore)
		store.GetReturns(encoded, nil)

		fetchCalled := false
		fetch := func(ctx context.Context) (string, error) {
			fetchCalled = true
			return "", nil
		}

		res, err := cache.GetOrFetch(context.Background(), store, key, time.Minute, codec, fetch)

		require.NoError(t, err)
		assert.Equal(t, val, res)
		assert.False(t, fetchCalled)
		assert.Equal(t, 1, store.GetCallCount())
	})

	t.Run("Cache miss and fetch success", func(t *testing.T) {
		store := new(mocks.FakeStore)
		store.GetReturns(nil, cache.ErrCacheMiss)

		setDone := make(chan struct{}) // Set is called in goroutine

		store.SetCalls(func(ctx context.Context, s string, b []byte, d time.Duration, s2 ...string) error {
			close(setDone)
			return nil
		})

		fetchCalled := false
		fetch := func(ctx context.Context) (string, error) {
			fetchCalled = true
			return val, nil
		}

		res, err := cache.GetOrFetch(context.Background(), store, key, time.Minute, codec, fetch)

		require.NoError(t, err)
		assert.Equal(t, val, res)
		assert.True(t, fetchCalled)

		<-setDone

		assert.Equal(t, 1, store.SetCallCount())
		_, k, v, ttl, _ := store.SetArgsForCall(0)
		assert.Equal(t, key, k)
		assert.Equal(t, encoded, v)
		assert.Equal(t, time.Minute, ttl)
	})

	t.Run("Singleflight coalesces concurrent requests", func(t *testing.T) {
		store := new(mocks.FakeStore)
		store.GetReturns(nil, cache.ErrCacheMiss)

		var fetchCount int32

		fetch := func(ctx context.Context) (string, error) {
			atomic.AddInt32(&fetchCount, 1)
			time.Sleep(100 * time.Millisecond)

			return val, nil
		}

		const concurrency = 5

		var wg sync.WaitGroup
		wg.Add(concurrency)

		for range concurrency {
			go func() {
				defer wg.Done()

				_, _ = cache.GetOrFetch(context.Background(), store, key, time.Minute, codec, fetch)
			}()
		}

		wg.Wait()

		assert.Equal(t, int32(1), atomic.LoadInt32(&fetchCount))
	})

	t.Run("fetch error", func(t *testing.T) {
		store := new(mocks.FakeStore)
		store.GetReturns(nil, cache.ErrCacheMiss)

		expectedErr := errors.New("fetch failed")
		fetch := func(ctx context.Context) (string, error) {
			return "", expectedErr
		}

		_, err := cache.GetOrFetch(context.Background(), store, key, time.Minute, codec, fetch)

		assert.ErrorIs(t, err, expectedErr)
	})
}
