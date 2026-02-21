package adapters

import (
	"context"
	"errors"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type WorkspaceUserProvider struct {
	Repo userDomain.UserRepository
}

func NewWorkspaceUserProvider(repo userDomain.UserRepository) *WorkspaceUserProvider {
	return &WorkspaceUserProvider{Repo: repo}
}

func (g *WorkspaceUserProvider) Exists(ctx context.Context, userID userDomain.UserID) (bool, error) {
	_, err := g.Repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
