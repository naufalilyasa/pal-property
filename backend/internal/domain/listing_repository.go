package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
)

type ListingRepository interface {
	Create(ctx context.Context, listing *entity.Listing) (*entity.Listing, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Listing, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Listing, error)
	Update(ctx context.Context, listing *entity.Listing, fields []string) (*entity.Listing, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListingFilter) ([]*entity.Listing, int64, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID, filter ListingFilter) ([]*entity.Listing, int64, error)
}

type ListingFilter struct {
	Status       string
	CategoryID   *uuid.UUID
	LocationCity string
	PriceMin     *int64
	PriceMax     *int64
	Page         int
	Limit        int
}
