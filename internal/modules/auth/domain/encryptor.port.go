package domain

// Encryptor defines how secrets are encrypted and decrypted.
type Encryptor interface {
	Encrypt(plaintext []byte, masterKey []byte) (ciphertext []byte, nonce []byte, err error)
	Decrypt(ciphertext []byte, nonce []byte, masterKey []byte) ([]byte, error)
}
