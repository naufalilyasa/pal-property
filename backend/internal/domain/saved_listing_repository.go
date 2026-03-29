package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
)

type SavedListingFilter struct {
	UserID uuid.UUID
	Page   int
	Limit  int
}

type SavedListingRepository interface {
	Save(ctx context.Context, savedListing *entity.SavedListing) (*entity.SavedListing, error)
	Remove(ctx context.Context, userID, listingID uuid.UUID) error
	Contains(ctx context.Context, userID uuid.UUID, listingIDs []uuid.UUID) ([]uuid.UUID, error)
	ListByUserID(ctx context.Context, filter SavedListingFilter) ([]*entity.SavedListing, int64, error)
}
