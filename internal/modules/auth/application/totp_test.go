package application_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/negrel/secrecy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto"
	infraCrypto "github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/application"
	authPg "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/postgres"
	authRedis "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/redis"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestTOTPFlow_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)

	redisClient := testutils.GetGlobalRedis(t).Connect(ctx, t)

	uow := sharedPg.NewUnitOfWork(pool)
	userRepo := userPg.NewUserRepo(pool, uow)
	authRepo := authPg.NewAuthRepo(pool, uow)
	totpGuard := authRedis.NewTOTPGuard(redisClient)
	masterKey := []byte("0123456789abcdef0123456789abcdef")
	hasher := infraCrypto.NewArgon2PasswordHasher()

	privKeyBytes, _ := os.ReadFile("../../../../private.pem")
	privKey, _ := jwt.ParseRSAPrivateKeyFromPEM(privKeyBytes)
	tokenIssuer := crypto.NewTokenIssuer(privKey, "test")

	uniqueEmail := fmt.Sprintf("auth-%s@example.com", uuid.New().String()[:8])

	registerHandler := application.NewRegisterHandler(userRepo, authRepo, hasher, sharedPg.NewUnitOfWork(pool))

	registerResp, err := registerHandler.Handle(ctx, application.RegisterCommand{
		Email:    uniqueEmail,
		Name:     "Auth User",
		Password: *secrecy.NewSecret("password123!"),
	})
	require.NoError(t, err)

	ctx = causation.WithMetadata(ctx, causation.Metadata{UserID: registerResp.ID.UUID()})

	initiateHandler := application.NewInitiateTOTPHandler(authRepo, masterKey)
	_ = application.NewVerifyTOTPHandler(authRepo, totpGuard, tokenIssuer, masterKey)

	uri, err := initiateHandler.Handle(ctx, struct{}{})
	require.NoError(t, err)
	assert.Contains(t, uri, "otpauth://totp/"+messaging.Keys.AppDisplayName())

	auth, err := authRepo.FindByUserID(ctx, registerResp.ID)
	require.NoError(t, err)
	assert.True(t, auth.IsTOTPPending())
}
