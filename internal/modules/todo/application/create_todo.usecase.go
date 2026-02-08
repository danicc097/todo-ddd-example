package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type CreateTodoCommand struct {
	Title string
}

type CreateTodoUseCase struct {
	repo      domain.TodoRepository
	publisher domain.EventPublisher
}

func NewCreateTodoUseCase(repo domain.TodoRepository, publisher domain.EventPublisher) *CreateTodoUseCase {
	return &CreateTodoUseCase{repo: repo, publisher: publisher}
}

func (uc *CreateTodoUseCase) Execute(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
	title, err := domain.NewTodoTitle(cmd.Title)
	if err != nil {
		return uuid.UUID{}, err
	}

	todo := domain.CreateTodo(title)

	if _, err := uc.repo.Save(ctx, todo); err != nil {
		return uuid.UUID{}, err
	}

	if err := uc.publisher.PublishTodoCreated(ctx, todo); err != nil {
		return uuid.UUID{}, err
	}

	return todo.ID(), nil
}
