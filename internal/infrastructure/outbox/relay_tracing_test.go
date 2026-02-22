package outbox_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

// nolint: paralleltest
func TestOutboxRelay_Tracing(t *testing.T) {
	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	exp := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exp))

	oldTP := otel.GetTracerProvider()

	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(oldTP) })

	oldPropagator := otel.GetTextMapPropagator()

	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(oldPropagator) })

	eventID := uuid.New()
	aggregateID := uuid.New()
	eventType := testutils.RandomEventType()

	processedCh := make(chan struct{})
	broker := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
		if args.AggID == aggregateID {
			close(processedCh)
		}

		return nil
	})

	relay := outbox.NewRelay(pool, broker)

	err := db.New().SaveOutboxEvent(ctx, pool, db.SaveOutboxEventParams{
		ID:            eventID,
		EventType:     eventType,
		AggregateType: "MOCK",
		AggregateID:   aggregateID,
		Payload:       []byte("{}"),
		Headers:       []byte("{}"),
	})
	require.NoError(t, err)

	relayCtx, cancel := context.WithCancel(ctx)
	go relay.Start(relayCtx)

	t.Cleanup(cancel)

	select {
	case <-processedCh:

	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for relay to process test event")
	}

	require.Eventually(t, func() bool {
		spans := exp.GetSpans()
		for _, s := range spans {
			if s.Name == "outbox.publish_event" {
				for _, attr := range s.Attributes {
					if attr.Key == "aggregate.id" && attr.Value.AsString() == aggregateID.String() {
						return true
					}
				}
			}
		}

		return false
	}, 1*time.Second, 10*time.Millisecond, "Expected span with matching aggregate ID")
}

// nolint: paralleltest
func TestOutboxRelay_TracingPropagation(t *testing.T) {
	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	exp := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exp))

	oldTP := otel.GetTracerProvider()

	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(oldTP) })

	oldPropagator := otel.GetTextMapPropagator()

	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(oldPropagator) })

	ctx, parentSpan := otel.Tracer("test").Start(ctx, "parent-operation")
	headers := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))
	parentSpan.End()

	headersJSON, _ := json.Marshal(headers)

	eventID := uuid.New()
	aggregateID := uuid.New()

	processedCh := make(chan struct{})
	broker := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
		if args.AggID == aggregateID {
			close(processedCh)
		}

		return nil
	})

	relay := outbox.NewRelay(pool, broker)

	err := db.New().SaveOutboxEvent(ctx, pool, db.SaveOutboxEventParams{
		ID:            eventID,
		EventType:     "test.propagated",
		AggregateType: "MOCK",
		AggregateID:   aggregateID,
		Payload:       []byte("{}"),
		Headers:       headersJSON,
	})
	require.NoError(t, err)

	relayCtx, cancel := context.WithCancel(context.Background())
	go relay.Start(relayCtx)

	t.Cleanup(cancel)

	select {
	case <-processedCh:
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for relay to process test event")
	}

	require.Eventually(t, func() bool {
		spans := exp.GetSpans()
		for _, s := range spans {
			if s.Name == "outbox.publish_event" {
				for _, attr := range s.Attributes {
					if attr.Key == "aggregate.id" && attr.Value.AsString() == aggregateID.String() {
						return s.SpanContext.TraceID() == parentSpan.SpanContext().TraceID()
					}
				}
			}
		}

		return false
	}, 1*time.Second, 10*time.Millisecond, "Expected to inherit parent TraceID")
}
