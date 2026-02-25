package application

import (
	"context"
	"errors"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type AssignTagToTodoCommand struct {
	TodoID domain.TodoID
	TagID  domain.TagID
}

type AssignTagToTodoResponse struct{}

type AssignTagToTodoHandler struct {
	todoRepo domain.TodoRepository
	tagRepo  domain.TagRepository
	uow      application.UnitOfWork
}

var _ application.RequestHandler[AssignTagToTodoCommand, AssignTagToTodoResponse] = (*AssignTagToTodoHandler)(nil)

func NewAssignTagToTodoHandler(
	todoRepo domain.TodoRepository,
	tagRepo domain.TagRepository,
	uow application.UnitOfWork,
) *AssignTagToTodoHandler {
	return &AssignTagToTodoHandler{
		todoRepo: todoRepo,
		tagRepo:  tagRepo,
		uow:      uow,
	}
}

func (h *AssignTagToTodoHandler) Handle(ctx context.Context, cmd AssignTagToTodoCommand) (AssignTagToTodoResponse, error) {
	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		todo, err := h.todoRepo.FindByID(ctx, cmd.TodoID)
		if err != nil {
			return err
		}

		tag, err := h.tagRepo.FindByID(ctx, cmd.TagID)
		if err != nil {
			return err
		}

		if tag.WorkspaceID() != todo.WorkspaceID() {
			return errors.New("tag does not belong to the todo's workspace")
		}

		todo.AddTag(cmd.TagID)

		return h.todoRepo.Save(ctx, todo)
	})
	if err != nil {
		return AssignTagToTodoResponse{}, err
	}

	return AssignTagToTodoResponse{}, nil
}
