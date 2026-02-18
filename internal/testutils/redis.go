package testutils

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	globalRedisOnce      sync.Once
	globalRedisContainer *RedisContainer
	globalRedisErr       error
)

type RedisContainer struct {
	container testcontainers.Container
	redisURI  string
}

func GetGlobalRedis(t *testing.T) *RedisContainer {
	ctx := context.Background()

	globalRedisOnce.Do(func() {
		globalRedisContainer, globalRedisErr = newRedisContainer(ctx)
	})

	if globalRedisErr != nil {
		t.Fatalf("Failed to initialize global redis container: %v", globalRedisErr)
	}

	return globalRedisContainer
}

func newRedisContainer(ctx context.Context) (*RedisContainer, error) {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	_ = os.Setenv("TESTCONTAINERS_REUSE_ENABLE", "true")

	req := testcontainers.ContainerRequest{
		Image: "redis:7-alpine",
		Name:  "todo-ddd-test-redis",
		Labels: map[string]string{
			"todo-ddd-test": "true", // cleanup watchdog
		},
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithStartupTimeout(15 * time.Second),
		SkipReaper: true,
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "6379")
	redisURI := fmt.Sprintf("redis://%s:%s/0", host, port.Port())

	return &RedisContainer{
		container: container,
		redisURI:  redisURI,
	}, nil
}

func (r *RedisContainer) Connect(ctx context.Context, t *testing.T) *redis.Client {
	t.Helper()

	var (
		client *redis.Client
		err    error
	)

	for range 50 {
		opt, parseErr := redis.ParseURL(r.redisURI)
		if parseErr != nil {
			t.Fatalf("failed to parse redis URL: %v", parseErr)
		}

		client = redis.NewClient(opt)
		if err = client.Ping(ctx).Err(); err == nil {
			t.Cleanup(func() {
				client.Close()
			})

			return client
		}

		client.Close()

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("failed to connect to redis after retries: %v", err)

	return nil
}
