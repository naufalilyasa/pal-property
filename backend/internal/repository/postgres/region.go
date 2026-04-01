package postgres

import (
	"context"
	"errors"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"gorm.io/gorm"
)

type regionRepository struct {
	db *gorm.DB
}

func NewRegionRepository(db *gorm.DB) domain.RegionRepository {
	return &regionRepository{db: db}
}

func (r *regionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.AdministrativeRegion{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *regionRepository) CreateBatch(ctx context.Context, regions []entity.AdministrativeRegion) error {
	if len(regions) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).CreateInBatches(regions, 1000).Error
}

func (r *regionRepository) FindByCode(ctx context.Context, code string) (*entity.AdministrativeRegion, error) {
	var region entity.AdministrativeRegion
	if err := r.db.WithContext(ctx).Where("code = ?", code).Take(&region).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &region, nil
}

func (r *regionRepository) ListByLevel(ctx context.Context, level int) ([]entity.AdministrativeRegion, error) {
	var regions []entity.AdministrativeRegion
	if err := r.db.WithContext(ctx).Where("level = ?", level).Order("name ASC").Find(&regions).Error; err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *regionRepository) ListByParent(ctx context.Context, parentCode string) ([]entity.AdministrativeRegion, error) {
	var regions []entity.AdministrativeRegion
	if err := r.db.WithContext(ctx).Where("parent_code = ?", parentCode).Order("name ASC").Find(&regions).Error; err != nil {
		return nil, err
	}

	return regions, nil
}
