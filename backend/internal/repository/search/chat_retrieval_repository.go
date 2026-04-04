package search

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
)

type ChatRetrievalRepository struct {
	index  string
	client *searchindex.Client
}

type elasticsearchSearchResponse[T any] struct {
	Hits struct {
		Hits []struct {
			Source T `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func NewChatRetrievalRepository(index string, client *searchindex.Client) (*ChatRetrievalRepository, error) {
	if index == "" {
		return nil, fmt.Errorf("chat retrieval repository: index is required")
	}
	if client == nil {
		return nil, fmt.Errorf("chat retrieval repository: search client is required")
	}
	return &ChatRetrievalRepository{index: index, client: client}, nil
}

func (r *ChatRetrievalRepository) FetchDocuments(ctx context.Context, filters domain.ChatRetrievalFilters, queryVector []float64, limit int) ([]domain.ChatRetrievalDocument, error) {
	if limit <= 0 {
		limit = 5
	}

	query := buildChatRetrievalQuery(filters, queryVector, limit)
	var raw elasticsearchSearchResponse[domain.ChatRetrievalDocument]
	if err := r.client.Search(ctx, r.index, query, &raw); err != nil {
		return nil, fmt.Errorf("search chat retrieval documents: %w", err)
	}

	documents := make([]domain.ChatRetrievalDocument, 0, len(raw.Hits.Hits))
	for _, hit := range raw.Hits.Hits {
		documents = append(documents, hit.Source)
	}
	return documents, nil
}

func (r *ChatRetrievalRepository) FetchDocumentByID(ctx context.Context, listingID uuid.UUID) (*domain.ChatRetrievalDocument, error) {
	query := map[string]any{
		"size": 1,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{"term": map[string]any{"listing_id": listingID.String()}},
				},
			},
		},
	}

	var raw elasticsearchSearchResponse[domain.ChatRetrievalDocument]
	if err := r.client.Search(ctx, r.index, query, &raw); err != nil {
		return nil, fmt.Errorf("search chat retrieval document by id: %w", err)
	}
	if len(raw.Hits.Hits) == 0 {
		return nil, domain.ErrNotFound
	}
	return &raw.Hits.Hits[0].Source, nil
}

func buildChatRetrievalQuery(filters domain.ChatRetrievalFilters, queryVector []float64, limit int) map[string]any {
	filterClauses := []any{
		map[string]any{"term": map[string]any{"status": "active"}},
	}
	if filters.TransactionType != "" {
		filterClauses = append(filterClauses, map[string]any{"term": map[string]any{"transaction_type": filters.TransactionType}})
	}
	if filters.CategoryID != nil {
		filterClauses = append(filterClauses, map[string]any{"term": map[string]any{"category.id": filters.CategoryID.String()}})
	}
	if filters.LocationProvince != "" {
		filterClauses = append(filterClauses, map[string]any{"term": map[string]any{"location_province": filters.LocationProvince}})
	}
	if filters.LocationCity != "" {
		filterClauses = append(filterClauses, map[string]any{"term": map[string]any{"location_city": filters.LocationCity}})
	}
	if filters.PriceMin != nil || filters.PriceMax != nil {
		rangeFilter := map[string]any{}
		if filters.PriceMin != nil {
			rangeFilter["gte"] = *filters.PriceMin
		}
		if filters.PriceMax != nil {
			rangeFilter["lte"] = *filters.PriceMax
		}
		filterClauses = append(filterClauses, map[string]any{"range": map[string]any{"price": rangeFilter}})
	}

	body := map[string]any{
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filterClauses,
			},
		},
	}

	if filters.Query != "" {
		boolQuery := body["query"].(map[string]any)["bool"].(map[string]any)
		boolQuery["must"] = []any{
			map[string]any{
				"multi_match": map[string]any{
					"query":  filters.Query,
					"fields": []string{"title^4", "lexical_search_text^3", "lexical_text^2", "description_excerpt^2", "location_province", "location_city", "category.name", "category.slug"},
				},
			},
		}
		if shouldClauses := buildChatRetrievalExactShouldClauses(filters.Query); len(shouldClauses) > 0 {
			boolQuery["should"] = shouldClauses
		}
	}

	if len(queryVector) > 0 {
		body["knn"] = map[string]any{
			"field":          "embedding",
			"query_vector":   queryVector,
			"k":              limit,
			"num_candidates": max(limit*4, 10),
			"boost":          0.65,
		}
	}

	return body
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func buildChatRetrievalExactShouldClauses(query string) []any {
	trimmedQuery := strings.TrimSpace(query)
	normalizedQuery := normalizeChatRetrievalQueryText(query)
	if trimmedQuery == "" || normalizedQuery == "" {
		return nil
	}

	querySlug := strings.ReplaceAll(normalizedQuery, " ", "-")
	shouldClauses := []any{
		map[string]any{
			"match_phrase": map[string]any{
				"title": map[string]any{
					"query": trimmedQuery,
					"boost": 20,
				},
			},
		},
		buildCaseInsensitiveTermQuery("category.name", trimmedQuery, 14),
		buildCaseInsensitiveTermQuery("location_city", trimmedQuery, 12),
		buildCaseInsensitiveTermQuery("location_province", trimmedQuery, 10),
	}
	if querySlug != "" {
		shouldClauses = append(shouldClauses,
			buildCaseInsensitiveTermQuery("slug", querySlug, 24),
			buildCaseInsensitiveTermQuery("category.slug", querySlug, 16),
		)
	}

	return shouldClauses
}

func buildCaseInsensitiveTermQuery(field string, value string, boost float64) map[string]any {
	return map[string]any{
		"term": map[string]any{
			field: map[string]any{
				"value":            value,
				"boost":            boost,
				"case_insensitive": true,
			},
		},
	}
}

func normalizeChatRetrievalQueryText(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return ""
	}

	var builder strings.Builder
	lastWasSpace := true
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
			lastWasSpace = false
			continue
		}
		if !lastWasSpace {
			builder.WriteRune(' ')
			lastWasSpace = true
		}
	}

	return strings.Join(strings.Fields(builder.String()), " ")
}
