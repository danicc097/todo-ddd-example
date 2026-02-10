package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestRegisterUserUseCase_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	pgContainer := testutils.NewPostgreSQLContainer(ctx, t)
	defer pgContainer.Close(ctx, t)

	pool := pgContainer.Connect(ctx, t)
	repo := userPg.NewUserRepo(pool)
	uc := application.NewRegisterUserUseCase(repo)

	t.Run("successfully registers", func(t *testing.T) {
		cmd := application.RegisterUserCommand{
			Email: "a@b.com",
			Name:  "user",
		}

		id, err := uc.Execute(ctx, cmd)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, cmd.Email, found.Email().String())
	})
}
