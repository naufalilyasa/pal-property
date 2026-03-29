package service_test

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/textproto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/mocks"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/naufalilyasa/pal-property-backend/pkg/mediaasset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

type fakeEventPublisher struct {
	listingEvents  []domain.ListingEvent
	categoryEvents []domain.CategoryEvent
}

func (f *fakeEventPublisher) PublishListingEvent(_ context.Context, event domain.ListingEvent) error {
	f.listingEvents = append(f.listingEvents, event)
	return nil
}

func (f *fakeEventPublisher) PublishCategoryEvent(_ context.Context, event domain.CategoryEvent) error {
	f.categoryEvents = append(f.categoryEvents, event)
	return nil
}

func TestListingService_Create_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	publisher := &fakeEventPublisher{}
	svc := service.NewListingServiceWithAuthzAndPublisher(repo, newTestAuthzService(t), publisher)

	userID := uuid.New()
	categoryID := uuid.New()
	createdID := uuid.New()
	req := &request.CreateListingRequest{
		CategoryID:   &categoryID,
		Title:        "Modern Villa in Bali",
		Description:  ptr("Luxury villa with pool"),
		Price:        5000000000,
		LocationCity: ptr("Badung"),
		Status:       "active",
		Specifications: request.Specifications{
			Bedrooms:    3,
			Bathrooms:   3,
			LandAreaSqm: 200,
		},
	}

	repo.On("ExistsBySlug", mock.Anything, "modern-villa-in-bali").Return(false, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.Listing) bool {
		return l.Title == req.Title && l.UserID == userID && l.Slug == "modern-villa-in-bali"
	})).Return(&entity.Listing{
		BaseEntity:     entity.BaseEntity{ID: createdID, CreatedAt: time.Now()},
		UserID:         userID,
		CategoryID:     req.CategoryID,
		Title:          req.Title,
		Slug:           "modern-villa-in-bali",
		Description:    req.Description,
		Price:          req.Price,
		LocationCity:   req.LocationCity,
		Status:         req.Status,
		Specifications: datatypes.JSON(`{"bedrooms":3,"bathrooms":3,"land_area_sqm":200,"building_area_sqm":0}`),
	}, nil)
	repo.On("FindByID", mock.Anything, createdID).Return(&entity.Listing{
		BaseEntity:     entity.BaseEntity{ID: createdID, CreatedAt: time.Now()},
		UserID:         userID,
		CategoryID:     req.CategoryID,
		Title:          req.Title,
		Slug:           "modern-villa-in-bali",
		Description:    req.Description,
		Price:          req.Price,
		LocationCity:   req.LocationCity,
		Status:         req.Status,
		Specifications: datatypes.JSON(`{"bedrooms":3,"bathrooms":3,"land_area_sqm":200,"building_area_sqm":0}`),
	}, nil)

	res, err := svc.Create(context.Background(), pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "modern-villa-in-bali", res.Slug)
	assert.Equal(t, req.Title, res.Title)
	assert.Len(t, publisher.listingEvents, 1)
	assert.Equal(t, domain.EventTypeListingCreated, publisher.listingEvents[0].Metadata.EventType)
	assert.Equal(t, req.Title, publisher.listingEvents[0].Payload.Title)
}

func TestListingService_Create_SlugCollision(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	userID := uuid.New()
	createdID := uuid.New()
	req := &request.CreateListingRequest{
		Title:  "Test Listing",
		Price:  1000000,
		Status: "active",
	}

	// First attempt exists, second one (with suffix) does not
	repo.On("ExistsBySlug", mock.Anything, "test-listing").Return(true, nil).Once()
	repo.On("ExistsBySlug", mock.Anything, mock.MatchedBy(func(s string) bool {
		return len(s) > len("test-listing")
	})).Return(false, nil).Once()

	repo.On("Create", mock.Anything, mock.Anything).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: createdID},
		Title:      req.Title,
		Slug:       "test-listing-random",
	}, nil)
	repo.On("FindByID", mock.Anything, createdID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: createdID},
		Title:      req.Title,
		Slug:       "test-listing-random",
		Price:      req.Price,
		Status:     req.Status,
	}, nil)

	res, err := svc.Create(context.Background(), pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEqual(t, "test-listing", res.Slug)
}

func TestListingService_Create_RepoError(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	userID := uuid.New()
	req := &request.CreateListingRequest{
		Title:  "Test Listing",
		Price:  1000000,
		Status: "active",
	}

	repo.On("ExistsBySlug", mock.Anything, mock.Anything).Return(false, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrConflict)

	res, err := svc.Create(context.Background(), pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrConflict, err)
}

func TestListingService_Create_ExpandedFields_PreservesCompatibilitySpecifications(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	userID := uuid.New()
	categoryID := uuid.New()
	createdID := uuid.New()
	negotiable := true
	province := "Jawa Barat"
	city := "Bogor"
	bedrooms := 4
	bathrooms := 3
	landArea := 180
	buildingArea := 140
	electricalPower := 5500
	transactionType := "sale"
	latitude := -6.5971
	longitude := 106.8060
	req := &request.CreateListingRequest{
		CategoryID:        &categoryID,
		Title:             "Rumah Keluarga Besar Bogor",
		Description:       ptr("Rumah luas dengan fasilitas lengkap"),
		TransactionType:   transactionType,
		Price:             1750000000,
		IsNegotiable:      &negotiable,
		SpecialOffers:     []string{"Promo", "Turun_Harga"},
		LocationProvince:  &province,
		LocationCity:      &city,
		Latitude:          &latitude,
		Longitude:         &longitude,
		BedroomCount:      &bedrooms,
		BathroomCount:     &bathrooms,
		LandAreaSqm:       &landArea,
		BuildingAreaSqm:   &buildingArea,
		ElectricalPowerVA: &electricalPower,
		Facilities:        []string{"AC", "CCTV"},
		Status:            "draft",
		Specifications: request.Specifications{
			Bedrooms:        1,
			Bathrooms:       1,
			LandAreaSqm:     20,
			BuildingAreaSqm: 15,
		},
	}

	repo.On("ExistsBySlug", mock.Anything, "rumah-keluarga-besar-bogor").Return(false, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.Listing) bool {
		if l.TransactionType != transactionType || !l.IsNegotiable || l.Status != "draft" {
			return false
		}
		if l.LocationProvince == nil || *l.LocationProvince != province {
			return false
		}
		if l.BedroomCount == nil || *l.BedroomCount != bedrooms {
			return false
		}
		if l.BuildingAreaSqm == nil || *l.BuildingAreaSqm != buildingArea {
			return false
		}
		var specs request.Specifications
		if err := json.Unmarshal(l.Specifications, &specs); err != nil {
			return false
		}
		return specs.Bedrooms == bedrooms && specs.Bathrooms == bathrooms && specs.LandAreaSqm == landArea && specs.BuildingAreaSqm == buildingArea
	})).Return(&entity.Listing{
		BaseEntity:        entity.BaseEntity{ID: createdID, CreatedAt: time.Now()},
		UserID:            userID,
		CategoryID:        &categoryID,
		Title:             req.Title,
		Slug:              "rumah-keluarga-besar-bogor",
		Description:       req.Description,
		TransactionType:   transactionType,
		Price:             req.Price,
		Currency:          "IDR",
		IsNegotiable:      true,
		SpecialOffers:     datatypes.JSON(`["Promo","Turun_Harga"]`),
		LocationProvince:  &province,
		LocationCity:      &city,
		Latitude:          &latitude,
		Longitude:         &longitude,
		BedroomCount:      &bedrooms,
		BathroomCount:     &bathrooms,
		LandAreaSqm:       &landArea,
		BuildingAreaSqm:   &buildingArea,
		ElectricalPowerVA: &electricalPower,
		Facilities:        datatypes.JSON(`["AC","CCTV"]`),
		Status:            "draft",
		Specifications:    datatypes.JSON(`{"bedrooms":4,"bathrooms":3,"land_area_sqm":180,"building_area_sqm":140}`),
	}, nil)
	repo.On("FindByID", mock.Anything, createdID).Return(&entity.Listing{
		BaseEntity:        entity.BaseEntity{ID: createdID, CreatedAt: time.Now()},
		UserID:            userID,
		CategoryID:        &categoryID,
		Title:             req.Title,
		Slug:              "rumah-keluarga-besar-bogor",
		Description:       req.Description,
		TransactionType:   transactionType,
		Price:             req.Price,
		Currency:          "IDR",
		IsNegotiable:      true,
		SpecialOffers:     datatypes.JSON(`["Promo","Turun_Harga"]`),
		LocationProvince:  &province,
		LocationCity:      &city,
		Latitude:          &latitude,
		Longitude:         &longitude,
		BedroomCount:      &bedrooms,
		BathroomCount:     &bathrooms,
		LandAreaSqm:       &landArea,
		BuildingAreaSqm:   &buildingArea,
		ElectricalPowerVA: &electricalPower,
		Facilities:        datatypes.JSON(`["AC","CCTV"]`),
		Status:            "draft",
		Specifications:    datatypes.JSON(`{"bedrooms":4,"bathrooms":3,"land_area_sqm":180,"building_area_sqm":140}`),
	}, nil)

	res, err := svc.Create(context.Background(), pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, transactionType, res.TransactionType)
	assert.Equal(t, province, *res.LocationProvince)
	assert.Equal(t, bedrooms, *res.BedroomCount)
	assert.Equal(t, buildingArea, *res.BuildingAreaSqm)
	assert.Equal(t, datatypes.JSON(`["Promo","Turun_Harga"]`), res.SpecialOffers)
}

func TestListingService_Update_ExpandedFields_RegeneratesCompatibilitySpecifications(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	userID := uuid.New()
	transactionType := "rent"
	province := "Banten"
	bedrooms := 5
	bathrooms := 4
	landArea := 250
	buildingArea := 210
	facilities := []string{"Wifi", "Water_Heater"}
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
		Status:     "active",
	}
	req := &request.UpdateListingRequest{
		TransactionType:  &transactionType,
		LocationProvince: &province,
		BedroomCount:     &bedrooms,
		BathroomCount:    &bathrooms,
		LandAreaSqm:      &landArea,
		BuildingAreaSqm:  &buildingArea,
		Facilities:       &facilities,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil).Once()
	repo.On("Update", mock.Anything, mock.MatchedBy(func(l *entity.Listing) bool {
		if l.TransactionType != transactionType {
			return false
		}
		if l.LocationProvince == nil || *l.LocationProvince != province {
			return false
		}
		if l.BedroomCount == nil || *l.BedroomCount != bedrooms {
			return false
		}
		var specs request.Specifications
		if err := json.Unmarshal(l.Specifications, &specs); err != nil {
			return false
		}
		return specs.Bedrooms == bedrooms && specs.Bathrooms == bathrooms && specs.LandAreaSqm == landArea && specs.BuildingAreaSqm == buildingArea
	}), mock.MatchedBy(func(fields []string) bool {
		required := map[string]bool{
			"transaction_type":  false,
			"location_province": false,
			"bedroom_count":     false,
			"bathroom_count":    false,
			"land_area_sqm":     false,
			"building_area_sqm": false,
			"facilities":        false,
			"specifications":    false,
		}
		for _, field := range fields {
			if _, ok := required[field]; ok {
				required[field] = true
			}
		}
		for _, seen := range required {
			if !seen {
				return false
			}
		}
		return true
	})).Return(&entity.Listing{
		BaseEntity:       entity.BaseEntity{ID: id},
		UserID:           userID,
		TransactionType:  transactionType,
		LocationProvince: &province,
		BedroomCount:     &bedrooms,
		BathroomCount:    &bathrooms,
		LandAreaSqm:      &landArea,
		BuildingAreaSqm:  &buildingArea,
		Facilities:       datatypes.JSON(`["Wifi","Water_Heater"]`),
		Status:           "active",
		Specifications:   datatypes.JSON(`{"bedrooms":5,"bathrooms":4,"land_area_sqm":250,"building_area_sqm":210}`),
	}, nil)

	res, err := svc.Update(context.Background(), id, pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, transactionType, res.TransactionType)
	assert.Equal(t, bedrooms, *res.BedroomCount)
	assert.Equal(t, buildingArea, *res.BuildingAreaSqm)
	assert.Equal(t, datatypes.JSON(`["Wifi","Water_Heater"]`), res.Facilities)
}

func TestListingService_GetByID_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		Title:      "Test",
		Status:     "active",
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil).Once()
	repo.On("IncrementViewCount", mock.Anything, id).Return(nil).Maybe()

	res, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, id, res.ID)
}

func TestListingService_GetByID_NotFound(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, domain.ErrNotFound)

	res, err := svc.GetByID(context.Background(), id)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrNotFound, err)
}

func TestListingService_GetBySlug_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	slugStr := "test-slug"
	id := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		Slug:       slugStr,
		Status:     "active",
	}

	repo.On("FindBySlug", mock.Anything, slugStr).Return(listing, nil)
	repo.On("IncrementViewCount", mock.Anything, id).Return(nil).Maybe()

	res, err := svc.GetBySlug(context.Background(), slugStr)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, slugStr, res.Slug)
}

func TestListingService_Update_SuccessOwner(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	publisher := &fakeEventPublisher{}
	svc := service.NewListingServiceWithAuthzAndPublisher(repo, newTestAuthzService(t), publisher)

	id := uuid.New()
	userID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
		Title:      "Old Title",
	}

	newTitle := "New Title"
	req := &request.UpdateListingRequest{
		Title: &newTitle,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil).Once()
	repo.On("ExistsBySlug", mock.Anything, "new-title").Return(false, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(l *entity.Listing) bool {
		return l.Title == newTitle
	}), []string{"title", "slug"}).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
		Title:      newTitle,
		Slug:       "new-title",
	}, nil)
	repo.On("FindByID", mock.Anything, id).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
		Title:      newTitle,
		Slug:       "new-title",
	}, nil).Once()

	res, err := svc.Update(context.Background(), id, pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, newTitle, res.Title)
	assert.Len(t, publisher.listingEvents, 1)
	assert.Equal(t, domain.EventTypeListingUpdated, publisher.listingEvents[0].Metadata.EventType)
}

func TestListingService_Update_SuccessAdmin(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	ownerID := uuid.New()
	adminID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     ownerID,
	}

	newPrice := int64(2000000)
	req := &request.UpdateListingRequest{
		Price: &newPrice,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)
	repo.On("Update", mock.Anything, mock.Anything, []string{"price"}).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     ownerID,
		Price:      newPrice,
	}, nil)

	res, err := svc.Update(context.Background(), id, pkgauthz.Principal{UserID: adminID, Role: "admin"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, newPrice, res.Price)
}

func TestListingService_Update_Forbidden(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     ownerID,
	}

	req := &request.UpdateListingRequest{}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)

	res, err := svc.Update(context.Background(), id, pkgauthz.Principal{UserID: otherUserID, Role: "user"}, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrForbidden, err)
}

func TestListingService_Update_NoChanges(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	userID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
		Title:      "Same",
	}

	req := &request.UpdateListingRequest{}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)

	res, err := svc.Update(context.Background(), id, pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Same", res.Title)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestListingService_Delete_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	publisher := &fakeEventPublisher{}
	svc := service.NewListingServiceWithAuthzAndPublisher(repo, newTestAuthzService(t), publisher)

	id := uuid.New()
	userID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id, pkgauthz.Principal{UserID: userID, Role: "user"})

	assert.NoError(t, err)
	assert.Len(t, publisher.listingEvents, 1)
	assert.Equal(t, domain.EventTypeListingDeleted, publisher.listingEvents[0].Metadata.EventType)
}

func TestListingService_Delete_Forbidden(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     ownerID,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)

	err := svc.Delete(context.Background(), id, pkgauthz.Principal{UserID: otherUserID, Role: "user"})

	assert.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)
}

func TestListingService_List_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	filter := domain.ListingFilter{Page: 1, Limit: 10}
	expectedFilter := filter
	expectedFilter.Statuses = []string{"active", "sold"}
	listings := []*entity.Listing{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "Listing 1"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "Listing 2"},
	}

	repo.On("List", mock.Anything, expectedFilter).Return(listings, int64(2), nil)

	res, err := svc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Data, 2)
	assert.Equal(t, int64(2), res.Total)
}

func TestListingService_List_Empty(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	filter := domain.ListingFilter{Page: 1, Limit: 10}
	expectedFilter := filter
	expectedFilter.Statuses = []string{"active", "sold"}
	repo.On("List", mock.Anything, expectedFilter).Return([]*entity.Listing{}, int64(0), nil)

	res, err := svc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Data, 0)
	assert.Equal(t, int64(0), res.Total)
	assert.Equal(t, 0, res.TotalPages)
}

func TestListingService_ListByUserID_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	userID := uuid.New()
	filter := domain.ListingFilter{Page: 1, Limit: 10}
	listings := []*entity.Listing{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, UserID: userID, Title: "User Listing"},
	}

	repo.On("FindByUserID", mock.Anything, userID, filter).Return(listings, int64(1), nil)

	res, err := svc.ListByUserID(context.Background(), pkgauthz.Principal{UserID: userID, Role: "user"}, filter)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Data, 1)
	assert.Equal(t, int64(1), res.Total)
}

// Edge case: Create with nil CategoryID
func TestListingService_Create_NilCategoryID(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	userID := uuid.New()
	createdID := uuid.New()
	req := &request.CreateListingRequest{
		CategoryID: nil,
		Title:      "No Category Listing",
		Price:      1000,
		Status:     "active",
	}

	repo.On("ExistsBySlug", mock.Anything, mock.Anything).Return(false, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.Listing) bool {
		return l.CategoryID == nil
	})).Return(&entity.Listing{BaseEntity: entity.BaseEntity{ID: createdID}, Title: req.Title}, nil)
	repo.On("FindByID", mock.Anything, createdID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: createdID},
		Title:      req.Title,
		Price:      req.Price,
		Status:     req.Status,
	}, nil)

	res, err := svc.Create(context.Background(), pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

// Edge case: Update only Specifications
func TestListingService_Update_OnlySpecifications(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	userID := uuid.New()
	bathrooms := 2
	landArea := 120
	buildingArea := 95
	listing := &entity.Listing{
		BaseEntity:     entity.BaseEntity{ID: id},
		UserID:         userID,
		Specifications: datatypes.JSON(`{"bedrooms":3,"bathrooms":2,"land_area_sqm":120,"building_area_sqm":95}`),
	}

	newSpecs := &request.Specifications{Bedrooms: 5}
	req := &request.UpdateListingRequest{
		Specifications: newSpecs,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)
	repo.On("Update", mock.Anything, mock.Anything, []string{"bedroom_count", "bathroom_count", "land_area_sqm", "building_area_sqm", "specifications"}).Return(&entity.Listing{
		BaseEntity:      entity.BaseEntity{ID: id},
		BedroomCount:    ptr(5),
		BathroomCount:   &bathrooms,
		LandAreaSqm:     &landArea,
		BuildingAreaSqm: &buildingArea,
		Specifications:  datatypes.JSON(`{"bedrooms":5,"bathrooms":2,"land_area_sqm":120,"building_area_sqm":95}`),
	}, nil)

	res, err := svc.Update(context.Background(), id, pkgauthz.Principal{UserID: userID, Role: "user"}, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	var s request.Specifications
	_ = json.Unmarshal(res.Specifications, &s)
	assert.Equal(t, 5, s.Bedrooms)
	assert.Equal(t, 2, s.Bathrooms)
	assert.Equal(t, 120, s.LandAreaSqm)
	assert.Equal(t, 95, s.BuildingAreaSqm)
	assert.Equal(t, 5, *res.BedroomCount)
	assert.Equal(t, 2, *res.BathroomCount)
}

func TestListingService_Update_OnlySpecifications_CanSetZeroWithoutDroppingSiblings(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	id := uuid.New()
	userID := uuid.New()
	bathrooms := 2
	landArea := 120
	buildingArea := 95
	listing := &entity.Listing{
		BaseEntity:     entity.BaseEntity{ID: id},
		UserID:         userID,
		Specifications: datatypes.JSON(`{"bedrooms":3,"bathrooms":2,"land_area_sqm":120,"building_area_sqm":95}`),
	}

	var req request.UpdateListingRequest
	err := json.Unmarshal([]byte(`{"specifications":{"bedrooms":0}}`), &req)
	assert.NoError(t, err)

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)
	repo.On("Update", mock.Anything, mock.Anything, []string{"bedroom_count", "bathroom_count", "land_area_sqm", "building_area_sqm", "specifications"}).Return(&entity.Listing{
		BaseEntity:      entity.BaseEntity{ID: id},
		BedroomCount:    ptr(0),
		BathroomCount:   &bathrooms,
		LandAreaSqm:     &landArea,
		BuildingAreaSqm: &buildingArea,
		Specifications:  datatypes.JSON(`{"bedrooms":0,"bathrooms":2,"land_area_sqm":120,"building_area_sqm":95}`),
	}, nil)

	res, err := svc.Update(context.Background(), id, pkgauthz.Principal{UserID: userID, Role: "user"}, &req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	var s request.Specifications
	_ = json.Unmarshal(res.Specifications, &s)
	assert.Equal(t, 0, s.Bedrooms)
	assert.Equal(t, 2, s.Bathrooms)
	assert.Equal(t, 120, s.LandAreaSqm)
	assert.Equal(t, 95, s.BuildingAreaSqm)
	assert.Equal(t, 0, *res.BedroomCount)
	assert.Equal(t, 2, *res.BathroomCount)
}

func TestListingService_UploadImage_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	storage.uploadResult = &mediaasset.UploadResult{
		AssetID:          "asset-123",
		PublicID:         "listing-asset-123",
		Version:          12,
		SecureURL:        "https://cdn.example.com/listing-asset-123.jpg",
		ResourceType:     mediaasset.DefaultResourceType,
		DeliveryType:     mediaasset.DefaultDeliveryType,
		Format:           "jpg",
		Bytes:            2048,
		Width:            1200,
		Height:           800,
		OriginalFilename: "villa.jpg",
	}
	publisher := &fakeEventPublisher{}
	svc := service.NewListingServiceWithAuthzAndPublisher(repo, newTestAuthzService(t), publisher, storage)

	listingID := uuid.New()
	userID := uuid.New()
	createdImageID := uuid.New()
	file := testListingImageFileHeader("villa.jpg", "image/jpeg")

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Once()
	repo.On("CreateImage", mock.Anything, mock.MatchedBy(func(img *entity.ListingImage) bool {
		return img.ListingID == listingID &&
			img.URL == storage.uploadResult.SecureURL &&
			img.PublicID != nil && *img.PublicID == storage.uploadResult.PublicID
	})).Return(&entity.ListingImage{ID: createdImageID}, nil).Once()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Images: []entity.ListingImage{{
			ID:        createdImageID,
			URL:       storage.uploadResult.SecureURL,
			IsPrimary: true,
			SortOrder: 0,
		}},
	}, nil).Once()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Images: []entity.ListingImage{{
			ID:        createdImageID,
			URL:       storage.uploadResult.SecureURL,
			IsPrimary: true,
			SortOrder: 0,
		}},
	}, nil).Once()

	res, err := svc.UploadImage(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, []*multipart.FileHeader{file})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Images, 1)
	assert.Equal(t, createdImageID, res.Images[0].ID)
	assert.True(t, res.Images[0].IsPrimary)
	assert.Equal(t, storage.uploadResult.SecureURL, res.Images[0].URL)
	assert.Len(t, storage.uploadCalls, 1)
	assert.Equal(t, file, storage.uploadCalls[0].File)
	assert.Equal(t, "listings/"+listingID.String(), storage.uploadCalls[0].Folder)
	assert.Equal(t, mediaasset.DefaultResourceType, storage.uploadCalls[0].ResourceType)
	assert.Equal(t, mediaasset.DefaultDeliveryType, storage.uploadCalls[0].DeliveryType)
	assert.NotEmpty(t, storage.uploadCalls[0].PublicID)
	assert.Len(t, storage.destroyCalls, 0)
	assert.Len(t, publisher.listingEvents, 1)
	assert.Equal(t, domain.EventTypeListingImagesChanged, publisher.listingEvents[0].Metadata.EventType)
	repo.AssertNotCalled(t, "ListActiveImagesByListingID", mock.Anything, mock.Anything)
}

func TestListingService_UploadImages_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	publisher := &fakeEventPublisher{}
	svc := service.NewListingServiceWithAuthzAndPublisher(repo, newTestAuthzService(t), publisher, storage)

	listingID := uuid.New()
	userID := uuid.New()
	imageIDs := []uuid.UUID{uuid.New(), uuid.New()}
	files := []*multipart.FileHeader{
		testListingImageFileHeader("front.jpg", "image/jpeg"),
		testListingImageFileHeader("living-room.jpg", "image/jpeg"),
	}

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Once()
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(&entity.ListingImage{ID: imageIDs[0]}, nil).Once()
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(&entity.ListingImage{ID: imageIDs[1]}, nil).Once()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Images: []entity.ListingImage{{
			ID:        imageIDs[0],
			URL:       "https://cdn.example.com/front.jpg",
			IsPrimary: true,
			SortOrder: 0,
		}, {
			ID:        imageIDs[1],
			URL:       "https://cdn.example.com/living-room.jpg",
			IsPrimary: false,
			SortOrder: 1,
		}},
	}, nil).Twice()

	res, err := svc.UploadImage(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, files)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Images, len(files))
	assert.Equal(t, imageIDs[0], res.Images[0].ID)
	assert.True(t, res.Images[0].IsPrimary)
	assert.False(t, res.Images[1].IsPrimary)
	assert.Len(t, storage.uploadCalls, len(files))
	assert.Len(t, storage.destroyCalls, 0)
	assert.Len(t, publisher.listingEvents, 1)
	assert.Equal(t, domain.EventTypeListingImagesChanged, publisher.listingEvents[0].Metadata.EventType)
}

func TestListingService_UploadImage_InvalidFile(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	res, err := svc.UploadImage(context.Background(), uuid.New(), pkgauthz.Principal{UserID: uuid.New(), Role: "user"}, nil)

	assert.ErrorIs(t, err, domain.ErrInvalidImageFile)
	assert.Nil(t, res)
	assert.Len(t, storage.uploadCalls, 0)
	repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}

func TestListingService_UploadImage_StorageUnset(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	res, err := svc.UploadImage(context.Background(), uuid.New(), pkgauthz.Principal{UserID: uuid.New(), Role: "user"}, []*multipart.FileHeader{testListingImageFileHeader("unit.jpg", "image/jpeg")})

	assert.ErrorIs(t, err, domain.ErrImageStorageUnset)
	assert.Nil(t, res)
	repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}

func TestListingService_UploadImages_OverflowBeforeUpload(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	listingID := uuid.New()
	userID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Images:     make([]entity.ListingImage, 10),
	}

	repo.On("FindByID", mock.Anything, listingID).Return(listing, nil).Once()

	res, err := svc.UploadImage(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, []*multipart.FileHeader{testListingImageFileHeader("unit.jpg", "image/jpeg")})

	assert.ErrorIs(t, err, domain.ErrImageLimitReached)
	assert.Nil(t, res)
	assert.Len(t, storage.uploadCalls, 0)
	repo.AssertNotCalled(t, "CreateImage", mock.Anything, mock.Anything)
}

func TestListingService_UploadImage_Forbidden(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	listingID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     ownerID,
	}, nil).Once()

	res, err := svc.UploadImage(context.Background(), listingID, pkgauthz.Principal{UserID: otherUserID, Role: "user"}, []*multipart.FileHeader{testListingImageFileHeader("unit.jpg", "image/jpeg")})

	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.Nil(t, res)
	assert.Len(t, storage.uploadCalls, 0)
}

func TestListingService_UploadImages_CreateImageFails_CleansUpAllUploads(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	listingID := uuid.New()
	userID := uuid.New()
	files := []*multipart.FileHeader{
		testListingImageFileHeader("first.jpg", "image/jpeg"),
		testListingImageFileHeader("second.jpg", "image/jpeg"),
	}

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Once()
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(&entity.ListingImage{}, nil).Once()
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(nil, domain.ErrConflict).Once()

	res, err := svc.UploadImage(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, files)

	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.Nil(t, res)
	assert.Len(t, storage.uploadCalls, len(files))
	assert.Len(t, storage.destroyCalls, len(files))
}

func TestListingService_DeleteImage_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	listingID := uuid.New()
	imageID := uuid.New()
	userID := uuid.New()
	publicID := "listing-image-public-id"
	resourceType := "image"
	deliveryType := "upload"

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Once()
	repo.On("FindImageByID", mock.Anything, imageID).Return(&entity.ListingImage{
		ID:           imageID,
		ListingID:    listingID,
		PublicID:     &publicID,
		ResourceType: &resourceType,
		Type:         &deliveryType,
	}, nil).Once()
	repo.On("DeleteImage", mock.Anything, listingID, imageID).Return(nil).Once()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Images:     []entity.ListingImage{},
	}, nil).Once()

	res, err := svc.DeleteImage(context.Background(), listingID, imageID, pkgauthz.Principal{UserID: userID, Role: "user"})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, storage.destroyCalls, 1)
	assert.Equal(t, publicID, storage.destroyCalls[0].PublicID)
	assert.Equal(t, resourceType, storage.destroyCalls[0].ResourceType)
	assert.Equal(t, deliveryType, storage.destroyCalls[0].DeliveryType)
	assert.True(t, storage.destroyCalls[0].Invalidate)
}

func TestListingService_UploadVideo_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	storage.videoUploadResult.Metadata = mediaasset.Metadata{"duration": float64(45)}
	publisher := &fakeEventPublisher{}
	svc := service.NewListingServiceWithAuthzAndPublisher(repo, newTestAuthzService(t), publisher, storage)

	listingID := uuid.New()
	userID := uuid.New()
	videoID := uuid.New()
	file := testListingVideoFileHeader("tour.mp4", "video/mp4", 20*1024*1024)

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Once()
	repo.On("CreateVideo", mock.Anything, mock.MatchedBy(func(v *entity.ListingVideo) bool {
		return v.ListingID == listingID && v.URL == storage.videoUploadResult.SecureURL
	})).Return(&entity.ListingVideo{ID: videoID}, nil).Once()
	videoEntity := &entity.ListingVideo{ID: videoID, ListingID: listingID, URL: storage.videoUploadResult.SecureURL}
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Video:      videoEntity,
	}, nil).Twice()

	res, err := svc.UploadVideo(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, file)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Video)
	assert.Equal(t, videoID, res.Video.ID)
	assert.Len(t, storage.videoUploads, 1)
	assert.Len(t, publisher.listingEvents, 1)
	assert.Equal(t, domain.EventTypeListingImagesChanged, publisher.listingEvents[0].Metadata.EventType)
}

func TestListingService_UploadVideo_AlreadyExists(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	listingID := uuid.New()
	userID := uuid.New()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Video: &entity.ListingVideo{
			ID:        uuid.New(),
			ListingID: listingID,
		},
	}, nil)

	res, err := svc.UploadVideo(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, testListingVideoFileHeader("tour.mp4", "video/mp4", 10*1024*1024))

	assert.ErrorIs(t, err, domain.ErrVideoAlreadyExists)
	assert.Nil(t, res)
	assert.Len(t, storage.videoUploads, 0)
}

func TestListingService_UploadVideo_InvalidFile(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	res, err := svc.UploadVideo(context.Background(), uuid.New(), pkgauthz.Principal{UserID: uuid.New(), Role: "user"}, nil)

	assert.ErrorIs(t, err, domain.ErrInvalidVideoFile)
	assert.Nil(t, res)
	repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}

func TestListingService_UploadVideo_TooLarge(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	largeFile := testListingVideoFileHeader("big.mov", "video/quicktime", 100*1024*1024+1)
	res, err := svc.UploadVideo(context.Background(), uuid.New(), pkgauthz.Principal{UserID: uuid.New(), Role: "user"}, largeFile)

	assert.ErrorIs(t, err, domain.ErrVideoTooLarge)
	assert.Nil(t, res)
	repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}

func TestListingService_UploadVideo_TooLong(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	storage.videoUploadResult.Metadata = mediaasset.Metadata{"duration": float64(70)}
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t), storage)

	listingID := uuid.New()
	userID := uuid.New()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil)
	res, err := svc.UploadVideo(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, testListingVideoFileHeader("long.mp4", "video/mp4", 5*1024*1024))

	assert.ErrorIs(t, err, domain.ErrVideoTooLong)
	assert.Nil(t, res)
	assert.Len(t, storage.videoUploads, 1)
	assert.Len(t, storage.videoDestroyCalls, 1)
}

func TestListingService_DeleteVideo_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	storage := newFakeListingImageStorage()
	publisher := &fakeEventPublisher{}
	svc := service.NewListingServiceWithAuthzAndPublisher(repo, newTestAuthzService(t), publisher, storage)

	listingID := uuid.New()
	userID := uuid.New()
	videoID := uuid.New()
	publicID := "listing-video-public"
	resourceType := mediaasset.DefaultVideoResourceType
	deliveryType := mediaasset.DefaultDeliveryType
	video := &entity.ListingVideo{
		ID:           videoID,
		ListingID:    listingID,
		PublicID:     &publicID,
		ResourceType: &resourceType,
		DeliveryType: &deliveryType,
	}
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Video:      video,
	}, nil).Once()
	repo.On("FindVideoByListingID", mock.Anything, listingID).Return(video, nil).Once()
	repo.On("DeleteVideoByListingID", mock.Anything, listingID).Return(nil).Once()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Twice()

	res, err := svc.DeleteVideo(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Nil(t, res.Video)
	assert.Len(t, storage.videoDestroyCalls, 1)
	assert.Len(t, publisher.listingEvents, 1)
}

func TestListingService_SetPrimaryImage_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	listingID := uuid.New()
	imageID := uuid.New()
	userID := uuid.New()

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Once()
	repo.On("FindImageByID", mock.Anything, imageID).Return(&entity.ListingImage{
		ID:        imageID,
		ListingID: listingID,
	}, nil).Once()
	repo.On("SetPrimaryImage", mock.Anything, listingID, imageID).Return(nil).Once()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Images: []entity.ListingImage{{
			ID:        imageID,
			IsPrimary: true,
			SortOrder: 0,
		}},
	}, nil).Once()

	res, err := svc.SetPrimaryImage(context.Background(), listingID, imageID, pkgauthz.Principal{UserID: userID, Role: "user"})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Images, 1)
	assert.Equal(t, imageID, res.Images[0].ID)
	assert.True(t, res.Images[0].IsPrimary)
}

func TestListingService_ReorderImages_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	listingID := uuid.New()
	userID := uuid.New()
	imageA := uuid.New()
	imageB := uuid.New()
	ordered := []uuid.UUID{imageB, imageA}

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Once()
	repo.On("ListActiveImagesByListingID", mock.Anything, listingID).Return([]*entity.ListingImage{
		{ID: imageA, ListingID: listingID, SortOrder: 0},
		{ID: imageB, ListingID: listingID, SortOrder: 1},
	}, nil).Once()
	repo.On("ReorderImages", mock.Anything, listingID, ordered).Return(nil).Once()
	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
		Images: []entity.ListingImage{
			{ID: imageB, SortOrder: 0, CreatedAt: time.Now().Add(-time.Minute)},
			{ID: imageA, SortOrder: 1, CreatedAt: time.Now()},
		},
	}, nil).Once()

	res, err := svc.ReorderImages(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, ordered)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Images, 2)
	assert.Equal(t, imageB, res.Images[0].ID)
	assert.Equal(t, imageA, res.Images[1].ID)
}

func TestListingService_ReorderImages_InvalidPayload(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingServiceWithAuthz(repo, newTestAuthzService(t))

	listingID := uuid.New()
	userID := uuid.New()
	imageA := uuid.New()
	imageB := uuid.New()

	repo.On("FindByID", mock.Anything, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		UserID:     userID,
	}, nil).Once()
	repo.On("ListActiveImagesByListingID", mock.Anything, listingID).Return([]*entity.ListingImage{
		{ID: imageA, ListingID: listingID, SortOrder: 0},
		{ID: imageB, ListingID: listingID, SortOrder: 1},
	}, nil).Once()

	res, err := svc.ReorderImages(context.Background(), listingID, pkgauthz.Principal{UserID: userID, Role: "user"}, []uuid.UUID{imageA})

	assert.ErrorIs(t, err, domain.ErrImageOrderInvalid)
	assert.Nil(t, res)
	repo.AssertNotCalled(t, "ReorderImages", mock.Anything, mock.Anything, mock.Anything)
}

type fakeListingImageStorage struct {
	uploadResult       *mediaasset.UploadResult
	uploadErr          error
	destroyResult      *mediaasset.DestroyResult
	destroyErr         error
	uploadCalls        []mediaasset.UploadInput
	destroyCalls       []mediaasset.DestroyInput
	videoUploadResult  *mediaasset.UploadResult
	videoUploadErr     error
	videoDestroyResult *mediaasset.DestroyResult
	videoDestroyErr    error
	videoUploads       []mediaasset.UploadInput
	videoDestroyCalls  []mediaasset.DestroyInput
}

func newFakeListingImageStorage() *fakeListingImageStorage {
	return &fakeListingImageStorage{
		uploadResult: &mediaasset.UploadResult{
			PublicID:     "test-public-id",
			SecureURL:    "https://cdn.example.com/test-public-id.jpg",
			ResourceType: mediaasset.DefaultResourceType,
			DeliveryType: mediaasset.DefaultDeliveryType,
		},
		destroyResult: &mediaasset.DestroyResult{Result: "ok"},
		videoUploadResult: &mediaasset.UploadResult{
			PublicID:     "test-video-public-id",
			SecureURL:    "https://cdn.example.com/test-video.mp4",
			ResourceType: mediaasset.DefaultVideoResourceType,
			DeliveryType: mediaasset.DefaultDeliveryType,
			Format:       "mp4",
			Bytes:        50 * 1024 * 1024,
			Metadata:     mediaasset.Metadata{"duration": float64(30)},
		},
		videoDestroyResult: &mediaasset.DestroyResult{Result: "ok"},
	}
}

func (f *fakeListingImageStorage) UploadListingImage(_ context.Context, input mediaasset.UploadInput) (*mediaasset.UploadResult, error) {
	f.uploadCalls = append(f.uploadCalls, input)
	return f.uploadResult, f.uploadErr
}

func (f *fakeListingImageStorage) DestroyListingImage(_ context.Context, input mediaasset.DestroyInput) (*mediaasset.DestroyResult, error) {
	f.destroyCalls = append(f.destroyCalls, input)
	return f.destroyResult, f.destroyErr
}

func (f *fakeListingImageStorage) UploadListingVideo(_ context.Context, input mediaasset.UploadInput) (*mediaasset.UploadResult, error) {
	f.videoUploads = append(f.videoUploads, input)
	return f.videoUploadResult, f.videoUploadErr
}

func (f *fakeListingImageStorage) DestroyListingVideo(_ context.Context, input mediaasset.DestroyInput) (*mediaasset.DestroyResult, error) {
	f.videoDestroyCalls = append(f.videoDestroyCalls, input)
	return f.videoDestroyResult, f.videoDestroyErr
}

func testListingImageFileHeader(filename, contentType string) *multipart.FileHeader {
	header := textproto.MIMEHeader{}
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}

	return &multipart.FileHeader{
		Filename: filename,
		Header:   header,
	}
}

func testListingVideoFileHeader(filename, contentType string, size int64) *multipart.FileHeader {
	header := textproto.MIMEHeader{}
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}

	return &multipart.FileHeader{
		Filename: filename,
		Header:   header,
		Size:     size,
	}
}

func ptr[T any](v T) *T {
	return &v
}
