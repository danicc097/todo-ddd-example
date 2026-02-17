package postgres_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

func TestTodoMapper_MapEvent(t *testing.T) {
	t.Parallel()

	mapper := &postgres.TodoMapper{}

	t.Run("TodoCreatedEvent", func(t *testing.T) {
		id := uuid.New()
		wsID := uuid.New()
		now := time.Now().Truncate(time.Second)

		evt := domain.TodoCreatedEvent{
			ID:          domain.TodoID{UUID: id},
			WorkspaceID: wsDomain.WorkspaceID{UUID: wsID},
			Title:       "Test Todo",
			Status:      "PENDING",
			CreatedAt:   now,
			Occurred:    now,
		}

		name, data, err := mapper.MapEvent(evt)
		require.NoError(t, err)
		assert.Equal(t, "todo.created", name)

		var payload map[string]any

		err = json.Unmarshal(data, &payload)
		require.NoError(t, err)

		assert.Equal(t, id.String(), payload["id"])
		assert.Equal(t, wsID.String(), payload["workspace_id"])
		assert.Equal(t, "Test Todo", payload["title"])
		assert.Equal(t, "PENDING", payload["status"])
	})

	t.Run("TagAddedEvent", func(t *testing.T) {
		todoID := uuid.New()
		tagID := uuid.New()
		now := time.Now()

		evt := domain.TagAddedEvent{
			TodoID:   domain.TodoID{UUID: todoID},
			TagID:    domain.TagID{UUID: tagID},
			Occurred: now,
		}

		name, data, err := mapper.MapEvent(evt)
		require.NoError(t, err)
		assert.Equal(t, "todo.tag_added", name)

		var payload map[string]any

		err = json.Unmarshal(data, &payload)
		require.NoError(t, err)

		assert.Equal(t, todoID.String(), payload["todo_id"])
		assert.Equal(t, tagID.String(), payload["tag_id"])
	})
}
