package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregateIntegrity(t *testing.T) {
	t.Parallel()

	t.Run("factory must emit creation event", func(t *testing.T) {
		title, _ := NewTodoTitle("New")
		todo := NewTodo(title)

		assert.Len(t, todo.Events(), 1)
		assert.IsType(t, TodoCreatedEvent{}, todo.Events()[0])
	})
}
