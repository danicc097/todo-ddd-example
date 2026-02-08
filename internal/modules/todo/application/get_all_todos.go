package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type GetAllTodosUseCase struct {
	repo domain.TodoRepository
}

func NewGetAllTodosUseCase(repo domain.TodoRepository) *GetAllTodosUseCase {
	return &GetAllTodosUseCase{repo: repo}
}

func (uc *GetAllTodosUseCase) Execute(ctx context.Context) ([]*domain.Todo, error) {
	return uc.repo.FindAll(ctx)
}
