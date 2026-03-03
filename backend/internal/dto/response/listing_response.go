package response

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ListingResponse struct {
	ID               uuid.UUID      `json:"id"`
	UserID           uuid.UUID      `json:"user_id"`
	CategoryID       *uuid.UUID     `json:"category_id"`
	Title            string         `json:"title"`
	Slug             string         `json:"slug"`
	Description      *string        `json:"description"`
	Price            int64          `json:"price"`
	Currency         string         `json:"currency"`
	LocationCity     *string        `json:"location_city"`
	LocationDistrict *string        `json:"location_district"`
	AddressDetail    *string        `json:"address_detail"`
	Status           string         `json:"status"`
	IsFeatured       bool           `json:"is_featured"`
	Specifications   datatypes.JSON `json:"specifications"`
	ViewCount        int            `json:"view_count"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type PaginatedListings struct {
	Data       []*ListingResponse `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}
