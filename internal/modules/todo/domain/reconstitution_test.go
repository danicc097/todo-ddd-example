package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

func TestAggregateIntegrity(t *testing.T) {
	t.Parallel()

	t.Run("factory must emit creation event", func(t *testing.T) {
		title, _ := NewTodoTitle("New")
		todo := NewTodo(title, wsDomain.WorkspaceID(uuid.New()))

		assert.Len(t, todo.Events(), 1)
		evt := todo.Events()[0].(TodoCreatedEvent)
		assert.WithinDuration(t, time.Now(), evt.OccurredAt(), time.Second)
		assert.WithinDuration(t, time.Now(), todo.CreatedAt(), time.Second)
	})
}
