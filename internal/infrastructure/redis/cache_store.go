package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
)

type CacheStore struct {
	client redis.UniversalClient
}

var _ cache.Store = (*CacheStore)(nil)

func NewCacheStore(client redis.UniversalClient) *CacheStore {
	return &CacheStore{client: client}
}

func (s *CacheStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrCacheMiss
		}

		return nil, err
	}

	return val, nil
}

func (s *CacheStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration, tags ...string) error {
	pipe := s.client.Pipeline()
	pipe.Set(ctx, key, value, ttl)

	for _, tag := range tags {
		tagKey := cache.Keys.TagSet(tag)
		pipe.SAdd(ctx, tagKey, key)
		pipe.Expire(ctx, tagKey, 24*time.Hour)
	}

	_, err := pipe.Exec(ctx)

	return err
}

func (s *CacheStore) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, k := range keys {
		pipe.Unlink(ctx, k)
	}

	_, err := pipe.Exec(ctx)

	return err
}

func (s *CacheStore) Invalidate(ctx context.Context, tags ...string) error {
	if len(tags) == 0 {
		return nil
	}

	var allKeys []string

	pipe := s.client.Pipeline()
	cmds := make([]*redis.StringSliceCmd, len(tags))

	for i, tag := range tags {
		tagKey := cache.Keys.TagSet(tag)
		allKeys = append(allKeys, tagKey)
		cmds[i] = pipe.SMembers(ctx, tagKey)
	}

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	for _, cmd := range cmds {
		if keys, err := cmd.Result(); err == nil {
			allKeys = append(allKeys, keys...)
		}
	}

	return s.Delete(ctx, allKeys...)
}

func (s *CacheStore) Incr(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}
