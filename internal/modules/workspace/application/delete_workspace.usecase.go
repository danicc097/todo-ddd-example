package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type DeleteWorkspaceUseCase interface {
	Execute(ctx context.Context, id uuid.UUID) error
}

type deleteWorkspaceUseCase struct {
	repo domain.WorkspaceRepository
}

func NewDeleteWorkspaceUseCase(repo domain.WorkspaceRepository) DeleteWorkspaceUseCase {
	return &deleteWorkspaceUseCase{repo: repo}
}

func (uc *deleteWorkspaceUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	meta := causation.FromContext(ctx)
	// would check permissions to delete with meta.UserID being owner or authorized user/system
	_ = meta

	// triggers the audit wrapper's Delete -> AuditDelete -> fetch -> delete -> log
	return uc.repo.Delete(ctx, id)
}
