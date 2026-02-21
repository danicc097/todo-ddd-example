package crypto

import (
	authDomain "github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

type Argon2PasswordHasher struct{}

func NewArgon2PasswordHasher() *Argon2PasswordHasher {
	return &Argon2PasswordHasher{}
}

func (h *Argon2PasswordHasher) Hash(password string) (string, error) {
	return crypto.HashPassword(password, crypto.DefaultArgon2Params)
}

func (h *Argon2PasswordHasher) Compare(password, hash string) (bool, error) {
	return crypto.ComparePassword(password, hash)
}

var _ authDomain.PasswordHasher = (*Argon2PasswordHasher)(nil)
