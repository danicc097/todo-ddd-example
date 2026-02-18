package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type CompleteTodoCommand struct {
	ID domain.TodoID
}

type CompleteTodoHandler struct {
	repo   domain.TodoRepository
	wsRepo wsDomain.WorkspaceRepository
}

var _ application.RequestHandler[CompleteTodoCommand, application.Void] = (*CompleteTodoHandler)(nil)

func NewCompleteTodoHandler(repo domain.TodoRepository, wsRepo wsDomain.WorkspaceRepository) *CompleteTodoHandler {
	return &CompleteTodoHandler{repo: repo, wsRepo: wsRepo}
}

func (h *CompleteTodoHandler) Handle(ctx context.Context, cmd CompleteTodoCommand) (application.Void, error) {
	meta := causation.FromContext(ctx)

	todo, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return application.Void{}, err
	}

	ws, err := h.wsRepo.FindByID(ctx, todo.WorkspaceID())
	if err != nil {
		return application.Void{}, err
	}

	_, isMember := ws.Members()[userDomain.UserID{UUID: meta.UserID}]

	if !isMember && !meta.IsSystem() {
		return application.Void{}, wsDomain.ErrNotOwner
	}

	if err := todo.Complete(); err != nil {
		return application.Void{}, err
	}

	return application.Void{}, h.repo.Update(ctx, todo)
}
