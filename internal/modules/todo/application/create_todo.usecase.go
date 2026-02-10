package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type CreateTodoCommand struct {
	Title  string
	TagIDs []uuid.UUID
}

type CreateTodoUseCase struct {
	tm db.TransactionManager
}

func NewCreateTodoUseCase(tm db.TransactionManager) *CreateTodoUseCase {
	return &CreateTodoUseCase{tm: tm}
}

func (uc *CreateTodoUseCase) Execute(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
	title, err := domain.NewTodoTitle(cmd.Title)
	if err != nil {
		return uuid.UUID{}, err
	}

	todo := domain.NewTodo(title)
	for _, tagID := range cmd.TagIDs {
		todo.AddTag(tagID)
	}

	err = uc.tm.Exec(ctx, func(p db.RepositoryProvider) error {
		return p.Todo().Save(ctx, todo)
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	return todo.ID(), nil
}
