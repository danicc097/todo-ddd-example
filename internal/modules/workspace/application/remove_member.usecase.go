package application

import (
	"context"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type RemoveWorkspaceMemberCommand struct {
	WorkspaceID domain.WorkspaceID
	MemberID    userDomain.UserID
}

type RemoveWorkspaceMemberHandler struct {
	repo domain.WorkspaceRepository
}

var _ application.RequestHandler[RemoveWorkspaceMemberCommand, application.Void] = (*RemoveWorkspaceMemberHandler)(nil)

func NewRemoveWorkspaceMemberHandler(repo domain.WorkspaceRepository) *RemoveWorkspaceMemberHandler {
	return &RemoveWorkspaceMemberHandler{repo: repo}
}

func (h *RemoveWorkspaceMemberHandler) Handle(ctx context.Context, cmd RemoveWorkspaceMemberCommand) (application.Void, error) {
	ws, err := h.repo.FindByID(ctx, cmd.WorkspaceID)
	if err != nil {
		return application.Void{}, err
	}

	if err := ws.RemoveMember(cmd.MemberID); err != nil {
		return application.Void{}, err
	}

	return application.Void{}, h.repo.Save(ctx, ws)
}
