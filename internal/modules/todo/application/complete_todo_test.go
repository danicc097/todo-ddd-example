package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCompleteTodoUseCase_Integration(t *testing.T) {
	ctx := context.Background()

	pgContainer := testutils.NewPostgreSQLContainer(ctx, t)
	defer pgContainer.Close(ctx, t)

	pool := pgContainer.Connect(ctx, t)
	tm := db.NewTransactionManager(pool)
	repo := todoPg.NewTodoRepo(pool)

	t.Run("completes", func(t *testing.T) {
		title, _ := domain.NewTodoTitle("Complete")
		todo := domain.NewTodo(title)
		require.NoError(t, repo.Save(ctx, todo))

		err := application.NewCompleteTodoUseCase(tm).Execute(ctx, todo.ID())
		require.NoError(t, err)

		found, _ := repo.FindByID(ctx, todo.ID())
		assert.Equal(t, domain.StatusCompleted, found.Status())
	})
}
