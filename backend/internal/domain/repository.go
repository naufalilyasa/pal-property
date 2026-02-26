package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity" // Adjust module path if needed
)

type AuthRepository interface {
	FindOAuthAccount(ctx context.Context, provider, providerUserID string) (*entity.OAuthAccount, error)
	CreateUserWithOAuth(ctx context.Context, user *entity.User, account *entity.OAuthAccount) (*entity.User, error)
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Updates(ctx context.Context, user *entity.User) error
}
