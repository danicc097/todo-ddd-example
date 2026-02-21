package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

func TestTodoMapper_MapEvent(t *testing.T) {
	t.Parallel()

	mapper := &postgres.TodoMapper{}

	t.Run("TodoCreatedEvent", func(t *testing.T) {
		id := uuid.New()
		wsID := uuid.New()
		now := time.Now().Truncate(time.Second)

		title, _ := domain.NewTodoTitle("Test Todo")

		evt := domain.TodoCreatedEvent{
			ID:        domain.TodoID(id),
			WsID:      wsDomain.WorkspaceID(wsID),
			Title:     title,
			Status:    domain.StatusPending,
			CreatedAt: now,
			Occurred:  now,
		}

		name, data, err := mapper.MapEvent(evt)
		require.NoError(t, err)
		assert.Equal(t, sharedDomain.TodoCreated, name)

		payload := data.(postgres.TodoOutboxDTO)

		assert.Equal(t, domain.TodoID(id), payload.ID)
		assert.Equal(t, wsDomain.WorkspaceID(wsID), payload.WorkspaceID)
		assert.Equal(t, "Test Todo", payload.Title)
		assert.Equal(t, "PENDING", payload.Status)
	})

	t.Run("TagAddedEvent", func(t *testing.T) {
		todoID := uuid.New()
		tagID := uuid.New()
		now := time.Now()

		evt := domain.TagAddedEvent{
			TodoID:   domain.TodoID(todoID),
			TagID:    domain.TagID(tagID),
			WsID:     wsDomain.WorkspaceID(uuid.New()), // adding missing field or keeping it consistent
			Occurred: now,
		}

		name, data, err := mapper.MapEvent(evt)
		require.NoError(t, err)
		assert.Equal(t, sharedDomain.TodoTagAdded, name)

		payload := data.(postgres.TagAddedOutboxDTO)

		assert.Equal(t, domain.TodoID(todoID), payload.TodoID)
		assert.Equal(t, domain.TagID(tagID), payload.TagID)
	})
}
