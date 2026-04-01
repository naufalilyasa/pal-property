package service

import (
	"context"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
)

type multiSearchProjector struct {
	projectors []domain.SearchProjector
}

func NewMultiSearchProjector(projectors ...domain.SearchProjector) domain.SearchProjector {
	filtered := make([]domain.SearchProjector, 0, len(projectors))
	for _, projector := range projectors {
		if projector != nil {
			filtered = append(filtered, projector)
		}
	}
	if len(filtered) == 0 {
		return NewNoopSearchProjector()
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return &multiSearchProjector{projectors: filtered}
}

func (m *multiSearchProjector) HandleListingEvent(ctx context.Context, event domain.ListingEvent) error {
	for _, projector := range m.projectors {
		if err := projector.HandleListingEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiSearchProjector) HandleCategoryEvent(ctx context.Context, event domain.CategoryEvent) error {
	for _, projector := range m.projectors {
		if err := projector.HandleCategoryEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
