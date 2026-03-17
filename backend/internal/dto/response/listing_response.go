package response

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ListingResponse struct {
	ID               uuid.UUID               `json:"id"`
	UserID           uuid.UUID               `json:"user_id"`
	CategoryID       *uuid.UUID              `json:"category_id"`
	Category         *CategoryShortResponse  `json:"category,omitempty"`
	Title            string                  `json:"title"`
	Slug             string                  `json:"slug"`
	Description      *string                 `json:"description"`
	Price            int64                   `json:"price"`
	Currency         string                  `json:"currency"`
	LocationCity     *string                 `json:"location_city"`
	LocationDistrict *string                 `json:"location_district"`
	AddressDetail    *string                 `json:"address_detail"`
	Status           string                  `json:"status"`
	IsFeatured       bool                    `json:"is_featured"`
	Specifications   datatypes.JSON          `json:"specifications"`
	ViewCount        int                     `json:"view_count"`
	Images           []*ListingImageResponse `json:"images"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

type ListingImageResponse struct {
	ID               uuid.UUID `json:"id"`
	URL              string    `json:"url"`
	Format           *string   `json:"format,omitempty"`
	Bytes            *int64    `json:"bytes,omitempty"`
	Width            *int      `json:"width,omitempty"`
	Height           *int      `json:"height,omitempty"`
	OriginalFilename *string   `json:"original_filename,omitempty"`
	IsPrimary        bool      `json:"is_primary"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
}

type PaginatedListings struct {
	Data       []*ListingResponse `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}
