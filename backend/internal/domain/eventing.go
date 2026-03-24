package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	AggregateTypeListing  = "listing"
	AggregateTypeCategory = "category"

	EventTypeListingCreated       = "listing.created"
	EventTypeListingUpdated       = "listing.updated"
	EventTypeListingDeleted       = "listing.deleted"
	EventTypeListingImagesChanged = "listing.images_changed"
	EventTypeCategoryCreated      = "category.created"
	EventTypeCategoryUpdated      = "category.updated"
	EventTypeCategoryDeleted      = "category.deleted"
)

type EventMetadata struct {
	EventID       uuid.UUID `json:"event_id"`
	EventType     string    `json:"event_type"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   uuid.UUID `json:"aggregate_id"`
	Version       int       `json:"version"`
	OccurredAt    time.Time `json:"occurred_at"`
}

type EventEnvelope[T any] struct {
	Metadata EventMetadata `json:"metadata"`
	Payload  T             `json:"payload"`
}

type ListingEvent struct {
	Metadata EventMetadata       `json:"metadata"`
	Payload  ListingEventPayload `json:"payload"`
}

type CategoryEvent struct {
	Metadata EventMetadata        `json:"metadata"`
	Payload  CategoryEventPayload `json:"payload"`
}

type ListingEventPayload struct {
	ID               uuid.UUID                `json:"id"`
	UserID           uuid.UUID                `json:"user_id"`
	CategoryID       *uuid.UUID               `json:"category_id,omitempty"`
	Category         *CategoryEventReference  `json:"category,omitempty"`
	Title            string                   `json:"title"`
	Slug             string                   `json:"slug"`
	Description      *string                  `json:"description,omitempty"`
	Price            int64                    `json:"price"`
	Currency         string                   `json:"currency"`
	LocationCity     *string                  `json:"location_city,omitempty"`
	LocationDistrict *string                  `json:"location_district,omitempty"`
	AddressDetail    *string                  `json:"address_detail,omitempty"`
	Status           string                   `json:"status"`
	IsFeatured       bool                     `json:"is_featured"`
	Specifications   json.RawMessage          `json:"specifications"`
	ViewCount        int                      `json:"view_count"`
	Images           []ListingImageEventImage `json:"images"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
	DeletedAt        *time.Time               `json:"deleted_at,omitempty"`
}

type ListingImageEventImage struct {
	ID               uuid.UUID  `json:"id"`
	URL              string     `json:"url"`
	AssetID          *string    `json:"asset_id,omitempty"`
	PublicID         *string    `json:"public_id,omitempty"`
	Format           *string    `json:"format,omitempty"`
	Bytes            *int64     `json:"bytes,omitempty"`
	Width            *int       `json:"width,omitempty"`
	Height           *int       `json:"height,omitempty"`
	OriginalFilename *string    `json:"original_filename,omitempty"`
	IsPrimary        bool       `json:"is_primary"`
	SortOrder        int        `json:"sort_order"`
	CreatedAt        time.Time  `json:"created_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CategoryEventPayload struct {
	ID        uuid.UUID                `json:"id"`
	Name      string                   `json:"name"`
	Slug      string                   `json:"slug"`
	ParentID  *uuid.UUID               `json:"parent_id,omitempty"`
	IconURL   *string                  `json:"icon_url,omitempty"`
	Parent    *CategoryEventReference  `json:"parent,omitempty"`
	Children  []CategoryEventReference `json:"children,omitempty"`
	CreatedAt time.Time                `json:"created_at"`
	DeletedAt *time.Time               `json:"deleted_at,omitempty"`
}

type CategoryEventReference struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	IconURL *string   `json:"icon_url,omitempty"`
}
