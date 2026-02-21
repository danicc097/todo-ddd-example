package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func mustCreateTodo(t *testing.T, title string, wsID wsDomain.WorkspaceID) *domain.Todo {
	t.Helper()

	tt, err := domain.NewTodoTitle(title)
	require.NoError(t, err)

	return domain.NewTodo(tt, wsID)
}

func TestTodoRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)
	repo := todoPg.NewTodoRepo(pool)

	user := fixtures.RandomUser(ctx, t)
	ws := fixtures.RandomWorkspace(ctx, t, user.ID())
	todo := mustCreateTodo(t, "Test Todo", ws.ID())

	t.Run("save and find", func(t *testing.T) {
		err := repo.Save(ctx, todo)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, todo.ID())
		require.NoError(t, err)
		assert.Equal(t, todo.ID(), found.ID())
	})

	t.Run("update", func(t *testing.T) {
		require.NoError(t, todo.Complete())
		err := repo.Save(ctx, todo)
		require.NoError(t, err)

		updated, err := repo.FindByID(ctx, todo.ID())
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, updated.Status())
	})

	t.Run("find all", func(t *testing.T) {
		todos, err := repo.FindAllByWorkspace(ctx, ws.ID())
		require.NoError(t, err)

		var found bool

		for _, td := range todos {
			if td.ID() == todo.ID() {
				found = true
				break
			}
		}

		assert.True(t, found)
	})

	t.Run("by id non-existent", func(t *testing.T) {
		_, err := repo.FindByID(ctx, domain.TodoID(uuid.New()))
		assert.ErrorIs(t, err, domain.ErrTodoNotFound)
	})

	t.Run("tags", func(t *testing.T) {
		taggedTodo := mustCreateTodo(t, "Todo with tags", ws.ID())
		tag1 := fixtures.RandomTag(ctx, t, ws.ID())
		tag2 := fixtures.RandomTag(ctx, t, ws.ID())

		taggedTodo.AddTag(tag1.ID())
		taggedTodo.AddTag(tag2.ID())

		err := repo.Save(ctx, taggedTodo)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, taggedTodo.ID())
		require.NoError(t, err)
		assert.ElementsMatch(t, []domain.TagID{tag1.ID(), tag2.ID()}, found.Tags())
	})
}
