package domain

import (
	"context"

	"github.com/naufalilyasa/pal-property-backend/pkg/mediaasset"
)

type ListingVideoStorage interface {
	UploadListingVideo(ctx context.Context, input mediaasset.UploadInput) (*mediaasset.UploadResult, error)
	DestroyListingVideo(ctx context.Context, input mediaasset.DestroyInput) (*mediaasset.DestroyResult, error)
}
