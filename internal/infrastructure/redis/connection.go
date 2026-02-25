package redis

import (
	"log/slog"
	"strings"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func NewClient(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: addr})

	if err := redisotel.InstrumentTracing(client,
		redisotel.WithAttributes(
			semconv.DBSystemNameRedis,
			semconv.PeerServiceKey.String("redis"),
		),
		redisotel.WithCommandFilter(func(cmd redis.Cmder) bool {
			name := strings.ToLower(cmd.Name())
			return name != "hello" && name != "client" && name != "ping" && name != "dial"
		}),
	); err != nil {
		slog.Error("failed to instrument redis", slog.String("error", err.Error()))
	}

	if err := redisotel.InstrumentMetrics(client); err != nil {
		slog.Error("failed to instrument redis metrics", slog.String("error", err.Error()))
	}

	return client
}
