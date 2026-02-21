package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

func TestWorkspaceMapper_MapEvent(t *testing.T) {
	t.Parallel()

	mapper := &postgres.WorkspaceMapper{}

	t.Run("WorkspaceCreatedEvent", func(t *testing.T) {
		id := uuid.New()
		ownerID := uuid.New()
		now := time.Now().Truncate(time.Second)

		wsName, _ := domain.NewWorkspaceName("Test WS")

		evt := domain.WorkspaceCreatedEvent{
			ID:       domain.WorkspaceID(id),
			Name:     wsName,
			OwnerID:  userDomain.UserID(ownerID),
			Occurred: now,
		}

		name, data, err := mapper.MapEvent(evt)
		require.NoError(t, err)
		assert.Equal(t, sharedDomain.WorkspaceCreated, name)

		payload := data.(postgres.WorkspaceCreatedDTO)

		assert.Equal(t, domain.WorkspaceID(id), payload.ID)
		assert.Equal(t, "Test WS", payload.Name)
		assert.Equal(t, userDomain.UserID(ownerID), payload.OwnerID)
	})

	t.Run("MemberAddedEvent", func(t *testing.T) {
		wsID := uuid.New()
		userID := uuid.New()
		now := time.Now().Truncate(time.Second)

		evt := domain.MemberAddedEvent{
			WsID:     domain.WorkspaceID(wsID),
			UserID:   userDomain.UserID(userID),
			Role:     domain.RoleOwner,
			Occurred: now,
		}

		name, data, err := mapper.MapEvent(evt)
		require.NoError(t, err)
		assert.Equal(t, sharedDomain.WorkspaceMemberAdded, name)

		payload := data.(postgres.MemberAddedDTO)

		assert.Equal(t, domain.WorkspaceID(wsID), payload.WorkspaceID)
		assert.Equal(t, userDomain.UserID(userID), payload.UserID)
		assert.Equal(t, "OWNER", payload.Role)
	})
}
