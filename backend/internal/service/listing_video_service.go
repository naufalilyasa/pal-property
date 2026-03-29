package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
)

type ListingVideoInput struct {
	ListingID       uuid.UUID
	DurationSeconds *int
}

type ListingVideoService interface {
	Upload(ctx context.Context, principal pkgauthz.Principal, input ListingVideoInput) (*entity.ListingVideo, error)
	Remove(ctx context.Context, principal pkgauthz.Principal, listingID uuid.UUID) error
}
