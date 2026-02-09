package postgres

import (
	"context"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: parallel and dynamic db as in https://github.com/danicc097/openapi-go-gin-postgres-sqlc/blob/main/internal/testutil/postgresql.go
func setupRepo(t *testing.T) (*TodoRepo, *pgxpool.Pool) {
	t.Helper()
	dbURL := "postgresql://postgres:postgres@localhost:5656/postgres"
	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)
	return NewTodoRepo(pool), pool
}

func mustCreateTodo(t *testing.T, title string) *domain.Todo {
	t.Helper()
	tt, err := domain.NewTodoTitle(title)
	require.NoError(t, err)
	return domain.CreateTodo(tt)
}

func TestTodoRepo_Integration(t *testing.T) {
	t.Skip("pending db setup")
	ctx := context.Background()
	repo, pool := setupRepo(t)
	defer pool.Close()

	todo := mustCreateTodo(t, "Integration")

	t.Run("Persistence", func(t *testing.T) {
		_, err := repo.Save(ctx, todo)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, todo.ID())
		assert.NoError(t, err)
		assert.Equal(t, todo.ID(), found.ID())

		require.NoError(t, todo.Complete())
		assert.NoError(t, repo.Update(ctx, todo))

		updated, _ := repo.FindByID(ctx, todo.ID())
		assert.Equal(t, domain.StatusCompleted, updated.Status())
	})

	t.Run("FindAll", func(t *testing.T) {
		todos, err := repo.FindAll(ctx)
		assert.NoError(t, err)

		var found bool
		for _, t := range todos {
			if t.ID() == todo.ID() {
				found = true
				break
			}
		}
		assert.True(t, found, "saved todo should be in the list")
	})
}
