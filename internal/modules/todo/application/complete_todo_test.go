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
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCompleteTodoUseCase_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	pgContainer := testutils.NewPostgreSQLContainer(ctx, t)
	defer pgContainer.Close(ctx, t)

	pool := pgContainer.Connect(ctx, t)
	repo := todoPg.NewTodoRepo(pool)

	baseHandler := application.NewCompleteTodoHandler(repo)
	handler := middleware.Transactional(pool, baseHandler)

	t.Run("completes", func(t *testing.T) {
		title, _ := domain.NewTodoTitle("Complete")
		todo := domain.NewTodo(title)
		require.NoError(t, repo.Save(ctx, todo))

		_, err := handler.Handle(ctx, application.CompleteTodoCommand{
			ID: todo.ID(),
		})
		require.NoError(t, err)

		found, _ := repo.FindByID(ctx, todo.ID())
		assert.Equal(t, domain.StatusCompleted, found.Status())

		var count int

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE event_type = 'todo.completed'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("fails if todo not found", func(t *testing.T) {
		_, err := handler.Handle(ctx, application.CompleteTodoCommand{
			ID: domain.TodoID{UUID: uuid.New()},
		})
		assert.ErrorIs(t, err, domain.ErrTodoNotFound)
	})
}
