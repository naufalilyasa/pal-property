package domain

import (
	"context"

	"github.com/username/pal-property-backend/internal/domain/entity" // Adjust module path if needed
)

type AuthRepository interface {
	FindOAuthAccount(ctx context.Context, provider, providerUserID string) (*entity.OAuthAccount, error)
	CreateUserWithOAuth(ctx context.Context, user *entity.User, account *entity.OAuthAccount) (*entity.User, error)
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	Updates(ctx context.Context, user *entity.User) error
}
