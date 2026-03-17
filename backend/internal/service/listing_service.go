package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	"github.com/naufalilyasa/pal-property-backend/pkg/mediaasset"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/slug"
	"gorm.io/datatypes"
)

type ListingService interface {
	Create(ctx context.Context, userID uuid.UUID, req *request.CreateListingRequest) (*response.ListingResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*response.ListingResponse, error)
	GetBySlug(ctx context.Context, slugStr string) (*response.ListingResponse, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string, req *request.UpdateListingRequest) (*response.ListingResponse, error)
	UploadImage(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string, file *multipart.FileHeader) (*response.ListingResponse, error)
	DeleteImage(ctx context.Context, id uuid.UUID, imageID uuid.UUID, userID uuid.UUID, userRole string) (*response.ListingResponse, error)
	SetPrimaryImage(ctx context.Context, id uuid.UUID, imageID uuid.UUID, userID uuid.UUID, userRole string) (*response.ListingResponse, error)
	ReorderImages(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string, orderedImageIDs []uuid.UUID) (*response.ListingResponse, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string) error
	List(ctx context.Context, filter domain.ListingFilter) (*response.PaginatedListings, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, filter domain.ListingFilter) (*response.PaginatedListings, error)
}

type listingService struct {
	repo    domain.ListingRepository
	storage domain.ListingImageStorage
}

func NewListingService(repo domain.ListingRepository, storage ...domain.ListingImageStorage) ListingService {
	service := &listingService{repo: repo}
	if len(storage) > 0 {
		service.storage = storage[0]
	}
	return service
}

func (s *listingService) Create(ctx context.Context, userID uuid.UUID, req *request.CreateListingRequest) (*response.ListingResponse, error) {
	if req.Title == "" || req.Price <= 0 || req.Status == "" {
		return nil, domain.ErrInvalidCredential
	}
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

func (s *listingService) UploadImage(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string, file *multipart.FileHeader) (*response.ListingResponse, error) {
	if err := validateListingImageFile(file); err != nil {
		return nil, err
	}
	if err := s.requireImageStorage(); err != nil {
		return nil, err
	}

	listing, err := s.getAuthorizedListing(ctx, id, userID, userRole)
	if err != nil {
		return nil, err
	}

	uploaded, err := s.storage.UploadListingImage(ctx, mediaasset.UploadInput{
		File:         file,
		Folder:       fmt.Sprintf("listings/%s", id.String()),
		PublicID:     uuid.NewString(),
		ResourceType: mediaasset.DefaultResourceType,
		DeliveryType: mediaasset.DefaultDeliveryType,
	})
	if err != nil {
		return nil, err
	}

	image := buildListingImageEntity(id, file, uploaded)
	if _, err := s.repo.CreateImage(ctx, image); err != nil {
		s.destroyUploadedImage(ctx, uploaded)
		return nil, err
	}

	listing.Images = append(listing.Images, *image)
	return s.fetchListingResponse(ctx, id)
}

func (s *listingService) DeleteImage(ctx context.Context, id uuid.UUID, imageID uuid.UUID, userID uuid.UUID, userRole string) (*response.ListingResponse, error) {
	if err := s.requireImageStorage(); err != nil {
		return nil, err
	}

	if _, err := s.getAuthorizedListing(ctx, id, userID, userRole); err != nil {
		return nil, err
	}

	image, err := s.repo.FindImageByID(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if image.ListingID != id {
		return nil, domain.ErrNotFound
	}

	if err := s.repo.DeleteImage(ctx, id, imageID); err != nil {
		return nil, err
	}

	s.destroyListingImageByEntity(ctx, image)

	return s.fetchListingResponse(ctx, id)
}

func (s *listingService) SetPrimaryImage(ctx context.Context, id uuid.UUID, imageID uuid.UUID, userID uuid.UUID, userRole string) (*response.ListingResponse, error) {
	if _, err := s.getAuthorizedListing(ctx, id, userID, userRole); err != nil {
		return nil, err
	}

	image, err := s.repo.FindImageByID(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if image.ListingID != id {
		return nil, domain.ErrNotFound
	}

	if err := s.repo.SetPrimaryImage(ctx, id, imageID); err != nil {
		return nil, err
	}

	return s.fetchListingResponse(ctx, id)
}

func (s *listingService) ReorderImages(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string, orderedImageIDs []uuid.UUID) (*response.ListingResponse, error) {
	if _, err := s.getAuthorizedListing(ctx, id, userID, userRole); err != nil {
		return nil, err
	}

	images, err := s.repo.ListActiveImagesByListingID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := validateListingImageOrder(images, orderedImageIDs); err != nil {
		return nil, err
	}

	if err := s.repo.ReorderImages(ctx, id, orderedImageIDs); err != nil {
		return nil, err
	}

	return s.fetchListingResponse(ctx, id)
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
		Images:           mapListingImagesToResponse(l.Images),
		Category: func() *response.CategoryShortResponse {
			if l.Category == nil {
				return nil
			}
			return &response.CategoryShortResponse{
				ID:      l.Category.ID,
				Name:    l.Category.Name,
				Slug:    l.Category.Slug,
				IconURL: l.Category.IconURL,
			}
		}(),
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func (s *listingService) getAuthorizedListing(ctx context.Context, id uuid.UUID, userID uuid.UUID, userRole string) (*entity.Listing, error) {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkOwnership(listing, userID, userRole); err != nil {
		return nil, err
	}

	return listing, nil
}

func (s *listingService) fetchListingResponse(ctx context.Context, id uuid.UUID) (*response.ListingResponse, error) {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(listing), nil
}

func (s *listingService) requireImageStorage() error {
	if s.storage == nil {
		return domain.ErrImageStorageUnset
	}

	return nil
}

func (s *listingService) destroyUploadedImage(ctx context.Context, uploaded *mediaasset.UploadResult) {
	if s.storage == nil || uploaded == nil || uploaded.PublicID == "" {
		return
	}

	resourceType := uploaded.ResourceType
	if resourceType == "" {
		resourceType = mediaasset.DefaultResourceType
	}
	deliveryType := uploaded.DeliveryType
	if deliveryType == "" {
		deliveryType = mediaasset.DefaultDeliveryType
	}

	_, _ = s.storage.DestroyListingImage(ctx, mediaasset.DestroyInput{
		PublicID:     uploaded.PublicID,
		ResourceType: resourceType,
		DeliveryType: deliveryType,
		Invalidate:   true,
	})
}

func (s *listingService) destroyListingImageByEntity(ctx context.Context, image *entity.ListingImage) {
	if s.storage == nil || image == nil || image.PublicID == nil || *image.PublicID == "" {
		return
	}

	resourceType := mediaasset.DefaultResourceType
	if image.ResourceType != nil && *image.ResourceType != "" {
		resourceType = *image.ResourceType
	}
	deliveryType := mediaasset.DefaultDeliveryType
	if image.Type != nil && *image.Type != "" {
		deliveryType = *image.Type
	}

	_, _ = s.storage.DestroyListingImage(ctx, mediaasset.DestroyInput{
		PublicID:     *image.PublicID,
		ResourceType: resourceType,
		DeliveryType: deliveryType,
		Invalidate:   true,
	})
}

func buildListingImageEntity(listingID uuid.UUID, file *multipart.FileHeader, uploaded *mediaasset.UploadResult) *entity.ListingImage {
	image := &entity.ListingImage{
		ListingID:    listingID,
		URL:          uploaded.SecureURL,
		AssetID:      stringPointer(uploaded.AssetID),
		PublicID:     stringPointer(uploaded.PublicID),
		Version:      int64Pointer(uploaded.Version),
		Format:       stringPointer(uploaded.Format),
		Bytes:        int64Pointer(uploaded.Bytes),
		Width:        intPointer(uploaded.Width),
		Height:       intPointer(uploaded.Height),
		ResourceType: stringPointer(uploaded.ResourceType),
		Type:         stringPointer(uploaded.DeliveryType),
	}

	originalFilename := uploaded.OriginalFilename
	if originalFilename == "" && file != nil {
		originalFilename = file.Filename
	}
	image.OriginalFilename = stringPointer(originalFilename)

	return image
}

func mapListingImagesToResponse(images []entity.ListingImage) []*response.ListingImageResponse {
	if len(images) == 0 {
		return []*response.ListingImageResponse{}
	}

	ordered := append([]entity.ListingImage(nil), images...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].SortOrder == ordered[j].SortOrder {
			return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
		}
		return ordered[i].SortOrder < ordered[j].SortOrder
	})

	res := make([]*response.ListingImageResponse, 0, len(ordered))
	for _, image := range ordered {
		img := image
		res = append(res, &response.ListingImageResponse{
			ID:               img.ID,
			URL:              img.URL,
			Format:           img.Format,
			Bytes:            img.Bytes,
			Width:            img.Width,
			Height:           img.Height,
			OriginalFilename: img.OriginalFilename,
			IsPrimary:        img.IsPrimary,
			SortOrder:        img.SortOrder,
			CreatedAt:        img.CreatedAt,
		})
	}

	return res
}

func validateListingImageFile(file *multipart.FileHeader) error {
	if file == nil {
		return domain.ErrInvalidImageFile
	}

	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	if strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return nil
	}

	switch strings.ToLower(filepath.Ext(file.Filename)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".avif", ".tif", ".tiff", ".heic", ".heif":
		return nil
	default:
		return domain.ErrInvalidImageFile
	}
}

func validateListingImageOrder(images []*entity.ListingImage, orderedImageIDs []uuid.UUID) error {
	if len(images) != len(orderedImageIDs) {
		return domain.ErrImageOrderInvalid
	}

	allowed := make(map[uuid.UUID]struct{}, len(images))
	for _, image := range images {
		allowed[image.ID] = struct{}{}
	}

	seen := make(map[uuid.UUID]struct{}, len(orderedImageIDs))
	for _, imageID := range orderedImageIDs {
		if _, ok := allowed[imageID]; !ok {
			return domain.ErrImageOrderInvalid
		}
		if _, ok := seen[imageID]; ok {
			return domain.ErrImageOrderInvalid
		}
		seen[imageID] = struct{}{}
	}

	return nil
}

func stringPointer(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return &value
}

func int64Pointer(value int64) *int64 {
	if value == 0 {
		return nil
	}

	return &value
}

func intPointer(value int) *int {
	if value == 0 {
		return nil
	}

	return &value
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
