package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachableEntity interface {
	ID() string
}

// GenericCache wraps any simple Find/Save operation with Redis caching.
type GenericCache[T any] struct {
	client *redis.Client
	ttl    time.Duration
}

func NewGenericCache[T any](client *redis.Client, ttl time.Duration) *GenericCache[T] {
	return &GenericCache[T]{
		client: client,
		ttl:    ttl,
	}
}

func (c *GenericCache[T]) Get(ctx context.Context, key string) (*T, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil // miss
	}

	if err != nil {
		return nil, err // redis error
	}

	var entity T
	if err := json.Unmarshal(val, &entity); err != nil {
		return nil, err
	}

	return &entity, nil
}

func (c *GenericCache[T]) Set(ctx context.Context, key string, entity *T) error {
	data, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *GenericCache[T]) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
