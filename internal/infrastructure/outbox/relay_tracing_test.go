package outbox_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestOutboxRelay_Tracing(t *testing.T) {
	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	exp := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exp))

	oldTP := otel.GetTracerProvider()

	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(oldTP)

	broker := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
		return nil
	})

	relay := outbox.NewRelay(pool, broker)

	eventID := uuid.New()
	eventType := sharedDomain.EventType("test.tracing." + eventID.String())

	_ = db.New().SaveOutboxEvent(ctx, pool, db.SaveOutboxEventParams{
		ID:            eventID,
		EventType:     eventType,
		AggregateType: "MOCK",
		AggregateID:   uuid.New(),
		Payload:       []byte("{}"),
		Headers:       []byte("{}"),
	})

	relayCtx, cancel := context.WithCancel(ctx)
	go relay.Start(relayCtx)

	defer cancel()

	require.Eventually(t, func() bool {
		spans := exp.GetSpans()
		foundBatch := false
		foundPublish := false

		for _, s := range spans {
			if s.Name == "outbox.relay_batch" {
				foundBatch = true
			}

			if s.Name == "outbox.publish_event" {
				foundPublish = true
			}
		}

		return foundBatch && foundPublish
	}, 5*time.Second, 100*time.Millisecond)
}
