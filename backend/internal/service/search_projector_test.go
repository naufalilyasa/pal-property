package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/mocks"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestElasticsearchSearchProjector_HandleListingEvent_UpsertsDocument(t *testing.T) {
	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client, err := searchindex.NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)
	projector, err := service.NewElasticsearchSearchProjector("listings", client, mocks.NewListingRepository(t))
	require.NoError(t, err)

	listingID := uuid.New()
	err = projector.HandleListingEvent(context.Background(), domain.ListingEvent{
		Metadata: domain.EventMetadata{EventType: domain.EventTypeListingUpdated, AggregateID: listingID, EventID: uuid.New(), AggregateType: domain.AggregateTypeListing, Version: 1, OccurredAt: time.Now().UTC()},
		Payload:  domain.ListingEventPayload{ID: listingID, Title: "Indexed Listing", Status: "active"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, method)
	assert.Equal(t, "/listings/_doc/"+listingID.String(), path)
}

func TestElasticsearchSearchProjector_HandleCategoryEvent_ReindexesListings(t *testing.T) {
	var puts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			puts++
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client, err := searchindex.NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)
	repo := mocks.NewListingRepository(t)
	categoryID := uuid.New()
	repo.On("FindByCategoryID", mock.Anything, categoryID).Return([]*entity.Listing{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "L1", Status: "active"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "L2", Status: "active"},
	}, nil)
	projector, err := service.NewElasticsearchSearchProjector("listings", client, repo)
	require.NoError(t, err)

	err = projector.HandleCategoryEvent(context.Background(), domain.CategoryEvent{
		Metadata: domain.EventMetadata{EventType: domain.EventTypeCategoryUpdated, AggregateID: categoryID, EventID: uuid.New(), AggregateType: domain.AggregateTypeCategory, Version: 1, OccurredAt: time.Now().UTC()},
		Payload:  domain.CategoryEventPayload{ID: categoryID, Name: "Residential"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, puts)
}

func TestRebuildListingIndex_UpsertsAllPages(t *testing.T) {
	var puts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			puts++
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client, err := searchindex.NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)
	repo := mocks.NewListingRepository(t)
	repo.On("List", mock.Anything, domain.ListingFilter{Page: 1, Limit: 2}).Return([]*entity.Listing{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "L1", Status: "active"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "L2", Status: "active"},
	}, int64(3), nil)
	repo.On("List", mock.Anything, domain.ListingFilter{Page: 2, Limit: 2}).Return([]*entity.Listing{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Title: "L3", Status: "active"},
	}, int64(3), nil)

	err = service.RebuildListingIndex(context.Background(), repo, client, "listings", 2)
	require.NoError(t, err)
	assert.Equal(t, 4, puts)
}
