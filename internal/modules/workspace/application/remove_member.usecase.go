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

type RemoveWorkspaceMemberResponse struct{}

type RemoveWorkspaceMemberHandler struct {
	repo domain.WorkspaceRepository
	uow  application.UnitOfWork
}

var _ application.RequestHandler[RemoveWorkspaceMemberCommand, RemoveWorkspaceMemberResponse] = (*RemoveWorkspaceMemberHandler)(nil)

func NewRemoveWorkspaceMemberHandler(repo domain.WorkspaceRepository, uow application.UnitOfWork) *RemoveWorkspaceMemberHandler {
	return &RemoveWorkspaceMemberHandler{repo: repo, uow: uow}
}

func (h *RemoveWorkspaceMemberHandler) Handle(ctx context.Context, cmd RemoveWorkspaceMemberCommand) (RemoveWorkspaceMemberResponse, error) {
	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		ws, err := h.repo.FindByID(ctx, cmd.WorkspaceID)
		if err != nil {
			return err
		}

		if err := ws.RemoveMember(cmd.MemberID); err != nil {
			return err
		}

		return h.repo.Save(ctx, ws)
	})
	if err != nil {
		return RemoveWorkspaceMemberResponse{}, err
	}

	return RemoveWorkspaceMemberResponse{}, nil
}
