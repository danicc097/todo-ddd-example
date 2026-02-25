package crypto_test

import (
	"testing"

	"github.com/alexedwards/argon2id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto"
)

func TestArgon2(t *testing.T) {
	t.Parallel()

	password := "my-secret-password"

	// keep small for fast tests
	params := argon2id.Params{
		Memory:      1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  8,
		KeyLength:   16,
	}

	t.Run("hash and compare success", func(t *testing.T) {
		hash, err := crypto.HashPassword(password, params)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Contains(t, hash, "$argon2id$")

		match, err := crypto.ComparePassword(password, hash)
		require.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("compare failure with wrong password", func(t *testing.T) {
		hash, err := crypto.HashPassword(password, params)
		require.NoError(t, err)

		match, err := crypto.ComparePassword("wrong-password", hash)
		require.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("compare error with invalid hash format", func(t *testing.T) {
		match, err := crypto.ComparePassword(password, "invalid-hash")
		require.ErrorIs(t, err, argon2id.ErrInvalidHash)
		assert.False(t, match)
	})
}
