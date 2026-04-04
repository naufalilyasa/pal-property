package search

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
	"github.com/stretchr/testify/require"
)

func TestChatRetrievalRepositoryBuildsHybridQuery(t *testing.T) {
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(`{"hits":{"hits":[{"_source":{"listing_id":"11111111-1111-1111-1111-111111111111","title":"Rumah Aktif","slug":"rumah-aktif","transaction_type":"sale","price":2500000000,"currency":"IDR","status":"active","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}]}}`))
	}))
	defer server.Close()

	client, err := searchindex.NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)
	repo, err := NewChatRetrievalRepository("chat-retrieval", client)
	require.NoError(t, err)
	categoryID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	priceMin := int64(1000000000)
	priceMax := int64(5000000000)

	documents, err := repo.FetchDocuments(t.Context(), domain.ChatRetrievalFilters{
		Query:            "rumah jakarta selatan",
		TransactionType:  "sale",
		CategoryID:       &categoryID,
		LocationProvince: "DKI Jakarta",
		LocationCity:     "Jakarta Selatan",
		PriceMin:         &priceMin,
		PriceMax:         &priceMax,
	}, []float64{0.1, 0.2, 0.3}, 3)
	require.NoError(t, err)
	require.Len(t, documents, 1)
	require.Contains(t, captured, "knn")
	require.Contains(t, captured, "query")

	queryMap, ok := captured["query"].(map[string]any)
	require.True(t, ok)
	boolQuery, ok := queryMap["bool"].(map[string]any)
	require.True(t, ok)
	filters, ok := boolQuery["filter"].([]any)
	require.True(t, ok)
	require.Len(t, filters, 6)
	mustClauses, ok := boolQuery["must"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, mustClauses)
	shouldClauses, ok := boolQuery["should"].([]any)
	require.True(t, ok)
	require.Len(t, shouldClauses, 6)
	multiMatch, ok := mustClauses[0].(map[string]any)["multi_match"].(map[string]any)
	require.True(t, ok)
	fields, ok := multiMatch["fields"].([]any)
	require.True(t, ok)
	require.Contains(t, fields, "lexical_search_text^3")
	require.Contains(t, fields, "lexical_text^2")
	require.Contains(t, fields, "category.name")

	query, ok := multiMatch["query"].(string)
	require.True(t, ok)
	require.Equal(t, "rumah jakarta selatan", query)

	fieldValues := make([]string, 0, len(fields))
	for _, field := range fields {
		if value, ok := field.(string); ok {
			fieldValues = append(fieldValues, value)
		}
	}
	require.ElementsMatch(t, []string{"title^4", "lexical_search_text^3", "lexical_text^2", "description_excerpt^2", "location_province", "location_city", "category.name", "category.slug"}, fieldValues)
	require.Contains(t, shouldClauses, map[string]any{"match_phrase": map[string]any{"title": map[string]any{"query": "rumah jakarta selatan", "boost": float64(20)}}})
	require.Contains(t, shouldClauses, map[string]any{"term": map[string]any{"slug": map[string]any{"value": "rumah-jakarta-selatan", "boost": float64(24), "case_insensitive": true}}})
	require.Contains(t, shouldClauses, map[string]any{"term": map[string]any{"category.slug": map[string]any{"value": "rumah-jakarta-selatan", "boost": float64(16), "case_insensitive": true}}})
	require.Contains(t, shouldClauses, map[string]any{"term": map[string]any{"category.name": map[string]any{"value": "rumah jakarta selatan", "boost": float64(14), "case_insensitive": true}}})
	require.Contains(t, shouldClauses, map[string]any{"term": map[string]any{"location_city": map[string]any{"value": "rumah jakarta selatan", "boost": float64(12), "case_insensitive": true}}})
	require.Contains(t, shouldClauses, map[string]any{"term": map[string]any{"location_province": map[string]any{"value": "rumah jakarta selatan", "boost": float64(10), "case_insensitive": true}}})

	knn := captured["knn"].(map[string]any)
	require.Equal(t, "embedding", knn["field"])
	require.Equal(t, float64(3), knn["k"])
}
