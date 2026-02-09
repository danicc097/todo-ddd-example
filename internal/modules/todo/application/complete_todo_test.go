package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain/domainfakes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCompleteTodoUseCase_Execute(t *testing.T) {
	setup := func() (*domainfakes.FakeTodoRepository, *application.CompleteTodoUseCase) {
		repo := &domainfakes.FakeTodoRepository{}
		tm := &FakeTransactionManager{repo: repo}
		return repo, application.NewCompleteTodoUseCase(tm)
	}

	t.Run("successfully completes todo and saves outbox event", func(t *testing.T) {
		repo, uc := setup()
		id := uuid.New()
		title, _ := domain.NewTodoTitle("Reliable Task")
		existingTodo := domain.NewTodo(id, title, domain.StatusPending, time.Now())

		repo.FindByIDReturns(existingTodo, nil)
		repo.UpdateReturns(nil)
		repo.SaveEventReturns(nil)

		err := uc.Execute(context.Background(), id)

		assert.NoError(t, err)
		assert.Equal(t, 1, repo.UpdateCallCount())
		assert.Equal(t, 1, repo.SaveEventCallCount())

		_, eventType, payload := repo.SaveEventArgsForCall(0)
		assert.Equal(t, "todo.completed", eventType)

		// Verify the entity itself was passed to the outbox
		assert.IsType(t, &domain.Todo{}, payload)
		assert.Equal(t, domain.StatusCompleted, payload.(*domain.Todo).Status())
	})

	t.Run("aborts if todo update fails", func(t *testing.T) {
		repo, uc := setup()
		id := uuid.New()
		title, _ := domain.NewTodoTitle("Task")
		repo.FindByIDReturns(domain.NewTodo(id, title, domain.StatusPending, time.Now()), nil)
		repo.UpdateReturns(assert.AnError)

		err := uc.Execute(context.Background(), id)

		assert.Error(t, err)
		assert.Equal(t, 0, repo.SaveEventCallCount())
	})

	t.Run("returns error if todo is not found", func(t *testing.T) {
		repo, uc := setup()
		id := uuid.New()
		repo.FindByIDReturns(nil, domain.ErrTodoNotFound)

		err := uc.Execute(context.Background(), id)

		assert.ErrorIs(t, err, domain.ErrTodoNotFound)
		assert.Equal(t, 0, repo.SaveEventCallCount())
	})
}
