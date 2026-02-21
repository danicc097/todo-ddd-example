package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type AssignTagToTodoCommand struct {
	TodoID domain.TodoID
	TagID  domain.TagID
}

type AssignTagToTodoResponse struct{}

type AssignTagToTodoHandler struct {
	repo domain.TodoRepository
}

var _ application.RequestHandler[AssignTagToTodoCommand, AssignTagToTodoResponse] = (*AssignTagToTodoHandler)(nil)

func NewAssignTagToTodoHandler(repo domain.TodoRepository) *AssignTagToTodoHandler {
	return &AssignTagToTodoHandler{repo: repo}
}

func (h *AssignTagToTodoHandler) Handle(ctx context.Context, cmd AssignTagToTodoCommand) (AssignTagToTodoResponse, error) {
	todo, err := h.repo.FindByID(ctx, cmd.TodoID)
	if err != nil {
		return AssignTagToTodoResponse{}, err
	}

	todo.AddTag(cmd.TagID)

	if err := h.repo.Save(ctx, todo); err != nil {
		return AssignTagToTodoResponse{}, err
	}

	return AssignTagToTodoResponse{}, nil
}
