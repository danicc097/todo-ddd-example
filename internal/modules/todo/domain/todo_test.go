package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

func TestTodo_Complete(t *testing.T) {
	t.Parallel()

	title, _ := NewTodoTitle("Task")
	wsID := wsDomain.WorkspaceID(uuid.New())
	actorID := userDomain.UserID(uuid.New())
	now := time.Now()

	t.Run("should transition to completed from pending", func(t *testing.T) {
		todo := NewTodo(title, wsID)
		err := todo.Complete(actorID, now)
		assert.NoError(t, err)
		assert.Equal(t, StatusCompleted, todo.Status())
	})

	t.Run("should fail transition if archived", func(t *testing.T) {
		todo := ReconstituteTodo(ReconstituteTodoArgs{
			ID:              TodoID(uuid.New()),
			Title:           title,
			Status:          StatusArchived,
			CreatedAt:       time.Now(),
			Tags:            nil,
			WorkspaceID:     wsID,
			DueDate:         nil,
			Recurrence:      nil,
			LastCompletedAt: nil,
			Sessions:        nil,
		})
		err := todo.Complete(actorID, now)
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("should rollover if recurrence is set", func(t *testing.T) {
		todo := NewTodo(title, wsID)
		rule, _ := NewRecurrenceRule("DAILY", 1)
		todo.SetRecurrence(&rule)

		dueDate := now.AddDate(0, 0, -1) // yesterday
		todo.SetDueDate(&dueDate)

		require.NoError(t, todo.Complete(actorID, now))
		assert.Equal(t, StatusPending, todo.Status())
		assert.Equal(t, now.Truncate(time.Minute), todo.DueDate().Truncate(time.Minute))
		assert.NotNil(t, todo.LastCompletedAt())
	})
}

func TestTodo_Focus(t *testing.T) {
	t.Parallel()

	title, _ := NewTodoTitle("Task")
	wsID := wsDomain.WorkspaceID(uuid.New())
	userID := userDomain.UserID(uuid.New())
	now := time.Now()

	t.Run("should start and stop focus", func(t *testing.T) {
		todo := NewTodo(title, wsID)
		sessionID := FocusSessionID(uuid.New())

		err := todo.StartFocus(userID, sessionID)
		assert.NoError(t, err)
		assert.NotNil(t, todo.ActiveFocusSession())

		err = todo.StopFocus(now.Add(time.Hour))
		assert.NoError(t, err)
		assert.Nil(t, todo.ActiveFocusSession())
		assert.Len(t, todo.Sessions(), 1)
		assert.NotNil(t, todo.Sessions()[0].EndTime())
	})

	t.Run("should fail if already focusing", func(t *testing.T) {
		todo := NewTodo(title, wsID)
		sessionID1 := FocusSessionID(uuid.New())
		_ = todo.StartFocus(userID, sessionID1)

		sessionID2 := FocusSessionID(uuid.New())
		err := todo.StartFocus(userID, sessionID2)
		assert.ErrorIs(t, err, ErrFocusSessionAlreadyActive)
	})
}
