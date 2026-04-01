package entity

import "time"

type AdministrativeRegion struct {
	Code       string    `gorm:"primaryKey;type:varchar(13)"`
	Name       string    `gorm:"type:varchar(100);not null"`
	Level      int       `gorm:"type:smallint;not null;index"`
	ParentCode *string   `gorm:"type:varchar(13);index"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`
	UpdatedAt  time.Time `gorm:"not null;default:now()"`
}

func (AdministrativeRegion) TableName() string {
	return "indonesia_regions"
}
