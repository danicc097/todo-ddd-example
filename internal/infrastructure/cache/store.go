package cache

import (
	"context"
	"errors"
	"time"
)

var ErrCacheMiss = errors.New("cache: key not found")

//counterfeiter:generate -o mocks/store.gen.go . Store

// Store defines the interface for a cache backend.
type Store interface {
	// Get retrieves a value from the cache.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache, optionally associating it with tags.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration, tags ...string) error

	// Delete removes one or more keys from the cache.
	Delete(ctx context.Context, keys ...string) error

	// Invalidate removes all keys associated with the given tags.
	Invalidate(ctx context.Context, tags ...string) error

	// Incr increments the integer value of a key by one.
	// It should initialize the key to 0 if it does not exist.
	Incr(ctx context.Context, key string) (int64, error)
}
