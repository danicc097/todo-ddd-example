package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

//go:generate go tool gowrap gen -g -i CompleteTodoUseCase -t ../../../../templates/transactional.gotmpl -o ../infrastructure/decorator/complete_todo_tx.gen.go
type CompleteTodoUseCase interface {
	Execute(ctx context.Context, id uuid.UUID) error
}

type completeTodoUseCase struct {
	repo domain.TodoRepository
}

func NewCompleteTodoUseCase(repo domain.TodoRepository) CompleteTodoUseCase {
	return &completeTodoUseCase{repo: repo}
}

func (uc *completeTodoUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	todo, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := todo.Complete(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, todo)
}
