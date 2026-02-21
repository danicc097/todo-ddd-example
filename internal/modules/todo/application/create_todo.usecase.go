package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type CreateTodoCommand struct {
	Title       string
	WorkspaceID wsDomain.WorkspaceID
	TagIDs      []domain.TagID
}

type CreateTodoResponse struct {
	ID domain.TodoID
}

type CreateTodoHandler struct {
	repo domain.TodoRepository
}

var _ application.RequestHandler[CreateTodoCommand, CreateTodoResponse] = (*CreateTodoHandler)(nil)

func NewCreateTodoHandler(repo domain.TodoRepository) *CreateTodoHandler {
	return &CreateTodoHandler{repo: repo}
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (CreateTodoResponse, error) {
	title, err := domain.NewTodoTitle(cmd.Title)
	if err != nil {
		return CreateTodoResponse{}, err
	}

	todo := domain.NewTodo(title, cmd.WorkspaceID)
	for _, tagID := range cmd.TagIDs {
		todo.AddTag(tagID)
	}

	if err := h.repo.Save(ctx, todo); err != nil {
		return CreateTodoResponse{}, err
	}

	return CreateTodoResponse{ID: todo.ID()}, nil
}
