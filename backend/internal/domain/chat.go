package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ChatRetrievalFilters keep the retrieval layer focused on the allowlisted property fields.
type ChatRetrievalFilters struct {
	Query            string     `json:"query"`
	TransactionType  string     `json:"transaction_type"`
	CategoryID       *uuid.UUID `json:"category_id"`
	LocationProvince string     `json:"location_province"`
	LocationCity     string     `json:"location_city"`
	PriceMin         *int64     `json:"price_min"`
	PriceMax         *int64     `json:"price_max"`
}

// ChatDocumentCategory carries the minimal category data exposed to the public chat surface.
type ChatDocumentCategory struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// ChatRetrievalDocument is the public-safe property schema that can be ground-truth for Gemini.
type ChatRetrievalDocument struct {
	ListingID          uuid.UUID             `json:"listing_id"`
	Category           *ChatDocumentCategory `json:"category,omitempty"`
	Title              string                `json:"title"`
	Slug               string                `json:"slug"`
	DescriptionExcerpt string                `json:"description_excerpt,omitempty"`
	TransactionType    string                `json:"transaction_type"`
	Price              int64                 `json:"price"`
	Currency           string                `json:"currency"`
	LocationProvince   *string               `json:"location_province,omitempty"`
	LocationCity       *string               `json:"location_city,omitempty"`
	LocationDistrict   *string               `json:"location_district,omitempty"`
	LocationVillage    *string               `json:"location_village,omitempty"`
	Status             string                `json:"status"`
	IsFeatured         bool                  `json:"is_featured"`
	PrimaryImageURL    *string               `json:"primary_image_url,omitempty"`
	ImageURLs          []string              `json:"image_urls,omitempty"`
	BedroomCount       *int                  `json:"bedroom_count,omitempty"`
	BathroomCount      *int                  `json:"bathroom_count,omitempty"`
	LandAreaSqm        *int                  `json:"land_area_sqm,omitempty"`
	BuildingAreaSqm    *int                  `json:"building_area_sqm,omitempty"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
}

// ChatGroundingMetadata records which retrieval snippets were provided to the model.
type ChatGroundingMetadata struct {
	DocumentID    uuid.UUID `json:"document_id"`
	ListingSlug   string    `json:"listing_slug,omitempty"`
	DocumentTitle string    `json:"document_title"`
	Source        string    `json:"source"`
	Section       string    `json:"section"`
	Excerpt       string    `json:"excerpt,omitempty"`
	RetrievedAt   time.Time `json:"retrieved_at"`
}

type ChatGrounding struct {
	ListingIDs     []uuid.UUID             `json:"listing_ids,omitempty"`
	ListingSlugs   []string                `json:"listing_slugs,omitempty"`
	Citations      []ChatGroundingMetadata `json:"citations,omitempty"`
	IsDegraded     bool                    `json:"is_degraded"`
	DegradedReason string                  `json:"degraded_reason,omitempty"`
}

type ChatRetrievalResult struct {
	Documents []ChatRetrievalDocument `json:"documents,omitempty"`
	Grounding ChatGrounding           `json:"grounding"`
}

// ChatRetrievalRepository exposes listing data allowed for public-facing chat.
// Implementations must only materialize property fields and must not rely on the legacy
// chat_rooms/chat_messages relationships used for earlier chat features.
type ChatRetrievalRepository interface {
	FetchDocuments(ctx context.Context, filters ChatRetrievalFilters, queryVector []float64, limit int) ([]ChatRetrievalDocument, error)
	FetchDocumentByID(ctx context.Context, listingID uuid.UUID) (*ChatRetrievalDocument, error)
}
