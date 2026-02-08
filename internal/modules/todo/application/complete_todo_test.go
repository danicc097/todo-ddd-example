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
	setup := func() (*domainfakes.FakeTodoRepository, *domainfakes.FakeEventPublisher, *application.CompleteTodoUseCase) {
		repo := &domainfakes.FakeTodoRepository{}
		pub := &domainfakes.FakeEventPublisher{}
		tm := &FakeTransactionManager{repo: repo}
		return repo, pub, application.NewCompleteTodoUseCase(tm, pub)
	}

	t.Run("successfully completes todo and publishes event", func(t *testing.T) {
		repo, pub, uc := setup()
		id := uuid.New()
		title, _ := domain.NewTodoTitle("Reliable Task")
		existingTodo := domain.NewTodo(id, title, domain.StatusPending, time.Now())

		repo.FindByIDReturns(existingTodo, nil)
		repo.UpdateReturns(nil)
		pub.PublishTodoUpdatedReturns(nil)

		err := uc.Execute(context.Background(), id)

		assert.NoError(t, err)

		assert.Equal(t, 1, repo.UpdateCallCount())
		assert.Equal(t, 1, pub.PublishTodoUpdatedCallCount())
	})

	t.Run("aborts if todo update fails", func(t *testing.T) {
		repo, pub, uc := setup()
		id := uuid.New()
		title, _ := domain.NewTodoTitle("Task")
		repo.FindByIDReturns(domain.NewTodo(id, title, domain.StatusPending, time.Now()), nil)
		repo.UpdateReturns(assert.AnError)

		err := uc.Execute(context.Background(), id)

		assert.Error(t, err)
		assert.Equal(t, 0, pub.PublishTodoUpdatedCallCount())
	})

	t.Run("returns error if todo is not found", func(t *testing.T) {
		repo, _, uc := setup()
		id := uuid.New()
		repo.FindByIDReturns(nil, domain.ErrTodoNotFound)

		err := uc.Execute(context.Background(), id)

		assert.ErrorIs(t, err, domain.ErrTodoNotFound)
	})
}
