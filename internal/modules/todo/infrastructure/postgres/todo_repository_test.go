package postgres

import (
	"context"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCreateTodo(t *testing.T, title string) *domain.Todo {
	t.Helper()
	tt, err := domain.NewTodoTitle(title)
	require.NoError(t, err)
	return domain.NewTodo(tt)
}

func TestTodoRepo_Integration(t *testing.T) {
	ctx := context.Background()

	postgres := testutils.NewPostgreSQLContainer(ctx, t)
	defer postgres.Close(ctx, t)

	pool := postgres.Connect(ctx, t)
	defer pool.Close()

	repo := NewTodoRepo(pool)
	tagRepo := NewTagRepo(pool)

	todo := mustCreateTodo(t, "Test Todo")

	t.Run("save and find", func(t *testing.T) {
		err := repo.Save(ctx, todo)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, todo.ID())
		require.NoError(t, err)
		assert.Equal(t, todo.ID(), found.ID())
	})

	t.Run("update", func(t *testing.T) {
		require.NoError(t, todo.Complete())
		err := repo.Update(ctx, todo)
		require.NoError(t, err)

		updated, err := repo.FindByID(ctx, todo.ID())
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, updated.Status())
	})

	t.Run("find all", func(t *testing.T) {
		todos, err := repo.FindAll(ctx)
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
		_, err := repo.FindByID(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrTodoNotFound)
	})

	t.Run("tags", func(t *testing.T) {
		taggedTodo := mustCreateTodo(t, "Todo with tags")

		tn1, _ := domain.NewTagName("tag-1")
		tag1 := domain.NewTag(tn1)
		require.NoError(t, tagRepo.Save(ctx, tag1))

		tn2, _ := domain.NewTagName("tag-2")
		tag2 := domain.NewTag(tn2)
		require.NoError(t, tagRepo.Save(ctx, tag2))

		taggedTodo.AddTag(tag1.ID())
		taggedTodo.AddTag(tag2.ID())

		err := repo.Save(ctx, taggedTodo)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, taggedTodo.ID())
		require.NoError(t, err)
		assert.ElementsMatch(t, []uuid.UUID{tag1.ID(), tag2.ID()}, found.Tags())
	})
}
