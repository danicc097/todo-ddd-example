package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
)

type testStruct struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestMsgpackCodec_Symmetry(t *testing.T) {
	t.Parallel()

	codec := cache.NewMsgpackCodec[testStruct]()

	t.Run("encodes and decodes correctly", func(t *testing.T) {
		input := testStruct{ID: 1, Name: "test"}

		data, err := codec.Marshal(input)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		output, err := codec.Unmarshal(data)
		require.NoError(t, err)
		assert.Equal(t, input, output)
	})
}

func TestCollectionCodec_Symmetry(t *testing.T) {
	t.Parallel()

	codec := cache.NewCollectionCodec[testStruct]()

	t.Run("encodes and decodes slices correctly", func(t *testing.T) {
		input := []testStruct{
			{ID: 1, Name: "a"},
			{ID: 2, Name: "b"},
		}

		data, err := codec.Marshal(input)
		require.NoError(t, err)

		output, err := codec.Unmarshal(data)
		require.NoError(t, err)
		assert.Equal(t, input, output)
	})
}
