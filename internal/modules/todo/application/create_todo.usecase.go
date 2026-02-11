package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type CreateTodoCommand struct {
	Title  string
	TagIDs []uuid.UUID
}

//go:generate go tool gowrap gen -g -i CreateTodoUseCase -t ../../../../templates/transactional.gotmpl -o ../infrastructure/decorator/create_todo_tx.gen.go
type CreateTodoUseCase interface {
	Execute(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error)
}

type createTodoUseCase struct {
	repo domain.TodoRepository
}

func NewCreateTodoUseCase(repo domain.TodoRepository) CreateTodoUseCase {
	return &createTodoUseCase{repo: repo}
}

func (uc *createTodoUseCase) Execute(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
	title, err := domain.NewTodoTitle(cmd.Title)
	if err != nil {
		return uuid.UUID{}, err
	}

	todo := domain.NewTodo(title)
	for _, tagID := range cmd.TagIDs {
		todo.AddTag(tagID)
	}

	if err := uc.repo.Save(ctx, todo); err != nil {
		return uuid.UUID{}, err
	}

	return todo.ID(), nil
}
