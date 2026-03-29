package response

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ListingResponse struct {
	ID                uuid.UUID               `json:"id"`
	UserID            uuid.UUID               `json:"user_id"`
	CategoryID        *uuid.UUID              `json:"category_id"`
	Category          *CategoryShortResponse  `json:"category,omitempty"`
	Title             string                  `json:"title"`
	Slug              string                  `json:"slug"`
	Description       *string                 `json:"description"`
	TransactionType   string                  `json:"transaction_type"`
	Price             int64                   `json:"price"`
	Currency          string                  `json:"currency"`
	IsNegotiable      bool                    `json:"is_negotiable"`
	SpecialOffers     datatypes.JSON          `json:"special_offers"`
	LocationProvince  *string                 `json:"location_province"`
	LocationCity      *string                 `json:"location_city"`
	LocationDistrict  *string                 `json:"location_district"`
	AddressDetail     *string                 `json:"address_detail"`
	Latitude          *float64                `json:"latitude"`
	Longitude         *float64                `json:"longitude"`
	BedroomCount      *int                    `json:"bedroom_count"`
	BathroomCount     *int                    `json:"bathroom_count"`
	FloorCount        *int                    `json:"floor_count"`
	CarportCapacity   *int                    `json:"carport_capacity"`
	LandAreaSqm       *int                    `json:"land_area_sqm"`
	BuildingAreaSqm   *int                    `json:"building_area_sqm"`
	CertificateType   *string                 `json:"certificate_type"`
	Condition         *string                 `json:"condition"`
	Furnishing        *string                 `json:"furnishing"`
	ElectricalPowerVA *int                    `json:"electrical_power_va"`
	FacingDirection   *string                 `json:"facing_direction"`
	YearBuilt         *int                    `json:"year_built"`
	Facilities        datatypes.JSON          `json:"facilities"`
	Status            string                  `json:"status"`
	IsFeatured        bool                    `json:"is_featured"`
	Specifications    datatypes.JSON          `json:"specifications"`
	ViewCount         int                     `json:"view_count"`
	Images            []*ListingImageResponse `json:"images"`
	Video             *ListingVideoResponse   `json:"video,omitempty"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
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


type ListingVideoResponse struct {
	ID               uuid.UUID `json:"id"`
	URL              string    `json:"url"`
	Format           *string   `json:"format,omitempty"`
	Bytes            *int64    `json:"bytes,omitempty"`
	Width            *int      `json:"width,omitempty"`
	Height           *int      `json:"height,omitempty"`
	DurationSeconds  *int      `json:"duration_seconds,omitempty"`
	OriginalFilename *string   `json:"original_filename,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
