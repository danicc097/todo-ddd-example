package application_test

import (
	"context"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain/domainfakes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type FakeTransactionManager struct {
	repo *domainfakes.FakeTodoRepository
}

func (f *FakeTransactionManager) Exec(ctx context.Context, fn func(domain.TodoRepository) error) error {
	return fn(f.repo)
}

func TestCreateTodoUseCase_Execute(t *testing.T) {
	t.Run("successfully create todo with tags and outbox event", func(t *testing.T) {
		fakeRepo := &domainfakes.FakeTodoRepository{}
		fakeTM := &FakeTransactionManager{repo: fakeRepo}
		uc := application.NewCreateTodoUseCase(fakeTM)

		tagID := uuid.New()
		cmd := application.CreateTodoCommand{
			Title:  "Senior Task",
			TagIDs: []uuid.UUID{tagID},
		}

		id, err := uc.Execute(context.Background(), cmd)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		assert.Equal(t, 1, fakeRepo.SaveCallCount())
		assert.Equal(t, 1, fakeRepo.AddTagCallCount())
		assert.Equal(t, 1, fakeRepo.SaveEventCallCount())

		_, tid, tTagID := fakeRepo.AddTagArgsForCall(0)
		assert.Equal(t, id, tid)
		assert.Equal(t, tagID, tTagID)

		_, eventType, _ := fakeRepo.SaveEventArgsForCall(0)
		assert.Equal(t, "todo.created", eventType)
	})

	t.Run("returns error when domain validation fails", func(t *testing.T) {
		fakeRepo := &domainfakes.FakeTodoRepository{}
		fakeTM := &FakeTransactionManager{repo: fakeRepo}
		uc := application.NewCreateTodoUseCase(fakeTM)

		cmd := application.CreateTodoCommand{Title: ""}
		_, err := uc.Execute(context.Background(), cmd)

		assert.Error(t, err)
		assert.Equal(t, 0, fakeRepo.SaveCallCount())
	})
}
