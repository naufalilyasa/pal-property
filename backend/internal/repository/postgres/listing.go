package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *listingRepository) CreateImage(ctx context.Context, image *entity.ListingImage) (*entity.ListingImage, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lockListing(tx.WithContext(ctx), image.ListingID); err != nil {
			return err
		}

		images, err := r.lockActiveImages(tx.WithContext(ctx), image.ListingID)
		if err != nil {
			return err
		}

		if len(images) >= 10 {
			return domain.ErrImageLimitReached
		}

		image.IsPrimary = len(images) == 0
		image.SortOrder = len(images)

		if err := tx.WithContext(ctx).Create(image).Error; err != nil {
			return fmt.Errorf("create image: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (r *listingRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Listing, error) {
	var listing entity.Listing
	err := r.db.WithContext(ctx).
		Preload("Images", r.activeImagePreload).
		Preload("Video").
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

func (r *listingRepository) FindImageByID(ctx context.Context, id uuid.UUID) (*entity.ListingImage, error) {
	var image entity.ListingImage
	err := r.db.WithContext(ctx).
		Where("listing_images.deleted_at IS NULL").
		First(&image, "listing_images.id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("find image by id: %w", err)
	}

	return &image, nil
}

func (r *listingRepository) CreateVideo(ctx context.Context, video *entity.ListingVideo) (*entity.ListingVideo, error) {
	err := r.db.WithContext(ctx).Create(video).Error
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return nil, domain.ErrVideoAlreadyExists
		}
		return nil, fmt.Errorf("create video: %w", err)
	}
	return video, nil
}

func (r *listingRepository) FindVideoByListingID(ctx context.Context, listingID uuid.UUID) (*entity.ListingVideo, error) {
	var video entity.ListingVideo
	err := r.db.WithContext(ctx).
		First(&video, "listing_id = ?", listingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find video by listing id: %w", err)
	}
	return &video, nil
}

func (r *listingRepository) DeleteVideoByListingID(ctx context.Context, listingID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("listing_id = ?", listingID).Delete(&entity.ListingVideo{})
	if result.Error != nil {
		return fmt.Errorf("delete video: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *listingRepository) FindBySlug(ctx context.Context, slug string) (*entity.Listing, error) {
	var listing entity.Listing
	err := r.db.WithContext(ctx).
		Preload("Images", r.activeImagePreload).
		Preload("Video").
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

func (r *listingRepository) ListActiveImagesByListingID(ctx context.Context, listingID uuid.UUID) ([]*entity.ListingImage, error) {
	var images []*entity.ListingImage
	err := r.db.WithContext(ctx).
		Model(&entity.ListingImage{}).
		Where("listing_id = ?", listingID).
		Where("deleted_at IS NULL").
		Order("sort_order ASC").
		Find(&images).Error
	if err != nil {
		return nil, fmt.Errorf("list active images by listing id: %w", err)
	}

	return images, nil
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

func (r *listingRepository) DeleteImage(ctx context.Context, listingID, imageID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		images, err := r.lockActiveImages(tx.WithContext(ctx), listingID)
		if err != nil {
			return err
		}

		var target *entity.ListingImage
		for _, image := range images {
			if image.ID == imageID {
				target = image
				break
			}
		}

		if target == nil {
			return domain.ErrNotFound
		}

		if err := tx.WithContext(ctx).
			Where("id = ? AND listing_id = ? AND deleted_at IS NULL", imageID, listingID).
			Delete(&entity.ListingImage{}).Error; err != nil {
			return fmt.Errorf("delete image: %w", err)
		}

		if err := r.normalizeActiveImageSortOrder(ctx, tx.WithContext(ctx), listingID, images, imageID); err != nil {
			return err
		}

		if !target.IsPrimary {
			return nil
		}

		for _, image := range images {
			if image.ID == imageID {
				continue
			}

			if err := tx.WithContext(ctx).
				Model(&entity.ListingImage{}).
				Where("id = ? AND listing_id = ? AND deleted_at IS NULL", image.ID, listingID).
				Update("is_primary", true).Error; err != nil {
				return fmt.Errorf("delete image: promote fallback primary: %w", err)
			}

			return nil
		}

		return nil
	})
}

func (r *listingRepository) normalizeActiveImageSortOrder(ctx context.Context, tx *gorm.DB, listingID uuid.UUID, images []*entity.ListingImage, deletedImageID uuid.UUID) error {
	remainingCount := len(images) - 1
	if remainingCount <= 0 {
		return nil
	}

	if err := tx.Model(&entity.ListingImage{}).
		Where("listing_id = ? AND deleted_at IS NULL", listingID).
		Update("sort_order", gorm.Expr("sort_order + ?", len(images))).Error; err != nil {
		return fmt.Errorf("delete image: offset remaining sort order: %w", err)
	}

	nextSortOrder := 0
	for _, image := range images {
		if image.ID == deletedImageID {
			continue
		}

		if err := tx.WithContext(ctx).
			Model(&entity.ListingImage{}).
			Where("id = ? AND listing_id = ? AND deleted_at IS NULL", image.ID, listingID).
			Update("sort_order", nextSortOrder).Error; err != nil {
			return fmt.Errorf("delete image: normalize remaining sort order: %w", err)
		}

		nextSortOrder++
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
	err = db.Preload("Images", r.activeImagePreload).
		Preload("Video").
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

func (r *listingRepository) ReorderImages(ctx context.Context, listingID uuid.UUID, orderedImageIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		images, err := r.lockActiveImages(tx.WithContext(ctx), listingID)
		if err != nil {
			return err
		}

		if len(images) != len(orderedImageIDs) {
			return fmt.Errorf("reorder images: image set mismatch")
		}

		imageByID := make(map[uuid.UUID]*entity.ListingImage, len(images))
		for _, image := range images {
			imageByID[image.ID] = image
		}

		seen := make(map[uuid.UUID]struct{}, len(orderedImageIDs))
		for _, imageID := range orderedImageIDs {
			if _, ok := imageByID[imageID]; !ok {
				return domain.ErrNotFound
			}
			if _, ok := seen[imageID]; ok {
				return fmt.Errorf("reorder images: duplicate image id %s", imageID)
			}
			seen[imageID] = struct{}{}
		}

		offset := len(images)
		if err := tx.WithContext(ctx).
			Model(&entity.ListingImage{}).
			Where("listing_id = ? AND deleted_at IS NULL", listingID).
			Update("sort_order", gorm.Expr("sort_order + ?", offset)).Error; err != nil {
			return fmt.Errorf("reorder images: offset existing sort order: %w", err)
		}

		for sortOrder, imageID := range orderedImageIDs {
			if err := tx.WithContext(ctx).
				Model(&entity.ListingImage{}).
				Where("id = ? AND listing_id = ? AND deleted_at IS NULL", imageID, listingID).
				Update("sort_order", sortOrder).Error; err != nil {
				return fmt.Errorf("reorder images: set sort order: %w", err)
			}
		}

		return nil
	})
}

func (r *listingRepository) SetPrimaryImage(ctx context.Context, listingID, imageID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		images, err := r.lockActiveImages(tx.WithContext(ctx), listingID)
		if err != nil {
			return err
		}

		targetFound := false
		for _, image := range images {
			if image.ID == imageID {
				targetFound = true
				break
			}
		}

		if !targetFound {
			return domain.ErrNotFound
		}

		if err := r.clearPrimaryImage(tx.WithContext(ctx), listingID); err != nil {
			return err
		}

		if err := tx.WithContext(ctx).
			Model(&entity.ListingImage{}).
			Where("id = ? AND listing_id = ? AND deleted_at IS NULL", imageID, listingID).
			Update("is_primary", true).Error; err != nil {
			return fmt.Errorf("set primary image: %w", err)
		}

		return nil
	})
}

func (r *listingRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Listing{}).Where("slug = ?", slug).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *listingRepository) FindByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]*entity.Listing, error) {
	var listings []*entity.Listing
	err := r.db.WithContext(ctx).
		Preload("Images", r.activeImagePreload).
		Preload("Video").
		Preload("Category").
		Joins("User").
		Where("listings.category_id = ?", categoryID).
		Where("listings.deleted_at IS NULL").
		Order("listings.created_at DESC").
		Find(&listings).Error
	if err != nil {
		return nil, fmt.Errorf("find listings by category id: %w", err)
	}
	return listings, nil
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
	err = db.Preload("Images", r.activeImagePreload).
		Preload("Video").
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
	} else if len(filter.Statuses) > 0 {
		db = db.Where("listings.status IN ?", filter.Statuses)
	}
	if filter.TransactionType != "" {
		db = db.Where("listings.transaction_type = ?", filter.TransactionType)
	}
	if filter.CategoryID != nil {
		db = db.Where("listings.category_id = ?", filter.CategoryID)
	}
	if filter.LocationProvince != "" {
		db = db.Where("listings.location_province ILIKE ?", fmt.Sprintf("%%%s%%", filter.LocationProvince))
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
	if filter.BedroomCount != nil {
		db = db.Where("listings.bedroom_count >= ?", *filter.BedroomCount)
	}
	if filter.BathroomCount != nil {
		db = db.Where("listings.bathroom_count >= ?", *filter.BathroomCount)
	}
	if filter.LandAreaMin != nil {
		db = db.Where("listings.land_area_sqm >= ?", *filter.LandAreaMin)
	}
	if filter.LandAreaMax != nil {
		db = db.Where("listings.land_area_sqm <= ?", *filter.LandAreaMax)
	}
	if filter.BuildingAreaMin != nil {
		db = db.Where("listings.building_area_sqm >= ?", *filter.BuildingAreaMin)
	}
	if filter.BuildingAreaMax != nil {
		db = db.Where("listings.building_area_sqm <= ?", *filter.BuildingAreaMax)
	}
	if filter.CertificateType != "" {
		db = db.Where("listings.certificate_type = ?", filter.CertificateType)
	}
	if filter.Condition != "" {
		db = db.Where("listings.condition = ?", filter.Condition)
	}
	if filter.Furnishing != "" {
		db = db.Where("listings.furnishing = ?", filter.Furnishing)
	}
	return db
}

func (r *listingRepository) activeImagePreload(db *gorm.DB) *gorm.DB {
	return db.Where("listing_images.deleted_at IS NULL").Order("listing_images.sort_order ASC")
}

func (r *listingRepository) clearPrimaryImage(tx *gorm.DB, listingID uuid.UUID) error {
	if err := tx.Model(&entity.ListingImage{}).
		Where("listing_id = ? AND deleted_at IS NULL AND is_primary = ?", listingID, true).
		Update("is_primary", false).Error; err != nil {
		return fmt.Errorf("clear primary image: %w", err)
	}

	return nil
}

func (r *listingRepository) lockListing(tx *gorm.DB, listingID uuid.UUID) error {
	var listing entity.Listing
	err := tx.Model(&entity.Listing{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&listing, "id = ?", listingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}

		return fmt.Errorf("lock listing: %w", err)
	}

	return nil
}

func (r *listingRepository) lockActiveImages(tx *gorm.DB, listingID uuid.UUID) ([]*entity.ListingImage, error) {
	var images []*entity.ListingImage
	err := tx.Model(&entity.ListingImage{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("listing_id = ?", listingID).
		Where("deleted_at IS NULL").
		Order("sort_order ASC").
		Find(&images).Error
	if err != nil {
		return nil, fmt.Errorf("lock active images: %w", err)
	}

	return images, nil
}
