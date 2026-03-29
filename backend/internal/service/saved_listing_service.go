package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
)

type SavedListingService interface {
	Save(ctx context.Context, principal pkgauthz.Principal, listingID uuid.UUID) error
	Remove(ctx context.Context, principal pkgauthz.Principal, listingID uuid.UUID) error
	Contains(ctx context.Context, principal pkgauthz.Principal, listingIDs []uuid.UUID) ([]uuid.UUID, error)
	ListByUserID(ctx context.Context, principal pkgauthz.Principal, filter domain.SavedListingFilter) (*response.PaginatedListings, error)
}
