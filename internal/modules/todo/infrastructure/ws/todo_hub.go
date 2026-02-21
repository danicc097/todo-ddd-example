package ws

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
)

func NewTodoHub(r *redis.Client, wsQuery application.WorkspaceQueryService) *Hub {
	permProvider := func(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
		workspaces, err := wsQuery.ListByUserID(ctx, userDomain.UserID(userID))
		if err != nil {
			return nil, err
		}

		ids := make([]uuid.UUID, len(workspaces))
		for i, w := range workspaces {
			ids[i] = w.ID.UUID()
		}

		return ids, nil // rooms for todo based on workspaces
	}

	// filter room id (workspace id) to route valid messages
	msgFilter := func(message []byte) (uuid.UUID, bool) {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(message, &envelope); err != nil {
			return uuid.Nil, false
		}

		var p struct {
			WorkspaceID uuid.UUID `json:"workspace_id"`
		}
		if err := json.Unmarshal(envelope.Data, &p); err != nil {
			return uuid.Nil, false
		}

		return p.WorkspaceID, true
	}

	return NewHub(r, cache.Keys.TodoAPIUpdatesChannel(), permProvider, msgFilter)
}
