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
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/middleware"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestAssignTagToTodoHandler_Handle_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)
	repo := todoPg.NewTodoRepo(pool)

	baseHandler := application.NewAssignTagToTodoHandler(repo)
	handler := middleware.Transactional(pool, baseHandler)

	t.Run("success", func(t *testing.T) {
		user := fixtures.RandomUser(ctx, t)
		ws := fixtures.RandomWorkspace(ctx, t, user.ID())
		tag := fixtures.RandomTag(ctx, t, ws.ID())
		todo := fixtures.RandomTodo(ctx, t, ws.ID())

		_, err := handler.Handle(ctx, application.AssignTagToTodoCommand{
			TodoID: todo.ID(),
			TagID:  tag.ID(),
		})

		require.NoError(t, err)

		updated, err := repo.FindByID(ctx, todo.ID())
		require.NoError(t, err)
		assert.Contains(t, updated.Tags(), tag.ID())
	})

	t.Run("failure - todo not found", func(t *testing.T) {
		_, err := handler.Handle(ctx, application.AssignTagToTodoCommand{
			TodoID: domain.TodoID(uuid.New()),
			TagID:  domain.TagID(uuid.New()),
		})

		assert.ErrorIs(t, err, domain.ErrTodoNotFound)
	})
}
