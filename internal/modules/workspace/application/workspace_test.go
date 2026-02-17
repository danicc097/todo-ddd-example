package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	userAdapters "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/adapters"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/middleware"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestWorkspaceUseCases_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)

	repo := wsPg.NewWorkspaceRepo(pool)
	ug := userAdapters.NewWorkspaceUserGateway(fixtures.UserRepo)

	t.Run("onboard workspace", func(t *testing.T) {
		owner := fixtures.RandomUser(ctx, t)
		member := fixtures.RandomUser(ctx, t)

		ctx := causation.WithMetadata(ctx, causation.Metadata{UserID: owner.ID().UUID})

		baseHandler := application.NewOnboardWorkspaceHandler(repo, ug)
		handler := middleware.Transactional(pool, baseHandler)

		cmd := application.OnboardWorkspaceCommand{
			Name: "Test Corp " + uuid.New().String(),
			Members: map[userDomain.UserID]application.MemberInitialState{
				member.ID(): {Role: wsDomain.RoleMember},
			},
			OwnerID: userDomain.UserID{UUID: uuid.Nil},
		}

		id, err := handler.Handle(ctx, cmd)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, cmd.Name, found.Name())
	})

	t.Run("remove member", func(t *testing.T) {
		owner := fixtures.RandomUser(ctx, t)
		member := fixtures.RandomUser(ctx, t)

		ws := fixtures.RandomWorkspace(ctx, t, owner.ID())
		require.NoError(t, ws.AddMember(member.ID(), wsDomain.RoleMember))
		require.NoError(t, repo.Save(ctx, ws))

		baseHandler := application.NewRemoveWorkspaceMemberHandler(repo)
		handler := middleware.Transactional(pool, baseHandler)

		cmd := application.RemoveWorkspaceMemberCommand{
			WorkspaceID: ws.ID(),
			MemberID:    member.ID(),
		}

		_, err := handler.Handle(ctx, cmd)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, ws.ID())
		require.NoError(t, err)
		assert.NotContains(t, found.Members(), member.ID())
	})

	t.Run("delete workspace requiring MFA", func(t *testing.T) {
		owner := fixtures.RandomUser(ctx, t)
		ws := fixtures.RandomWorkspace(ctx, t, owner.ID())

		baseHandler := application.NewDeleteWorkspaceHandler(repo)
		handler := middleware.Transactional(pool, baseHandler)

		cmd := application.DeleteWorkspaceCommand{ID: ws.ID()}

		// without MFA
		ctxNoMFA := causation.WithMetadata(ctx, causation.Metadata{UserID: owner.ID().UUID, MFAVerified: false})
		_, err := handler.Handle(ctxNoMFA, cmd)
		require.Error(t, err)

		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.MFARequired, appErr.Code)

		// with MFA
		ctxWithMFA := causation.WithMetadata(ctx, causation.Metadata{UserID: owner.ID().UUID, MFAVerified: true})
		_, err = handler.Handle(ctxWithMFA, cmd)
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, ws.ID())
		assert.ErrorIs(t, err, wsDomain.ErrWorkspaceNotFound)
	})
}
