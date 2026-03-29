package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListingVideo struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ListingID        uuid.UUID `gorm:"type:uuid;not null;unique" json:"listing_id"`
	URL              string    `gorm:"type:text;not null" json:"url"`
	AssetID          *string   `gorm:"type:varchar(255)" json:"asset_id,omitempty"`
	PublicID         *string   `gorm:"type:varchar(255)" json:"public_id,omitempty"`
	Version          *int64    `gorm:"type:bigint" json:"version,omitempty"`
	ResourceType     *string   `gorm:"type:varchar(50)" json:"resource_type,omitempty"`
	DeliveryType     *string   `gorm:"type:varchar(50)" json:"delivery_type,omitempty"`
	Format           *string   `gorm:"type:varchar(50)" json:"format,omitempty"`
	Bytes            *int64    `gorm:"type:bigint" json:"bytes,omitempty"`
	Width            *int      `gorm:"type:int" json:"width,omitempty"`
	Height           *int      `gorm:"type:int" json:"height,omitempty"`
	DurationSeconds  *int      `gorm:"type:int" json:"duration_seconds,omitempty"`
	OriginalFilename *string   `gorm:"type:varchar(255)" json:"original_filename,omitempty"`
	CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	Listing *Listing `gorm:"foreignKey:ListingID" json:"listing,omitempty"`
}

func (lv *ListingVideo) BeforeCreate(tx *gorm.DB) error {
	if lv.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		lv.ID = id
	}
	return nil
}
