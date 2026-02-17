package service

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"github.com/username/pal-property-backend/internal/domain"
	"github.com/username/pal-property-backend/internal/domain/entity"
	"gorm.io/gorm"
)

type AuthService interface {
	// BeginAuth initiates the OAuth2 flow.
	BeginAuth(c *gin.Context, provider string)
	// CompleteAuth handles the OAuth2 callback.
	CompleteAuth(c *gin.Context, provider string) (*entity.User, error)
}

type authService struct {
	repo domain.AuthRepository
}

func NewAuthService(repo domain.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) BeginAuth(c *gin.Context, provider string) {
	// Add provider to context for Gothic
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (s *authService) CompleteAuth(c *gin.Context, provider string) (*entity.User, error) {
	// Add provider to context for Gothic
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		return nil, err
	}

	// 1. Check if OAuthAccount exists
	_, err = s.repo.FindOAuthAccount(c.Request.Context(), provider, gothUser.UserID)
	if err == nil {
		// Account exists, return user
		user, err := s.repo.FindUserByEmail(c.Request.Context(), gothUser.Email)
		if err != nil {
			return nil, errors.New("oauth account exists but user not found")
		}
		return user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err // DB Error
	}

	// 2. Check if User exists by email
	user, err := s.repo.FindUserByEmail(c.Request.Context(), gothUser.Email)
	if err == nil {
		// User exists
		// TODO: Link OAuth account to existing user here if desired.
		// For now, just return the user to log them in.
		return user, nil
	}

	// 3. User does not exist, create new
	newUser := &entity.User{
		Name:       gothUser.Name,
		Email:      gothUser.Email,
		AvatarURL:  &gothUser.AvatarURL,
		Role:       "user",
		IsVerified: true, // OAuth usually verified
	}

	newAccount := &entity.OAuthAccount{
		Provider:       provider,
		ProviderUserID: gothUser.UserID,
		AccessToken:    &gothUser.AccessToken,
		RefreshToken:   &gothUser.RefreshToken,
	}

	return s.repo.CreateUserWithOAuth(c.Request.Context(), newUser, newAccount)
}
