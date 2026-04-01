package response

import (
	"time"

	"github.com/google/uuid"
)

type SearchCategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type SearchListingCardResponse struct {
	ID                 uuid.UUID               `json:"id"`
	CategoryID         *uuid.UUID              `json:"category_id,omitempty"`
	Category           *SearchCategoryResponse `json:"category,omitempty"`
	Title              string                  `json:"title"`
	Slug               string                  `json:"slug"`
	DescriptionExcerpt string                  `json:"description_excerpt,omitempty"`
	TransactionType    string                  `json:"transaction_type"`
	Price              int64                   `json:"price"`
	Currency           string                  `json:"currency"`
	LocationProvince   *string                 `json:"location_province,omitempty"`
	LocationCity       *string                 `json:"location_city,omitempty"`
	LocationDistrict   *string                 `json:"location_district,omitempty"`
	LocationVillage    *string                 `json:"location_village,omitempty"`
	Latitude           *float64                `json:"latitude,omitempty"`
	Longitude          *float64                `json:"longitude,omitempty"`
	BedroomCount       *int                    `json:"bedroom_count,omitempty"`
	BathroomCount      *int                    `json:"bathroom_count,omitempty"`
	LandAreaSqm        *int                    `json:"land_area_sqm,omitempty"`
	BuildingAreaSqm    *int                    `json:"building_area_sqm,omitempty"`
	Status             string                  `json:"status"`
	IsFeatured         bool                    `json:"is_featured"`
	PrimaryImageURL    *string                 `json:"primary_image_url,omitempty"`
	ImageURLs          []string                `json:"image_urls,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

type SearchListingsPageResponse struct {
	Items      []*SearchListingCardResponse `json:"items"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	Limit      int                          `json:"limit"`
	TotalPages int                          `json:"total_pages"`
}
