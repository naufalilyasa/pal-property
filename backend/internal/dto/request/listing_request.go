package request

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Specifications struct {
	Bedrooms        int `json:"bedrooms" validate:"gte=0,lte=50"`
	Bathrooms       int `json:"bathrooms" validate:"gte=0,lte=50"`
	LandAreaSqm     int `json:"land_area_sqm" validate:"gte=0"`
	BuildingAreaSqm int `json:"building_area_sqm" validate:"gte=0"`

	hasBedrooms        bool
	hasBathrooms       bool
	hasLandAreaSqm     bool
	hasBuildingAreaSqm bool
}

func (s *Specifications) UnmarshalJSON(data []byte) error {
	type alias Specifications
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*s = Specifications(decoded)
	_, s.hasBedrooms = raw["bedrooms"]
	_, s.hasBathrooms = raw["bathrooms"]
	_, s.hasLandAreaSqm = raw["land_area_sqm"]
	_, s.hasBuildingAreaSqm = raw["building_area_sqm"]
	return nil
}

func (s Specifications) HasBedrooms() bool {
	return s.hasBedrooms
}

func (s Specifications) HasBathrooms() bool {
	return s.hasBathrooms
}

func (s Specifications) HasLandAreaSqm() bool {
	return s.hasLandAreaSqm
}

func (s Specifications) HasBuildingAreaSqm() bool {
	return s.hasBuildingAreaSqm
}

func (s Specifications) MarshalJSON() ([]byte, error) {
	payload := map[string]int{}
	if s.hasBedrooms || s.Bedrooms != 0 {
		payload["bedrooms"] = s.Bedrooms
	}
	if s.hasBathrooms || s.Bathrooms != 0 {
		payload["bathrooms"] = s.Bathrooms
	}
	if s.hasLandAreaSqm || s.LandAreaSqm != 0 {
		payload["land_area_sqm"] = s.LandAreaSqm
	}
	if s.hasBuildingAreaSqm || s.BuildingAreaSqm != 0 {
		payload["building_area_sqm"] = s.BuildingAreaSqm
	}
	return json.Marshal(payload)
}

func (s *Specifications) SetBedrooms(value int) {
	s.Bedrooms = value
	s.hasBedrooms = true
}

func (s *Specifications) SetBathrooms(value int) {
	s.Bathrooms = value
	s.hasBathrooms = true
}

func (s *Specifications) SetLandAreaSqm(value int) {
	s.LandAreaSqm = value
	s.hasLandAreaSqm = true
}

func (s *Specifications) SetBuildingAreaSqm(value int) {
	s.BuildingAreaSqm = value
	s.hasBuildingAreaSqm = true
}

type CreateListingRequest struct {
	CategoryID        *uuid.UUID     `json:"category_id"`
	Title             string         `json:"title" validate:"required,min=5,max=255"`
	Description       *string        `json:"description" validate:"omitempty,max=5000"`
	TransactionType   string         `json:"transaction_type" validate:"omitempty,oneof=sale rent"`
	Price             int64          `json:"price" validate:"required,gt=0"`
	Currency          *string        `json:"currency" validate:"omitempty,oneof=IDR"`
	IsNegotiable      *bool          `json:"is_negotiable"`
	SpecialOffers     []string       `json:"special_offers" validate:"omitempty,dive,oneof=Promo DP_0 Aset_Bank Turun_Harga"`
	LocationProvince  *string        `json:"location_province" validate:"omitempty,max=100"`
	LocationCity      *string        `json:"location_city" validate:"omitempty,max=100"`
	LocationDistrict  *string        `json:"location_district" validate:"omitempty,max=100"`
	AddressDetail     *string        `json:"address_detail" validate:"omitempty,max=5000"`
	Latitude          *float64       `json:"latitude" validate:"omitempty,gte=-90,lte=90"`
	Longitude         *float64       `json:"longitude" validate:"omitempty,gte=-180,lte=180"`
	BedroomCount      *int           `json:"bedroom_count" validate:"omitempty,gte=0,lte=50"`
	BathroomCount     *int           `json:"bathroom_count" validate:"omitempty,gte=0,lte=50"`
	FloorCount        *int           `json:"floor_count" validate:"omitempty,gte=0,lte=200"`
	CarportCapacity   *int           `json:"carport_capacity" validate:"omitempty,gte=0,lte=100"`
	LandAreaSqm       *int           `json:"land_area_sqm" validate:"omitempty,gte=0"`
	BuildingAreaSqm   *int           `json:"building_area_sqm" validate:"omitempty,gte=0"`
	CertificateType   *string        `json:"certificate_type" validate:"omitempty,oneof=SHM HGB Strata Girik AJB SHMRS"`
	Condition         *string        `json:"condition" validate:"omitempty,oneof=new second"`
	Furnishing        *string        `json:"furnishing" validate:"omitempty,oneof=unfurnished semi fully"`
	ElectricalPowerVA *int           `json:"electrical_power_va" validate:"omitempty,gte=0"`
	FacingDirection   *string        `json:"facing_direction" validate:"omitempty,oneof=north south east west northeast northwest southeast southwest"`
	YearBuilt         *int           `json:"year_built" validate:"omitempty,gte=1800,lte=2100"`
	Facilities        []string       `json:"facilities" validate:"omitempty,dive,oneof=AC CCTV Wifi Water_Heater Carport Garden Pool Gym Playground Security"`
	Status            string         `json:"status" validate:"required,oneof=active inactive sold draft archived"`
	Specifications    Specifications `json:"specifications"`
}

type UpdateListingRequest struct {
	CategoryID        *uuid.UUID      `json:"category_id"`
	Title             *string         `json:"title" validate:"omitempty,min=5,max=255"`
	Description       *string         `json:"description" validate:"omitempty,max=5000"`
	TransactionType   *string         `json:"transaction_type" validate:"omitempty,oneof=sale rent"`
	Price             *int64          `json:"price" validate:"omitempty,gt=0"`
	Currency          *string         `json:"currency" validate:"omitempty,oneof=IDR"`
	IsNegotiable      *bool           `json:"is_negotiable"`
	SpecialOffers     *[]string       `json:"special_offers" validate:"omitempty,dive,oneof=Promo DP_0 Aset_Bank Turun_Harga"`
	LocationProvince  *string         `json:"location_province" validate:"omitempty,max=100"`
	LocationCity      *string         `json:"location_city" validate:"omitempty,max=100"`
	LocationDistrict  *string         `json:"location_district" validate:"omitempty,max=100"`
	AddressDetail     *string         `json:"address_detail" validate:"omitempty,max=5000"`
	Latitude          *float64        `json:"latitude" validate:"omitempty,gte=-90,lte=90"`
	Longitude         *float64        `json:"longitude" validate:"omitempty,gte=-180,lte=180"`
	BedroomCount      *int            `json:"bedroom_count" validate:"omitempty,gte=0,lte=50"`
	BathroomCount     *int            `json:"bathroom_count" validate:"omitempty,gte=0,lte=50"`
	FloorCount        *int            `json:"floor_count" validate:"omitempty,gte=0,lte=200"`
	CarportCapacity   *int            `json:"carport_capacity" validate:"omitempty,gte=0,lte=100"`
	LandAreaSqm       *int            `json:"land_area_sqm" validate:"omitempty,gte=0"`
	BuildingAreaSqm   *int            `json:"building_area_sqm" validate:"omitempty,gte=0"`
	CertificateType   *string         `json:"certificate_type" validate:"omitempty,oneof=SHM HGB Strata Girik AJB SHMRS"`
	Condition         *string         `json:"condition" validate:"omitempty,oneof=new second"`
	Furnishing        *string         `json:"furnishing" validate:"omitempty,oneof=unfurnished semi fully"`
	ElectricalPowerVA *int            `json:"electrical_power_va" validate:"omitempty,gte=0"`
	FacingDirection   *string         `json:"facing_direction" validate:"omitempty,oneof=north south east west northeast northwest southeast southwest"`
	YearBuilt         *int            `json:"year_built" validate:"omitempty,gte=1800,lte=2100"`
	Facilities        *[]string       `json:"facilities" validate:"omitempty,dive,oneof=AC CCTV Wifi Water_Heater Carport Garden Pool Gym Playground Security"`
	Status            *string         `json:"status" validate:"omitempty,oneof=active inactive sold draft archived"`
	Specifications    *Specifications `json:"specifications"`
}

type ReorderListingImagesRequest struct {
	OrderedImageIDs []uuid.UUID `json:"ordered_image_ids"`
}
