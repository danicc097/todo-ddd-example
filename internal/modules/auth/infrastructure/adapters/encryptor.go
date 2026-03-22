package adapters

import (
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
)

type AESGCMEncryptor struct{}

var _ domain.Encryptor = (*AESGCMEncryptor)(nil)

func NewAESGCMEncryptor() *AESGCMEncryptor {
	return &AESGCMEncryptor{}
}

func (e *AESGCMEncryptor) Encrypt(plaintext []byte, masterKey []byte) ([]byte, []byte, error) {
	return crypto.Encrypt(plaintext, masterKey)
}

func (e *AESGCMEncryptor) Decrypt(ciphertext []byte, nonce []byte, masterKey []byte) ([]byte, error) {
	return crypto.Decrypt(ciphertext, nonce, masterKey)
}
