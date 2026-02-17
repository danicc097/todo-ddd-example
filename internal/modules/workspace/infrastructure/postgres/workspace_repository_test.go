package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestWorkspaceRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)
	repo := wsPg.NewWorkspaceRepo(pool)

	owner := fixtures.RandomUser(ctx, t)
	member := fixtures.RandomUser(ctx, t)

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

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE event_type = 'workspace.created' AND (payload ->> 'id')::uuid = $1", ws.ID().UUID).Scan(&count)
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
