package domain

import (
	"context"

	"github.com/naufalilyasa/pal-property-backend/pkg/mediaasset"
)

type ListingImageStorage interface {
	UploadListingImage(ctx context.Context, input mediaasset.UploadInput) (*mediaasset.UploadResult, error)
	DestroyListingImage(ctx context.Context, input mediaasset.DestroyInput) (*mediaasset.DestroyResult, error)
}
