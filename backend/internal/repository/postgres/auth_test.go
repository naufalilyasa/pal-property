package postgres

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/crypto"
)

var testTokenEncryptionKey = []byte("01234567890123456789012345678901")

func setTestEncryptionKey(t *testing.T) {
	t.Helper()
	key := make([]byte, len(testTokenEncryptionKey))
	copy(key, testTokenEncryptionKey)
	config.Env.OAuthTokenEncryptionKey = key
}

func TestAuthRepository_DecryptEncrypted(t *testing.T) {
	setTestEncryptionKey(t)
	plaintext := "plaintext-token"
	encrypted, err := crypto.Encrypt(plaintext, config.Env.OAuthTokenEncryptionKey)
	if err != nil {
		t.Fatalf("failed to encrypt token: %v", err)
	}

	result, err := decryptOAuthToken(encrypted, "access token")
	if err != nil {
		t.Fatalf("unexpected error decrypting encrypted token: %v", err)
	}
	if result == nil || *result != plaintext {
		t.Fatalf("expected decrypted token %q, got %v", plaintext, result)
	}
}

func TestAuthRepository_DecryptPlaintextLegacy(t *testing.T) {
	setTestEncryptionKey(t)
	plaintext := "legacy-plaintext-token"

	result, err := decryptOAuthToken(plaintext, "access token")
	if err != nil {
		t.Fatalf("unexpected error decrypting legacy plaintext: %v", err)
	}
	if result == nil || *result != plaintext {
		t.Fatalf("expected plaintext token %q, got %v", plaintext, result)
	}
}

func TestAuthRepository_DecryptBase64LikePlaintextLegacy(t *testing.T) {
	setTestEncryptionKey(t)
	plaintext := "c29tZS10b2tlbg=="

	result, err := decryptOAuthToken(plaintext, "access token")
	if err != nil {
		t.Fatalf("unexpected error decrypting base64-like plaintext: %v", err)
	}
	if result == nil || *result != plaintext {
		t.Fatalf("expected plaintext token %q, got %v", plaintext, result)
	}
}

func TestAuthRepository_DecryptInvalidCiphertext(t *testing.T) {
	setTestEncryptionKey(t)
	encrypted, err := crypto.Encrypt("value", config.Env.OAuthTokenEncryptionKey)
	if err != nil {
		t.Fatalf("failed to encrypt token: %v", err)
	}

	decoded, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("failed to decode encrypted token: %v", err)
	}
	decoded[len(decoded)-1] ^= 0xff
	invalid := base64.URLEncoding.EncodeToString(decoded)

	result, err := decryptOAuthToken(invalid, "refresh token")
	if err == nil {
		t.Fatalf("expected error for invalid ciphertext, got %v", result)
	}
	if !strings.Contains(err.Error(), "failed to decrypt refresh token") {
		t.Fatalf("unexpected error for tampered ciphertext: %v", err)
	}
}
