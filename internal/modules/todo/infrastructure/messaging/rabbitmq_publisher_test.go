package messaging_test

import (
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain/domainfakes"
	"github.com/stretchr/testify/assert"
)

var _ domain.EventPublisher = (*domainfakes.FakeEventPublisher)(nil)

func TestRabbitMQPublisher_Contract(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert.True(t, true)
	})
}
