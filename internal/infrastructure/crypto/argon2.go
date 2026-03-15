package crypto

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

// Minimum params as per https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
var DefaultArgon2Params = argon2id.Params{
	Memory:      9 * 1024,
	Iterations:  4,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

// HashPassword hashes a password using Argon2id.
func HashPassword(password string, p argon2id.Params) (string, error) {
	hash, err := argon2id.CreateHash(password, &p)
	if err != nil {
		return "", fmt.Errorf("argon2id.CreateHash: %w", err)
	}

	return hash, nil
}

// ComparePassword verifies a password against an Argon2id hash.
func ComparePassword(password, encodedHash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, encodedHash)
	if err != nil {
		return false, fmt.Errorf("argon2id.ComparePasswordAndHash: %w", err)
	}

	return match, nil
}
