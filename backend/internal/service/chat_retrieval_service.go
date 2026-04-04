package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	requestdto "github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/gemini"
	validator "github.com/naufalilyasa/pal-property-backend/pkg/validator"
)

type chatQueryEmbedder interface {
	EmbedQuery(ctx context.Context, inputs ...string) ([]gemini.EmbeddingResult, error)
}

type ChatRetrievalService interface {
	Retrieve(ctx context.Context, req requestdto.ChatRequest) (*domain.ChatRetrievalResult, error)
}

type chatRetrievalService struct {
	repo     domain.ChatRetrievalRepository
	embedder chatQueryEmbedder
	limit    int
	timeout  time.Duration
}

func NewChatRetrievalService(repo domain.ChatRetrievalRepository, embedder chatQueryEmbedder, limit int) (ChatRetrievalService, error) {
	if repo == nil {
		return nil, fmt.Errorf("chat retrieval service: repository is required")
	}
	if embedder == nil {
		return nil, fmt.Errorf("chat retrieval service: embedder is required")
	}
	if limit <= 0 {
		limit = 5
	}
	timeout := time.Duration(config.Env.ChatRetrievalTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 1500 * time.Millisecond
	}
	return &chatRetrievalService{repo: repo, embedder: embedder, limit: limit, timeout: timeout}, nil
}

func (s *chatRetrievalService) Retrieve(ctx context.Context, req requestdto.ChatRequest) (*domain.ChatRetrievalResult, error) {
	if err := validator.Validate.Struct(req); err != nil {
		return nil, err
	}

	maxDocs := req.MaxDocuments
	if maxDocs <= 0 {
		maxDocs = s.limit
	}

	filters := domain.ChatRetrievalFilters{
		Query:            req.Message,
		TransactionType:  req.Filters.TransactionType,
		CategoryID:       req.Filters.CategoryID,
		LocationProvince: req.Filters.LocationProvince,
		LocationCity:     req.Filters.LocationCity,
		PriceMin:         req.Filters.PriceMin,
		PriceMax:         req.Filters.PriceMax,
	}

	retrievalCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	documents := make([]domain.ChatRetrievalDocument, 0, maxDocs)
	if req.ListingID != nil {
		document, err := s.repo.FetchDocumentByID(retrievalCtx, *req.ListingID)
		switch {
		case err == nil && document != nil && isPublicListingStatus(document.Status):
			documents = append(documents, *document)
		case err == nil || errors.Is(err, domain.ErrNotFound):
		case err != nil:
			return nil, fmt.Errorf("fetch chat retrieval document by id: %w", err)
		}
	}

	if len(documents) < maxDocs {
		embeddings, err := s.embedder.EmbedQuery(ctx, req.Message)
		if err != nil {
			return nil, err
		}
		if len(embeddings) == 0 || len(embeddings[0].Values) == 0 {
			if len(documents) == 0 {
				return &domain.ChatRetrievalResult{Grounding: domain.ChatGrounding{IsDegraded: true, DegradedReason: "empty_query_embedding"}}, nil
			}
			return buildChatRetrievalResult(documents), nil
		}

		fetchedDocuments, err := s.repo.FetchDocuments(retrievalCtx, filters, embeddings[0].Values, maxDocs)
		if err != nil {
			return nil, fmt.Errorf("fetch chat retrieval documents: %w", err)
		}
		documents = mergeChatRetrievalDocuments(documents, rerankChatRetrievalDocuments(fetchedDocuments, req.Message, filters), maxDocs)
	}

	result := buildChatRetrievalResult(documents)
	if len(result.Grounding.ListingIDs) == 0 {
		result.Grounding.IsDegraded = true
		result.Grounding.DegradedReason = "no_active_grounded_results"
	}

	return result, nil
}

func buildChatRetrievalResult(documents []domain.ChatRetrievalDocument) *domain.ChatRetrievalResult {
	result := &domain.ChatRetrievalResult{
		Documents: documents,
		Grounding: domain.ChatGrounding{
			ListingIDs:   make([]uuid.UUID, 0, len(documents)),
			ListingSlugs: make([]string, 0, len(documents)),
			Citations:    make([]domain.ChatGroundingMetadata, 0, len(documents)),
		},
	}

	now := time.Now().UTC()
	for _, document := range documents {
		if !isPublicListingStatus(document.Status) {
			continue
		}
		result.Grounding.ListingIDs = append(result.Grounding.ListingIDs, document.ListingID)
		result.Grounding.ListingSlugs = append(result.Grounding.ListingSlugs, document.Slug)
		result.Grounding.Citations = append(result.Grounding.Citations, domain.ChatGroundingMetadata{
			DocumentID:    document.ListingID,
			ListingSlug:   document.Slug,
			DocumentTitle: document.Title,
			Source:        "chat_retrieval_index",
			Section:       "listing",
			Excerpt:       document.DescriptionExcerpt,
			RetrievedAt:   now,
		})
	}

	return result
}

func rerankChatRetrievalDocuments(documents []domain.ChatRetrievalDocument, query string, filters domain.ChatRetrievalFilters) []domain.ChatRetrievalDocument {
	type rankedDocument struct {
		document domain.ChatRetrievalDocument
		score    int
	}

	normalizedQuery := normalizeChatRetrievalSearchText(query)
	querySlug := strings.ReplaceAll(normalizedQuery, " ", "-")
	ranked := make([]rankedDocument, 0, len(documents))
	for _, document := range documents {
		ranked = append(ranked, rankedDocument{
			document: document,
			score:    chatRetrievalDocumentScore(document, normalizedQuery, querySlug, filters),
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	reordered := make([]domain.ChatRetrievalDocument, 0, len(ranked))
	for _, item := range ranked {
		reordered = append(reordered, item.document)
	}

	return reordered
}

func mergeChatRetrievalDocuments(prioritized []domain.ChatRetrievalDocument, fetched []domain.ChatRetrievalDocument, limit int) []domain.ChatRetrievalDocument {
	if limit <= 0 {
		limit = len(prioritized) + len(fetched)
	}

	merged := make([]domain.ChatRetrievalDocument, 0, min(limit, len(prioritized)+len(fetched)))
	seen := make(map[uuid.UUID]struct{}, len(prioritized)+len(fetched))
	appendUnique := func(document domain.ChatRetrievalDocument) {
		if len(merged) >= limit {
			return
		}
		if _, ok := seen[document.ListingID]; ok {
			return
		}
		seen[document.ListingID] = struct{}{}
		merged = append(merged, document)
	}

	for _, document := range prioritized {
		appendUnique(document)
	}
	for _, document := range fetched {
		appendUnique(document)
	}

	return merged
}

func chatRetrievalDocumentScore(document domain.ChatRetrievalDocument, normalizedQuery string, querySlug string, filters domain.ChatRetrievalFilters) int {
	score := 0

	if querySlug != "" && strings.EqualFold(document.Slug, querySlug) {
		score += 480
	}
	score += weightedExactishMatchScore(normalizedQuery, normalizeChatRetrievalSearchText(document.Title), 360, 220)

	if document.Category != nil {
		if querySlug != "" && strings.EqualFold(document.Category.Slug, querySlug) {
			score += 260
		}
		score += weightedExactishMatchScore(normalizedQuery, normalizeChatRetrievalSearchText(document.Category.Name), 220, 170)
		if filters.CategoryID != nil && document.Category.ID == *filters.CategoryID {
			score += 120
		}
	}

	if document.LocationCity != nil {
		score += weightedExactishMatchScore(normalizedQuery, normalizeChatRetrievalSearchText(*document.LocationCity), 200, 150)
		if filters.LocationCity != "" && strings.EqualFold(*document.LocationCity, filters.LocationCity) {
			score += 90
		}
	}
	if document.LocationProvince != nil {
		score += weightedExactishMatchScore(normalizedQuery, normalizeChatRetrievalSearchText(*document.LocationProvince), 180, 130)
		if filters.LocationProvince != "" && strings.EqualFold(*document.LocationProvince, filters.LocationProvince) {
			score += 70
		}
	}
	if filters.TransactionType != "" && strings.EqualFold(document.TransactionType, filters.TransactionType) {
		score += 50
	}

	titleTokens := tokenOverlapCount(normalizedQuery, normalizeChatRetrievalSearchText(document.Title))
	score += min(titleTokens, 4) * 20
	if normalizedQuery != "" && strings.Contains(normalizeChatRetrievalSearchText(document.DescriptionExcerpt), normalizedQuery) {
		score += 25
	}
	if document.IsFeatured {
		score += 5
	}

	return score
}

func weightedExactishMatchScore(query string, field string, exactScore int, containedScore int) int {
	if query == "" || field == "" {
		return 0
	}
	if query == field {
		return exactScore
	}
	if strings.Contains(field, query) || strings.Contains(query, field) {
		return containedScore
	}
	return 0
}

func tokenOverlapCount(query string, field string) int {
	if query == "" || field == "" {
		return 0
	}
	count := 0
	seen := make(map[string]struct{})
	for _, token := range strings.Fields(query) {
		if len(token) < 3 {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		if strings.Contains(field, token) {
			count++
			seen[token] = struct{}{}
		}
	}
	return count
}

func normalizeChatRetrievalSearchText(input string) string {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
