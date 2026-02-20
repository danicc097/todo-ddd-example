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

	t.Run("symmetry", func(t *testing.T) {
		todoID := domain.TodoID(uuid.New())
		wsID := wsDomain.WorkspaceID(uuid.New())
		title, _ := domain.NewTodoTitle("Complex Task")
		tags := []domain.TagID{domain.TagID(uuid.New()), domain.TagID(uuid.New())}

		now := time.Now().Truncate(time.Second)

		original := domain.ReconstituteTodo(
			todoID,
			title,
			domain.StatusPending,
			now,
			tags,
			wsID,
		)

		data, err := codec.Marshal(original)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		reconstituted, err := codec.Unmarshal(data)
		require.NoError(t, err)

		assert.Equal(t, original.ID(), reconstituted.ID())
		assert.Equal(t, original.Title().String(), reconstituted.Title().String())
		assert.Equal(t, original.Status(), reconstituted.Status())
		assert.Equal(t, original.CreatedAt().Unix(), reconstituted.CreatedAt().Unix())
		assert.ElementsMatch(t, original.Tags(), reconstituted.Tags())
		assert.Equal(t, original.WorkspaceID(), reconstituted.WorkspaceID())
	})
}
