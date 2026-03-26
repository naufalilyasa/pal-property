package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	handler "github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSearchReadService struct {
	req    request.SearchListingsRequest
	res    *response.SearchListingsPageResponse
	err    error
	called bool
}

func (f *fakeSearchReadService) SearchListings(_ context.Context, req request.SearchListingsRequest) (*response.SearchListingsPageResponse, error) {
	f.called = true
	f.req = req
	return f.res, f.err
}

func TestSearchHandler_SearchListings_Success(t *testing.T) {
	fake := &fakeSearchReadService{res: &response.SearchListingsPageResponse{Items: []*response.SearchListingCardResponse{{ID: uuid.New(), Title: "Search Result", Slug: "search-result", TransactionType: "sale", Price: 1000, Currency: "IDR", Status: "active"}}, Total: 1, Page: 1, Limit: 20, TotalPages: 1}}
	app := fiber.New()
	app.Get("/api/search/listings", handler.NewSearchHandler(fake).SearchListings)

	req := httptest.NewRequest(http.MethodGet, "/api/search/listings?q=jakarta&transaction_type=sale&location_city=Jakarta", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.True(t, fake.called)
	assert.Equal(t, "jakarta", fake.req.Query)
	assert.Equal(t, "sale", fake.req.TransactionType)
	assert.Equal(t, "Jakarta", fake.req.LocationCity)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSearchHandler_SearchListings_InvalidCategoryID(t *testing.T) {
	fake := &fakeSearchReadService{}
	app := fiber.New()
	app.Get("/api/search/listings", handler.NewSearchHandler(fake).SearchListings)

	req := httptest.NewRequest(http.MethodGet, "/api/search/listings?category_id=bad-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.False(t, fake.called)
}
