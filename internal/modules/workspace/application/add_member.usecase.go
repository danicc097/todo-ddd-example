package application

import (
	"context"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type AddWorkspaceMemberCommand struct {
	WorkspaceID domain.WorkspaceID
	UserID      userDomain.UserID
	Role        domain.WorkspaceRole
}

type AddWorkspaceMemberHandler struct {
	repo domain.WorkspaceRepository
}

var _ application.RequestHandler[AddWorkspaceMemberCommand, application.Void] = (*AddWorkspaceMemberHandler)(nil)

func NewAddWorkspaceMemberHandler(repo domain.WorkspaceRepository) *AddWorkspaceMemberHandler {
	return &AddWorkspaceMemberHandler{repo: repo}
}

func (h *AddWorkspaceMemberHandler) Handle(ctx context.Context, cmd AddWorkspaceMemberCommand) (application.Void, error) {
	meta := causation.FromContext(ctx)

	ws, err := h.repo.FindByID(ctx, cmd.WorkspaceID)
	if err != nil {
		return application.Void{}, err
	}

	if !ws.IsOwner(userDomain.UserID(meta.UserID)) && !meta.IsSystem() {
		return application.Void{}, domain.ErrNotOwner
	}

	if err := ws.AddMember(cmd.UserID, cmd.Role); err != nil {
		return application.Void{}, err
	}

	if err := h.repo.Save(ctx, ws); err != nil {
		return application.Void{}, err
	}

	return application.Void{}, nil
}
