package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCreateTodoUseCase_Integration(t *testing.T) {
	ctx := context.Background()

	pgContainer := testutils.NewPostgreSQLContainer(ctx, t)
	defer pgContainer.Close(ctx, t)

	pool := pgContainer.Connect(ctx, t)
	tm := db.NewTransactionManager(pool)
	repo := todoPg.NewTodoRepo(pool)
	tagRepo := todoPg.NewTagRepo(pool)
	uc := application.NewCreateTodoUseCase(tm)

	t.Run("creates", func(t *testing.T) {
		tn, _ := domain.NewTagName("urgent")
		tag := domain.NewTag(tn)
		require.NoError(t, tagRepo.Save(ctx, tag))

		const title = "Integration"

		cmd := application.CreateTodoCommand{
			Title:  title,
			TagIDs: []uuid.UUID{tag.ID()},
		}

		id, err := uc.Execute(ctx, cmd)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, title, found.Title().String())
		assert.Contains(t, found.Tags(), tag.ID())
	})
}
