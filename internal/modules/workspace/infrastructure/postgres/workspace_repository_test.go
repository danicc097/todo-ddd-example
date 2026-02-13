package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func createTestUser(ctx context.Context, t *testing.T, pool *testutils.PostgreSQLContainer, email string) *userDomain.User {
	repo := userPg.NewUserRepo(pool.Pool())
	e, _ := userDomain.NewUserEmail(email)
	u := userDomain.CreateUser(e, "test")
	require.NoError(t, repo.Save(ctx, u))

	return u
}

func TestWorkspaceRepo_Integration(t *testing.T) {
	ctx := context.Background()

	pg := testutils.NewPostgreSQLContainer(ctx, t)
	defer pg.Close(ctx, t)

	pool := pg.Connect(ctx, t)
	repo := NewWorkspaceRepo(pool)

	owner := createTestUser(ctx, t, pg, "owner@mail.com")
	member := createTestUser(ctx, t, pg, "member@mail.com")

	t.Run("save and find", func(t *testing.T) {
		ws := wsDomain.NewWorkspace("My Workspace", "Desc", owner.ID())
		err := repo.Save(ctx, ws)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, ws.ID())
		require.NoError(t, err)
		assert.Equal(t, ws.ID(), found.ID())
		assert.Equal(t, ws.Name(), found.Name())
		assert.Len(t, found.Members(), 1)

		var count int

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE event_type = 'workspace.created'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("add member", func(t *testing.T) {
		ws := wsDomain.NewWorkspace("Team WS", "Desc", owner.ID())
		require.NoError(t, repo.Save(ctx, ws))

		require.NoError(t, ws.AddMember(member.ID(), wsDomain.RoleMember))
		require.NoError(t, repo.Save(ctx, ws))

		found, err := repo.FindByID(ctx, ws.ID())
		require.NoError(t, err)
		assert.Len(t, found.Members(), 2)
		assert.Equal(t, wsDomain.RoleMember, found.Members()[member.ID()])
	})

	t.Run("remove member", func(t *testing.T) {
		ws := wsDomain.NewWorkspace("Delete WS", "Desc", owner.ID())
		require.NoError(t, ws.AddMember(member.ID(), wsDomain.RoleMember))
		require.NoError(t, repo.Save(ctx, ws))

		require.NoError(t, ws.RemoveMember(member.ID()))
		require.NoError(t, repo.Save(ctx, ws))

		found, err := repo.FindByID(ctx, ws.ID())
		require.NoError(t, err)
		assert.Len(t, found.Members(), 1)
		assert.NotContains(t, found.Members(), member.ID())
	})

	t.Run("delete workspace", func(t *testing.T) {
		ws := wsDomain.NewWorkspace("Gone WS", "Desc", owner.ID())
		require.NoError(t, repo.Save(ctx, ws))

		err := repo.Delete(ctx, ws.ID())
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, ws.ID())
		assert.ErrorIs(t, err, wsDomain.ErrWorkspaceNotFound)
	})
}
