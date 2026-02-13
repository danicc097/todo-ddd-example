package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/middleware"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCreateTodoUseCase_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	pgContainer := testutils.NewPostgreSQLContainer(ctx, t)
	defer pgContainer.Close(ctx, t)

	pool := pgContainer.Connect(ctx, t)
	repo := todoPg.NewTodoRepo(pool)
	tagRepo := todoPg.NewTagRepo(pool)

	baseHandler := application.NewCreateTodoHandler(repo)
	handler := middleware.Transactional(pool, baseHandler)

	t.Run("creates", func(t *testing.T) {
		tn, _ := domain.NewTagName("urgent")
		tag := domain.NewTag(tn)
		require.NoError(t, tagRepo.Save(ctx, tag))

		const title = "Integration"

		cmd := application.CreateTodoCommand{
			Title:  title,
			TagIDs: []domain.TagID{tag.ID()},
		}

		id, err := handler.Handle(ctx, cmd)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, title, found.Title().String())
		assert.Contains(t, found.Tags(), tag.ID())

		var count int

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE event_type = 'todo.created'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}
