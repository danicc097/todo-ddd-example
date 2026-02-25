package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

func TestDailySchedule_CommitTask(t *testing.T) {
	t.Parallel()

	userID := userDomain.UserID(uuid.New())
	date := NewScheduleDate(time.Now())

	t.Run("should commit task within capacity", func(t *testing.T) {
		s, _ := NewDailySchedule(userID, date, 10)
		todoID := todoDomain.TodoID(uuid.New())
		cost, _ := NewEnergyCost(3)

		err := s.CommitTask(todoID, cost)
		require.NoError(t, err)
		assert.Len(t, s.CommittedTasks(), 1)
		assert.Equal(t, cost, s.CommittedTasks()[todoID])
	})

	t.Run("should fail if capacity exceeded", func(t *testing.T) {
		s, _ := NewDailySchedule(userID, date, 5)
		todo1 := todoDomain.TodoID(uuid.New())
		todo2 := todoDomain.TodoID(uuid.New())
		cost, _ := NewEnergyCost(3)

		_ = s.CommitTask(todo1, cost)
		err := s.CommitTask(todo2, cost)

		assert.ErrorIs(t, err, ErrDailyCapacityExceeded)
	})
}
