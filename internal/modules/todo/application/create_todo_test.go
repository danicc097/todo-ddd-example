package application_test

import (
	"context"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain/domainfakes"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type FakeRepositoryProvider struct {
	todoRepo *domainfakes.FakeTodoRepository
}

func (p *FakeRepositoryProvider) Todo() domain.TodoRepository { return p.todoRepo }
func (p *FakeRepositoryProvider) User() userDomain.UserRepository { return nil }

type FakeTransactionManager struct {
	repo *domainfakes.FakeTodoRepository
}

func (f *FakeTransactionManager) Exec(ctx context.Context, fn func(db.RepositoryProvider) error) error {
	return fn(&FakeRepositoryProvider{todoRepo: f.repo})
}

func TestCreateTodoUseCase_Execute(t *testing.T) {
	setup := func() (*domainfakes.FakeTodoRepository, *application.CreateTodoUseCase) {
		repo := &domainfakes.FakeTodoRepository{}
		tm := &FakeTransactionManager{repo: repo}
		return repo, application.NewCreateTodoUseCase(tm)
	}

	t.Run("successfully create todo with tags and outbox event", func(t *testing.T) {
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
		assert.Equal(t, 1, repo.AddTagCallCount())
		assert.Equal(t, 1, repo.SaveEventCallCount())

		_, tid, tTagID := repo.AddTagArgsForCall(0)
		assert.Equal(t, id, tid)
		assert.Equal(t, tagID, tTagID)

		_, eventType, payload := repo.SaveEventArgsForCall(0)
		assert.Equal(t, "todo.created", eventType)
		assert.IsType(t, &domain.Todo{}, payload)
	})

	t.Run("returns error when domain validation fails", func(t *testing.T) {
		repo, uc := setup()

		cmd := application.CreateTodoCommand{Title: ""}
		_, err := uc.Execute(context.Background(), cmd)

		assert.Error(t, err)
		assert.Equal(t, 0, repo.SaveCallCount())
		assert.Equal(t, 0, repo.SaveEventCallCount())
	})
}
