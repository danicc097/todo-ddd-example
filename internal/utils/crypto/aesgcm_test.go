package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

func TestAESGCM(t *testing.T) {
	t.Parallel()

	masterKey := []byte("01234567890123456789012345678901") // 32 bytes
	payload := []byte("hello world")

	t.Run("encrypt and decrypt success", func(t *testing.T) {
		t.Parallel()

		ciphertext, nonce, err := crypto.Encrypt(payload, masterKey)
		require.NoError(t, err)
		assert.NotEqual(t, payload, ciphertext)

		decrypted, err := crypto.Decrypt(ciphertext, nonce, masterKey)
		require.NoError(t, err)
		assert.Equal(t, payload, decrypted)
	})

	t.Run("encrypt same payload twice results in different ciphertexts", func(t *testing.T) {
		t.Parallel()

		c1, n1, err := crypto.Encrypt(payload, masterKey)
		require.NoError(t, err)
		c2, n2, err := crypto.Encrypt(payload, masterKey)
		require.NoError(t, err)

		assert.NotEqual(t, n1, n2)
		assert.NotEqual(t, c1, c2)
	})

	t.Run("decrypt fails with wrong key", func(t *testing.T) {
		t.Parallel()

		ciphertext, nonce, err := crypto.Encrypt(payload, masterKey)
		require.NoError(t, err)

		wrongKey := []byte("wrong-key-0123456789012345678901")
		_, err = crypto.Decrypt(ciphertext, nonce, wrongKey)
		assert.Error(t, err)
	})

	t.Run("encrypt fails with invalid key length", func(t *testing.T) {
		t.Parallel()

		shortKey := []byte("short")
		_, _, err := crypto.Encrypt(payload, shortKey)
		assert.ErrorIs(t, err, crypto.ErrInvalidKey)
	})
}
