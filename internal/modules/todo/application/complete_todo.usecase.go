package application

import (
	"context"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type CompleteTodoCommand struct {
	ID domain.TodoID
}

type CompleteTodoResponse struct{}

type WorkspaceProvider interface {
	IsMember(ctx context.Context, wsID wsDomain.WorkspaceID, userID userDomain.UserID) (bool, error)
}

type CompleteTodoHandler struct {
	repo   domain.TodoRepository
	wsProv WorkspaceProvider
	uow    application.UnitOfWork
}

var _ application.RequestHandler[CompleteTodoCommand, CompleteTodoResponse] = (*CompleteTodoHandler)(nil)

func NewCompleteTodoHandler(repo domain.TodoRepository, wsProv WorkspaceProvider, uow application.UnitOfWork) *CompleteTodoHandler {
	return &CompleteTodoHandler{repo: repo, wsProv: wsProv, uow: uow}
}

func (h *CompleteTodoHandler) Handle(ctx context.Context, cmd CompleteTodoCommand) (CompleteTodoResponse, error) {
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

		if err := todo.Complete(userDomain.UserID(meta.UserID), time.Now()); err != nil {
			return err
		}

		return h.repo.Save(ctx, todo)
	})
	if err != nil {
		return CompleteTodoResponse{}, err
	}

	return CompleteTodoResponse{}, nil
}
