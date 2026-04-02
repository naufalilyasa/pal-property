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
	"github.com/markbates/goth"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/mocks"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"github.com/redis/go-redis/v9"
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

	config.Env.JwtAccessExpiration = 900
	config.Env.JwtRefreshExpiration = 604800
	config.Env.OAuthTokenEncryptionKey = make([]byte, 32)
	config.Env.JwtRefreshExpiration = 604800

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
	assert.False(t, res.SellerCapabilities.CanAccessDashboard)
	assert.False(t, res.SellerCapabilities.RequiresOnboarding)
	mockRepo.AssertExpectations(t)
}

func TestGetMe_NotFound(t *testing.T) {
	mockRepo, _, svc := setupTest()

	userID, _ := uuid.NewV7()
	mockRepo.On("FindUserByID", mock.Anything, userID).Return((*entity.User)(nil), domain.ErrNotFound)

	res, err := svc.GetMe(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, domain.ErrNotFound)
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
	mockCache.On("SaveRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string"), userID, time.Duration(config.Env.JwtRefreshExpiration)*time.Second).Return(nil)

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
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestRefreshToken_Revoked(t *testing.T) {
	_, mockCache, svc := setupTest()

	userID, _ := uuid.NewV7()
	_, refToken, jti, err := jwt.GenerateTokens(userID)
	assert.NoError(t, err)

	// Mock token revoked (not in cache)
	mockCache.On("ValidateRefreshTokenJTI", mock.Anything, jti, userID).Return(redis.Nil)

	res, err := svc.RefreshToken(context.Background(), refToken)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_CacheInfraError(t *testing.T) {
	_, mockCache, svc := setupTest()

	userID, _ := uuid.NewV7()
	_, refToken, jti, err := jwt.GenerateTokens(userID)
	assert.NoError(t, err)

	infraErr := errors.New("redis timeout")
	mockCache.On("ValidateRefreshTokenJTI", mock.Anything, jti, userID).Return(infraErr)

	res, err := svc.RefreshToken(context.Background(), refToken)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.NotErrorIs(t, err, domain.ErrUnauthorized)
	assert.ErrorIs(t, err, infraErr)
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

func TestCompleteAuth_ExistingOAuthAccount(t *testing.T) {
	mockRepo, _, svc := setupTest()
	userID, _ := uuid.NewV7()
	gothUser := goth.User{UserID: "provider-uid-123", Email: "existing@example.com", Provider: "google"}

	account := &entity.OAuthAccount{UserID: userID}
	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "provider-uid-123").Return(account, nil)
	mockRepo.On("FindUserByID", mock.Anything, userID).Return(&entity.User{BaseEntity: entity.BaseEntity{ID: userID}, Email: "existing@example.com"}, nil)

	res, err := svc.CompleteAuth(context.Background(), "google", gothUser)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "existing@example.com", res.Email)
	mockRepo.AssertExpectations(t)
}

func TestCompleteAuth_OAuthAccountWithoutUser(t *testing.T) {
	mockRepo, _, svc := setupTest()
	gothUser := goth.User{UserID: "existing-oauth", Email: "missing@example.com", Provider: "google"}

	account := &entity.OAuthAccount{UserID: uuid.New()}
	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "existing-oauth").Return(account, nil)
	mockRepo.On("FindUserByID", mock.Anything, account.UserID).Return((*entity.User)(nil), domain.ErrNotFound)

	res, err := svc.CompleteAuth(context.Background(), "google", gothUser)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestCompleteAuth_NewUserBrandNew(t *testing.T) {
	mockRepo, _, svc := setupTest()
	newID, _ := uuid.NewV7()
	gothUser := goth.User{UserID: "new-uid", Email: "new@example.com", Name: "New User", Provider: "google"}

	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "new-uid").Return((*entity.OAuthAccount)(nil), domain.ErrNotFound)
	mockRepo.On("FindUserByEmail", mock.Anything, "new@example.com").Return((*entity.User)(nil), domain.ErrNotFound)
	mockRepo.On("CreateUserWithOAuth", mock.Anything, mock.AnythingOfType("*entity.User"), mock.AnythingOfType("*entity.OAuthAccount")).Return(&entity.User{BaseEntity: entity.BaseEntity{ID: newID}, Email: "new@example.com", Name: "New User"}, nil)

	res, err := svc.CompleteAuth(context.Background(), "google", gothUser)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "new@example.com", res.Email)
	mockRepo.AssertExpectations(t)
}

func TestCompleteAuth_FindUserByEmailError(t *testing.T) {
	mockRepo, _, svc := setupTest()
	gothUser := goth.User{UserID: "new-oauth-uid", Email: "existing@example.com", Provider: "google"}
	errDB := errors.New("failed to reach db")

	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "new-oauth-uid").Return((*entity.OAuthAccount)(nil), domain.ErrNotFound)
	mockRepo.On("FindUserByEmail", mock.Anything, "existing@example.com").Return((*entity.User)(nil), errDB)

	res, err := svc.CompleteAuth(context.Background(), "google", gothUser)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, errDB)
	mockRepo.AssertExpectations(t)
}

func TestCompleteAuth_EmailExistsNoOAuth(t *testing.T) {
	mockRepo, _, svc := setupTest()
	userID, _ := uuid.NewV7()
	gothUser := goth.User{UserID: "new-oauth-uid", Email: "existing@example.com", Provider: "google", AccessToken: "access-token", RefreshToken: "refresh-token"}

	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "new-oauth-uid").Return((*entity.OAuthAccount)(nil), domain.ErrNotFound)
	mockRepo.On("FindUserByEmail", mock.Anything, "existing@example.com").Return(&entity.User{BaseEntity: entity.BaseEntity{ID: userID}, Email: "existing@example.com"}, nil)
	mockRepo.On("CreateOAuthAccount", mock.Anything, mock.MatchedBy(func(account *entity.OAuthAccount) bool {
		if account == nil {
			return false
		}
		return account.UserID == userID &&
			account.Provider == "google" &&
			account.ProviderUserID == "new-oauth-uid" &&
			account.AccessToken != nil && *account.AccessToken == "access-token" &&
			account.RefreshToken != nil && *account.RefreshToken == "refresh-token"
	})).Return(nil)

	res, err := svc.CompleteAuth(context.Background(), "google", gothUser)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "existing@example.com", res.Email)
	mockRepo.AssertExpectations(t)
}

func TestCompleteAuth_DBError(t *testing.T) {
	mockRepo, _, svc := setupTest()
	gothUser := goth.User{UserID: "uid", Provider: "google"}

	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "uid").Return((*entity.OAuthAccount)(nil), errors.New("connection refused"))

	res, err := svc.CompleteAuth(context.Background(), "google", gothUser)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "connection refused")
	mockRepo.AssertExpectations(t)
}

func TestGetMe_MapsAllFields(t *testing.T) {
	mockRepo, _, svc := setupTest()
	userID, _ := uuid.NewV7()
	avatarURL := "https://example.com/avatar.png"
	expectedUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: userID},
		Email:      "test@example.com",
		Name:       "Test User",
		AvatarURL:  &avatarURL,
		Role:       "admin",
	}

	mockRepo.On("FindUserByID", mock.Anything, userID).Return(expectedUser, nil)

	res, err := svc.GetMe(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/avatar.png", *res.AvatarURL)
	assert.Equal(t, "admin", res.Role)
	assert.True(t, res.SellerCapabilities.CanAccessDashboard)
	assert.False(t, res.SellerCapabilities.RequiresOnboarding)
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_CacheSaveFailure(t *testing.T) {
	_, mockCache, svc := setupTest()
	userID, _ := uuid.NewV7()
	user := &entity.User{BaseEntity: entity.BaseEntity{ID: userID}, Email: "test@example.com"}

	mockCache.On("SaveRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string"), userID, mock.Anything).Return(errors.New("redis down"))

	res, err := svc.LoginUser(context.Background(), user)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "failed to cache refresh token")
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_DeleteJTIError(t *testing.T) {
	_, mockCache, svc := setupTest()
	userID, _ := uuid.NewV7()
	_, refToken, jti, err := jwt.GenerateTokens(userID)
	assert.NoError(t, err)

	mockCache.On("ValidateRefreshTokenJTI", mock.Anything, jti, userID).Return(nil)
	mockCache.On("DeleteRefreshTokenJTI", mock.Anything, jti).Return(errors.New("redis error"))
	mockCache.On("SaveRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string"), userID, mock.Anything).Return(nil)

	res, err := svc.RefreshToken(context.Background(), refToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_NewJTISaveFails(t *testing.T) {
	_, mockCache, svc := setupTest()
	userID, _ := uuid.NewV7()
	_, refToken, jti, err := jwt.GenerateTokens(userID)
	assert.NoError(t, err)

	mockCache.On("ValidateRefreshTokenJTI", mock.Anything, jti, userID).Return(nil)
	mockCache.On("DeleteRefreshTokenJTI", mock.Anything, jti).Return(nil)
	mockCache.On("SaveRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string"), userID, mock.Anything).Return(errors.New("redis save error"))

	_, err = svc.RefreshToken(context.Background(), refToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to cache new refresh token")
	assert.NotErrorIs(t, err, domain.ErrUnauthorized)
	mockCache.AssertExpectations(t)
}

func TestCompleteAuth_ExistingAccount(t *testing.T) {
	mockRepo, _, svc := setupTest()
	ctx := context.Background()
	userID := uuid.New()
	gothUser := goth.User{
		UserID:   "google-uid-123",
		Email:    "existing@example.com",
		Name:     "Existing User",
		Provider: "google",
	}
	existingUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: userID},
		Email:      "existing@example.com",
		Name:       "Existing User",
	}
	account := &entity.OAuthAccount{UserID: userID}
	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "google-uid-123").Return(account, nil)
	mockRepo.On("FindUserByID", mock.Anything, userID).Return(existingUser, nil)

	user, err := svc.CompleteAuth(ctx, "google", gothUser)

	assert.NoError(t, err)
	assert.Equal(t, "existing@example.com", user.Email)
	mockRepo.AssertExpectations(t)
}

func TestCompleteAuth_BrandNewUser(t *testing.T) {
	mockRepo, _, svc := setupTest()
	ctx := context.Background()
	newID := uuid.New()
	gothUser := goth.User{
		UserID:       "google-uid-new",
		Email:        "new@example.com",
		Name:         "New User",
		Provider:     "google",
		AvatarURL:    "https://example.com/avatar.png",
		AccessToken:  "goog-access",
		RefreshToken: "goog-refresh",
	}
	createdUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: newID},
		Email:      "new@example.com",
		Name:       "New User",
	}
	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "google-uid-new").Return(nil, domain.ErrNotFound)
	mockRepo.On("FindUserByEmail", mock.Anything, "new@example.com").Return(nil, domain.ErrNotFound)
	mockRepo.On("CreateUserWithOAuth", mock.Anything, mock.AnythingOfType("*entity.User"), mock.AnythingOfType("*entity.OAuthAccount")).Return(createdUser, nil)

	user, err := svc.CompleteAuth(ctx, "google", gothUser)

	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", user.Email)
	mockRepo.AssertExpectations(t)
}

func TestCompleteAuth_CreateUserFails(t *testing.T) {
	mockRepo, _, svc := setupTest()
	ctx := context.Background()
	gothUser := goth.User{
		UserID:   "google-uid-fail",
		Email:    "fail@example.com",
		Provider: "google",
	}
	mockRepo.On("FindOAuthAccount", mock.Anything, "google", "google-uid-fail").Return(nil, domain.ErrNotFound)
	mockRepo.On("FindUserByEmail", mock.Anything, "fail@example.com").Return(nil, domain.ErrNotFound)
	mockRepo.On("CreateUserWithOAuth", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("unique constraint violation"))

	user, err := svc.CompleteAuth(ctx, "google", gothUser)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_CacheSaveFails(t *testing.T) {
	_, mockCache, svc := setupTest()
	ctx := context.Background()
	user := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New()},
		Email:      "user@example.com",
	}
	mockCache.On("SaveRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("uuid.UUID"), mock.Anything).Return(errors.New("redis down"))

	tokens, err := svc.LoginUser(ctx, user)

	assert.Error(t, err)
	assert.Nil(t, tokens)
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_DeleteJTIErrorIsIgnored(t *testing.T) {
	_, mockCache, svc := setupTest()
	ctx := context.Background()
	userID := uuid.New()
	_, refreshToken, _, err := jwt.GenerateTokens(userID)
	if err != nil {
		t.Fatal(err)
	}
	mockCache.On("ValidateRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string"), userID).Return(nil)
	mockCache.On("DeleteRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("redis error"))
	mockCache.On("SaveRefreshTokenJTI", mock.Anything, mock.AnythingOfType("string"), userID, mock.Anything).Return(nil)

	tokens, err := svc.RefreshToken(ctx, refreshToken)

	assert.NoError(t, err, "delete JTI error should be silently ignored")
	assert.NotEmpty(t, tokens.AccessToken)
	mockCache.AssertExpectations(t)
}
