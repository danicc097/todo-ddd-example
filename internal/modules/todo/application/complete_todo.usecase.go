package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/google/uuid"
)

type CompleteTodoUseCase struct {
	tm db.TransactionManager
}

func NewCompleteTodoUseCase(tm db.TransactionManager) *CompleteTodoUseCase {
	return &CompleteTodoUseCase{tm: tm}
}

func (uc *CompleteTodoUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.tm.Exec(ctx, func(p db.RepositoryProvider) error {
		repo := p.Todo()

		todo, err := repo.FindByID(ctx, id)
		if err != nil {
			return err
		}

		if err := todo.Complete(); err != nil {
			return err
		}

		return repo.Update(ctx, todo)
	})
}
