package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type StartFocusCommand struct {
	ID domain.TodoID
}

type StartFocusResponse struct{}

type StartFocusHandler struct {
	repo   domain.TodoRepository
	wsProv WorkspaceProvider
	uow    application.UnitOfWork
}

var _ application.RequestHandler[StartFocusCommand, StartFocusResponse] = (*StartFocusHandler)(nil)

func NewStartFocusHandler(repo domain.TodoRepository, wsProv WorkspaceProvider, uow application.UnitOfWork) *StartFocusHandler {
	return &StartFocusHandler{repo: repo, wsProv: wsProv, uow: uow}
}

func (h *StartFocusHandler) Handle(ctx context.Context, cmd StartFocusCommand) (StartFocusResponse, error) {
	meta := causation.FromContext(ctx)

	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		todo, err := h.repo.FindByID(ctx, cmd.ID)
		if err != nil {
			return err
		}

		isMember, err := h.wsProv.IsMember(ctx, todo.WorkspaceID(), userDomain.UserID(meta.UserID))
		if err != nil {
			return err
		}

		if !isMember && !meta.IsSystem() {
			return wsDomain.ErrNotOwner
		}

		sessionID := domain.FocusSessionID(uuid.New())
		if err := todo.StartFocus(userDomain.UserID(meta.UserID), sessionID); err != nil {
			return err
		}

		return h.repo.Save(ctx, todo)
	})
	if err != nil {
		return StartFocusResponse{}, err
	}

	return StartFocusResponse{}, nil
}
