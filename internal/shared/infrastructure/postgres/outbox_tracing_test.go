package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestSaveDomainEvents_Tracing(t *testing.T) {
	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	queries := db.New()

	exp := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exp))

	oldTP := otel.GetTracerProvider()

	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(oldTP)

	t.Run("records SaveDomainEvents span", func(t *testing.T) {
		agg := &mockAggregate{
			events: []domain.DomainEvent{
				mockEvent{id: uuid.New(), name: "test.event"},
			},
		}

		err := postgres.SaveDomainEvents(ctx, queries, pool, &mockMapper{}, agg)
		require.NoError(t, err)

		spans := exp.GetSpans()
		found := false

		for _, s := range spans {
			if s.Name == "SaveDomainEvents" {
				found = true
				break
			}
		}

		assert.True(t, found)
	})
}
