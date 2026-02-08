package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type GetTodoUseCase struct {
	repo domain.TodoRepository
}

func NewGetTodoUseCase(repo domain.TodoRepository) *GetTodoUseCase {
	return &GetTodoUseCase{repo: repo}
}

func (uc *GetTodoUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Todo, error) {
	return uc.repo.FindByID(ctx, id)
}
