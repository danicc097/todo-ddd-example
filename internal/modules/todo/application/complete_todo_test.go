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
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCompleteTodoUseCase_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)
	uow := sharedPg.NewUnitOfWork(pool)
	repo := todoPg.NewTodoRepo(pool, uow)
	wsRepo := wsPg.NewWorkspaceRepo(pool, uow)
	wsProv := wsAdapters.NewTodoWorkspaceProvider(wsRepo)
	uow = sharedPg.NewUnitOfWork(pool)

	handler := application.NewCompleteTodoHandler(repo, wsProv, uow)

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

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE event_type = $1 AND aggregate_id = $2", sharedDomain.TodoCompleted, todo.ID().UUID()).Scan(&count)
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
