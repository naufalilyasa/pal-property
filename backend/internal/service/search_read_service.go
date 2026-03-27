package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
	pkgvalidator "github.com/naufalilyasa/pal-property-backend/pkg/validator"
)

type SearchReadService interface {
	SearchListings(ctx context.Context, req request.SearchListingsRequest) (*response.SearchListingsPageResponse, error)
}

type elasticsearchHit[T any] struct {
	Source T `json:"_source"`
}

type elasticsearchSearchResponse[T any] struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []elasticsearchHit[T] `json:"hits"`
	} `json:"hits"`
}

type searchReadService struct {
	index  string
	client *searchindex.Client
}

func NewSearchReadService(index string, client *searchindex.Client) (SearchReadService, error) {
	if index == "" {
		return nil, fmt.Errorf("search read service: index is required")
	}
	if client == nil {
		return nil, fmt.Errorf("search read service: client is required")
	}
	return &searchReadService{index: index, client: client}, nil
}

func (s *searchReadService) SearchListings(ctx context.Context, req request.SearchListingsRequest) (*response.SearchListingsPageResponse, error) {
	if err := pkgvalidator.Validate.Struct(req); err != nil {
		return nil, err
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Sort == "" {
		if req.Query != "" {
			req.Sort = "relevance"
		} else {
			req.Sort = "newest"
		}
	}

	query := buildSearchListingsQuery(req)
	var raw elasticsearchSearchResponse[response.SearchListingCardResponse]
	if err := s.client.Search(ctx, s.index, query, &raw); err != nil {
		if isMissingSearchIndexError(err) {
			if ensureErr := s.client.EnsureIndex(ctx, s.index, ListingIndexMapping()); ensureErr != nil {
				return nil, ensureErr
			}
			return &response.SearchListingsPageResponse{Items: []*response.SearchListingCardResponse{}, Total: 0, Page: req.Page, Limit: req.Limit, TotalPages: 0}, nil
		}
		return nil, err
	}
	items := make([]*response.SearchListingCardResponse, 0, len(raw.Hits.Hits))
	for _, hit := range raw.Hits.Hits {
		item := hit.Source
		items = append(items, &item)
	}
	totalPages := int((raw.Hits.Total.Value + int64(req.Limit) - 1) / int64(req.Limit))
	return &response.SearchListingsPageResponse{Items: items, Total: raw.Hits.Total.Value, Page: req.Page, Limit: req.Limit, TotalPages: totalPages}, nil
}

func buildSearchListingsQuery(req request.SearchListingsRequest) map[string]any {
	filter := []map[string]any{{"term": map[string]any{"status": "active"}}}
	must := make([]map[string]any, 0)
	if req.Query != "" {
		must = append(must, map[string]any{"multi_match": map[string]any{"query": req.Query, "fields": []string{"title^3", "description_excerpt", "location_city", "location_province", "category.name"}}})
	}
	if req.TransactionType != "" {
		filter = append(filter, map[string]any{"term": map[string]any{"transaction_type": req.TransactionType}})
	}
	if req.CategoryID != nil {
		filter = append(filter, map[string]any{"term": map[string]any{"category_id": req.CategoryID.String()}})
	}
	if req.LocationProvince != "" {
		filter = append(filter, map[string]any{"term": map[string]any{"location_province": req.LocationProvince}})
	}
	if req.LocationCity != "" {
		filter = append(filter, map[string]any{"term": map[string]any{"location_city": req.LocationCity}})
	}
	if req.PriceMin != nil || req.PriceMax != nil {
		rangeBody := map[string]any{}
		if req.PriceMin != nil {
			rangeBody["gte"] = *req.PriceMin
		}
		if req.PriceMax != nil {
			rangeBody["lte"] = *req.PriceMax
		}
		filter = append(filter, map[string]any{"range": map[string]any{"price": rangeBody}})
	}
	query := map[string]any{"bool": map[string]any{"filter": filter}}
	if len(must) > 0 {
		query["bool"].(map[string]any)["must"] = must
	}
	return map[string]any{
		"from":  (req.Page - 1) * req.Limit,
		"size":  req.Limit,
		"query": query,
		"sort":  buildSearchSort(req.Sort),
	}
}

func buildSearchSort(sort string) []map[string]any {
	switch sort {
	case "price_asc":
		return []map[string]any{{"price": map[string]any{"order": "asc"}}}
	case "price_desc":
		return []map[string]any{{"price": map[string]any{"order": "desc"}}}
	case "newest":
		return []map[string]any{{"created_at": map[string]any{"order": "desc"}}}
	default:
		return []map[string]any{{"_score": map[string]any{"order": "desc"}}, {"created_at": map[string]any{"order": "desc"}}}
	}
}

func isMissingSearchIndexError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "index_not_found_exception") || strings.Contains(message, "status 404")
}
