package application_test

import (
	"context"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db/dbfakes"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain/domainfakes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type FakeTransactionManager struct {
	repo *domainfakes.FakeTodoRepository
}

func (f *FakeTransactionManager) Exec(ctx context.Context, fn func(db.RepositoryProvider) error) error {
	fakeProvider := &dbfakes.FakeRepositoryProvider{}
	fakeProvider.TodoReturns(f.repo)
	return fn(fakeProvider)
}

func TestCreateTodoUseCase_Execute(t *testing.T) {
	setup := func() (*domainfakes.FakeTodoRepository, *application.CreateTodoUseCase) {
		repo := &domainfakes.FakeTodoRepository{}
		tm := &FakeTransactionManager{repo: repo}
		return repo, application.NewCreateTodoUseCase(tm)
	}

	t.Run("successfully create todo with tags", func(t *testing.T) {
		repo, uc := setup()

		tagID := uuid.New()
		cmd := application.CreateTodoCommand{
			Title:  "Senior Task",
			TagIDs: []uuid.UUID{tagID},
		}

		id, err := uc.Execute(context.Background(), cmd)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		assert.Equal(t, 1, repo.SaveCallCount())

		// Verify tags were added to the entity passed to Save
		_, savedTodo := repo.SaveArgsForCall(0)
		assert.Contains(t, savedTodo.Tags(), tagID)
	})

	t.Run("returns error when domain validation fails", func(t *testing.T) {
		repo, uc := setup()

		cmd := application.CreateTodoCommand{Title: ""}
		_, err := uc.Execute(context.Background(), cmd)

		assert.Error(t, err)
		assert.Equal(t, 0, repo.SaveCallCount())
	})
}
