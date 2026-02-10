package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisContainer struct {
	container *tcRedis.RedisContainer
	client    *redis.Client
}

func NewRedisContainer(ctx context.Context, t *testing.T) *RedisContainer {
	t.Helper()

	container, err := tcRedis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.NewLogStrategy("Ready to accept connections"),
		),
	)
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	return &RedisContainer{container: container}
}

func (r *RedisContainer) Connect(ctx context.Context, t *testing.T) *redis.Client {
	t.Helper()

	var (
		client *redis.Client
		err    error
	)

	for range 10 {
		uri, err := r.container.ConnectionString(ctx)
		if err != nil {
			t.Fatalf("failed to get redis connection string: %v", err)
		}

		opt, err := redis.ParseURL(uri)
		if err != nil {
			t.Fatalf("failed to parse redis URL: %v", err)
		}

		client = redis.NewClient(opt)
		if err := client.Ping(ctx).Err(); err == nil {
			r.client = client
			return client
		}

		time.Sleep(500 * time.Millisecond)
	}

	t.Fatalf("failed to connect to redis after retries: %v", err)

	return nil
}

func (r *RedisContainer) Close(ctx context.Context, t *testing.T) {
	t.Helper()

	if r.client != nil {
		r.client.Close()
	}

	if err := r.container.Terminate(ctx); err != nil {
		t.Logf("failed to terminate redis container: %v", err)
	}
}

func (r *RedisContainer) Client() *redis.Client {
	return r.client
}
