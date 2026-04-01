package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"mime/multipart"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/mediaasset"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/slug"
	pkgvalidator "github.com/naufalilyasa/pal-property-backend/pkg/validator"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ListingService interface {
	Create(ctx context.Context, principal pkgauthz.Principal, req *request.CreateListingRequest) (*response.ListingResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*response.ListingResponse, error)
	GetBySlug(ctx context.Context, slugStr string) (*response.ListingResponse, error)
	Update(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, req *request.UpdateListingRequest) (*response.ListingResponse, error)
	UploadImage(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, files []*multipart.FileHeader) (*response.ListingResponse, error)
	DeleteImage(ctx context.Context, id uuid.UUID, imageID uuid.UUID, principal pkgauthz.Principal) (*response.ListingResponse, error)
	SetPrimaryImage(ctx context.Context, id uuid.UUID, imageID uuid.UUID, principal pkgauthz.Principal) (*response.ListingResponse, error)
	ReorderImages(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, orderedImageIDs []uuid.UUID) (*response.ListingResponse, error)
	UploadVideo(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, file *multipart.FileHeader) (*response.ListingResponse, error)
	DeleteVideo(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal) (*response.ListingResponse, error)
	Delete(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal) error
	List(ctx context.Context, filter domain.ListingFilter) (*response.PaginatedListings, error)
	ListByUserID(ctx context.Context, principal pkgauthz.Principal, filter domain.ListingFilter) (*response.PaginatedListings, error)
}

type listingService struct {
	repo         domain.ListingRepository
	storage      domain.ListingImageStorage
	videoStorage domain.ListingVideoStorage
	authz        AuthzService
	regions      RegionLookupService
	publish      domain.EventPublisher
	jobs         domain.SearchIndexJobRepository
	txm          domain.SearchIndexTransactionManager
}

type normalizedListingLocation struct {
	ProvinceName *string
	ProvinceCode *string
	CityName     *string
	CityCode     *string
	DistrictName *string
	DistrictCode *string
	VillageName  *string
	VillageCode  *string
}

func (s *listingService) applyStorage(storage []domain.ListingImageStorage) {
	if len(storage) == 0 {
		return
	}
	s.storage = storage[0]
	if vs, ok := storage[0].(domain.ListingVideoStorage); ok {
		s.videoStorage = vs
	}
}

func NewListingService(repo domain.ListingRepository, storage ...domain.ListingImageStorage) ListingService {
	service := &listingService{repo: repo}
	service.applyStorage(storage)
	return service
}

func NewListingServiceWithAuthz(repo domain.ListingRepository, authzService AuthzService, storage ...domain.ListingImageStorage) ListingService {
	service := &listingService{repo: repo, authz: authzService}
	service.applyStorage(storage)
	return service
}

func NewListingServiceWithAuthzAndPublisher(repo domain.ListingRepository, authzService AuthzService, publisher domain.EventPublisher, storage ...domain.ListingImageStorage) ListingService {
	service := &listingService{repo: repo, authz: authzService, publish: publisher}
	service.applyStorage(storage)
	return service
}

func NewListingServiceWithAuthzAndJobs(repo domain.ListingRepository, authzService AuthzService, jobs domain.SearchIndexJobRepository, storage ...domain.ListingImageStorage) ListingService {
	service := &listingService{repo: repo, authz: authzService, jobs: jobs}
	service.applyStorage(storage)
	return service
}

func NewListingServiceWithAuthzJobsAndTransactions(repo domain.ListingRepository, authzService AuthzService, jobs domain.SearchIndexJobRepository, txm domain.SearchIndexTransactionManager, storage ...domain.ListingImageStorage) ListingService {
	service := &listingService{repo: repo, authz: authzService, jobs: jobs, txm: txm}
	service.applyStorage(storage)
	return service
}

func NewListingServiceWithAuthzPublisherJobsAndTransactions(repo domain.ListingRepository, authzService AuthzService, publisher domain.EventPublisher, jobs domain.SearchIndexJobRepository, txm domain.SearchIndexTransactionManager, storage ...domain.ListingImageStorage) ListingService {
	service := &listingService{repo: repo, authz: authzService, publish: publisher, jobs: jobs, txm: txm}
	service.applyStorage(storage)
	return service
}

func NewListingServiceWithAuthzPublisherAndJobs(repo domain.ListingRepository, authzService AuthzService, publisher domain.EventPublisher, jobs domain.SearchIndexJobRepository, storage ...domain.ListingImageStorage) ListingService {
	service := &listingService{repo: repo, authz: authzService, publish: publisher, jobs: jobs}
	service.applyStorage(storage)
	return service
}

func WithRegionLookupService(svc ListingService, regions RegionLookupService) ListingService {
	impl, ok := svc.(*listingService)
	if !ok {
		return svc
	}
	impl.regions = regions
	return impl
}

func (s *listingService) Create(ctx context.Context, principal pkgauthz.Principal, req *request.CreateListingRequest) (*response.ListingResponse, error) {
	if req.Title == "" || req.Price <= 0 || req.Status == "" {
		return nil, domain.ErrInvalidCredential
	}
	if err := pkgvalidator.Validate.Struct(req); err != nil {
		return nil, domain.ErrInvalidCredential
	}
	// Generate unique slug
	listingSlug := slug.GenerateUnique(req.Title, func(candidate string) bool {
		exists, _ := s.repo.ExistsBySlug(ctx, candidate)
		return exists
	})

	specsJSON, err := json.Marshal(buildCompatibilitySpecifications(req.BedroomCount, req.BathroomCount, req.LandAreaSqm, req.BuildingAreaSqm, &req.Specifications))
	if err != nil {
		return nil, err
	}
	specialOffersJSON, err := marshalStringArray(req.SpecialOffers)
	if err != nil {
		return nil, err
	}
	facilitiesJSON, err := marshalStringArray(req.Facilities)
	if err != nil {
		return nil, err
	}
	location, err := s.normalizeLocation(ctx, req.LocationProvince, req.LocationProvinceCode, req.LocationCity, req.LocationCityCode, req.LocationDistrict, req.LocationDistrictCode, req.LocationVillage, req.LocationVillageCode)
	if err != nil {
		return nil, err
	}

	listing := &entity.Listing{
		UserID:               principal.UserID,
		CategoryID:           req.CategoryID,
		Title:                req.Title,
		Slug:                 listingSlug,
		Description:          req.Description,
		TransactionType:      defaultString(req.TransactionType, "sale"),
		Price:                req.Price,
		Currency:             defaultStringPointer(req.Currency, "IDR"),
		IsNegotiable:         defaultBool(req.IsNegotiable),
		SpecialOffers:        datatypes.JSON(specialOffersJSON),
		LocationProvince:     location.ProvinceName,
		LocationProvinceCode: location.ProvinceCode,
		LocationCity:         location.CityName,
		LocationCityCode:     location.CityCode,
		LocationDistrict:     location.DistrictName,
		LocationDistrictCode: location.DistrictCode,
		LocationVillage:      location.VillageName,
		LocationVillageCode:  location.VillageCode,
		AddressDetail:        req.AddressDetail,
		Latitude:             req.Latitude,
		Longitude:            req.Longitude,
		BedroomCount:         resolveIntFieldWithLegacy(req.BedroomCount, req.Specifications.Bedrooms, req.Specifications.HasBedrooms()),
		BathroomCount:        resolveIntFieldWithLegacy(req.BathroomCount, req.Specifications.Bathrooms, req.Specifications.HasBathrooms()),
		FloorCount:           req.FloorCount,
		CarportCapacity:      req.CarportCapacity,
		LandAreaSqm:          resolveIntFieldWithLegacy(req.LandAreaSqm, req.Specifications.LandAreaSqm, req.Specifications.HasLandAreaSqm()),
		BuildingAreaSqm:      resolveIntFieldWithLegacy(req.BuildingAreaSqm, req.Specifications.BuildingAreaSqm, req.Specifications.HasBuildingAreaSqm()),
		CertificateType:      req.CertificateType,
		Condition:            req.Condition,
		Furnishing:           req.Furnishing,
		ElectricalPowerVA:    req.ElectricalPowerVA,
		FacingDirection:      req.FacingDirection,
		YearBuilt:            req.YearBuilt,
		Facilities:           datatypes.JSON(facilitiesJSON),
		Status:               req.Status,
		Specifications:       datatypes.JSON(specsJSON),
	}

	var created *entity.Listing
	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			var txErr error
			created, txErr = store.Listings().Create(ctx, listing)
			if txErr != nil {
				return txErr
			}
			created, txErr = store.Listings().FindByID(ctx, created.ID)
			if txErr != nil {
				return txErr
			}
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingCreated, created)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		created, err = s.repo.Create(ctx, listing)
		if err == nil {
			created, err = s.repo.FindByID(ctx, created.ID)
		}
		if err == nil && s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingCreated, created)
		}
	}
	if err != nil {
		return nil, err
	}
	s.publishListingEvent(ctx, domain.EventTypeListingCreated, created)

	return s.mapToResponse(created), nil
}

func (s *listingService) GetByID(ctx context.Context, id uuid.UUID) (*response.ListingResponse, error) {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !isPublicListingStatus(listing.Status) {
		return nil, domain.ErrNotFound
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
	if !isPublicListingStatus(listing.Status) {
		return nil, domain.ErrNotFound
	}

	// Increment view count atomically
	_ = s.repo.IncrementViewCount(ctx, listing.ID)
	listing.ViewCount++ // Update local object for response mapping

	return s.mapToResponse(listing), nil
}

func (s *listingService) Update(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, req *request.UpdateListingRequest) (*response.ListingResponse, error) {
	if err := pkgvalidator.Validate.Struct(req); err != nil {
		return nil, domain.ErrInvalidCredential
	}

	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.authorizeListingAction(listing, principal, pkgauthz.ActionUpdate); err != nil {
		return nil, err
	}

	fields := []string{}

	if req.CategoryID != nil {
		listing.CategoryID = req.CategoryID
		fields = append(fields, "category_id")
	}

	if req.TransactionType != nil {
		listing.TransactionType = *req.TransactionType
		fields = append(fields, "transaction_type")
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

	if req.Currency != nil {
		listing.Currency = *req.Currency
		fields = append(fields, "currency")
	}

	if req.IsNegotiable != nil {
		listing.IsNegotiable = *req.IsNegotiable
		fields = append(fields, "is_negotiable")
	}

	if req.SpecialOffers != nil {
		specialOffersJSON, err := marshalStringArray(*req.SpecialOffers)
		if err != nil {
			return nil, err
		}
		listing.SpecialOffers = datatypes.JSON(specialOffersJSON)
		fields = append(fields, "special_offers")
	}

	if hasLocationUpdates(req) {
		location, err := s.normalizeLocation(ctx, req.LocationProvince, req.LocationProvinceCode, req.LocationCity, req.LocationCityCode, req.LocationDistrict, req.LocationDistrictCode, req.LocationVillage, req.LocationVillageCode)
		if err != nil {
			return nil, err
		}
		listing.LocationProvince = location.ProvinceName
		listing.LocationProvinceCode = location.ProvinceCode
		listing.LocationCity = location.CityName
		listing.LocationCityCode = location.CityCode
		listing.LocationDistrict = location.DistrictName
		listing.LocationDistrictCode = location.DistrictCode
		listing.LocationVillage = location.VillageName
		listing.LocationVillageCode = location.VillageCode
		fields = append(fields, "location_province", "location_province_code", "location_city", "location_city_code", "location_district", "location_district_code", "location_village", "location_village_code")
	}

	if req.AddressDetail != nil {
		listing.AddressDetail = req.AddressDetail
		fields = append(fields, "address_detail")
	}

	if req.Latitude != nil {
		listing.Latitude = req.Latitude
		fields = append(fields, "latitude")
	}

	if req.Longitude != nil {
		listing.Longitude = req.Longitude
		fields = append(fields, "longitude")
	}

	compatSpecs := mergeCompatibilitySpecifications(readCompatibilitySpecifications(listing.Specifications), req.Specifications, req.BedroomCount, req.BathroomCount, req.LandAreaSqm, req.BuildingAreaSqm)

	if req.BedroomCount != nil || req.Specifications != nil {
		listing.BedroomCount = resolveIntFieldWithLegacy(req.BedroomCount, compatSpecs.Bedrooms, compatSpecs.HasBedrooms())
		fields = append(fields, "bedroom_count")
	}

	if req.BathroomCount != nil || req.Specifications != nil {
		listing.BathroomCount = resolveIntFieldWithLegacy(req.BathroomCount, compatSpecs.Bathrooms, compatSpecs.HasBathrooms())
		fields = append(fields, "bathroom_count")
	}

	if req.FloorCount != nil {
		listing.FloorCount = req.FloorCount
		fields = append(fields, "floor_count")
	}

	if req.CarportCapacity != nil {
		listing.CarportCapacity = req.CarportCapacity
		fields = append(fields, "carport_capacity")
	}

	if req.LandAreaSqm != nil || req.Specifications != nil {
		listing.LandAreaSqm = resolveIntFieldWithLegacy(req.LandAreaSqm, compatSpecs.LandAreaSqm, compatSpecs.HasLandAreaSqm())
		fields = append(fields, "land_area_sqm")
	}

	if req.BuildingAreaSqm != nil || req.Specifications != nil {
		listing.BuildingAreaSqm = resolveIntFieldWithLegacy(req.BuildingAreaSqm, compatSpecs.BuildingAreaSqm, compatSpecs.HasBuildingAreaSqm())
		fields = append(fields, "building_area_sqm")
	}

	if req.CertificateType != nil {
		listing.CertificateType = req.CertificateType
		fields = append(fields, "certificate_type")
	}

	if req.Condition != nil {
		listing.Condition = req.Condition
		fields = append(fields, "condition")
	}

	if req.Furnishing != nil {
		listing.Furnishing = req.Furnishing
		fields = append(fields, "furnishing")
	}

	if req.ElectricalPowerVA != nil {
		listing.ElectricalPowerVA = req.ElectricalPowerVA
		fields = append(fields, "electrical_power_va")
	}

	if req.FacingDirection != nil {
		listing.FacingDirection = req.FacingDirection
		fields = append(fields, "facing_direction")
	}

	if req.YearBuilt != nil {
		listing.YearBuilt = req.YearBuilt
		fields = append(fields, "year_built")
	}

	if req.Facilities != nil {
		facilitiesJSON, err := marshalStringArray(*req.Facilities)
		if err != nil {
			return nil, err
		}
		listing.Facilities = datatypes.JSON(facilitiesJSON)
		fields = append(fields, "facilities")
	}

	if req.Status != nil {
		listing.Status = *req.Status
		fields = append(fields, "status")
	}

	if req.Specifications != nil || req.BedroomCount != nil || req.BathroomCount != nil || req.LandAreaSqm != nil || req.BuildingAreaSqm != nil {
		specsJSON, err := json.Marshal(compatSpecs)
		if err != nil {
			return nil, err
		}
		listing.Specifications = datatypes.JSON(specsJSON)
		fields = append(fields, "specifications")
	}

	if len(fields) == 0 {
		return s.mapToResponse(listing), nil
	}

	var updated *entity.Listing
	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			var txErr error
			updated, txErr = store.Listings().Update(ctx, listing, fields)
			if txErr != nil {
				return txErr
			}
			listingForJob, txErr := store.Listings().FindByID(ctx, updated.ID)
			if txErr != nil {
				listingForJob = updated
			}
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingUpdated, listingForJob)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		updated, err = s.repo.Update(ctx, listing, fields)
		if err == nil && s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingUpdated, s.loadListingForEvent(ctx, updated.ID, updated))
		}
	}
	if err != nil {
		return nil, err
	}
	if s.publish != nil {
		s.publishListingEvent(ctx, domain.EventTypeListingUpdated, s.loadListingForEvent(ctx, updated.ID, updated))
	}

	return s.mapToResponse(updated), nil
}

func (s *listingService) Delete(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal) error {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.authorizeListingAction(listing, principal, pkgauthz.ActionDelete); err != nil {
		return err
	}

	deleted := *listing
	now := time.Now().UTC()
	deleted.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			if txErr := store.Listings().Delete(ctx, id); txErr != nil {
				return txErr
			}
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingDeleted, &deleted)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		err = s.repo.Delete(ctx, id)
		if err == nil && s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingDeleted, &deleted)
		}
	}
	if err != nil {
		return err
	}
	s.publishListingEvent(ctx, domain.EventTypeListingDeleted, &deleted)
	return nil
}

func (s *listingService) UploadImage(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, files []*multipart.FileHeader) (*response.ListingResponse, error) {
	if len(files) == 0 {
		return nil, domain.ErrInvalidImageFile
	}
	if err := s.requireImageStorage(); err != nil {
		return nil, err
	}

	listing, err := s.getAuthorizedListing(ctx, id, principal, pkgauthz.ActionUploadImage)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if err := validateListingImageFile(file); err != nil {
			return nil, err
		}
	}

	if len(listing.Images)+len(files) > 10 {
		return nil, domain.ErrImageLimitReached
	}

	uploads := make([]*mediaasset.UploadResult, 0, len(files))
	images := make([]*entity.ListingImage, 0, len(files))
	folder := fmt.Sprintf("listings/%s", id.String())
	for _, file := range files {
		uploaded, uploadErr := s.storage.UploadListingImage(ctx, mediaasset.UploadInput{
			File:         file,
			Folder:       folder,
			PublicID:     uuid.NewString(),
			ResourceType: mediaasset.DefaultResourceType,
			DeliveryType: mediaasset.DefaultDeliveryType,
		})
		if uploadErr != nil {
			s.destroyUploadedImages(ctx, uploads)
			return nil, uploadErr
		}
		uploads = append(uploads, uploaded)
		images = append(images, buildListingImageEntity(id, file, uploaded))
	}

	for _, image := range images {
		listing.Images = append(listing.Images, *image)
	}

	var persistErr error
	if s.txm != nil && s.jobs != nil {
		persistErr = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			for _, image := range images {
				if _, txErr := store.Listings().CreateImage(ctx, image); txErr != nil {
					return txErr
				}
			}
			listingForJob, txErr := store.Listings().FindByID(ctx, id)
			if txErr != nil {
				listingForJob = listing
			}
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingImagesChanged, listingForJob)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		for _, image := range images {
			if _, createErr := s.repo.CreateImage(ctx, image); createErr != nil {
				persistErr = createErr
				break
			}
		}
		if persistErr == nil && s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, listing))
		}
	}
	if persistErr != nil {
		s.destroyUploadedImages(ctx, uploads)
		return nil, persistErr
	}
	if s.publish != nil {
		s.publishListingEvent(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, listing))
	}
	return s.fetchListingResponse(ctx, id)
}
func (s *listingService) UploadVideo(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, file *multipart.FileHeader) (*response.ListingResponse, error) {
	if err := validateListingVideoFile(file); err != nil {
		return nil, err
	}
	if err := s.requireVideoStorage(); err != nil {
		return nil, err
	}

	listing, err := s.getAuthorizedListing(ctx, id, principal, pkgauthz.ActionUploadVideo)
	if err != nil {
		return nil, err
	}

	if listing.Video != nil {
		return nil, domain.ErrVideoAlreadyExists
	}

	uploaded, err := s.videoStorage.UploadListingVideo(ctx, mediaasset.UploadInput{
		File:         file,
		Folder:       fmt.Sprintf("listings/%s", id.String()),
		PublicID:     uuid.NewString(),
		ResourceType: mediaasset.DefaultVideoResourceType,
		DeliveryType: mediaasset.DefaultDeliveryType,
	})
	if err != nil {
		return nil, err
	}

	duration := uploaded.DurationSeconds
	if duration == nil {
		duration = extractDurationSeconds(uploaded.Metadata)
	}
	if duration != nil && *duration > maxVideoDurationSeconds {
		s.destroyUploadedVideo(ctx, uploaded)
		return nil, domain.ErrVideoTooLong
	}

	video := buildListingVideoEntity(id, file, uploaded, duration)

	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			if _, txErr := store.Listings().CreateVideo(ctx, video); txErr != nil {
				return txErr
			}
			listingForJob, txErr := store.Listings().FindByID(ctx, id)
			if txErr != nil {
				listingForJob = listing
				listingForJob.Video = video
			}
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingImagesChanged, listingForJob)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		if _, err := s.repo.CreateVideo(ctx, video); err != nil {
			s.destroyUploadedVideo(ctx, uploaded)
			return nil, err
		}
		if s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, listing))
		}
	}
	if err != nil {
		s.destroyUploadedVideo(ctx, uploaded)
		return nil, err
	}
	if s.publish != nil {
		s.publishListingEvent(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, listing))
	}

	return s.fetchListingResponse(ctx, id)
}

func (s *listingService) DeleteImage(ctx context.Context, id uuid.UUID, imageID uuid.UUID, principal pkgauthz.Principal) (*response.ListingResponse, error) {
	if err := s.requireImageStorage(); err != nil {
		return nil, err
	}

	if _, err := s.getAuthorizedListing(ctx, id, principal, pkgauthz.ActionDeleteImage); err != nil {
		return nil, err
	}

	image, err := s.repo.FindImageByID(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if image.ListingID != id {
		return nil, domain.ErrNotFound
	}

	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			if txErr := store.Listings().DeleteImage(ctx, id, imageID); txErr != nil {
				return txErr
			}
			listingForJob, txErr := store.Listings().FindByID(ctx, id)
			if txErr != nil {
				return txErr
			}
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingImagesChanged, listingForJob)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		if err := s.repo.DeleteImage(ctx, id, imageID); err != nil {
			return nil, err
		}
		if s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, nil))
		}
	}
	if err != nil {
		return nil, err
	}
	s.destroyListingImageByEntity(ctx, image)
	if s.publish != nil {
		s.publishListingEvent(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, nil))
	}

	return s.fetchListingResponse(ctx, id)
}
func (s *listingService) DeleteVideo(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal) (*response.ListingResponse, error) {
	if err := s.requireVideoStorage(); err != nil {
		return nil, err
	}

	listing, err := s.getAuthorizedListing(ctx, id, principal, pkgauthz.ActionDeleteVideo)
	if err != nil {
		return nil, err
	}

	video, err := s.repo.FindVideoByListingID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			if txErr := store.Listings().DeleteVideoByListingID(ctx, id); txErr != nil {
				return txErr
			}
			listingForJob, txErr := store.Listings().FindByID(ctx, id)
			if txErr != nil {
				listingForJob = listing
			}
			listingForJob.Video = nil
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingImagesChanged, listingForJob)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		if err := s.repo.DeleteVideoByListingID(ctx, id); err != nil {
			return nil, err
		}
		if s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, listing))
		}
	}
	if err != nil {
		return nil, err
	}
	s.destroyListingVideoByEntity(ctx, video)
	if s.publish != nil {
		s.publishListingEvent(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, listing))
	}
	return s.fetchListingResponse(ctx, id)
}

func (s *listingService) SetPrimaryImage(ctx context.Context, id uuid.UUID, imageID uuid.UUID, principal pkgauthz.Principal) (*response.ListingResponse, error) {
	if _, err := s.getAuthorizedListing(ctx, id, principal, pkgauthz.ActionSetPrimaryImage); err != nil {
		return nil, err
	}

	image, err := s.repo.FindImageByID(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if image.ListingID != id {
		return nil, domain.ErrNotFound
	}

	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			if txErr := store.Listings().SetPrimaryImage(ctx, id, imageID); txErr != nil {
				return txErr
			}
			listingForJob, txErr := store.Listings().FindByID(ctx, id)
			if txErr != nil {
				return txErr
			}
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingImagesChanged, listingForJob)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		if err := s.repo.SetPrimaryImage(ctx, id, imageID); err != nil {
			return nil, err
		}
		if s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, nil))
		}
	}
	if err != nil {
		return nil, err
	}
	if s.publish != nil {
		s.publishListingEvent(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, nil))
	}

	return s.fetchListingResponse(ctx, id)
}

func (s *listingService) ReorderImages(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, orderedImageIDs []uuid.UUID) (*response.ListingResponse, error) {
	if _, err := s.getAuthorizedListing(ctx, id, principal, pkgauthz.ActionReorderImages); err != nil {
		return nil, err
	}

	images, err := s.repo.ListActiveImagesByListingID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := validateListingImageOrder(images, orderedImageIDs); err != nil {
		return nil, err
	}

	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			if txErr := store.Listings().ReorderImages(ctx, id, orderedImageIDs); txErr != nil {
				return txErr
			}
			listingForJob, txErr := store.Listings().FindByID(ctx, id)
			if txErr != nil {
				return txErr
			}
			job, txErr := buildSearchIndexJobFromListing(domain.EventTypeListingImagesChanged, listingForJob)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		if err := s.repo.ReorderImages(ctx, id, orderedImageIDs); err != nil {
			return nil, err
		}
		if s.jobs != nil {
			s.enqueueListingIndexJob(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, nil))
		}
	}
	if err != nil {
		return nil, err
	}
	if s.publish != nil {
		s.publishListingEvent(ctx, domain.EventTypeListingImagesChanged, s.loadListingForEvent(ctx, id, nil))
	}

	return s.fetchListingResponse(ctx, id)
}

func (s *listingService) List(ctx context.Context, filter domain.ListingFilter) (*response.PaginatedListings, error) {
	if filter.Status != "" && !isPublicListingStatus(filter.Status) {
		return &response.PaginatedListings{
			Data:       []*response.ListingResponse{},
			Total:      0,
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalPages: 0,
		}, nil
	}
	if filter.Status == "" {
		filter.Statuses = publicListingStatuses()
	}
	listings, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return s.mapToPaginatedResponse(listings, total, filter.Page, filter.Limit), nil
}

func (s *listingService) ListByUserID(ctx context.Context, principal pkgauthz.Principal, filter domain.ListingFilter) (*response.PaginatedListings, error) {
	listings, total, err := s.repo.FindByUserID(ctx, principal.UserID, filter)
	if err != nil {
		return nil, err
	}

	return s.mapToPaginatedResponse(listings, total, filter.Page, filter.Limit), nil
}

func (s *listingService) authorizeListingAction(listing *entity.Listing, principal pkgauthz.Principal, action string) error {
	if s.authz == nil {
		return fmt.Errorf("authz service unavailable")
	}

	return s.authz.EnforceListingAction(principal, listing, action)
}

func (s *listingService) mapToResponse(l *entity.Listing) *response.ListingResponse {
	if l == nil {
		return nil
	}
	compatSpecs := readCompatibilitySpecifications(l.Specifications)
	return &response.ListingResponse{
		ID:                   l.ID,
		UserID:               l.UserID,
		CategoryID:           l.CategoryID,
		Title:                l.Title,
		Slug:                 l.Slug,
		Description:          l.Description,
		TransactionType:      l.TransactionType,
		Price:                l.Price,
		Currency:             l.Currency,
		IsNegotiable:         l.IsNegotiable,
		SpecialOffers:        l.SpecialOffers,
		LocationProvince:     l.LocationProvince,
		LocationProvinceCode: l.LocationProvinceCode,
		LocationCity:         l.LocationCity,
		LocationCityCode:     l.LocationCityCode,
		LocationDistrict:     l.LocationDistrict,
		LocationDistrictCode: l.LocationDistrictCode,
		LocationVillage:      l.LocationVillage,
		LocationVillageCode:  l.LocationVillageCode,
		AddressDetail:        l.AddressDetail,
		Latitude:             l.Latitude,
		Longitude:            l.Longitude,
		BedroomCount:         resolveIntFieldWithLegacy(l.BedroomCount, compatSpecs.Bedrooms, compatSpecs.HasBedrooms()),
		BathroomCount:        resolveIntFieldWithLegacy(l.BathroomCount, compatSpecs.Bathrooms, compatSpecs.HasBathrooms()),
		FloorCount:           l.FloorCount,
		CarportCapacity:      l.CarportCapacity,
		LandAreaSqm:          resolveIntFieldWithLegacy(l.LandAreaSqm, compatSpecs.LandAreaSqm, compatSpecs.HasLandAreaSqm()),
		BuildingAreaSqm:      resolveIntFieldWithLegacy(l.BuildingAreaSqm, compatSpecs.BuildingAreaSqm, compatSpecs.HasBuildingAreaSqm()),
		CertificateType:      l.CertificateType,
		Condition:            l.Condition,
		Furnishing:           l.Furnishing,
		ElectricalPowerVA:    l.ElectricalPowerVA,
		FacingDirection:      l.FacingDirection,
		YearBuilt:            l.YearBuilt,
		Facilities:           l.Facilities,
		Status:               l.Status,
		IsFeatured:           l.IsFeatured,
		Specifications:       l.Specifications,
		ViewCount:            l.ViewCount,
		Images:               mapListingImagesToResponse(l.Images),
		Video:                mapListingVideoToResponse(l.Video),
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

func buildCompatibilitySpecifications(bedroomCount, bathroomCount, landAreaSqm, buildingAreaSqm *int, legacy *request.Specifications) request.Specifications {
	specs := request.Specifications{}
	if legacy != nil {
		specs = *legacy
	}
	if bedroomCount != nil {
		specs.SetBedrooms(*bedroomCount)
	}
	if bathroomCount != nil {
		specs.SetBathrooms(*bathroomCount)
	}
	if landAreaSqm != nil {
		specs.SetLandAreaSqm(*landAreaSqm)
	}
	if buildingAreaSqm != nil {
		specs.SetBuildingAreaSqm(*buildingAreaSqm)
	}
	return specs
}

func mergeCompatibilitySpecifications(existing request.Specifications, legacy *request.Specifications, bedroomCount, bathroomCount, landAreaSqm, buildingAreaSqm *int) request.Specifications {
	specs := existing
	if legacy != nil {
		if legacy.HasBedrooms() || legacy.Bedrooms != 0 {
			specs.SetBedrooms(legacy.Bedrooms)
		}
		if legacy.HasBathrooms() || legacy.Bathrooms != 0 {
			specs.SetBathrooms(legacy.Bathrooms)
		}
		if legacy.HasLandAreaSqm() || legacy.LandAreaSqm != 0 {
			specs.SetLandAreaSqm(legacy.LandAreaSqm)
		}
		if legacy.HasBuildingAreaSqm() || legacy.BuildingAreaSqm != 0 {
			specs.SetBuildingAreaSqm(legacy.BuildingAreaSqm)
		}
	}
	if bedroomCount != nil {
		specs.SetBedrooms(*bedroomCount)
	}
	if bathroomCount != nil {
		specs.SetBathrooms(*bathroomCount)
	}
	if landAreaSqm != nil {
		specs.SetLandAreaSqm(*landAreaSqm)
	}
	if buildingAreaSqm != nil {
		specs.SetBuildingAreaSqm(*buildingAreaSqm)
	}
	return specs
}

func readCompatibilitySpecifications(raw datatypes.JSON) request.Specifications {
	if len(raw) == 0 {
		return request.Specifications{}
	}
	var specs request.Specifications
	if err := json.Unmarshal(raw, &specs); err != nil {
		return request.Specifications{}
	}
	return specs
}

func marshalStringArray(values []string) ([]byte, error) {
	if len(values) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(values)
}

func resolveIntField(explicit *int, legacy int) *int {
	if explicit != nil {
		return explicit
	}
	if legacy == 0 {
		return nil
	}
	value := legacy
	return &value
}

func resolveIntFieldWithLegacy(explicit *int, legacy int, legacyPresent bool) *int {
	if explicit != nil {
		return explicit
	}
	if !legacyPresent && legacy == 0 {
		return nil
	}
	value := legacy
	return &value
}

func defaultBool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func defaultStringPointer(value *string, fallback string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}
	return *value
}

func hasLocationUpdates(req *request.UpdateListingRequest) bool {
	return req.LocationProvince != nil || req.LocationProvinceCode != nil || req.LocationCity != nil || req.LocationCityCode != nil || req.LocationDistrict != nil || req.LocationDistrictCode != nil || req.LocationVillage != nil || req.LocationVillageCode != nil
}

func (s *listingService) normalizeLocation(ctx context.Context, provinceName, provinceCode, cityName, cityCode, districtName, districtCode, villageName, villageCode *string) (*normalizedListingLocation, error) {
	if hasAnyLocationCode(provinceCode, cityCode, districtCode, villageCode) {
		if s.regions == nil {
			return nil, domain.ErrInvalidLocation
		}
		selection, err := s.regions.ResolveHierarchy(ctx, normalizePointerValue(provinceCode), normalizePointerValue(cityCode), normalizePointerValue(districtCode), normalizePointerValue(villageCode))
		if err != nil {
			return nil, err
		}
		return &normalizedListingLocation{
			ProvinceName: stringPointer(selection.Province.Name),
			ProvinceCode: stringPointer(selection.Province.Code),
			CityName:     stringPointer(selection.City.Name),
			CityCode:     stringPointer(selection.City.Code),
			DistrictName: stringPointer(selection.District.Name),
			DistrictCode: stringPointer(selection.District.Code),
			VillageName:  stringPointer(selection.Village.Name),
			VillageCode:  stringPointer(selection.Village.Code),
		}, nil
	}

	return &normalizedListingLocation{
		ProvinceName: normalizeOptionalPointer(provinceName),
		ProvinceCode: nil,
		CityName:     normalizeOptionalPointer(cityName),
		CityCode:     nil,
		DistrictName: normalizeOptionalPointer(districtName),
		DistrictCode: nil,
		VillageName:  normalizeOptionalPointer(villageName),
		VillageCode:  nil,
	}, nil
}

func hasAnyLocationCode(values ...*string) bool {
	for _, value := range values {
		if strings.TrimSpace(normalizePointerValue(value)) != "" {
			return true
		}
	}
	return false
}

func normalizeOptionalPointer(value *string) *string {
	trimmed := normalizePointerValue(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizePointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func stringPointer(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func publicListingStatuses() []string {
	return []string{"active", "sold"}
}

func isPublicListingStatus(status string) bool {
	return slices.Contains(publicListingStatuses(), status)
}

func (s *listingService) publishListingEvent(ctx context.Context, eventType string, listing *entity.Listing) {
	if s.publish == nil || listing == nil {
		return
	}
	if err := s.publish.PublishListingEvent(ctx, buildListingEvent(eventType, listing)); err != nil {
		logger.Log.Warn("Failed to publish listing event", zap.String("event_type", eventType), zap.String("listing_id", listing.ID.String()), zap.Error(err))
	}
}

func (s *listingService) enqueueListingIndexJob(ctx context.Context, eventType string, listing *entity.Listing) {
	if s.jobs == nil || listing == nil {
		return
	}
	job, err := buildSearchIndexJobFromListing(eventType, listing)
	if err != nil {
		logger.Log.Warn("Failed to build search index job for listing", zap.String("event_type", eventType), zap.String("listing_id", listing.ID.String()), zap.Error(err))
		return
	}
	if _, err := s.jobs.Enqueue(ctx, job); err != nil {
		logger.Log.Warn("Failed to enqueue search index job for listing", zap.String("event_type", eventType), zap.String("listing_id", listing.ID.String()), zap.Error(err))
	}
}

func (s *listingService) loadListingForEvent(ctx context.Context, id uuid.UUID, fallback *entity.Listing) *entity.Listing {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fallback
	}
	return listing
}

func (s *listingService) getAuthorizedListing(ctx context.Context, id uuid.UUID, principal pkgauthz.Principal, action string) (*entity.Listing, error) {
	listing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.authorizeListingAction(listing, principal, action); err != nil {
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

func (s *listingService) destroyUploadedImages(ctx context.Context, uploads []*mediaasset.UploadResult) {
	for _, uploaded := range uploads {
		s.destroyUploadedImage(ctx, uploaded)
	}
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

func mapListingVideoToResponse(video *entity.ListingVideo) *response.ListingVideoResponse {
	if video == nil {
		return nil
	}

	return &response.ListingVideoResponse{
		ID:               video.ID,
		URL:              video.URL,
		Format:           video.Format,
		Bytes:            video.Bytes,
		Width:            video.Width,
		Height:           video.Height,
		DurationSeconds:  video.DurationSeconds,
		OriginalFilename: video.OriginalFilename,
		CreatedAt:        video.CreatedAt,
	}
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

const (
	maxVideoBytes           = 100 * 1024 * 1024
	maxVideoDurationSeconds = 60
)

var allowedVideoExtensions = map[string]struct{}{
	".mp4":  {},
	".mov":  {},
	".m4v":  {},
	".webm": {},
	".mkv":  {},
	".flv":  {},
	".avi":  {},
	".mpg":  {},
	".mpeg": {},
	".ogv":  {},
}

func validateListingVideoFile(file *multipart.FileHeader) error {
	if file == nil {
		return domain.ErrInvalidVideoFile
	}

	if file.Size > maxVideoBytes {
		return domain.ErrVideoTooLarge
	}

	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	if strings.HasPrefix(strings.ToLower(contentType), "video/") {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedVideoExtensions[ext]; ok {
		return nil
	}

	return domain.ErrInvalidVideoFile
}

func (s *listingService) requireVideoStorage() error {
	if s.videoStorage == nil {
		return domain.ErrVideoStorageUnset
	}
	return nil
}

func (s *listingService) destroyUploadedVideo(ctx context.Context, uploaded *mediaasset.UploadResult) {
	if s.videoStorage == nil || uploaded == nil || uploaded.PublicID == "" {
		return
	}

	resourceType := uploaded.ResourceType
	if resourceType == "" {
		resourceType = mediaasset.DefaultVideoResourceType
	}
	deliveryType := uploaded.DeliveryType
	if deliveryType == "" {
		deliveryType = mediaasset.DefaultDeliveryType
	}

	_, _ = s.videoStorage.DestroyListingVideo(ctx, mediaasset.DestroyInput{
		PublicID:     uploaded.PublicID,
		ResourceType: resourceType,
		DeliveryType: deliveryType,
		Invalidate:   true,
	})
}

func (s *listingService) destroyListingVideoByEntity(ctx context.Context, video *entity.ListingVideo) {
	if s.videoStorage == nil || video == nil || video.PublicID == nil || *video.PublicID == "" {
		return
	}

	resourceType := mediaasset.DefaultVideoResourceType
	if video.ResourceType != nil && *video.ResourceType != "" {
		resourceType = *video.ResourceType
	}
	deliveryType := mediaasset.DefaultDeliveryType
	if video.DeliveryType != nil && *video.DeliveryType != "" {
		deliveryType = *video.DeliveryType
	}

	_, _ = s.videoStorage.DestroyListingVideo(ctx, mediaasset.DestroyInput{
		PublicID:     *video.PublicID,
		ResourceType: resourceType,
		DeliveryType: deliveryType,
		Invalidate:   true,
	})
}

func buildListingVideoEntity(listingID uuid.UUID, file *multipart.FileHeader, uploaded *mediaasset.UploadResult, durationSeconds *int) *entity.ListingVideo {
	video := &entity.ListingVideo{
		ListingID:       listingID,
		URL:             uploaded.SecureURL,
		AssetID:         stringPointer(uploaded.AssetID),
		PublicID:        stringPointer(uploaded.PublicID),
		Version:         int64Pointer(uploaded.Version),
		Format:          stringPointer(uploaded.Format),
		Bytes:           int64Pointer(uploaded.Bytes),
		Width:           intPointer(uploaded.Width),
		Height:          intPointer(uploaded.Height),
		ResourceType:    stringPointer(uploaded.ResourceType),
		DeliveryType:    stringPointer(uploaded.DeliveryType),
		DurationSeconds: durationSeconds,
	}
	original := uploaded.OriginalFilename
	if original == "" && file != nil {
		original = file.Filename
	}
	if original != "" {
		video.OriginalFilename = stringPointer(original)
	}
	return video
}

func extractDurationSeconds(metadata mediaasset.Metadata) *int {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata["duration"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case int:
		return &v
	case int64:
		value := int(v)
		return &value
	case float64:
		value := int(math.Ceil(v))
		return &value
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil
		}
		value := int(math.Ceil(parsed))
		return &value
	default:
		return nil
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
