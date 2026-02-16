package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var ErrInvalidKey = errors.New("encryption key must be exactly 32 bytes for AES-256")

// Encrypt payload using AES-256-GCM. Returns the ciphertext and nonce.
func Encrypt(payload []byte, masterKey []byte) (ciphertext []byte, nonce []byte, err error) {
	if len(masterKey) != 32 {
		return nil, nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, payload, nil)

	return ciphertext, nonce, nil
}

// Decrypt ciphertext using AES-256-GCM and the provided nonce.
func Decrypt(ciphertext []byte, nonce []byte, masterKey []byte) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}
