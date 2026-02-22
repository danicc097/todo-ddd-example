package redis_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/redis"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestPublisher_Tracing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rdb := testutils.GetGlobalRedis(t).Connect(ctx, t)

	exp := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exp))

	oldTP := otel.GetTracerProvider()

	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(oldTP)

	pub := redis.NewPublisher(rdb)

	err := pub.Publish(ctx, messaging.PublishArgs{
		EventType: sharedDomain.EventType("test.event"),
		AggID:     uuid.New(),
		Payload:   []byte("{}"),
		Headers:   make(map[messaging.Header]string),
	})
	require.NoError(t, err)

	spans := exp.GetSpans()
	found := false

	for _, s := range spans {
		if s.Name == "redis.publish_event" {
			found = true
			break
		}
	}

	assert.True(t, found)
}
