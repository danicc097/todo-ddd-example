package crypto_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

func TestJWT(t *testing.T) {
	t.Parallel()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pubKey := &privKey.PublicKey

	issuer := "test-issuer"
	userID := uuid.New()

	ti := crypto.NewTokenIssuer(privKey, issuer)
	tv := crypto.NewTokenVerifier(pubKey)

	t.Run("issue and verify success", func(t *testing.T) {
		token, err := ti.Issue(userID, true, time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := tv.Verify(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.True(t, claims.MFAVerified)
		assert.Equal(t, issuer, claims.Issuer)
	})

	t.Run("verify failure with expired token", func(t *testing.T) {
		token, err := ti.Issue(userID, false, -time.Hour)
		require.NoError(t, err)

		_, err = tv.Verify(token)
		assert.Error(t, err)
	})

	t.Run("verify failure with wrong key", func(t *testing.T) {
		token, err := ti.Issue(userID, false, time.Hour)
		require.NoError(t, err)

		otherPriv, _ := rsa.GenerateKey(rand.Reader, 2048)
		otherTv := crypto.NewTokenVerifier(&otherPriv.PublicKey)

		_, err = otherTv.Verify(token)
		assert.Error(t, err)
	})
}
