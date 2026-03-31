package postgres

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/crypto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestAuthRepository_RejectsMalformedCiphertextPlaintextFallback(t *testing.T) {
	setTestEncryptionKey(t)
	encrypted, err := crypto.Encrypt("value", config.Env.OAuthTokenEncryptionKey)
	if err != nil {
		t.Fatalf("failed to encrypt token: %v", err)
	}

	decoded, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("failed to decode encrypted token: %v", err)
	}
	short := decoded[:20]
	short[0] = 0
	invalid := base64.URLEncoding.EncodeToString(short)

	result, err := decryptOAuthToken(invalid, "refresh token")
	if err == nil {
		t.Fatalf("expected error for malformed ciphertext, got %v", result)
	}
	if !strings.Contains(err.Error(), "failed to decrypt refresh token") {
		t.Fatalf("unexpected error for malformed ciphertext: %v", err)
	}
}

func TestAuthRepository_RejectsNonPlaintextGarbage(t *testing.T) {
	setTestEncryptionKey(t)
	garbage := "not plaintext"
	result, err := decryptOAuthToken(garbage, "access token")
	if err == nil {
		t.Fatalf("expected error for non-plaintext fallback, got %v", result)
	}
	if !strings.Contains(err.Error(), "failed to decrypt access token") {
		t.Fatalf("unexpected error for non-plaintext garbage: %v", err)
	}
}

func TestAuthRepository_FindOAuthAccount_IgnoresUndecryptableTokens(t *testing.T) {
	setTestEncryptionKey(t)
	originalKey := append([]byte(nil), config.Env.OAuthTokenEncryptionKey...)
	t.Cleanup(func() {
		config.Env.OAuthTokenEncryptionKey = originalKey
	})

	db := newInMemoryDB(t)
	repo := NewAuthRepository(db)

	userID := uuid.New()
	user := &entity.User{
		BaseEntity: entity.BaseEntity{ID: userID},
		Name:       "Test User",
		Email:      "test@example.com",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	plaintext := "stale-token"
	encrypted, err := crypto.Encrypt(plaintext, config.Env.OAuthTokenEncryptionKey)
	if err != nil {
		t.Fatalf("failed to encrypt token: %v", err)
	}

	account := &entity.OAuthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       "google",
		ProviderUserID: "provider-uid",
		AccessToken:    &encrypted,
		RefreshToken:   &encrypted,
	}
	if err := db.Create(account).Error; err != nil {
		t.Fatalf("failed to create oauth account: %v", err)
	}

	rotatedKey := append([]byte(nil), config.Env.OAuthTokenEncryptionKey...)
	rotatedKey[0] ^= 0xff
	config.Env.OAuthTokenEncryptionKey = rotatedKey

	found, err := repo.FindOAuthAccount(context.Background(), "google", "provider-uid")
	if err != nil {
		t.Fatalf("expected oauth account even when tokens stale: %v", err)
	}
	if found == nil {
		t.Fatalf("expected oauth account, got nil")
	}
	if found.UserID != userID {
		t.Fatalf("expected user id %s, got %s", userID, found.UserID)
	}
	if found.AccessToken != nil {
		t.Fatalf("expected access token to be nil, got %v", *found.AccessToken)
	}
	if found.RefreshToken != nil {
		t.Fatalf("expected refresh token to be nil, got %v", *found.RefreshToken)
	}
}

func newInMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	if err := db.AutoMigrate(&entity.User{}, &entity.OAuthAccount{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}
	return db
}
