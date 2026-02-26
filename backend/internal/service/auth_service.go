package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"gorm.io/gorm"
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
	_, err := s.repo.FindOAuthAccount(ctx, provider, gothUser.UserID)
	if err == nil {
		// Account exists, return user
		user, err := s.repo.FindUserByEmail(ctx, gothUser.Email)
		if err != nil {
			return nil, errors.New("oauth account exists but user not found")
		}
		return user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err // DB Error
	}

	// 2. Check if User exists by email
	user, err := s.repo.FindUserByEmail(ctx, gothUser.Email)
	if err == nil {
		// User exists
		return user, nil
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
		return nil, errors.New("failed to generate tokens: " + err.Error())
	}

	// Save jti to Redis
	err = s.cache.SaveRefreshTokenJTI(ctx, jti, user.ID, config.Env.JwtRefreshExpiration)
	if err != nil {
		return nil, errors.New("failed to cache refresh token: " + err.Error())
	}

	return &response.AuthTokens{
		AccessToken:  accToken,
		RefreshToken: refToken,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
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
		return nil, errors.New("invalid refresh token: " + err.Error())
	}

	// 2. Validate JTI against Redis cache (check if revoked/expired)
	err = s.cache.ValidateRefreshTokenJTI(ctx, jti, userID)
	if err != nil {
		return nil, errors.New("refresh token session expired or revoked")
	}

	// 3. Delete old JTI from Redis (Refresh token rotation)
	_ = s.cache.DeleteRefreshTokenJTI(ctx, jti)

	// 4. Generate new tokens
	accToken, newRefToken, newJTI, err := jwt.GenerateTokens(userID)
	if err != nil {
		return nil, errors.New("failed to generate new tokens: " + err.Error())
	}

	// 5. Save new JTI to Redis
	err = s.cache.SaveRefreshTokenJTI(ctx, newJTI, userID, config.Env.JwtRefreshExpiration)
	if err != nil {
		return nil, errors.New("failed to cache new refresh token: " + err.Error())
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
