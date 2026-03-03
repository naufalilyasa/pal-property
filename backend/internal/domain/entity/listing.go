package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
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

type Listing struct {
	BaseEntity
	UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CategoryID  *uuid.UUID `gorm:"type:uuid" json:"category_id"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Slug        string     `gorm:"type:varchar(255);unique;not null" json:"slug"`
	Description *string    `gorm:"type:text" json:"description"`
	// Price is stored in the smallest currency unit (Indonesian Rupiah, no decimal).
	// Example: Rp 500.000.000 is stored as 500000000.
	Price int64 `gorm:"not null" json:"price"`
	Currency    string     `gorm:"type:varchar(3);default:'IDR'" json:"currency"`

	LocationCity     *string `gorm:"type:varchar(100)" json:"location_city"`
	LocationDistrict *string `gorm:"type:varchar(100)" json:"location_district"`
	// LocationCoordinates is handled as point type, usually needs custom GORM data type or raw SQL.
	// For simplicity in struct, we might use a custom struct or interface.
	// Putting placeholder here, might need PostGIS or similar if advanced.
	// But schema says POINT.

	AddressDetail *string `gorm:"type:text" json:"address_detail"`

	Status     string `gorm:"type:varchar(20);default:'active'" json:"status"`
	IsFeatured bool   `gorm:"default:false" json:"is_featured"`

	Specifications datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"specifications"`

	ViewCount int `gorm:"default:0" json:"view_count"`

	User     *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images   []ListingImage `gorm:"foreignKey:ListingID" json:"images,omitempty"`
}

type ListingImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ListingID uuid.UUID `gorm:"type:uuid;not null" json:"listing_id"`
	URL       string    `gorm:"type:text;not null" json:"url"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
