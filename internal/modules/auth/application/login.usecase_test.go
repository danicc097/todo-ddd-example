package application_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/negrel/secrecy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/auth/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	authPg "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/postgres"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

func TestLoginHandler_Handle_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)

	userRepo := userPg.NewUserRepo(pool)
	authRepo := authPg.NewAuthRepo(pool)

	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := crypto.NewTokenIssuer(privKey, "test")

	handler := application.NewLoginHandler(userRepo, authRepo, issuer)

	t.Run("success", func(t *testing.T) {
		user := fixtures.RandomUser(ctx, t)
		password := "password123"
		hash, _ := crypto.HashPassword(password, crypto.DefaultArgon2Params)
		auth := domain.NewUserAuth(user.ID(), hash)
		require.NoError(t, authRepo.Save(ctx, auth))

		resp, err := handler.Handle(ctx, application.LoginCommand{
			Email:    user.Email().String(),
			Password: *secrecy.NewSecret(password),
		})

		require.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("failure - user not found", func(t *testing.T) {
		_, err := handler.Handle(ctx, application.LoginCommand{
			Email:    "nonexistent@example.com",
			Password: *secrecy.NewSecret("any"),
		})

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("failure - wrong password", func(t *testing.T) {
		user := fixtures.RandomUser(ctx, t)
		hash, _ := crypto.HashPassword("correct-password", crypto.DefaultArgon2Params)
		auth := domain.NewUserAuth(user.ID(), hash)
		require.NoError(t, authRepo.Save(ctx, auth))

		_, err := handler.Handle(ctx, application.LoginCommand{
			Email:    user.Email().String(),
			Password: *secrecy.NewSecret("wrong-password"),
		})

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})
}
