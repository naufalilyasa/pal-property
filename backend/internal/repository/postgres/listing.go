package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"gorm.io/gorm"
)

type listingRepository struct {
	db *gorm.DB
}

func NewListingRepository(db *gorm.DB) domain.ListingRepository {
	return &listingRepository{db: db}
}

func (r *listingRepository) Create(ctx context.Context, listing *entity.Listing) (*entity.Listing, error) {
	err := r.db.WithContext(ctx).Create(listing).Error
	if err != nil {
		return nil, err
	}
	return listing, nil
}

func (r *listingRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Listing, error) {
	var listing entity.Listing
	err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("Category").
		Joins("User").
		First(&listing, "listings.id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &listing, nil
}

func (r *listingRepository) FindBySlug(ctx context.Context, slug string) (*entity.Listing, error) {
	var listing entity.Listing
	err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("Category").
		Joins("User").
		First(&listing, "listings.slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &listing, nil
}

func (r *listingRepository) Update(ctx context.Context, listing *entity.Listing, fields []string) (*entity.Listing, error) {
	err := r.db.WithContext(ctx).
		Model(listing).
		Select(fields).
		Updates(listing).Error
	if err != nil {
		return nil, err
	}
	return listing, nil
}

func (r *listingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Listing{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *listingRepository) List(ctx context.Context, filter domain.ListingFilter) ([]*entity.Listing, int64, error) {
	var listings []*entity.Listing
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Listing{})
	db = r.applyFilter(db, filter)

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err = db.Preload("Images").
		Preload("Category").
		Joins("User").
		Order("listings.created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&listings).Error
	if err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}

func (r *listingRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Listing{}).Where("slug = ?", slug).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *listingRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Listing{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *listingRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter domain.ListingFilter) ([]*entity.Listing, int64, error) {
	var listings []*entity.Listing
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Listing{}).Where("user_id = ?", userID)
	db = r.applyFilter(db, filter)

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err = db.Preload("Images").
		Preload("Category").
		Joins("User").
		Order("listings.created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&listings).Error
	if err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}

func (r *listingRepository) applyFilter(db *gorm.DB, filter domain.ListingFilter) *gorm.DB {
	if filter.Status != "" {
		db = db.Where("listings.status = ?", filter.Status)
	}
	if filter.CategoryID != nil {
		db = db.Where("listings.category_id = ?", filter.CategoryID)
	}
	if filter.LocationCity != "" {
		db = db.Where("listings.location_city ILIKE ?", fmt.Sprintf("%%%s%%", filter.LocationCity))
	}
	if filter.PriceMin != nil {
		db = db.Where("listings.price >= ?", *filter.PriceMin)
	}
	if filter.PriceMax != nil {
		db = db.Where("listings.price <= ?", *filter.PriceMax)
	}
	return db
}
