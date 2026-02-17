package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseEntity contains common columns for all tables.
type BaseEntity struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUIDv7 if not present.
func (base *BaseEntity) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
