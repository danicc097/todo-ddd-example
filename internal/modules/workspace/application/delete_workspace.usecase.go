package application

import (
	"context"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type DeleteWorkspaceCommand struct {
	ID domain.WorkspaceID
}

type DeleteWorkspaceHandler struct {
	repo domain.WorkspaceRepository
}

var _ application.RequestHandler[DeleteWorkspaceCommand, application.Void] = (*DeleteWorkspaceHandler)(nil)

func NewDeleteWorkspaceHandler(repo domain.WorkspaceRepository) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{repo: repo}
}

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, cmd DeleteWorkspaceCommand) (application.Void, error) {
	meta := causation.FromContext(ctx)

	ws, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return application.Void{}, err
	}

	if !ws.IsOwner(userDomain.UserID{UUID: meta.UserID}) && !meta.IsSystem() {
		return application.Void{}, domain.ErrNotOwner
	}

	// repo decorators handle audit, etc.
	return application.Void{}, h.repo.Delete(ctx, cmd.ID)
}
