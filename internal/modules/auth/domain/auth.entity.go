package domain

import (
	"errors"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrInvalidOTP         = shared.NewDomainError(apperrors.Unauthorized, "invalid OTP code")
	ErrAuthNotFound       = shared.NewDomainError(apperrors.NotFound, "auth record not found")
	ErrInvalidCredentials = shared.NewDomainError(apperrors.Unauthorized, "invalid email or password")
)

const (
	TOTPDisabled = "DISABLED"
	TOTPPending  = "PENDING"
	TOTPActive   = "ACTIVE"
)

type UserAuth struct {
	shared.AggregateRoot

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

type ReconstituteUserAuthArgs struct {
	ID           userDomain.UserID
	Status       string
	Cipher       []byte
	Nonce        []byte
	PasswordHash string
}

func ReconstituteUserAuth(args ReconstituteUserAuthArgs) *UserAuth {
	return &UserAuth{
		userID:           args.ID,
		totpStatus:       args.Status,
		totpSecretCipher: args.Cipher,
		totpSecretNonce:  args.Nonce,
		passwordHash:     args.PasswordHash,
	}
}

func (a *UserAuth) UserID() userDomain.UserID         { return a.userID }
func (a *UserAuth) PasswordHash() string              { return a.passwordHash }
func (a *UserAuth) IsTOTPActive() bool                { return a.totpStatus == TOTPActive }
func (a *UserAuth) IsTOTPPending() bool               { return a.totpStatus == TOTPPending }
func (a *UserAuth) TOTPCredentials() ([]byte, []byte) { return a.totpSecretCipher, a.totpSecretNonce }
func (a *UserAuth) TOTPStatus() string                { return a.totpStatus }

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
