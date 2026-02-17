package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

type mockEvent struct {
	id   uuid.UUID
	name string
}

func (e mockEvent) EventName() string      { return e.name }
func (e mockEvent) OccurredAt() time.Time  { return time.Now() }
func (e mockEvent) AggregateID() uuid.UUID { return e.id }

type mockAggregate struct {
	events []domain.DomainEvent
}

func (a *mockAggregate) Events() []domain.DomainEvent { return a.events }
func (a *mockAggregate) ClearEvents()                 { a.events = nil }

type mockMapper struct{}

func (m *mockMapper) MapEvent(e domain.DomainEvent) (string, []byte, error) {
	return e.EventName(), []byte(`{"foo":"bar"}`), nil
}

func TestSaveDomainEvents_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	queries := db.New()

	t.Run("saves events to outbox", func(t *testing.T) {
		uniqueEventType := "test.event." + uuid.New().String()

		agg := &mockAggregate{
			events: []domain.DomainEvent{
				mockEvent{id: uuid.New(), name: uniqueEventType},
				mockEvent{id: uuid.New(), name: uniqueEventType},
			},
		}

		err := postgres.SaveDomainEvents(ctx, queries, pool, &mockMapper{}, agg)
		require.NoError(t, err)

		assert.Empty(t, agg.Events())

		var count int

		err = pool.QueryRow(ctx, "SELECT count(*) FROM outbox WHERE event_type = $1", uniqueEventType).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}
