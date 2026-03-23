package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"github.com/redis/go-redis/v9"
)

type AuthService interface {
	CompleteAuth(ctx context.Context, provider string, gothUser goth.User) (*entity.User, error)
	LoginUser(ctx context.Context, user *entity.User) (*response.AuthTokens, error)
	GetMe(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*response.AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	repo  domain.AuthRepository
	cache domain.CacheRepository
}

func NewAuthService(repo domain.AuthRepository, cache domain.CacheRepository) AuthService {
	return &authService{repo: repo, cache: cache}
}

func (s *authService) CompleteAuth(ctx context.Context, provider string, gothUser goth.User) (*entity.User, error) {

	// 1. Check if OAuthAccount exists
	// Convert ProviderUserID string to UUID if needed, but in entity it is varchar(255)
	oauthAccount, err := s.repo.FindOAuthAccount(ctx, provider, gothUser.UserID)
	if err == nil {
		user, err := s.repo.FindUserByID(ctx, oauthAccount.UserID)
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err // DB Error
	}

	// 2. Check if User exists by email
	user, err := s.repo.FindUserByEmail(ctx, gothUser.Email)
	if err == nil {
		account := &entity.OAuthAccount{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: gothUser.UserID,
			AccessToken:    &gothUser.AccessToken,
			RefreshToken:   &gothUser.RefreshToken,
		}
		if err := s.repo.CreateOAuthAccount(ctx, account); err != nil {
			return nil, err
		}
		return user, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	// 3. User does not exist, create new
	userID, _ := uuid.NewV7()
	newUser := &entity.User{
		BaseEntity: entity.BaseEntity{
			ID: userID,
		},
		Name:       gothUser.Name,
		Email:      gothUser.Email,
		AvatarURL:  &gothUser.AvatarURL,
		Role:       "user",
		IsVerified: true, // OAuth usually verified
	}

	accountID, _ := uuid.NewV7()
	newAccount := &entity.OAuthAccount{
		ID:             accountID,
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: gothUser.UserID,
		AccessToken:    &gothUser.AccessToken,
		RefreshToken:   &gothUser.RefreshToken,
	}

	createdUser, err := s.repo.CreateUserWithOAuth(ctx, newUser, newAccount)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *authService) LoginUser(ctx context.Context, user *entity.User) (*response.AuthTokens, error) {
	accToken, refToken, jti, err := jwt.GenerateTokens(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Save jti to Redis
	err = s.cache.SaveRefreshTokenJTI(ctx, jti, user.ID, time.Duration(config.Env.JwtRefreshExpiration)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to cache refresh token: %w", err)
	}

	return &response.AuthTokens{
		AccessToken:  accToken,
		RefreshToken: refToken,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &response.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*response.AuthTokens, error) {
	// 1. Validate refresh token structure and signature
	userID, jti, err := jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", domain.ErrUnauthorized)
	}

	// 2. Validate JTI against Redis cache (check if revoked/expired)
	err = s.cache.ValidateRefreshTokenJTI(ctx, jti, userID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("refresh token session expired or revoked: %w", domain.ErrUnauthorized)
		}
		return nil, fmt.Errorf("failed to validate refresh token session: %w", err)
	}

	// 3. Delete old JTI from Redis (Refresh token rotation)
	_ = s.cache.DeleteRefreshTokenJTI(ctx, jti)

	// 4. Generate new tokens
	accToken, newRefToken, newJTI, err := jwt.GenerateTokens(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// 5. Save new JTI to Redis
	err = s.cache.SaveRefreshTokenJTI(ctx, newJTI, userID, time.Duration(config.Env.JwtRefreshExpiration)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to cache new refresh token: %w", err)
	}

	return &response.AuthTokens{
		AccessToken:  accToken,
		RefreshToken: newRefToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	// We only need the jti from the token to invalidate it
	_, jti, err := jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		// Even if token is invalid, we proceed so the clear cookie runs on handler side
		return nil
	}

	// Delete from Redis
	return s.cache.DeleteRefreshTokenJTI(ctx, jti)
}
