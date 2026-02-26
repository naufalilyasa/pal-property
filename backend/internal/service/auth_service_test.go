package service_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/mocks"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTest() (*mocks.AuthRepository, *mocks.CacheRepository, service.AuthService) {
	// Setup RSA Keys for JWT Utility
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	config.Env.JwtPrivateKeyBase64 = base64.StdEncoding.EncodeToString(privPEM)

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	config.Env.JwtPublicKeyBase64 = base64.StdEncoding.EncodeToString(pubPEM)

	config.Env.JwtAccessExpiration = time.Minute * 15
	config.Env.JwtRefreshExpiration = time.Hour * 168

	mockAuthRepo := new(mocks.AuthRepository)
	mockCacheRepo := new(mocks.CacheRepository)
	svc := service.NewAuthService(mockAuthRepo, mockCacheRepo)

	return mockAuthRepo, mockCacheRepo, svc
}

func TestGetMe_Success(t *testing.T) {
	mockRepo, _, svc := setupTest()

	userID, _ := uuid.NewV7()
	expectedUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: userID},
		Email:      "test@example.com",
		Name:       "Test User",
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(expectedUser, nil)

	res, err := svc.GetMe(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, userID, res.ID)
	assert.Equal(t, "test@example.com", res.Email)
	mockRepo.AssertExpectations(t)
}

func TestGetMe_NotFound(t *testing.T) {
	mockRepo, _, svc := setupTest()

	userID, _ := uuid.NewV7()
	mockRepo.On("FindUserByID", mock.Anything, userID).Return(&entity.User{}, errors.New("db error"))

	res, err := svc.GetMe(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_Success(t *testing.T) {
	_, mockCache, svc := setupTest()

	userID, _ := uuid.NewV7()
	_, refToken, jti, err := jwt.GenerateTokens(userID)
	assert.NoError(t, err)

	// Mock validating old JTI
	mockCache.On("ValidateRefreshTokenJTI", mock.Anything, jti, userID).Return(nil)
	// Mock deleting old JTI
	mockCache.On("DeleteRefreshTokenJTI", mock.Anything, jti).Return(nil)
	// Mock saving new JTI
	mockCache.On("SaveRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string"), userID, config.Env.JwtRefreshExpiration).Return(nil)

	res, err := svc.RefreshToken(context.Background(), refToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	_, _, svc := setupTest()
	res, err := svc.RefreshToken(context.Background(), "invalid-token")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

func TestRefreshToken_Revoked(t *testing.T) {
	_, mockCache, svc := setupTest()

	userID, _ := uuid.NewV7()
	_, refToken, jti, err := jwt.GenerateTokens(userID)
	assert.NoError(t, err)

	// Mock token revoked (not in cache)
	mockCache.On("ValidateRefreshTokenJTI", mock.Anything, jti, userID).Return(errors.New("redis nil"))

	res, err := svc.RefreshToken(context.Background(), refToken)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "refresh token session expired or revoked")
	mockCache.AssertExpectations(t)
}

func TestLogout_Success(t *testing.T) {
	_, mockCache, svc := setupTest()

	userID, _ := uuid.NewV7()
	_, refToken, jti, err := jwt.GenerateTokens(userID)
	assert.NoError(t, err)

	mockCache.On("DeleteRefreshTokenJTI", mock.Anything, jti).Return(nil)

	err = svc.Logout(context.Background(), refToken)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestLogout_InvalidTokenIgnored(t *testing.T) {
	_, mockCache, svc := setupTest()

	// Should not hit cache because token parsing validates fail but we return nil so we can clear cookies.
	err := svc.Logout(context.Background(), "invalid-token")

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}
