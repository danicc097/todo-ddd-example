package adapters

import (
	"context"
	"errors"

	todoApp "github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type TodoWorkspaceProvider struct {
	Repo wsDomain.WorkspaceRepository
}

var _ todoApp.WorkspaceProvider = (*TodoWorkspaceProvider)(nil)

func NewTodoWorkspaceProvider(repo wsDomain.WorkspaceRepository) *TodoWorkspaceProvider {
	return &TodoWorkspaceProvider{Repo: repo}
}

func (g *TodoWorkspaceProvider) IsMember(ctx context.Context, wsID wsDomain.WorkspaceID, userID userDomain.UserID) (bool, error) {
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
