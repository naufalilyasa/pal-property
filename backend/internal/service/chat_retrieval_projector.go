package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/pkg/gemini"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
)

const chatEmbeddingDimensions = 768

type chatRetrievalProjector struct {
	index    string
	client   *searchindex.Client
	listings domain.ListingRepository
	embedder chatDocumentEmbedder
}

type chatDocumentEmbedder interface {
	EmbedDocument(ctx context.Context, inputs ...string) ([]gemini.EmbeddingResult, error)
}

type chatCategoryDocument struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type chatRetrievalSearchDocument struct {
	ListingID           uuid.UUID             `json:"listing_id"`
	Category            *chatCategoryDocument `json:"category,omitempty"`
	Title               string                `json:"title"`
	Slug                string                `json:"slug"`
	LexicalText         string                `json:"lexical_text,omitempty"`
	LexicalSearchText   string                `json:"lexical_search_text,omitempty"`
	DescriptionExcerpt  *string               `json:"description_excerpt,omitempty"`
	TransactionType     string                `json:"transaction_type"`
	Price               int64                 `json:"price"`
	Currency            string                `json:"currency"`
	LocationProvince    *string               `json:"location_province,omitempty"`
	LocationCity        *string               `json:"location_city,omitempty"`
	LocationDistrict    *string               `json:"location_district,omitempty"`
	LocationVillage     *string               `json:"location_village,omitempty"`
	Status              string                `json:"status"`
	IsFeatured          bool                  `json:"is_featured"`
	PrimaryImageURL     *string               `json:"primary_image_url,omitempty"`
	ImageURLs           []string              `json:"image_urls,omitempty"`
	BedroomCount        *int                  `json:"bedroom_count,omitempty"`
	BathroomCount       *int                  `json:"bathroom_count,omitempty"`
	LandAreaSqm         *int                  `json:"land_area_sqm,omitempty"`
	BuildingAreaSqm     *int                  `json:"building_area_sqm,omitempty"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	Embedding           []float32             `json:"embedding,omitempty"`
	facilityTokens      []string
	specificationTokens []string
}

func NewChatRetrievalProjector(index string, client *searchindex.Client, listings domain.ListingRepository, embedder chatDocumentEmbedder) (domain.SearchProjector, error) {
	if index == "" {
		return nil, fmt.Errorf("chat retrieval projector: index is required")
	}
	if client == nil {
		return nil, fmt.Errorf("chat retrieval projector: search client is required")
	}
	if listings == nil {
		return nil, fmt.Errorf("chat retrieval projector: listing repository is required")
	}
	if embedder == nil {
		return nil, fmt.Errorf("chat retrieval projector: document embedder is required")
	}

	return &chatRetrievalProjector{index: index, client: client, listings: listings, embedder: embedder}, nil
}

func (p *chatRetrievalProjector) HandleListingEvent(ctx context.Context, event domain.ListingEvent) error {
	if event.Metadata.EventType == domain.EventTypeListingDeleted || !isPublicListingStatus(event.Payload.Status) {
		return p.client.DeleteDocument(ctx, p.index, event.Payload.ID.String())
	}

	document, err := p.buildDocumentFromEvent(ctx, event)
	if err != nil {
		return err
	}
	return p.client.UpsertDocument(ctx, p.index, event.Payload.ID.String(), document)
}

func (p *chatRetrievalProjector) HandleCategoryEvent(ctx context.Context, event domain.CategoryEvent) error {
	if event.Payload.ID == uuid.Nil {
		return nil
	}

	listings, err := p.listings.FindByCategoryID(ctx, event.Payload.ID)
	if err != nil {
		return err
	}

	for _, listing := range listings {
		if !isPublicListingStatus(listing.Status) {
			if err := p.client.DeleteDocument(ctx, p.index, listing.ID.String()); err != nil {
				return err
			}
			continue
		}

		document, err := p.buildDocumentFromEntity(ctx, listing)
		if err != nil {
			return err
		}
		if err := p.client.UpsertDocument(ctx, p.index, listing.ID.String(), document); err != nil {
			return err
		}
	}

	return nil
}

func RebuildChatRetrievalIndex(ctx context.Context, repo domain.ListingRepository, client *searchindex.Client, embedder chatDocumentEmbedder, index string, pageSize int) error {
	if repo == nil {
		return fmt.Errorf("chat retrieval projector: listing repository is required")
	}
	if client == nil {
		return fmt.Errorf("chat retrieval projector: search client is required")
	}
	if embedder == nil {
		return fmt.Errorf("chat retrieval projector: document embedder is required")
	}
	if index == "" {
		return fmt.Errorf("chat retrieval projector: index is required")
	}
	if pageSize <= 0 {
		pageSize = 200
	}

	if err := client.RecreateIndex(ctx, index, ChatRetrievalIndexMapping()); err != nil {
		return err
	}

	for page := 1; ; page++ {
		listings, total, err := repo.List(ctx, domain.ListingFilter{Page: page, Limit: pageSize})
		if err != nil {
			return err
		}

		for _, listing := range listings {
			if !isPublicListingStatus(listing.Status) {
				continue
			}
			document, err := buildChatRetrievalDocumentFromEntity(ctx, embedder, listing)
			if err != nil {
				return err
			}
			if err := client.UpsertDocument(ctx, index, listing.ID.String(), document); err != nil {
				return err
			}
		}

		if int64(page*pageSize) >= total || len(listings) == 0 {
			return nil
		}
	}
}

func ChatRetrievalIndexMapping() map[string]any {
	return map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"listing_id":          map[string]any{"type": "keyword"},
				"category":            map[string]any{"properties": map[string]any{"id": map[string]any{"type": "keyword"}, "name": map[string]any{"type": "keyword"}, "slug": map[string]any{"type": "keyword"}}},
				"title":               map[string]any{"type": "text"},
				"slug":                map[string]any{"type": "keyword"},
				"lexical_text":        map[string]any{"type": "text"},
				"lexical_search_text": map[string]any{"type": "text"},
				"description_excerpt": map[string]any{"type": "text"},
				"transaction_type":    map[string]any{"type": "keyword"},
				"price":               map[string]any{"type": "long"},
				"currency":            map[string]any{"type": "keyword"},
				"location_province":   map[string]any{"type": "keyword"},
				"location_city":       map[string]any{"type": "keyword"},
				"location_district":   map[string]any{"type": "keyword"},
				"location_village":    map[string]any{"type": "keyword"},
				"status":              map[string]any{"type": "keyword"},
				"is_featured":         map[string]any{"type": "boolean"},
				"primary_image_url":   map[string]any{"type": "keyword", "index": false},
				"image_urls":          map[string]any{"type": "keyword", "index": false},
				"bedroom_count":       map[string]any{"type": "integer"},
				"bathroom_count":      map[string]any{"type": "integer"},
				"land_area_sqm":       map[string]any{"type": "integer"},
				"building_area_sqm":   map[string]any{"type": "integer"},
				"created_at":          map[string]any{"type": "date"},
				"updated_at":          map[string]any{"type": "date"},
				"embedding": map[string]any{
					"type":       "dense_vector",
					"dims":       chatEmbeddingDimensions,
					"index":      true,
					"similarity": "cosine",
				},
			},
		},
	}
}

func chatRetrievalDocumentFromEvent(event domain.ListingEvent) chatRetrievalSearchDocument {
	document := chatRetrievalSearchDocument{
		ListingID:          event.Payload.ID,
		Category:           mapChatCategoryReference(event.Payload.Category),
		Title:              event.Payload.Title,
		Slug:               event.Payload.Slug,
		DescriptionExcerpt: event.Payload.Description,
		TransactionType:    event.Payload.TransactionType,
		Price:              event.Payload.Price,
		Currency:           event.Payload.Currency,
		LocationProvince:   event.Payload.LocationProvince,
		LocationCity:       event.Payload.LocationCity,
		LocationDistrict:   event.Payload.LocationDistrict,
		LocationVillage:    event.Payload.LocationVillage,
		Status:             event.Payload.Status,
		IsFeatured:         event.Payload.IsFeatured,
		PrimaryImageURL:    primaryImageURLFromEvent(event.Payload.Images),
		ImageURLs:          imageURLsFromEvent(event.Payload.Images),
		BedroomCount:       event.Payload.BedroomCount,
		BathroomCount:      event.Payload.BathroomCount,
		LandAreaSqm:        event.Payload.LandAreaSqm,
		BuildingAreaSqm:    event.Payload.BuildingAreaSqm,
		CreatedAt:          event.Payload.CreatedAt,
		UpdatedAt:          event.Payload.UpdatedAt,
	}
	enrichChatDocumentTokens(&document, event.Payload.Facilities, event.Payload.Specifications)
	document.LexicalText = buildChatLexicalText(document)
	document.LexicalSearchText = buildChatLexicalSearchText(document)
	return document
}

func chatRetrievalDocumentFromEntity(listing *entity.Listing) chatRetrievalSearchDocument {
	document := chatRetrievalSearchDocument{
		ListingID:          listing.ID,
		Category:           mapChatCategoryEntity(listing.Category),
		Title:              listing.Title,
		Slug:               listing.Slug,
		DescriptionExcerpt: listing.Description,
		TransactionType:    listing.TransactionType,
		Price:              listing.Price,
		Currency:           listing.Currency,
		LocationProvince:   listing.LocationProvince,
		LocationCity:       listing.LocationCity,
		LocationDistrict:   listing.LocationDistrict,
		LocationVillage:    listing.LocationVillage,
		Status:             listing.Status,
		IsFeatured:         listing.IsFeatured,
		PrimaryImageURL:    primaryImageURLFromEntities(listing.Images),
		ImageURLs:          imageURLsFromEntities(listing.Images),
		BedroomCount:       listing.BedroomCount,
		BathroomCount:      listing.BathroomCount,
		LandAreaSqm:        listing.LandAreaSqm,
		BuildingAreaSqm:    listing.BuildingAreaSqm,
		CreatedAt:          listing.CreatedAt,
		UpdatedAt:          listing.UpdatedAt,
	}
	enrichChatDocumentTokens(&document, json.RawMessage(listing.Facilities), json.RawMessage(listing.Specifications))
	document.LexicalText = buildChatLexicalText(document)
	document.LexicalSearchText = buildChatLexicalSearchText(document)
	return document
}

func (p *chatRetrievalProjector) buildDocumentFromEvent(ctx context.Context, event domain.ListingEvent) (chatRetrievalSearchDocument, error) {
	document := chatRetrievalDocumentFromEvent(event)
	return p.attachEmbedding(ctx, document)
}

func (p *chatRetrievalProjector) buildDocumentFromEntity(ctx context.Context, listing *entity.Listing) (chatRetrievalSearchDocument, error) {
	return buildChatRetrievalDocumentFromEntity(ctx, p.embedder, listing)
}

func buildChatRetrievalDocumentFromEntity(ctx context.Context, embedder chatDocumentEmbedder, listing *entity.Listing) (chatRetrievalSearchDocument, error) {
	document := chatRetrievalDocumentFromEntity(listing)
	embeddings, err := embedder.EmbedDocument(ctx, buildChatEmbeddingInput(document))
	if err != nil {
		return chatRetrievalSearchDocument{}, fmt.Errorf("embed chat retrieval document: %w", err)
	}
	if len(embeddings) > 0 {
		document.Embedding = toFloat32Slice(embeddings[0].Values)
	}
	return document, nil
}

func (p *chatRetrievalProjector) attachEmbedding(ctx context.Context, document chatRetrievalSearchDocument) (chatRetrievalSearchDocument, error) {
	embeddings, err := p.embedder.EmbedDocument(ctx, buildChatEmbeddingInput(document))
	if err != nil {
		return chatRetrievalSearchDocument{}, fmt.Errorf("embed chat retrieval document: %w", err)
	}
	if len(embeddings) > 0 {
		document.Embedding = toFloat32Slice(embeddings[0].Values)
	}
	return document, nil
}

func buildChatEmbeddingInput(document chatRetrievalSearchDocument) string {
	parts := []string{
		labeledChatEmbeddingValue("Title", document.Title),
		labeledChatEmbeddingValue("Slug", document.Slug),
		labeledChatEmbeddingValue("Transaction type", document.TransactionType),
		labeledChatEmbeddingValue("Category", chatCategoryEmbeddingValue(document.Category)),
		labeledChatEmbeddingValue("Location", chatLocationPath(document)),
		labeledChatEmbeddingValue("Price", chatPriceEmbeddingValue(document.Price, document.Currency)),
		labeledChatEmbeddingValue("Bedrooms", optionalIntEmbeddingValue(document.BedroomCount)),
		labeledChatEmbeddingValue("Bathrooms", optionalIntEmbeddingValue(document.BathroomCount)),
		labeledChatEmbeddingValue("Land area sqm", optionalIntEmbeddingValue(document.LandAreaSqm)),
		labeledChatEmbeddingValue("Building area sqm", optionalIntEmbeddingValue(document.BuildingAreaSqm)),
		labeledChatEmbeddingValue("Description excerpt", optionalStringEmbeddingValue(document.DescriptionExcerpt)),
		labeledChatEmbeddingValue("Lexical search text", document.LexicalSearchText),
	}
	return strings.Join(nonEmptyChatParts(parts...), "\n")
}

func buildChatLexicalText(document chatRetrievalSearchDocument) string {
	parts := []string{
		document.Title,
		document.Slug,
		document.TransactionType,
		chatCategoryEmbeddingValue(document.Category),
		chatLocationPath(document),
		chatPriceEmbeddingValue(document.Price, document.Currency),
		optionalIntEmbeddingValue(document.BedroomCount),
		optionalIntEmbeddingValue(document.BathroomCount),
		optionalIntEmbeddingValue(document.LandAreaSqm),
		optionalIntEmbeddingValue(document.BuildingAreaSqm),
		optionalStringEmbeddingValue(document.DescriptionExcerpt),
	}
	if facilityText := chatFacilitySentence(document); facilityText != "" {
		parts = append(parts, facilityText)
	}
	if specificationText := chatSpecificationSentence(document); specificationText != "" {
		parts = append(parts, specificationText)
	}
	return strings.Join(nonEmptyChatParts(parts...), "\n")
}

func buildChatLexicalSearchText(document chatRetrievalSearchDocument) string {
	parts := []string{
		labeledChatEmbeddingValue("Title", document.Title),
		labeledChatEmbeddingValue("Transaction type", document.TransactionType),
		labeledChatEmbeddingValue("Category", chatCategoryEmbeddingValue(document.Category)),
		labeledChatEmbeddingValue("Location", chatLocationPath(document)),
		labeledChatEmbeddingValue("Province", optionalStringEmbeddingValue(document.LocationProvince)),
		labeledChatEmbeddingValue("City", optionalStringEmbeddingValue(document.LocationCity)),
		labeledChatEmbeddingValue("District", optionalStringEmbeddingValue(document.LocationDistrict)),
		labeledChatEmbeddingValue("Village", optionalStringEmbeddingValue(document.LocationVillage)),
		labeledChatEmbeddingValue("Price", chatPriceEmbeddingValue(document.Price, document.Currency)),
		labeledChatEmbeddingValue("Bedrooms", optionalIntEmbeddingValue(document.BedroomCount)),
		labeledChatEmbeddingValue("Bathrooms", optionalIntEmbeddingValue(document.BathroomCount)),
		labeledChatEmbeddingValue("Land area sqm", optionalIntEmbeddingValue(document.LandAreaSqm)),
		labeledChatEmbeddingValue("Building area sqm", optionalIntEmbeddingValue(document.BuildingAreaSqm)),
		labeledChatEmbeddingValue("Description", optionalStringEmbeddingValue(document.DescriptionExcerpt)),
		labeledChatEmbeddingValue("Facilities", joinNormalizedTokens(document.facilityTokens)),
		labeledChatEmbeddingValue("Specifications", joinNormalizedTokens(document.specificationTokens)),
	}
	return strings.Join(nonEmptyChatParts(parts...), "\n")
}

var specificationFieldOrder = []string{
	"bedrooms",
	"bathrooms",
	"land_area_sqm",
	"building_area_sqm",
}

var specificationFieldLabels = map[string]string{
	"bedrooms":          "Bedrooms",
	"bathrooms":         "Bathrooms",
	"land_area_sqm":     "Land area sqm",
	"building_area_sqm": "Building area sqm",
}

func enrichChatDocumentTokens(document *chatRetrievalSearchDocument, facilities, specifications json.RawMessage) {
	if document == nil {
		return
	}
	document.facilityTokens = parseFacilityTokens(facilities)
	document.specificationTokens = parseSpecificationTokens(specifications)
}

func chatFacilitySentence(document chatRetrievalSearchDocument) string {
	if text := joinNormalizedTokens(document.facilityTokens); text != "" {
		return "Facilities: " + text
	}
	return ""
}

func chatSpecificationSentence(document chatRetrievalSearchDocument) string {
	if text := joinNormalizedTokens(document.specificationTokens); text != "" {
		return "Specifications: " + text
	}
	return ""
}

func joinNormalizedTokens(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	return strings.Join(tokens, ", ")
}

func parseFacilityTokens(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	tokens := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := normalizeFacilityValue(value)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		tokens = append(tokens, normalized)
	}
	return tokens
}

func parseSpecificationTokens(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	tokens := make([]string, 0, len(values))
	for _, field := range specificationFieldOrder {
		data, ok := values[field]
		if !ok {
			continue
		}
		var value int
		if err := json.Unmarshal(data, &value); err != nil {
			continue
		}
		label := specificationFieldLabels[field]
		tokens = append(tokens, fmt.Sprintf("%s: %d", label, value))
	}
	return tokens
}

func normalizeFacilityValue(value string) string {
	if value == "" {
		return ""
	}
	cleaned := strings.ReplaceAll(value, "_", " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	return strings.TrimSpace(cleaned)
}

func nonEmptyChatParts(parts ...string) []string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		filtered = append(filtered, trimmed)
	}
	return filtered
}

func labeledChatEmbeddingValue(label, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%s: %s", label, trimmed)
}

func chatCategoryEmbeddingValue(category *chatCategoryDocument) string {
	if category == nil {
		return ""
	}
	return strings.Join(nonEmptyChatParts(category.Name, category.Slug), " ")
}

func chatLocationPath(document chatRetrievalSearchDocument) string {
	return strings.Join(nonEmptyChatParts(
		optionalStringEmbeddingValue(document.LocationVillage),
		optionalStringEmbeddingValue(document.LocationDistrict),
		optionalStringEmbeddingValue(document.LocationCity),
		optionalStringEmbeddingValue(document.LocationProvince),
	), ", ")
}

func chatPriceEmbeddingValue(price int64, currency string) string {
	trimmedCurrency := strings.TrimSpace(currency)
	if trimmedCurrency == "" {
		return fmt.Sprintf("%d", price)
	}
	return fmt.Sprintf("%s %d", trimmedCurrency, price)
}

func optionalIntEmbeddingValue(value *int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}

func optionalStringEmbeddingValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func toFloat32Slice(values []float64) []float32 {
	converted := make([]float32, 0, len(values))
	for _, value := range values {
		converted = append(converted, float32(value))
	}
	return converted
}

func mapChatCategoryReference(ref *domain.CategoryEventReference) *chatCategoryDocument {
	if ref == nil {
		return nil
	}
	return &chatCategoryDocument{ID: ref.ID, Name: ref.Name, Slug: ref.Slug}
}

func mapChatCategoryEntity(category *entity.Category) *chatCategoryDocument {
	if category == nil {
		return nil
	}
	return &chatCategoryDocument{ID: category.ID, Name: category.Name, Slug: category.Slug}
}
