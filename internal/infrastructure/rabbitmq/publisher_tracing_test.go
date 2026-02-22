package rabbitmq_test

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
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/rabbitmq"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestPublisher_Tracing(t *testing.T) {
	ctx := context.Background()
	rmq := testutils.GetGlobalRabbitMQ(t)

	conn := rmq.Connect(ctx, t)
	defer conn.Close()

	exp := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exp))

	oldTP := otel.GetTracerProvider()

	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(oldTP)

	pub, err := rabbitmq.NewPublisher(conn, "test-tracing")
	require.NoError(t, err)

	defer pub.Close()

	err = pub.Publish(ctx, messaging.PublishArgs{
		EventType: sharedDomain.EventType("test.event"),
		AggID:     uuid.New(),
		Payload:   []byte("{}"),
		Headers:   nil,
	})
	require.NoError(t, err)

	spans := exp.GetSpans()
	found := false

	for _, s := range spans {
		if s.Name == "rabbitmq.publish" {
			found = true
			break
		}
	}

	assert.True(t, found, "rabbitmq.publish span not found")
}
