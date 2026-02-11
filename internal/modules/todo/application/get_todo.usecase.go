package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type GetTodoUseCase interface {
	Execute(ctx context.Context, id uuid.UUID) (*domain.Todo, error)
}

type getTodoUseCase struct {
	repo domain.TodoRepository
}

func NewGetTodoUseCase(repo domain.TodoRepository) GetTodoUseCase {
	return &getTodoUseCase{repo: repo}
}

func (uc *getTodoUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Todo, error) {
	return uc.repo.FindByID(ctx, id)
}
