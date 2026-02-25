package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
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
	uow := sharedPg.NewUnitOfWork(pool)
	repo := todoPg.NewTodoRepo(pool, uow)

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
		now := time.Now()
		require.NoError(t, todo.Complete(userDomain.UserID(user.ID()), now))
		err := repo.Save(ctx, todo)
		require.NoError(t, err)

		updated, err := repo.FindByID(ctx, todo.ID())
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, updated.Status())
	})

	t.Run("focus sessions", func(t *testing.T) {
		ftodo := mustCreateTodo(t, "Focus Todo", ws.ID())
		sessionID := domain.FocusSessionID(uuid.New())
		require.NoError(t, ftodo.StartFocus(user.ID(), sessionID))
		require.NoError(t, repo.Save(ctx, ftodo))

		found, err := repo.FindByID(ctx, ftodo.ID())
		require.NoError(t, err)
		assert.Len(t, found.Sessions(), 1)
		assert.True(t, found.Sessions()[0].IsActive())

		require.NoError(t, found.StopFocus(time.Now()))
		require.NoError(t, repo.Save(ctx, found))

		found2, _ := repo.FindByID(ctx, ftodo.ID())
		assert.False(t, found2.Sessions()[0].IsActive())
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
