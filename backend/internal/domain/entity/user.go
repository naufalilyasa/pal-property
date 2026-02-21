package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	BaseEntity
	Name         string  `gorm:"type:varchar(255);not null" json:"name"`
	Email        string  `gorm:"type:varchar(255);unique;not null" json:"email"`
	Phone        *string `gorm:"type:varchar(20);unique" json:"phone"`
	PasswordHash *string `gorm:"type:varchar(255)" json:"-"`
	AvatarURL    *string `gorm:"type:text" json:"avatar_url"`
	Role         string  `gorm:"type:varchar(20);default:'user'" json:"role"`
	IsVerified   bool    `gorm:"default:false" json:"is_verified"`

	OAuthAccounts []OAuthAccount `gorm:"foreignKey:UserID" json:"oauth_accounts,omitempty"`
	Listings      []Listing      `gorm:"foreignKey:UserID" json:"listings,omitempty"`
}

type OAuthAccount struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Provider       string    `gorm:"type:varchar(50);not null" json:"provider"`
	ProviderUserID string    `gorm:"type:varchar(255);not null" json:"provider_user_id"`
	AccessToken    *string   `gorm:"type:text" json:"-"`
	RefreshToken   *string   `gorm:"type:text" json:"-"`
	CreatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}
