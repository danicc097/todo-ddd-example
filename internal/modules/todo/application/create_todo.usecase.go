package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type CreateTodoCommand struct {
	Title  string
	TagIDs []domain.TagID
}

type CreateTodoHandler struct {
	repo domain.TodoRepository
}

var _ application.RequestHandler[CreateTodoCommand, domain.TodoID] = (*CreateTodoHandler)(nil)

func NewCreateTodoHandler(repo domain.TodoRepository) *CreateTodoHandler {
	return &CreateTodoHandler{repo: repo}
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (domain.TodoID, error) {
	title, err := domain.NewTodoTitle(cmd.Title)
	if err != nil {
		return domain.TodoID{}, err
	}

	todo := domain.NewTodo(title)
	for _, tagID := range cmd.TagIDs {
		todo.AddTag(tagID)
	}

	if err := h.repo.Save(ctx, todo); err != nil {
		return domain.TodoID{}, err
	}

	return todo.ID(), nil
}
