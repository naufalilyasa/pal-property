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
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindAll(ctx context.Context) ([]entity.Category, error) {
	var cats []entity.Category
	err := r.db.WithContext(ctx).
		Preload("Children").
		Where("parent_id IS NULL").
		Order("name ASC").
		Find(&cats).Error
	if err != nil {
		return nil, fmt.Errorf("find all categories: %w", err)
	}
	return cats, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var cat entity.Category
	err := r.db.WithContext(ctx).
		Preload("Children").
		Preload("Parent").
		First(&cat, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find category by id: %w", err)
	}
	return &cat, nil
}

func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	var cat entity.Category
	err := r.db.WithContext(ctx).
		Preload("Children").
		Preload("Parent").
		First(&cat, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find category by slug: %w", err)
	}
	return &cat, nil
}

func (r *categoryRepository) Create(ctx context.Context, cat *entity.Category) (*entity.Category, error) {
	err := r.db.WithContext(ctx).Create(cat).Error
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return nil, domain.ErrConflict
		}
		return nil, fmt.Errorf("create category: %w", err)
	}
	return cat, nil
}

func (r *categoryRepository) Update(ctx context.Context, cat *entity.Category, fields []string) (*entity.Category, error) {
	err := r.db.WithContext(ctx).
		Model(cat).
		Select(fields).
		Updates(cat).Error
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return nil, domain.ErrConflict
		}
		return nil, fmt.Errorf("update category: %w", err)
	}
	return cat, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Category{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *categoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Category{}).
		Where("slug = ?", slug).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("exists by slug: %w", err)
	}
	return count > 0, nil
}

func (r *categoryRepository) CountListingsByCategory(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Listing{}).
		Where("category_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count listings by category: %w", err)
	}
	return count, nil
}

func (r *categoryRepository) CountChildrenByParent(ctx context.Context, parentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Category{}).
		Where("parent_id = ?", parentID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count children by parent: %w", err)
	}
	return count, nil
}
