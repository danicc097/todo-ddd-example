package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type DeleteWorkspaceCommand struct {
	ID domain.WorkspaceID
}

type DeleteWorkspaceResponse struct{}

type DeleteWorkspaceHandler struct {
	repo domain.WorkspaceRepository
}

var _ application.RequestHandler[DeleteWorkspaceCommand, DeleteWorkspaceResponse] = (*DeleteWorkspaceHandler)(nil)

func NewDeleteWorkspaceHandler(repo domain.WorkspaceRepository) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{repo: repo}
}

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, cmd DeleteWorkspaceCommand) (DeleteWorkspaceResponse, error) {
	meta := causation.FromContext(ctx)

	// domain agnostic step-up auth check
	if !meta.MFAVerified {
		return DeleteWorkspaceResponse{}, apperrors.New(apperrors.MFARequired, "MFA required for this privileged action")
	}

	ws, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return DeleteWorkspaceResponse{}, err
	}

	if !ws.IsOwner(userDomain.UserID(meta.UserID)) && !meta.IsSystem() {
		return DeleteWorkspaceResponse{}, domain.ErrNotOwner
	}

	// repo decorators handle audit, etc.
	return DeleteWorkspaceResponse{}, h.repo.Delete(ctx, cmd.ID)
}
