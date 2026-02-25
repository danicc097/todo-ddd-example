package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/schedule/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/schedule/domain"
	schedulePg "github.com/danicc097/todo-ddd-example/internal/modules/schedule/infrastructure/postgres"
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestCommitTaskHandler_Handle_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)

	uow := sharedPg.NewUnitOfWork(pool)
	scheduleRepo := schedulePg.NewScheduleRepo(pool, uow)
	todoRepo := todoPg.NewTodoRepo(pool, uow)

	handler := sharedApp.Retry(
		application.NewCommitTaskHandler(scheduleRepo, todoRepo, uow),
		3,
	)

	t.Run("success", func(t *testing.T) {
		user := fixtures.RandomUser(ctx, t)
		ws := fixtures.RandomWorkspace(ctx, t, user.ID())
		todo := fixtures.RandomTodo(ctx, t, ws.ID())

		userCtx := causation.WithMetadata(ctx, causation.Metadata{UserID: user.ID().UUID()})
		date := time.Now().Format(time.DateOnly)

		cmd := application.CommitTaskCommand{
			TodoID: todo.ID().UUID(),
			Cost:   3,
			Date:   date,
		}

		_, err := handler.Handle(userCtx, cmd)
		require.NoError(t, err)

		s, err := scheduleRepo.FindByUserAndDate(ctx, user.ID(), domain.ScheduleDate(date))
		require.NoError(t, err)
		assert.Len(t, s.CommittedTasks(), 1)
		assert.Equal(t, 3, int(s.CommittedTasks()[todo.ID()]))
	})

	t.Run("failure - capacity exceeded", func(t *testing.T) {
		user := fixtures.RandomUser(ctx, t)
		ws := fixtures.RandomWorkspace(ctx, t, user.ID())
		userCtx := causation.WithMetadata(ctx, causation.Metadata{UserID: user.ID().UUID()})
		date := time.Now().AddDate(0, 0, 1).Format(time.DateOnly)

		var err error

		for _, cost := range []int{5, 5, 1} {
			todo := fixtures.RandomTodo(ctx, t, ws.ID())
			_, err = handler.Handle(userCtx, application.CommitTaskCommand{
				TodoID: todo.ID().UUID(),
				Cost:   cost,
				Date:   date,
			})
		}

		assert.ErrorIs(t, err, domain.ErrDailyCapacityExceeded)
	})
}

func TestCommitTaskHandler_Handle_Concurrency(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := testutils.GetGlobalPostgresPool(t)
	fixtures := testfixtures.NewFixtures(pool)

	uow := sharedPg.NewUnitOfWork(pool)
	scheduleRepo := schedulePg.NewScheduleRepo(pool, uow)
	todoRepo := todoPg.NewTodoRepo(pool, uow)

	handler := sharedApp.Retry(
		application.NewCommitTaskHandler(scheduleRepo, todoRepo, uow),
		10,
	)

	user := fixtures.RandomUser(ctx, t)
	ws := fixtures.RandomWorkspace(ctx, t, user.ID())
	userCtx := causation.WithMetadata(ctx, causation.Metadata{UserID: user.ID().UUID()})
	date := time.Now().Format(time.DateOnly)

	concurrency := 10
	errChan := make(chan error, concurrency)

	todos := make([]*todoDomain.Todo, concurrency)
	for i := range concurrency {
		todos[i] = fixtures.RandomTodo(ctx, t, ws.ID())
	}

	initialSchedule, _ := domain.NewDailySchedule(user.ID(), domain.ScheduleDate(date), 100)
	require.NoError(t, scheduleRepo.Save(ctx, initialSchedule))

	for i := range concurrency {
		go func(idx int) {
			_, err := handler.Handle(userCtx, application.CommitTaskCommand{
				TodoID: todos[idx].ID().UUID(),
				Cost:   1,
				Date:   date,
			})
			errChan <- err
		}(i)
	}

	for range concurrency {
		require.NoError(t, <-errChan)
	}

	updatedSchedule, err := scheduleRepo.FindByUserAndDate(ctx, user.ID(), domain.ScheduleDate(date))
	require.NoError(t, err)
	assert.Len(t, updatedSchedule.CommittedTasks(), concurrency)
	assert.Equal(t, concurrency+1, updatedSchedule.Version())
}
