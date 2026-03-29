package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/mocks"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSavedListingService_Save_SucceedsAndIdempotent(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	listingID := uuid.New()
	listingRepo := mocks.NewListingRepository(t)
	savedRepo := newFakeSavedListingRepository(t)
	svc := service.NewSavedListingService(savedRepo, listingRepo)

	listingRepo.On("FindByID", ctx, listingID).Return(&entity.Listing{
		BaseEntity: entity.BaseEntity{ID: listingID},
		Status:     "active",
	}, nil)
	savedRepo.On("Save", ctx, mock.MatchedBy(func(sl *entity.SavedListing) bool {
		return sl.UserID == userID && sl.ListingID == listingID
	})).Return(&entity.SavedListing{UserID: userID, ListingID: listingID}, nil)

	err := svc.Save(ctx, pkgauthz.Principal{UserID: userID, Role: "user"}, listingID)
	assert.NoError(t, err)
	savedRepo.AssertExpectations(t)
	listingRepo.AssertExpectations(t)
}

func TestSavedListingService_Save_NonActiveListingReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	listingID := uuid.New()
	listingRepo := mocks.NewListingRepository(t)
	savedRepo := newFakeSavedListingRepository(t)
	svc := service.NewSavedListingService(savedRepo, listingRepo)

	listingRepo.On("FindByID", ctx, listingID).Return(&entity.Listing{BaseEntity: entity.BaseEntity{ID: listingID}, Status: "draft"}, nil)

	err := svc.Save(ctx, pkgauthz.Principal{UserID: userID, Role: "user"}, listingID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	savedRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	listingRepo.AssertExpectations(t)
}

func TestSavedListingService_Remove_Idempotent(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	listingID := uuid.New()
	savedRepo := newFakeSavedListingRepository(t)
	listingRepo := mocks.NewListingRepository(t)
	svc := service.NewSavedListingService(savedRepo, listingRepo)

	savedRepo.On("Remove", ctx, userID, listingID).Return(nil)

	err := svc.Remove(ctx, pkgauthz.Principal{UserID: userID, Role: "user"}, listingID)
	assert.NoError(t, err)
	savedRepo.AssertExpectations(t)
	listingRepo.AssertExpectations(t)
}

func TestSavedListingService_Contains_ReturnsOrderedSubset(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	savedRepo := newFakeSavedListingRepository(t)
	listingRepo := mocks.NewListingRepository(t)
	svc := service.NewSavedListingService(savedRepo, listingRepo)

	savedRepo.On("Contains", ctx, userID, ids).Return([]uuid.UUID{ids[2], ids[0]}, nil)

	res, err := svc.Contains(ctx, pkgauthz.Principal{UserID: userID, Role: "user"}, ids)
	assert.NoError(t, err)
	assert.Equal(t, []uuid.UUID{ids[0], ids[2]}, res)
	savedRepo.AssertExpectations(t)
	listingRepo.AssertExpectations(t)
}

func TestSavedListingService_Contains_LimitExceeded(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ids := make([]uuid.UUID, service.SavedListingContainsLimit+1)
	for i := range ids {
		ids[i] = uuid.New()
	}
	savedRepo := newFakeSavedListingRepository(t)
	listingRepo := mocks.NewListingRepository(t)
	svc := service.NewSavedListingService(savedRepo, listingRepo)

	res, err := svc.Contains(ctx, pkgauthz.Principal{UserID: userID, Role: "user"}, ids)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at most")
	savedRepo.AssertNotCalled(t, "Contains", mock.Anything, mock.Anything, mock.Anything)
	listingRepo.AssertExpectations(t)
}

func TestSavedListingService_ListByUserID_ReturnsPaginatedActiveListings(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	savedRepo := newFakeSavedListingRepository(t)
	listingRepo := mocks.NewListingRepository(t)
	svc := service.NewSavedListingService(savedRepo, listingRepo)

	listingA := &entity.Listing{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Status: "active"}
	listingB := &entity.Listing{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Status: "active"}
	savedEntries := []*entity.SavedListing{
		{Listing: listingA},
		{Listing: listingB},
	}
	savedRepo.On("ListByUserID", ctx, mock.MatchedBy(func(filter domain.SavedListingFilter) bool {
		return filter.UserID == userID && filter.Page == 1 && filter.Limit == 20
	})).Return(savedEntries, int64(len(savedEntries)), nil)

	res, err := svc.ListByUserID(ctx, pkgauthz.Principal{UserID: userID, Role: "user"}, domain.SavedListingFilter{Page: 1, Limit: 20})
	assert.NoError(t, err)
	assert.Equal(t, int64(len(savedEntries)), res.Total)
	assert.Len(t, res.Data, len(savedEntries))
	assert.Equal(t, listingA.ID, res.Data[0].ID)
	assert.Equal(t, listingB.ID, res.Data[1].ID)
	savedRepo.AssertExpectations(t)
	listingRepo.AssertExpectations(t)
}

type fakeSavedListingRepository struct {
	mock.Mock
}

func newFakeSavedListingRepository(t *testing.T) *fakeSavedListingRepository {
	t.Helper()
	return &fakeSavedListingRepository{}
}

func (f *fakeSavedListingRepository) Save(ctx context.Context, savedListing *entity.SavedListing) (*entity.SavedListing, error) {
	args := f.Called(ctx, savedListing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SavedListing), args.Error(1)
}

func (f *fakeSavedListingRepository) Remove(ctx context.Context, userID, listingID uuid.UUID) error {
	return f.Called(ctx, userID, listingID).Error(0)
}

func (f *fakeSavedListingRepository) Contains(ctx context.Context, userID uuid.UUID, listingIDs []uuid.UUID) ([]uuid.UUID, error) {
	args := f.Called(ctx, userID, listingIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (f *fakeSavedListingRepository) ListByUserID(ctx context.Context, filter domain.SavedListingFilter) ([]*entity.SavedListing, int64, error) {
	args := f.Called(ctx, filter)
	var listings []*entity.SavedListing
	if args.Get(0) != nil {
		listings = args.Get(0).([]*entity.SavedListing)
	}
	return listings, args.Get(1).(int64), args.Error(2)
}
