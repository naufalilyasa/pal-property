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
}
