package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

var ErrInvalidKey = fmt.Errorf("encryption key must be exactly %d bytes for AES-256", MasterKeyLength)

const MasterKeyLength = 32

// Encrypt payload using AES-256-GCM. Returns the ciphertext and nonce.
func Encrypt(payload []byte, masterKey []byte) (ciphertext []byte, nonce []byte, err error) {
	if len(masterKey) != MasterKeyLength {
		return nil, nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, nil, fmt.Errorf("aes.NewCipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("cipher.NewGCM failed: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("io.ReadFull failed: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, payload, nil)

	return ciphertext, nonce, nil
}

// Decrypt ciphertext using AES-256-GCM and the provided nonce.
func Decrypt(ciphertext []byte, nonce []byte, masterKey []byte) ([]byte, error) {
	if len(masterKey) != MasterKeyLength {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM failed: %w", err)
	}

	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("gcm.Open failed: %w", err)
	}

	return decrypted, nil
}
