package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SavedListing struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ListingID uuid.UUID `gorm:"type:uuid;not null;index" json:"listing_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Listing *Listing `gorm:"foreignKey:ListingID" json:"listing,omitempty"`
}

func (sl *SavedListing) BeforeCreate(tx *gorm.DB) error {
	if sl.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		sl.ID = id
	}
	return nil
}
