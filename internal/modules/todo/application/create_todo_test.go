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
	t.Run("successfully create todo", func(t *testing.T) {
		fakeRepo := &domainfakes.FakeTodoRepository{}
		uc := application.NewCreateTodoUseCase(fakeRepo)

		cmd := application.CreateTodoCommand{Title: "Senior Task"}
		id, err := uc.Execute(context.Background(), cmd)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		assert.Equal(t, 1, fakeRepo.SaveCallCount())

		_, savedTodo := fakeRepo.SaveArgsForCall(0)

		assert.Equal(t, savedTodo.ID(), id)
		assert.Equal(t, "Senior Task", savedTodo.Title().String())
	})

	t.Run("returns error when domain validation fails", func(t *testing.T) {
		fakeRepo := &domainfakes.FakeTodoRepository{}
		uc := application.NewCreateTodoUseCase(fakeRepo)

		cmd := application.CreateTodoCommand{Title: ""}
		_, err := uc.Execute(context.Background(), cmd)

		assert.Error(t, err)
		assert.Equal(t, 0, fakeRepo.SaveCallCount())
	})
}
