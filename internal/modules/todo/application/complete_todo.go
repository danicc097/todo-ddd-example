package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type CompleteTodoUseCase struct {
	repo domain.TodoRepository
}

func NewCompleteTodoUseCase(repo domain.TodoRepository) *CompleteTodoUseCase {
	return &CompleteTodoUseCase{repo: repo}
}

func (uc *CompleteTodoUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	todo, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	todo.Complete()

	return uc.repo.Update(ctx, todo)
}
