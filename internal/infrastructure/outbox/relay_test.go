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
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestOutboxRelay_RetryLogic(t *testing.T) {
	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	q := db.New()
	eventID := uuid.New()
	eventType := "test.poison.message." + eventID.String() // Unique per test
	payload := []byte(`{"data": "bad"}`)

	err := q.SaveOutboxEvent(ctx, pool, db.SaveOutboxEventParams{
		ID:        eventID,
		EventType: eventType,
		Payload:   payload,
	})
	require.NoError(t, err)

	relay := outbox.NewRelay(pool)

	mockErr := errors.New("simulated transient failure")

	relay.Register(eventType, func(ctx context.Context, p []byte) error {
		return mockErr
	})

	relayCtx, cancel := context.WithCancel(ctx)
	go relay.Start(relayCtx)

	defer cancel()

	const qry = "SELECT retries,last_error FROM outbox WHERE id = $1"

	require.Eventually(t, func() bool {
		var (
			retries   int
			lastError *string
		)

		if err := pool.QueryRow(ctx, qry, eventID).Scan(&retries, &lastError); err != nil {
			return false
		}

		return retries >= 1 && lastError != nil && *lastError == mockErr.Error()
	}, 10*time.Second, 100*time.Millisecond)
}

func TestOutboxRelay_GracefulShutdown(t *testing.T) {
	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	relay := outbox.NewRelay(pool)

	handlerStarted := make(chan struct{})
	blockHandler := make(chan struct{})

	eventID := uuid.New()
	eventType := "test.slow." + eventID.String()

	relay.Register(eventType, func(ctx context.Context, p []byte) error {
		close(handlerStarted)
		<-blockHandler // block until test allows proceeding

		return nil
	})

	_ = db.New().SaveOutboxEvent(ctx, pool, db.SaveOutboxEventParams{
		ID:        eventID,
		EventType: eventType,
		Payload:   []byte("{}"),
	})

	relayCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		relay.Start(relayCtx)
		close(done)
	}()

	select {
	case <-handlerStarted:
		// event is being processed
	case <-time.After(5 * time.Second):
		t.Fatal("Relay did not pick up event")
	}

	cancel()

	select {
	case <-done:
		t.Fatal("Relay stopped before handler finished")
	default:
		// still running
	}

	close(blockHandler) // unblock

	select {
	case <-done:
		// exits
	case <-time.After(5 * time.Second):
		t.Fatal("Relay did not stop after active tasks finished")
	}

	count := 0
	_ = pool.QueryRow(ctx, "SELECT count(*) FROM outbox WHERE id = $1 AND processed_at IS NOT NULL", eventID).Scan(&count)
	assert.Equal(t, 1, count)
}
