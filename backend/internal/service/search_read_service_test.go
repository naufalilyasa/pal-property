package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchReadService_SearchListings_DefaultsToNewestWithoutQuery(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hits":{"total":{"value":1},"hits":[{"_source":{"id":"` + uuid.NewString() + `","title":"Public Listing","slug":"public-listing","transaction_type":"sale","price":1000,"currency":"IDR","status":"active","is_featured":false,"created_at":"2026-03-17T00:00:00Z","updated_at":"2026-03-17T00:00:00Z"}}]}}`))
	}))
	defer server.Close()

	client, err := searchindex.NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)
	svc, err := service.NewSearchReadService("listings", client)
	require.NoError(t, err)

	res, err := svc.SearchListings(context.Background(), request.SearchListingsRequest{Page: 0, Limit: 0})
	require.NoError(t, err)
	assert.Equal(t, 1, res.Page)
	assert.Equal(t, 20, res.Limit)
	assert.Len(t, res.Items, 1)
	sortBody := body["sort"].([]any)
	assert.Equal(t, "created_at", mapsFirstKey(sortBody[0].(map[string]any)))
}

func TestSearchReadService_SearchListings_UsesRelevanceWhenQueryPresent(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hits":{"total":{"value":0},"hits":[]}}`))
	}))
	defer server.Close()

	client, err := searchindex.NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)
	svc, err := service.NewSearchReadService("listings", client)
	require.NoError(t, err)

	_, err = svc.SearchListings(context.Background(), request.SearchListingsRequest{Query: "jakarta", Sort: ""})
	require.NoError(t, err)
	boolQuery := body["query"].(map[string]any)["bool"].(map[string]any)
	assert.NotEmpty(t, boolQuery["must"])
	sortBody := body["sort"].([]any)
	assert.Equal(t, "_score", mapsFirstKey(sortBody[0].(map[string]any)))
}

func mapsFirstKey(m map[string]any) string {
	for key := range m {
		return key
	}
	return ""
}
