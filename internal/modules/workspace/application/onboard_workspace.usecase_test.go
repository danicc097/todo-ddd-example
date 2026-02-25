package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	userAdapters "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/adapters"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestOnboardWorkspaceHandler_Handle_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)

	uow := sharedPg.NewUnitOfWork(pool)
	repo := wsPg.NewWorkspaceRepo(pool, uow)
	up := userAdapters.NewWorkspaceUserProvider(fixtures.UserRepo)

	handler := application.NewOnboardWorkspaceHandler(repo, up, uow)

	t.Run("system created", func(t *testing.T) {
		owner := fixtures.RandomUser(ctx, t)

		cmd := application.OnboardWorkspaceCommand{
			Name:    "System Workspace " + uuid.New().String(),
			OwnerID: owner.ID(),
		}

		resp, err := handler.Handle(ctx, cmd)
		require.NoError(t, err)

		savedWs, err := repo.FindByID(ctx, resp.ID)
		require.NoError(t, err)
		assert.Equal(t, cmd.Name, savedWs.Name().String())
		assert.True(t, savedWs.IsOwner(owner.ID()))
	})

	t.Run("user created", func(t *testing.T) {
		user := fixtures.RandomUser(ctx, t)
		userCtx := causation.WithMetadata(ctx, causation.Metadata{UserID: user.ID().UUID()})

		cmd := application.OnboardWorkspaceCommand{
			Name: "User Workspace " + uuid.New().String(),
			// OwnerID is empty, should fallback to context UserID
		}

		resp, err := handler.Handle(userCtx, cmd)
		require.NoError(t, err)

		savedWs, err := repo.FindByID(ctx, resp.ID)
		require.NoError(t, err)
		assert.Equal(t, cmd.Name, savedWs.Name().String())
		assert.True(t, savedWs.IsOwner(user.ID()))
	})

	t.Run("member initialization", func(t *testing.T) {
		owner := fixtures.RandomUser(ctx, t)
		member := fixtures.RandomUser(ctx, t)

		cmd := application.OnboardWorkspaceCommand{
			Name:    "Team Workspace " + uuid.New().String(),
			OwnerID: owner.ID(),
			Members: map[userDomain.UserID]application.MemberInitialState{
				member.ID(): {Role: domain.RoleMember},
			},
		}

		resp, err := handler.Handle(ctx, cmd)
		require.NoError(t, err)

		savedWs, err := repo.FindByID(ctx, resp.ID)
		require.NoError(t, err)
		assert.Contains(t, savedWs.Members(), owner.ID())
		assert.Contains(t, savedWs.Members(), member.ID())
		assert.Equal(t, domain.RoleMember, savedWs.Members()[member.ID()])
	})
}
