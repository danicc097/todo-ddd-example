package crypto

import (
	authDomain "github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
)

type Argon2PasswordHasher struct{}

func NewArgon2PasswordHasher() *Argon2PasswordHasher {
	return &Argon2PasswordHasher{}
}

func (h *Argon2PasswordHasher) Hash(password string) (string, error) {
	return HashPassword(password, DefaultArgon2Params)
}

func (h *Argon2PasswordHasher) Compare(password, hash string) (bool, error) {
	return ComparePassword(password, hash)
}

var _ authDomain.PasswordHasher = (*Argon2PasswordHasher)(nil)
