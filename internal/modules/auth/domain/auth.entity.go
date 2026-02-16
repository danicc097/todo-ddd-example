package domain

import (
	"errors"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

var (
	ErrInvalidOTP         = errors.New("invalid OTP code")
	ErrAuthNotFound       = errors.New("auth record not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

const (
	TOTPDisabled = "DISABLED"
	TOTPPending  = "PENDING"
	TOTPActive   = "ACTIVE"
)

type UserAuth struct {
	userID           userDomain.UserID
	totpStatus       string
	totpSecretCipher []byte
	totpSecretNonce  []byte
	passwordHash     string
}

func NewUserAuth(userID userDomain.UserID, passwordHash string) *UserAuth {
	return &UserAuth{
		userID:       userID,
		totpStatus:   TOTPDisabled,
		passwordHash: passwordHash,
	}
}

func ReconstituteUserAuth(id userDomain.UserID, status string, cipher, nonce []byte, passwordHash string) *UserAuth {
	return &UserAuth{
		userID:           id,
		totpStatus:       status,
		totpSecretCipher: cipher,
		totpSecretNonce:  nonce,
		passwordHash:     passwordHash,
	}
}

func (a *UserAuth) UserID() userDomain.UserID { return a.userID }
func (a *UserAuth) PasswordHash() string      { return a.passwordHash }
func (a *UserAuth) IsTOTPActive() bool        { return a.totpStatus == TOTPActive }
func (a *UserAuth) IsTOTPPending() bool       { return a.totpStatus == TOTPPending }
func (a *UserAuth) TOTPCredentials() ([]byte, []byte) { return a.totpSecretCipher, a.totpSecretNonce }
func (a *UserAuth) TOTPStatus() string        { return a.totpStatus }

func (a *UserAuth) InitiateTOTP(cipher, nonce []byte) {
	a.totpStatus = TOTPPending
	a.totpSecretCipher = cipher
	a.totpSecretNonce = nonce
}

func (a *UserAuth) ActivateTOTP() error {
	if a.totpStatus != TOTPPending {
		return errors.New("cannot activate TOTP: not pending")
	}
	a.totpStatus = TOTPActive
	return nil
}
