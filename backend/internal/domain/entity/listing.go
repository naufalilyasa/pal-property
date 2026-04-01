package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Category struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	Slug      string     `gorm:"type:varchar(100);unique;not null" json:"slug"`
	ParentID  *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	IconURL   *string    `gorm:"type:text" json:"icon_url"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Listings []Listing  `gorm:"foreignKey:CategoryID" json:"listings,omitempty"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		c.ID = id
	}
	return nil
}

type Listing struct {
	BaseEntity
	UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CategoryID *uuid.UUID `gorm:"type:uuid" json:"category_id"`

	Title       string  `gorm:"type:varchar(255);not null" json:"title"`
	Slug        string  `gorm:"type:varchar(255);unique;not null" json:"slug"`
	Description *string `gorm:"type:text" json:"description"`

	// Price is stored in the smallest currency unit (Indonesian Rupiah, no decimal).
	// Example: Rp 500.000.000 is stored as 500000000.
	TransactionType string `gorm:"type:varchar(20);not null;default:'sale';index" json:"transaction_type"`
	Price           int64  `gorm:"not null" json:"price"`
	Currency        string `gorm:"type:varchar(3);default:'IDR'" json:"currency"`
	IsNegotiable    bool   `gorm:"default:false" json:"is_negotiable"`

	SpecialOffers datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"special_offers"`

	LocationProvince     *string  `gorm:"type:varchar(100);index" json:"location_province"`
	LocationProvinceCode *string  `gorm:"type:varchar(13);index" json:"location_province_code"`
	LocationCity         *string  `gorm:"type:varchar(100)" json:"location_city"`
	LocationCityCode     *string  `gorm:"type:varchar(13);index" json:"location_city_code"`
	LocationDistrict     *string  `gorm:"type:varchar(100)" json:"location_district"`
	LocationDistrictCode *string  `gorm:"type:varchar(13);index" json:"location_district_code"`
	LocationVillage      *string  `gorm:"type:varchar(100)" json:"location_village"`
	LocationVillageCode  *string  `gorm:"type:varchar(13);index" json:"location_village_code"`
	AddressDetail        *string  `gorm:"type:text" json:"address_detail"`
	Latitude             *float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude            *float64 `gorm:"type:decimal(11,8)" json:"longitude"`

	BedroomCount    *int `gorm:"type:int;index" json:"bedroom_count"`
	BathroomCount   *int `gorm:"type:int;index" json:"bathroom_count"`
	FloorCount      *int `gorm:"type:int" json:"floor_count"`
	CarportCapacity *int `gorm:"type:int" json:"carport_capacity"`

	LandAreaSqm     *int `gorm:"type:int;index" json:"land_area_sqm"`
	BuildingAreaSqm *int `gorm:"type:int;index" json:"building_area_sqm"`

	CertificateType   *string `gorm:"type:varchar(50);index" json:"certificate_type"`
	Condition         *string `gorm:"type:varchar(50);index" json:"condition"`
	Furnishing        *string `gorm:"type:varchar(50);index" json:"furnishing"`
	ElectricalPowerVA *int    `gorm:"type:int" json:"electrical_power_va"`
	FacingDirection   *string `gorm:"type:varchar(50)" json:"facing_direction"`
	YearBuilt         *int    `gorm:"type:int" json:"year_built"`

	Facilities     datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"facilities"`
	Specifications datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"specifications"`

	Status     string `gorm:"type:varchar(20);default:'active';index" json:"status"`
	IsFeatured bool   `gorm:"default:false" json:"is_featured"`
	ViewCount  int    `gorm:"default:0" json:"view_count"`

	User     *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images   []ListingImage `gorm:"foreignKey:ListingID" json:"images,omitempty"`
	Video    *ListingVideo  `gorm:"foreignKey:ListingID" json:"video,omitempty"`
}

type ListingImage struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ListingID        uuid.UUID      `gorm:"type:uuid;not null" json:"listing_id"`
	URL              string         `gorm:"type:text;not null" json:"url"`
	AssetID          *string        `gorm:"type:varchar(255)" json:"asset_id,omitempty"`
	PublicID         *string        `gorm:"type:varchar(255)" json:"public_id,omitempty"`
	Version          *int64         `gorm:"type:bigint" json:"version,omitempty"`
	ResourceType     *string        `gorm:"type:varchar(50)" json:"resource_type,omitempty"`
	Type             *string        `gorm:"type:varchar(50)" json:"type,omitempty"`
	Format           *string        `gorm:"type:varchar(50)" json:"format,omitempty"`
	Bytes            *int64         `gorm:"type:bigint" json:"bytes,omitempty"`
	Width            *int           `gorm:"type:int" json:"width,omitempty"`
	Height           *int           `gorm:"type:int" json:"height,omitempty"`
	OriginalFilename *string        `gorm:"type:varchar(255)" json:"original_filename,omitempty"`
	IsPrimary        bool           `gorm:"default:false" json:"is_primary"`
	SortOrder        int            `gorm:"default:0" json:"sort_order"`
	CreatedAt        time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (li *ListingImage) BeforeCreate(tx *gorm.DB) error {
	if li.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		li.ID = id
	}

	return nil
}
