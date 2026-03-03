package request

import "github.com/google/uuid"

type Specifications struct {
	Bedrooms        int `json:"bedrooms"`
	Bathrooms       int `json:"bathrooms"`
	LandAreaSqm     int `json:"land_area_sqm"`
	BuildingAreaSqm int `json:"building_area_sqm"`
}

type CreateListingRequest struct {
	CategoryID       *uuid.UUID     `json:"category_id"`
	Title            string         `json:"title" validate:"required,min=5,max=255"`
	Description      *string        `json:"description"`
	Price            int64          `json:"price" validate:"required,gt=0"`
	LocationCity     *string        `json:"location_city"`
	LocationDistrict *string        `json:"location_district"`
	AddressDetail    *string        `json:"address_detail"`
	Status           string         `json:"status" validate:"required,oneof=active inactive sold"`
	Specifications   Specifications `json:"specifications"`
}

type UpdateListingRequest struct {
	CategoryID       *uuid.UUID      `json:"category_id"`
	Title            *string         `json:"title" validate:"omitempty,min=5,max=255"`
	Description      *string         `json:"description"`
	Price            *int64          `json:"price" validate:"omitempty,gt=0"`
	LocationCity     *string         `json:"location_city"`
	LocationDistrict *string         `json:"location_district"`
	AddressDetail    *string         `json:"address_detail"`
	Status           *string         `json:"status" validate:"omitempty,oneof=active inactive sold"`
	Specifications   *Specifications `json:"specifications"`
}
