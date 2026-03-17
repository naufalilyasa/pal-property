package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt encrypts plaintext using AES-256-GCM.
// key MUST be exactly 32 bytes (256-bit). Pass via config (OAUTH_TOKEN_ENCRYPTION_KEY env var).
// Returns a base64url-encoded string: nonce (12 bytes) + ciphertext + auth tag.
func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("crypto: key must be 32 bytes for AES-256")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64url-encoded AES-256-GCM ciphertext produced by Encrypt.
func Decrypt(encoded string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("crypto: key must be 32 bytes for AES-256")
	}
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("crypto: ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
