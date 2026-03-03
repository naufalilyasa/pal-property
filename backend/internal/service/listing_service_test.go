package service_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/mocks"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

func TestListingService_Create_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	userID := uuid.New()
	categoryID := uuid.New()
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
		BaseEntity:     entity.BaseEntity{ID: uuid.New(), CreatedAt: time.Now()},
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

	res, err := svc.Create(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "modern-villa-in-bali", res.Slug)
	assert.Equal(t, req.Title, res.Title)
}

func TestListingService_Create_SlugCollision(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	userID := uuid.New()
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
		Title: req.Title,
		Slug:  "test-listing-random",
	}, nil)

	res, err := svc.Create(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEqual(t, "test-listing", res.Slug)
}

func TestListingService_Create_RepoError(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	userID := uuid.New()
	req := &request.CreateListingRequest{
		Title:  "Test Listing",
		Price:  1000000,
		Status: "active",
	}

	repo.On("ExistsBySlug", mock.Anything, mock.Anything).Return(false, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrConflict)

	res, err := svc.Create(context.Background(), userID, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrConflict, err)
}

func TestListingService_GetByID_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	id := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		Title:      "Test",
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)
	repo.On("IncrementViewCount", mock.Anything, id).Return(nil).Maybe()

	res, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, id, res.ID)
}

func TestListingService_GetByID_NotFound(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, domain.ErrNotFound)

	res, err := svc.GetByID(context.Background(), id)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrNotFound, err)
}

func TestListingService_GetBySlug_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	slugStr := "test-slug"
	id := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		Slug:       slugStr,
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
	svc := service.NewListingService(repo)

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

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)
	repo.On("ExistsBySlug", mock.Anything, "new-title").Return(false, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(l *entity.Listing) bool {
		return l.Title == newTitle
	}), []string{"title", "slug"}).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
		Title:      newTitle,
		Slug:       "new-title",
	}, nil)

	res, err := svc.Update(context.Background(), id, userID, "user", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, newTitle, res.Title)
}

func TestListingService_Update_SuccessAdmin(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

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

	res, err := svc.Update(context.Background(), id, adminID, "admin", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, newPrice, res.Price)
}

func TestListingService_Update_Forbidden(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	id := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     ownerID,
	}

	req := &request.UpdateListingRequest{}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)

	res, err := svc.Update(context.Background(), id, otherUserID, "user", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrForbidden, err)
}

func TestListingService_Update_NoChanges(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	id := uuid.New()
	userID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
		Title:      "Same",
	}

	req := &request.UpdateListingRequest{}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)

	res, err := svc.Update(context.Background(), id, userID, "user", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Same", res.Title)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestListingService_Delete_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	id := uuid.New()
	userID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id, userID, "user")

	assert.NoError(t, err)
}

func TestListingService_Delete_Forbidden(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	id := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     ownerID,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)

	err := svc.Delete(context.Background(), id, otherUserID, "user")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)
}

func TestListingService_List_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	filter := domain.ListingFilter{Page: 1, Limit: 10}
	listings := []*entity.Listing{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "Listing 1"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "Listing 2"},
	}

	repo.On("List", mock.Anything, filter).Return(listings, int64(2), nil)

	res, err := svc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Data, 2)
	assert.Equal(t, int64(2), res.Total)
}

func TestListingService_List_Empty(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	filter := domain.ListingFilter{Page: 1, Limit: 10}
	repo.On("List", mock.Anything, filter).Return([]*entity.Listing{}, int64(0), nil)

	res, err := svc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Data, 0)
	assert.Equal(t, int64(0), res.Total)
	assert.Equal(t, 0, res.TotalPages)
}

func TestListingService_ListByUserID_Success(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	userID := uuid.New()
	filter := domain.ListingFilter{Page: 1, Limit: 10}
	listings := []*entity.Listing{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, UserID: userID, Title: "User Listing"},
	}

	repo.On("FindByUserID", mock.Anything, userID, filter).Return(listings, int64(1), nil)

	res, err := svc.ListByUserID(context.Background(), userID, filter)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Data, 1)
	assert.Equal(t, int64(1), res.Total)
}

// Edge case: Create with nil CategoryID
func TestListingService_Create_NilCategoryID(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	userID := uuid.New()
	req := &request.CreateListingRequest{
		CategoryID: nil,
		Title:      "No Category Listing",
		Price:      1000,
		Status:     "active",
	}

	repo.On("ExistsBySlug", mock.Anything, mock.Anything).Return(false, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.Listing) bool {
		return l.CategoryID == nil
	})).Return(&entity.Listing{Title: req.Title}, nil)

	res, err := svc.Create(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

// Edge case: Update only Specifications
func TestListingService_Update_OnlySpecifications(t *testing.T) {
	repo := mocks.NewListingRepository(t)
	svc := service.NewListingService(repo)

	id := uuid.New()
	userID := uuid.New()
	listing := &entity.Listing{
		BaseEntity: entity.BaseEntity{ID: id},
		UserID:     userID,
	}

	newSpecs := &request.Specifications{Bedrooms: 5}
	req := &request.UpdateListingRequest{
		Specifications: newSpecs,
	}

	repo.On("FindByID", mock.Anything, id).Return(listing, nil)
	repo.On("Update", mock.Anything, mock.Anything, []string{"specifications"}).Return(&entity.Listing{
		BaseEntity:     entity.BaseEntity{ID: id},
		Specifications: datatypes.JSON(`{"bedrooms":5}`),
	}, nil)

	res, err := svc.Update(context.Background(), id, userID, "user", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	var s request.Specifications
	_ = json.Unmarshal(res.Specifications, &s)
	assert.Equal(t, 5, s.Bedrooms)
}

func ptr[T any](v T) *T {
	return &v
}
