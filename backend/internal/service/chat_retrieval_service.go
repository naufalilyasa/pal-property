package service

import (
	"context"
	"fmt"
	"time"

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

	embeddings, err := s.embedder.EmbedQuery(ctx, req.Message)
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 || len(embeddings[0].Values) == 0 {
		return &domain.ChatRetrievalResult{Grounding: domain.ChatGrounding{IsDegraded: true, DegradedReason: "empty_query_embedding"}}, nil
	}

	retrievalCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	documents, err := s.repo.FetchDocuments(retrievalCtx, filters, embeddings[0].Values, maxDocs)
	if err != nil {
		return nil, fmt.Errorf("fetch chat retrieval documents: %w", err)
	}

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

	if len(result.Grounding.ListingIDs) == 0 {
		result.Grounding.IsDegraded = true
		result.Grounding.DegradedReason = "no_active_grounded_results"
	}

	return result, nil
}
