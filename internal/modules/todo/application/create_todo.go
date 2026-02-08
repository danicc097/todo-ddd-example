package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type CreateTodoCommand struct {
	Title string
}

type CreateTodoUseCase struct {
	repo domain.TodoRepository
}

func NewCreateTodoUseCase(repo domain.TodoRepository) *CreateTodoUseCase {
	return &CreateTodoUseCase{repo: repo}
}

func (uc *CreateTodoUseCase) Execute(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
	title, err := domain.NewTodoTitle(cmd.Title)
	if err != nil {
		return uuid.UUID{}, err
	}

	todo := domain.CreateTodo(title)

	return uc.repo.Save(ctx, todo)
}
