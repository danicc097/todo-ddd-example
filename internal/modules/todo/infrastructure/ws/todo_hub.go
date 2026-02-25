package ws

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
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

	return NewHub(r, permProvider, Config{
		GlobalChannel:          messaging.Keys.TodoAPIUpdatesChannel(),
		WorkspaceChannelPrefix: messaging.Keys.WorkspaceTodoAPIUpdatesChannelPrefix(),
	})
}
