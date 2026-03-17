package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
)

type CategoryRepository interface {
	FindAll(ctx context.Context) ([]entity.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Category, error)
	Create(ctx context.Context, category *entity.Category) (*entity.Category, error)
	Update(ctx context.Context, category *entity.Category, fields []string) (*entity.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	CountListingsByCategory(ctx context.Context, id uuid.UUID) (int64, error)
	CountChildrenByParent(ctx context.Context, parentID uuid.UUID) (int64, error)
}
