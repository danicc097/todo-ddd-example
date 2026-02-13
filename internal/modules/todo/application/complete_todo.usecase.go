package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type CompleteTodoCommand struct {
	ID domain.TodoID
}

type CompleteTodoHandler struct {
	repo domain.TodoRepository
}

var _ application.RequestHandler[CompleteTodoCommand, application.Void] = (*CompleteTodoHandler)(nil)

func NewCompleteTodoHandler(repo domain.TodoRepository) *CompleteTodoHandler {
	return &CompleteTodoHandler{repo: repo}
}

func (h *CompleteTodoHandler) Handle(ctx context.Context, cmd CompleteTodoCommand) (application.Void, error) {
	todo, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return application.Void{}, err
	}

	if err := todo.Complete(); err != nil {
		return application.Void{}, err
	}

	return application.Void{}, h.repo.Update(ctx, todo)
}
