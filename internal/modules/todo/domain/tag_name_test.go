package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTagName(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		name, err := NewTagName("  work  ")
		require.NoError(t, err)
		assert.Equal(t, "work", name.String())
	})

	t.Run("empty", func(t *testing.T) {
		_, err := NewTagName("   ")
		assert.ErrorIs(t, err, ErrTagNameEmpty)
	})

	t.Run("too long", func(t *testing.T) {
		longName := strings.Repeat("a", tagMaxLen+1)
		_, err := NewTagName(longName)
		assert.ErrorIs(t, err, ErrTagNameTooLong)
	})

	t.Run("max chars", func(t *testing.T) {
		nameStr := strings.Repeat("a", tagMaxLen)
		name, err := NewTagName(nameStr)
		require.NoError(t, err)
		assert.Equal(t, nameStr, name.String())
	})
}
