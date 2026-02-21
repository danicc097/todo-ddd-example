package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	wsAdapters "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/adapters"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/middleware"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCompleteTodoUseCase_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)
	repo := todoPg.NewTodoRepo(pool)
	wsRepo := wsPg.NewWorkspaceRepo(pool)
	wsProv := wsAdapters.NewTodoWorkspaceGateway(wsRepo)

	baseHandler := application.NewCompleteTodoHandler(repo, wsProv)
	handler := middleware.Transactional(pool, baseHandler)

	t.Run("completes", func(t *testing.T) {
		user := fixtures.RandomUser(ctx, t)
		ws := fixtures.RandomWorkspace(ctx, t, user.ID())
		todo := fixtures.RandomTodo(ctx, t, ws.ID())

		userCtx := causation.WithMetadata(ctx, causation.Metadata{UserID: user.ID().UUID()})

		_, err := handler.Handle(userCtx, application.CompleteTodoCommand{
			ID: todo.ID(),
		})
		require.NoError(t, err)

		found, _ := repo.FindByID(ctx, todo.ID())
		assert.Equal(t, domain.StatusCompleted, found.Status())

		var count int

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE event_type = 'todo.completed' AND aggregate_id = $1", todo.ID().UUID()).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("fails if todo not found", func(t *testing.T) {
		_, err := handler.Handle(ctx, application.CompleteTodoCommand{
			ID: domain.TodoID(uuid.New()),
		})
		assert.ErrorIs(t, err, domain.ErrTodoNotFound)
	})
}
