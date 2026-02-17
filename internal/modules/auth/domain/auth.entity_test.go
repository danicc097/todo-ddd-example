package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

func TestUserAuth_TOTP(t *testing.T) {
	t.Parallel()

	userID := userDomain.UserID{UUID: uuid.New()}
	auth := domain.NewUserAuth(userID, "hash")

	t.Run("initiate totp", func(t *testing.T) {
		cipher := []byte("cipher")
		nonce := []byte("nonce")

		auth.InitiateTOTP(cipher, nonce)

		assert.Equal(t, domain.TOTPPending, auth.TOTPStatus())
		assert.True(t, auth.IsTOTPPending())
		c, n := auth.TOTPCredentials()
		assert.Equal(t, cipher, c)
		assert.Equal(t, nonce, n)
	})

	t.Run("activate totp - success", func(t *testing.T) {
		userID := userDomain.UserID{UUID: uuid.New()}
		auth := domain.NewUserAuth(userID, "hash")
		auth.InitiateTOTP([]byte("c"), []byte("n"))

		err := auth.ActivateTOTP()
		require.NoError(t, err)
		assert.Equal(t, domain.TOTPActive, auth.TOTPStatus())
		assert.True(t, auth.IsTOTPActive())
	})

	t.Run("activate totp - failure", func(t *testing.T) {
		userID := userDomain.UserID{UUID: uuid.New()}
		auth := domain.NewUserAuth(userID, "hash")

		// status is DISABLED
		err := auth.ActivateTOTP()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not pending")

		// status is already ACTIVE
		auth.InitiateTOTP([]byte("c"), []byte("n"))
		_ = auth.ActivateTOTP()
		err = auth.ActivateTOTP()
		assert.Error(t, err)
	})
}
