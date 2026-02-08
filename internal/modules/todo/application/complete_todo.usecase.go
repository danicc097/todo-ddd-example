package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type CompleteTodoUseCase struct {
	tm        db.TransactionManager
	publisher domain.EventPublisher
}

func NewCompleteTodoUseCase(tm db.TransactionManager, publisher domain.EventPublisher) *CompleteTodoUseCase {
	return &CompleteTodoUseCase{tm: tm, publisher: publisher}
}

func (uc *CompleteTodoUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	var todo *domain.Todo

	err := uc.tm.Exec(ctx, func(repo domain.TodoRepository) error {
		var err error
		todo, err = repo.FindByID(ctx, id)
		if err != nil {
			return err
		}

		if err := todo.Complete(); err != nil {
			return err
		}

		if err := repo.Update(ctx, todo); err != nil {
			return err
		}

		return repo.SaveEvent(ctx, "todo.completed", map[string]any{"id": id})
	})

	if err != nil {
		return err
	}

	return uc.publisher.PublishTodoUpdated(ctx, todo)
}
