package application_test

import (
	"context"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain/domainfakes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateTodoUseCase_Execute(t *testing.T) {
	setup := func() (*domainfakes.FakeTodoRepository, *domainfakes.FakeEventPublisher, *application.CreateTodoUseCase) {
		repo := &domainfakes.FakeTodoRepository{}
		pub := &domainfakes.FakeEventPublisher{}
		return repo, pub, application.NewCreateTodoUseCase(repo, pub)
	}

	t.Run("successfully create todo and publish event", func(t *testing.T) {
		repo, pub, uc := setup()

		repo.SaveReturns(uuid.New(), nil)
		pub.PublishTodoCreatedReturns(nil)

		cmd := application.CreateTodoCommand{Title: "RabbitMQ Task"}
		id, err := uc.Execute(context.Background(), cmd)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		assert.Equal(t, 1, repo.SaveCallCount())
		assert.Equal(t, 1, pub.PublishTodoCreatedCallCount())

		_, savedTodo := pub.PublishTodoCreatedArgsForCall(0)
		assert.Equal(t, id, savedTodo.ID())
		assert.Equal(t, "RabbitMQ Task", savedTodo.Title().String())
	})

	t.Run("returns error when domain validation fails", func(t *testing.T) {
		repo, _, uc := setup()

		cmd := application.CreateTodoCommand{Title: ""}
		_, err := uc.Execute(context.Background(), cmd)

		assert.Error(t, err)
		assert.Equal(t, 0, repo.SaveCallCount())
	})

	t.Run("fail if publisher fails", func(t *testing.T) {
		repo, pub, uc := setup()

		repo.SaveReturns(uuid.New(), nil)
		pub.PublishTodoCreatedReturns(assert.AnError)

		cmd := application.CreateTodoCommand{Title: "Task"}
		_, err := uc.Execute(context.Background(), cmd)

		assert.Error(t, err)
		assert.Equal(t, 1, repo.SaveCallCount())
	})
}
