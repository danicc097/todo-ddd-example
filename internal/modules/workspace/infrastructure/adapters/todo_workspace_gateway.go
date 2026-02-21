package adapters

import (
	"context"
	"errors"

	todoApp "github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type TodoWorkspaceGateway struct {
	Repo wsDomain.WorkspaceRepository
}

func NewTodoWorkspaceGateway(repo wsDomain.WorkspaceRepository) *TodoWorkspaceGateway {
	return &TodoWorkspaceGateway{Repo: repo}
}

func (g *TodoWorkspaceGateway) IsMember(ctx context.Context, wsID wsDomain.WorkspaceID, userID userDomain.UserID) (bool, error) {
	ws, err := g.Repo.FindByID(ctx, wsID)
	if err != nil {
		if errors.Is(err, wsDomain.ErrWorkspaceNotFound) {
			return false, nil
		}

		return false, err
	}

	_, isMember := ws.Members()[userID]

	return isMember, nil
}

var _ todoApp.WorkspaceProvider = (*TodoWorkspaceGateway)(nil)
