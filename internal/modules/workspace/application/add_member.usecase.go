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

type AddWorkspaceMemberResponse struct{}

type AddWorkspaceMemberHandler struct {
	repo domain.WorkspaceRepository
	uow  application.UnitOfWork
}

var _ application.RequestHandler[AddWorkspaceMemberCommand, AddWorkspaceMemberResponse] = (*AddWorkspaceMemberHandler)(nil)

func NewAddWorkspaceMemberHandler(repo domain.WorkspaceRepository, uow application.UnitOfWork) *AddWorkspaceMemberHandler {
	return &AddWorkspaceMemberHandler{repo: repo, uow: uow}
}

func (h *AddWorkspaceMemberHandler) Handle(ctx context.Context, cmd AddWorkspaceMemberCommand) (AddWorkspaceMemberResponse, error) {
	meta := causation.FromContext(ctx)

	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		ws, err := h.repo.FindByID(ctx, cmd.WorkspaceID)
		if err != nil {
			return err
		}

		if !ws.IsOwner(userDomain.UserID(meta.UserID)) && !meta.IsSystem() {
			return domain.ErrNotOwner
		}

		if err := ws.AddMember(cmd.UserID, cmd.Role); err != nil {
			return err
		}

		return h.repo.Save(ctx, ws)
	})
	if err != nil {
		return AddWorkspaceMemberResponse{}, err
	}

	return AddWorkspaceMemberResponse{}, nil
}
