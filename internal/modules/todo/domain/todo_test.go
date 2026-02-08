package domain

import (
	"testing"
	"time"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTodo_Complete(t *testing.T) {
	title, _ := NewTodoTitle("Task")

	t.Run("should transition to completed from pending", func(t *testing.T) {
		todo := NewTodo(uuid.New(), title, StatusPending, time.Now())
		err := todo.Complete()
		assert.NoError(t, err)
		assert.Equal(t, StatusCompleted, todo.Status())
	})

	t.Run("should fail transition if archived", func(t *testing.T) {
		todo := NewTodo(uuid.New(), title, StatusArchived, time.Now())
		err := todo.Complete()
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})
}
