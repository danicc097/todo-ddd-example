package application_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestRegisterUserUseCase_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	repo := userPg.NewUserRepo(pool)
	uc := application.NewRegisterUserUseCase(repo)

	t.Run("successfully registers", func(t *testing.T) {
		uid := uuid.New().String()[:8]
		email := fmt.Sprintf("new-%s@example.com", uid)
		cmd := application.RegisterUserCommand{
			Email: email,
			Name:  "user " + uid,
		}

		resp, err := uc.Execute(ctx, cmd)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, resp.ID)
		require.NoError(t, err)
		assert.Equal(t, cmd.Email, found.Email().String())
	})
}
