package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

func TestTodo_Complete(t *testing.T) {
	title, _ := NewTodoTitle("Task")
	wsID := wsDomain.WorkspaceID{UUID: uuid.New()}

	t.Run("should transition to completed from pending", func(t *testing.T) {
		todo := NewTodo(title, wsID)
		err := todo.Complete()
		assert.NoError(t, err)
		assert.Equal(t, StatusCompleted, todo.Status())
	})

	t.Run("should fail transition if archived", func(t *testing.T) {
		todo := ReconstituteTodo(TodoID{UUID: uuid.New()}, title, StatusArchived, time.Now(), nil, wsID)
		err := todo.Complete()
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})
}
