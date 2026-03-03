package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/slug"
	"gorm.io/datatypes"
)

type ListingService interface {
	Create(ctx context.Context, userID uuid.UUID, req *request.CreateListingRequest) (*response.ListingResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*response.ListingResponse, error)
	GetBySlug(ctx context.Context, slugStr string) (*response.ListingResponse, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string, req *request.UpdateListingRequest) (*response.ListingResponse, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string) error
	List(ctx context.Context, filter domain.ListingFilter) (*response.PaginatedListings, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, filter domain.ListingFilter) (*response.PaginatedListings, error)
}

type listingService struct {
	repo domain.ListingRepository
}

func NewListingService(repo domain.ListingRepository) ListingService {
	return &listingService{repo: repo}
}

func (s *listingService) Create(ctx context.Context, userID uuid.UUID, req *request.CreateListingRequest) (*response.ListingResponse, error) {
	if req.Title == "" || req.Price <= 0 || req.Status == "" {
		return nil, domain.ErrInvalidCredential
	}
	// Generate unique slug
	// Generate unique slug
	listingSlug := slug.GenerateUnique(req.Title, func(candidate string) bool {
		exists, _ := s.repo.ExistsBySlug(ctx, candidate)
		return exists
	})

	specsJSON, err := json.Marshal(req.Specifications)
	if err != nil {
		return nil, err
	}

	listing := &entity.Listing{
		UserID:           userID,
		CategoryID:       req.CategoryID,
		Title:            req.Title,
		Slug:             listingSlug,
		Description:      req.Description,
		Price:            req.Price,
		LocationCity:     req.LocationCity,
		LocationDistrict: req.LocationDistrict,
		AddressDetail:    req.AddressDetail,
		Status:           req.Status,
		Specifications:   datatypes.JSON(specsJSON),
	}

	created, err := s.repo.Create(ctx, listing)
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(created), nil
}

func (s *listingService) GetByID(ctx context.Context, id uuid.UUID) (*response.ListingResponse, error) {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Increment view count atomically
	_ = s.repo.IncrementViewCount(ctx, id)
	listing.ViewCount++ // Update local object for response mapping

	return s.mapToResponse(listing), nil
	_ = s.repo.IncrementViewCount(ctx, id)

	return s.mapToResponse(listing), nil
}

func (s *listingService) GetBySlug(ctx context.Context, slugStr string) (*response.ListingResponse, error) {
	listing, err := s.repo.FindBySlug(ctx, slugStr)
	if err != nil {
		return nil, err
	}

	// Increment view count atomically
	_ = s.repo.IncrementViewCount(ctx, listing.ID)
	listing.ViewCount++ // Update local object for response mapping

	return s.mapToResponse(listing), nil
	_ = s.repo.IncrementViewCount(ctx, listing.ID)

	return s.mapToResponse(listing), nil
}

func (s *listingService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string, req *request.UpdateListingRequest) (*response.ListingResponse, error) {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkOwnership(listing, userID, userRole); err != nil {
		return nil, err
	}

	fields := []string{}

	if req.CategoryID != nil {
		listing.CategoryID = req.CategoryID
		fields = append(fields, "category_id")
	}

	if req.Title != nil && *req.Title != listing.Title {
		listing.Title = *req.Title
		listing.Slug = slug.GenerateUnique(*req.Title, func(candidate string) bool {
			exists, _ := s.repo.ExistsBySlug(ctx, candidate)
			return exists
		})
		fields = append(fields, "title", "slug")
	}

	if req.Description != nil {
		listing.Description = req.Description
		fields = append(fields, "description")
	}

	if req.Price != nil {
		listing.Price = *req.Price
		fields = append(fields, "price")
	}

	if req.LocationCity != nil {
		listing.LocationCity = req.LocationCity
		fields = append(fields, "location_city")
	}

	if req.LocationDistrict != nil {
		listing.LocationDistrict = req.LocationDistrict
		fields = append(fields, "location_district")
	}

	if req.AddressDetail != nil {
		listing.AddressDetail = req.AddressDetail
		fields = append(fields, "address_detail")
	}

	if req.Status != nil {
		listing.Status = *req.Status
		fields = append(fields, "status")
	}

	if req.Specifications != nil {
		specsJSON, err := json.Marshal(req.Specifications)
		if err != nil {
			return nil, err
		}
		listing.Specifications = datatypes.JSON(specsJSON)
		fields = append(fields, "specifications")
	}

	if len(fields) == 0 {
		return s.mapToResponse(listing), nil
	}

	updated, err := s.repo.Update(ctx, listing, fields)
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(updated), nil
}

func (s *listingService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string) error {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.checkOwnership(listing, userID, userRole); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *listingService) List(ctx context.Context, filter domain.ListingFilter) (*response.PaginatedListings, error) {
	listings, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return s.mapToPaginatedResponse(listings, total, filter.Page, filter.Limit), nil
}

func (s *listingService) ListByUserID(ctx context.Context, userID uuid.UUID, filter domain.ListingFilter) (*response.PaginatedListings, error) {
	listings, total, err := s.repo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	return s.mapToPaginatedResponse(listings, total, filter.Page, filter.Limit), nil
}

func (s *listingService) checkOwnership(listing *entity.Listing, userID uuid.UUID, userRole string) error {
	if userRole == "admin" {
		return nil // admin bypass
	}
	if listing.UserID != userID {
		return domain.ErrForbidden
	}
	return nil
}

func (s *listingService) mapToResponse(l *entity.Listing) *response.ListingResponse {
	if l == nil {
		return nil
	}
	return &response.ListingResponse{
		ID:               l.ID,
		UserID:           l.UserID,
		CategoryID:       l.CategoryID,
		Title:            l.Title,
		Slug:             l.Slug,
		Description:      l.Description,
		Price:            l.Price,
		Currency:         l.Currency,
		LocationCity:     l.LocationCity,
		LocationDistrict: l.LocationDistrict,
		AddressDetail:    l.AddressDetail,
		Status:           l.Status,
		IsFeatured:       l.IsFeatured,
		Specifications:   l.Specifications,
		ViewCount:        l.ViewCount,
		CreatedAt:        l.CreatedAt,
		UpdatedAt:        l.UpdatedAt,
	}
}

func (s *listingService) mapToPaginatedResponse(listings []*entity.Listing, total int64, page, limit int) *response.PaginatedListings {
	data := make([]*response.ListingResponse, len(listings))
	for i, l := range listings {
		data[i] = s.mapToResponse(l)
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &response.PaginatedListings{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
