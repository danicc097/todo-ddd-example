package redis_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

func TestTodoCacheCodec_Symmetry(t *testing.T) {
	t.Parallel()

	codec := redis.NewTodoCacheCodec()

	t.Run("basic mapping smoke test", func(t *testing.T) {
		todoID := domain.TodoID(uuid.New())
		wsID := wsDomain.WorkspaceID(uuid.New())
		title, _ := domain.NewTodoTitle("smoke Test")

		original := domain.ReconstituteTodo(domain.ReconstituteTodoArgs{
			ID:              todoID,
			WorkspaceID:     wsID,
			Title:           title,
			Status:          domain.StatusPending,
			CreatedAt:       time.Now(),
			Tags:            nil,
			DueDate:         nil,
			Recurrence:      nil,
			LastCompletedAt: nil,
			Sessions:        nil,
		})

		data, err := codec.Marshal(original)
		require.NoError(t, err)

		reconstituted, err := codec.Unmarshal(data)
		require.NoError(t, err)

		assert.Equal(t, original.ID(), reconstituted.ID())
		assert.Equal(t, original.Title().String(), reconstituted.Title().String())
	})
}
