package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
)

const SavedListingContainsLimit = 50

type savedListingService struct {
	repo        domain.SavedListingRepository
	listingRepo domain.ListingRepository
}

func NewSavedListingService(savedRepo domain.SavedListingRepository, listingRepo domain.ListingRepository) SavedListingService {
	return &savedListingService{repo: savedRepo, listingRepo: listingRepo}
}

func (s *savedListingService) Save(ctx context.Context, principal pkgauthz.Principal, listingID uuid.UUID) error {
	listing, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil {
		return err
	}
	if listing == nil || listing.Status != "active" {
		return domain.ErrNotFound
	}
	_, err = s.repo.Save(ctx, &entity.SavedListing{
		UserID:    principal.UserID,
		ListingID: listingID,
	})
	return err
}

func (s *savedListingService) Remove(ctx context.Context, principal pkgauthz.Principal, listingID uuid.UUID) error {
	return s.repo.Remove(ctx, principal.UserID, listingID)
}

func (s *savedListingService) Contains(ctx context.Context, principal pkgauthz.Principal, listingIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(listingIDs) > SavedListingContainsLimit {
		return nil, fmt.Errorf("contains: at most %d listing ids are allowed", SavedListingContainsLimit)
	}
	saved, err := s.repo.Contains(ctx, principal.UserID, listingIDs)
	if err != nil {
		return nil, err
	}
	contains := make(map[uuid.UUID]struct{}, len(saved))
	for _, id := range saved {
		contains[id] = struct{}{}
	}
	res := make([]uuid.UUID, 0, len(saved))
	for _, id := range listingIDs {
		if _, ok := contains[id]; ok {
			res = append(res, id)
		}
	}
	return res, nil
}

func (s *savedListingService) ListByUserID(ctx context.Context, principal pkgauthz.Principal, filter domain.SavedListingFilter) (*response.PaginatedListings, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	filter.UserID = principal.UserID
	saved, total, err := s.repo.ListByUserID(ctx, filter)
	if err != nil {
		return nil, err
	}
	listings := make([]*entity.Listing, 0, len(saved))
	for _, sl := range saved {
		if sl != nil && sl.Listing != nil {
			listings = append(listings, sl.Listing)
		}
	}
	return (&listingService{}).mapToPaginatedResponse(listings, total, filter.Page, filter.Limit), nil
}
