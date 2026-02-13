package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/middleware"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func createTestUser(ctx context.Context, t *testing.T, pool *testutils.PostgreSQLContainer, email string) *userDomain.User {
	repo := userPg.NewUserRepo(pool.Pool())
	e, _ := userDomain.NewUserEmail(email)
	u := userDomain.CreateUser(e, "test")
	require.NoError(t, repo.Save(ctx, u))

	return u
}

type mockUserGateway struct{}

func (m *mockUserGateway) Exists(ctx context.Context, userID userDomain.UserID) (bool, error) {
	return true, nil
}

func TestWorkspaceUseCases_Integration(t *testing.T) {
	ctx := context.Background()

	pg := testutils.NewPostgreSQLContainer(ctx, t)
	defer pg.Close(ctx, t)

	pool := pg.Connect(ctx, t)
	repo := wsPg.NewWorkspaceRepo(pool)
	ug := &mockUserGateway{}

	owner := createTestUser(ctx, t, pg, "owner@app.com")
	member := createTestUser(ctx, t, pg, "member@app.com")

	ctx = causation.WithMetadata(ctx, causation.Metadata{UserID: owner.ID().UUID})

	t.Run("onboard workspace", func(t *testing.T) {
		baseHandler := application.NewOnboardWorkspaceHandler(repo, ug)
		handler := middleware.Transactional(pool, baseHandler)

		cmd := application.OnboardWorkspaceCommand{
			Name: "Test Corp",
			Members: map[userDomain.UserID]application.MemberInitialState{
				member.ID(): {Role: wsDomain.RoleMember},
			},
			OwnerID: userDomain.UserID{UUID: uuid.Nil},
		}

		id, err := handler.Handle(ctx, cmd)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "Test Corp", found.Name())
	})

	t.Run("remove member", func(t *testing.T) {
		ws := wsDomain.NewWorkspace("Removal", "Desc", userDomain.UserID{UUID: owner.ID().UUID})
		require.NoError(t, ws.AddMember(userDomain.UserID{UUID: member.ID().UUID}, wsDomain.RoleMember))
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
		assert.NotContains(t, found.Members(), userDomain.UserID{UUID: member.ID().UUID})
	})
}
