package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	wsAdapters "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/adapters"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/middleware"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCreateTodoUseCase_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)

	repo := todoPg.NewTodoRepo(pool)
	wsRepo := wsPg.NewWorkspaceRepo(pool)
	wsProv := wsAdapters.NewTodoWorkspaceProvider(wsRepo)

	baseHandler := application.NewCreateTodoHandler(repo, wsProv)
	handler := middleware.Transactional(pool, baseHandler)

	t.Run("creates", func(t *testing.T) {
		user := fixtures.RandomUser(ctx, t)
		ws := fixtures.RandomWorkspace(ctx, t, user.ID())
		tag := fixtures.RandomTag(ctx, t, ws.ID())

		userCtx := causation.WithMetadata(ctx, causation.Metadata{UserID: user.ID().UUID()})

		title := "Integration " + user.ID().String()

		cmd := application.CreateTodoCommand{
			Title:       title,
			WorkspaceID: ws.ID(),
			TagIDs:      []domain.TagID{tag.ID()},
		}

		resp, err := handler.Handle(userCtx, cmd)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, resp.ID)
		require.NoError(t, err)
		assert.Equal(t, title, found.Title().String())
		assert.Contains(t, found.Tags(), tag.ID())

		var count int

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE event_type = $1 AND aggregate_id = $2", sharedDomain.TodoCreated, resp.ID.UUID()).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("fails if user is not in workspace", func(t *testing.T) {
		owner := fixtures.RandomUser(ctx, t)
		ws := fixtures.RandomWorkspace(ctx, t, owner.ID())

		stranger := fixtures.RandomUser(ctx, t)
		strangerCtx := causation.WithMetadata(ctx, causation.Metadata{UserID: stranger.ID().UUID()})

		cmd := application.CreateTodoCommand{
			Title:       "Todo",
			WorkspaceID: ws.ID(),
		}

		_, err := handler.Handle(strangerCtx, cmd)
		assert.ErrorIs(t, err, wsDomain.ErrNotOwner)
	})
}
