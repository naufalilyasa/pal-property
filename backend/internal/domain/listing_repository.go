package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
)

type ListingRepository interface {
	Create(ctx context.Context, listing *entity.Listing) (*entity.Listing, error)
	CreateImage(ctx context.Context, image *entity.ListingImage) (*entity.ListingImage, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Listing, error)
	FindImageByID(ctx context.Context, id uuid.UUID) (*entity.ListingImage, error)
	CreateVideo(ctx context.Context, video *entity.ListingVideo) (*entity.ListingVideo, error)
	FindVideoByListingID(ctx context.Context, listingID uuid.UUID) (*entity.ListingVideo, error)
	DeleteVideoByListingID(ctx context.Context, listingID uuid.UUID) error
	FindBySlug(ctx context.Context, slug string) (*entity.Listing, error)
	ListActiveImagesByListingID(ctx context.Context, listingID uuid.UUID) ([]*entity.ListingImage, error)
	Update(ctx context.Context, listing *entity.Listing, fields []string) (*entity.Listing, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteImage(ctx context.Context, listingID, imageID uuid.UUID) error
	List(ctx context.Context, filter ListingFilter) ([]*entity.Listing, int64, error)
	ReorderImages(ctx context.Context, listingID uuid.UUID, orderedImageIDs []uuid.UUID) error
	SetPrimaryImage(ctx context.Context, listingID, imageID uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID, filter ListingFilter) ([]*entity.Listing, int64, error)
	FindByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]*entity.Listing, error)
}

type ListingFilter struct {
	Status           string
	Statuses         []string
	TransactionType  string
	CategoryID       *uuid.UUID
	LocationProvince string
	LocationCity     string
	PriceMin         *int64
	PriceMax         *int64
	BedroomCount     *int
	BathroomCount    *int
	LandAreaMin      *int
	LandAreaMax      *int
	BuildingAreaMin  *int
	BuildingAreaMax  *int
	CertificateType  string
	Condition        string
	Furnishing       string
	Page             int
	Limit            int
}
