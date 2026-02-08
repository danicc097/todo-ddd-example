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
		return repo, pub, application.NewCompleteTodoUseCase(repo, pub)
	}

	t.Run("successfully completes todo and publishes event", func(t *testing.T) {
		repo, pub, uc := setup()
		id := uuid.New()
		title, _ := domain.NewTodoTitle("Task")
		existingTodo := domain.NewTodo(id, title, domain.StatusPending, time.Now())

		repo.FindByIDReturns(existingTodo, nil)
		repo.UpdateReturns(nil)
		pub.PublishTodoUpdatedReturns(nil)

		err := uc.Execute(context.Background(), id)

		assert.NoError(t, err)
		assert.Equal(t, 1, repo.UpdateCallCount())
		assert.Equal(t, 1, pub.PublishTodoUpdatedCallCount())
	})
}
