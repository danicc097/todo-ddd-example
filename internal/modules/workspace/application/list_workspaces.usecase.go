package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type ListWorkspacesUseCase interface {
	Execute(ctx context.Context) ([]*domain.Workspace, error)
}

type listWorkspacesUseCase struct {
	repo domain.WorkspaceRepository
}

func NewListWorkspacesUseCase(repo domain.WorkspaceRepository) ListWorkspacesUseCase {
	return &listWorkspacesUseCase{repo: repo}
}

func (uc *listWorkspacesUseCase) Execute(ctx context.Context) ([]*domain.Workspace, error) {
	return uc.repo.FindAll(ctx)
}
