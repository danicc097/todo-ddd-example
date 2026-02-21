package outbox_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestOutboxRelay_RetryLogic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	q := db.New()
	eventID := uuid.New()
	eventType := sharedDomain.EventType("test.poison.message." + eventID.String())
	payload := []byte(`{"data": "bad"}`)

	err := q.SaveOutboxEvent(ctx, pool, db.SaveOutboxEventParams{
		ID:            eventID,
		EventType:     eventType,
		AggregateType: "MOCK",
		AggregateID:   uuid.New(),
		Payload:       payload,
		Headers:       []byte("{}"),
	})
	require.NoError(t, err)

	mockErr := errors.New("simulated transient failure")
	broker := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
		return mockErr
	})

	relay := outbox.NewRelay(pool, broker)

	relayCtx, cancel := context.WithCancel(ctx)
	go relay.Start(relayCtx)

	defer cancel()

	const qry = "SELECT retries, last_error, last_attempted_at FROM outbox WHERE id = $1"

	require.Eventually(t, func() bool {
		var (
			retries         int
			lastError       *string
			lastAttemptedAt *time.Time
		)

		if err := pool.QueryRow(ctx, qry, eventID).Scan(&retries, &lastError, &lastAttemptedAt); err != nil {
			return false
		}

		return retries >= 1 && lastError != nil && *lastError == mockErr.Error() && lastAttemptedAt != nil
	}, 10*time.Second, 100*time.Millisecond)
}

func TestOutboxRelay_FatalErrorBackoff(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	q := db.New()
	eventID := uuid.New()
	eventType := sharedDomain.EventType("test.fatal." + eventID.String())

	err := q.SaveOutboxEvent(ctx, pool, db.SaveOutboxEventParams{
		ID:            eventID,
		EventType:     eventType,
		AggregateType: "MOCK",
		AggregateID:   uuid.New(),
		Payload:       []byte("{}"),
		Headers:       []byte("[\"invalid\", \"type\"]"),
	})
	require.NoError(t, err)

	broker := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
		return nil
	})

	relay := outbox.NewRelay(pool, broker)

	relayCtx, cancel := context.WithCancel(ctx)
	go relay.Start(relayCtx)

	defer cancel()

	const qry = "SELECT retries, last_error FROM outbox WHERE id = $1"

	require.Eventually(t, func() bool {
		var (
			retries   int
			lastError *string
		)

		if err := pool.QueryRow(ctx, qry, eventID).Scan(&retries, &lastError); err != nil {
			return false
		}

		return retries >= 1 && lastError != nil && *lastError == "fatal: invalid header JSON"
	}, 10*time.Second, 100*time.Millisecond)
}

func TestOutboxRelay_GracefulShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	handlerStarted := make(chan struct{})
	blockHandler := make(chan struct{})

	broker := messaging.BrokerPublishFunc(func(ctx context.Context, args messaging.PublishArgs) error {
		close(handlerStarted)
		<-blockHandler

		return nil
	})

	relay := outbox.NewRelay(pool, broker)

	eventID := uuid.New()
	eventType := sharedDomain.EventType("test.slow." + eventID.String())

	_ = db.New().SaveOutboxEvent(ctx, pool, db.SaveOutboxEventParams{
		ID:            eventID,
		EventType:     eventType,
		AggregateType: "MOCK",
		AggregateID:   uuid.New(),
		Payload:       []byte("{}"),
		Headers:       []byte("{}"),
	})

	relayCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		relay.Start(relayCtx)
		close(done)
	}()

	select {
	case <-handlerStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("Relay did not pick up event")
	}

	cancel()

	select {
	case <-done:
		t.Fatal("Relay stopped before handler finished")
	default:
	}

	close(blockHandler)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Relay did not stop after active tasks finished")
	}

	count := 0
	_ = pool.QueryRow(ctx, "SELECT count(*) FROM outbox WHERE id = $1 AND processed_at IS NOT NULL", eventID).Scan(&count)
	assert.Equal(t, 1, count)
}
