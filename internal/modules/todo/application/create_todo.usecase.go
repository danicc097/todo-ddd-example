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

	todo := domain.CreateTodo(title)

	err = uc.tm.Exec(ctx, func(repo domain.TodoRepository) error {
		if _, err := repo.Save(ctx, todo); err != nil {
			return err
		}

		for _, tagID := range cmd.TagIDs {
			if err := repo.AddTag(ctx, todo.ID(), tagID); err != nil {
				return err
			}
		}

		return repo.SaveEvent(ctx, "todo.created", todo)
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	return todo.ID(), nil
}
