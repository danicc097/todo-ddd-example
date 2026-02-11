package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type GetAllTodosUseCase interface {
	Execute(ctx context.Context) ([]*domain.Todo, error)
}

type getAllTodosUseCase struct {
	repo domain.TodoRepository
}

func NewGetAllTodosUseCase(repo domain.TodoRepository) GetAllTodosUseCase {
	return &getAllTodosUseCase{repo: repo}
}

func (uc *getAllTodosUseCase) Execute(ctx context.Context) ([]*domain.Todo, error) {
	return uc.repo.FindAll(ctx)
}
