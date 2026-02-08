package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type CompleteTodoUseCase struct {
	repo      domain.TodoRepository
	publisher domain.EventPublisher
}

func NewCompleteTodoUseCase(repo domain.TodoRepository, publisher domain.EventPublisher) *CompleteTodoUseCase {
	return &CompleteTodoUseCase{repo: repo, publisher: publisher}
}

func (uc *CompleteTodoUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	todo, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := todo.Complete(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, todo); err != nil {
		return err
	}

	// Trigger the side effect (Redis/WebSockets)
	return uc.publisher.PublishTodoUpdated(ctx, todo)
}
