package postgres_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
)

func TestWorkspaceMapper_MapEvent(t *testing.T) {
	t.Parallel()

	mapper := &postgres.WorkspaceMapper{}

	t.Run("WorkspaceCreatedEvent", func(t *testing.T) {
		id := uuid.New()
		ownerID := uuid.New()
		now := time.Now().Truncate(time.Second)

		evt := domain.WorkspaceCreatedEvent{
			ID:       domain.WorkspaceID{UUID: id},
			Name:     "Test WS",
			OwnerID:  userDomain.UserID{UUID: ownerID},
			Occurred: now,
		}

		name, data, err := mapper.MapEvent(evt)
		require.NoError(t, err)
		assert.Equal(t, "workspace.created", name)

		var payload map[string]any

		err = json.Unmarshal(data, &payload)
		require.NoError(t, err)

		assert.Equal(t, id.String(), payload["id"])
		assert.Equal(t, "Test WS", payload["name"])
		assert.Equal(t, ownerID.String(), payload["owner_id"])
	})

	t.Run("MemberAddedEvent", func(t *testing.T) {
		wsID := uuid.New()
		userID := uuid.New()
		now := time.Now().Truncate(time.Second)

		evt := domain.MemberAddedEvent{
			WorkspaceID: domain.WorkspaceID{UUID: wsID},
			UserID:      userDomain.UserID{UUID: userID},
			Role:        "OWNER",
			Occurred:    now,
		}

		name, data, err := mapper.MapEvent(evt)
		require.NoError(t, err)
		assert.Equal(t, "workspace.member_added", name)

		var payload map[string]any

		err = json.Unmarshal(data, &payload)
		require.NoError(t, err)

		assert.Equal(t, wsID.String(), payload["workspace_id"])
		assert.Equal(t, userID.String(), payload["user_id"])
		assert.Equal(t, "OWNER", payload["role"])
	})
}
