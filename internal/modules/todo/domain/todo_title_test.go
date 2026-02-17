package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTodoTitle(t *testing.T) {
	t.Parallel()

	t.Run("should create valid title", func(t *testing.T) {
		title, err := NewTodoTitle("Valid Task")
		assert.NoError(t, err)
		assert.Equal(t, "Valid Task", title.String())
	})

	t.Run("should trim whitespace", func(t *testing.T) {
		title, err := NewTodoTitle("  Trim Me  ")
		assert.NoError(t, err)
		assert.Equal(t, "Trim Me", title.String())
	})

	t.Run("should fail with empty title", func(t *testing.T) {
		_, err := NewTodoTitle("")
		assert.ErrorIs(t, err, ErrTitleEmpty)
	})

	t.Run("should fail with only whitespace", func(t *testing.T) {
		_, err := NewTodoTitle("   ")
		assert.ErrorIs(t, err, ErrTitleEmpty)
	})

	t.Run("should fail if title too long", func(t *testing.T) {
		longTitle := strings.Repeat("a", 101)
		_, err := NewTodoTitle(longTitle)
		assert.ErrorIs(t, err, ErrTitleTooLong)
	})
}
