package domain

import (
	"context"

	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
)

type RegionRepository interface {
	Count(ctx context.Context) (int64, error)
	CreateBatch(ctx context.Context, regions []entity.AdministrativeRegion) error
	FindByCode(ctx context.Context, code string) (*entity.AdministrativeRegion, error)
	ListByLevel(ctx context.Context, level int) ([]entity.AdministrativeRegion, error)
	ListByParent(ctx context.Context, parentCode string) ([]entity.AdministrativeRegion, error)
}
