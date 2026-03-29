package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type savedListingRepository struct {
	db *gorm.DB
}

func NewSavedListingRepository(db *gorm.DB) domain.SavedListingRepository {
	return &savedListingRepository{db: db}
}

func (r *savedListingRepository) Save(ctx context.Context, savedListing *entity.SavedListing) (*entity.SavedListing, error) {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "listing_id"}}, DoNothing: true}).
		Create(savedListing).Error; err != nil {
		return nil, fmt.Errorf("save saved listing: %w", err)
	}
	var result entity.SavedListing
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND listing_id = ?", savedListing.UserID, savedListing.ListingID).
		First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("fetch saved listing: %w", err)
	}
	return &result, nil
}

func (r *savedListingRepository) Remove(ctx context.Context, userID, listingID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND listing_id = ?", userID, listingID).
		Delete(&entity.SavedListing{}).Error; err != nil {
		return fmt.Errorf("remove saved listing: %w", err)
	}
	return nil
}

func (r *savedListingRepository) Contains(ctx context.Context, userID uuid.UUID, listingIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(listingIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var saved []*entity.SavedListing
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND listing_id IN ?", userID, listingIDs).
		Find(&saved).Error; err != nil {
		return nil, fmt.Errorf("contains saved listing: %w", err)
	}
	result := make([]uuid.UUID, 0, len(saved))
	for _, sl := range saved {
		result = append(result, sl.ListingID)
	}
	return result, nil
}

func (r *savedListingRepository) ListByUserID(ctx context.Context, filter domain.SavedListingFilter) ([]*entity.SavedListing, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	var savedListings []*entity.SavedListing
	var total int64
	db := r.db.WithContext(ctx).
		Model(&entity.SavedListing{}).
		Joins("Listing").
		Where("saved_listings.user_id = ?", filter.UserID).
		Where("\"Listing\".status = ?", "active")
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count saved listings: %w", err)
	}
	offset := (filter.Page - 1) * filter.Limit
	if err := db.
		Preload("Listing.Images", r.listingImagesPreload).
		Preload("Listing.Category").
		Order("saved_listings.created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&savedListings).Error; err != nil {
		return nil, 0, fmt.Errorf("list saved listings: %w", err)
	}
	return savedListings, total, nil
}

func (r *savedListingRepository) listingImagesPreload(db *gorm.DB) *gorm.DB {
	return db.Where("listing_images.deleted_at IS NULL").Order("listing_images.sort_order ASC")
}
