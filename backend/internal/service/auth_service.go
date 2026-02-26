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
