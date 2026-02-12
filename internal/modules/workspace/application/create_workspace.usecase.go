package application

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type CreateWorkspaceCommand struct {
	Name        string
	Description string
}

type CreateWorkspaceUseCase interface {
	Execute(ctx context.Context, cmd CreateWorkspaceCommand) (uuid.UUID, error)
}

type createWorkspaceUseCase struct {
	repo domain.WorkspaceRepository
}

func NewCreateWorkspaceUseCase(repo domain.WorkspaceRepository) CreateWorkspaceUseCase {
	return &createWorkspaceUseCase{repo: repo}
}

func (uc *createWorkspaceUseCase) Execute(ctx context.Context, cmd CreateWorkspaceCommand) (uuid.UUID, error) {
	meta := causation.FromContext(ctx)
	if !meta.IsAuthenticated() {
		return uuid.Nil, errors.New("unauthenticated: actor required for auditing")
	}

	ws := domain.NewWorkspace(cmd.Name, cmd.Description, meta.UserID)

	// triggers audit log
	if err := uc.repo.Save(ctx, ws); err != nil {
		return uuid.Nil, err
	}

	return ws.ID(), nil
}
