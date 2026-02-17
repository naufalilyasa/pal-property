package entity

import (
	"time"

	"github.com/google/uuid"
)

type Wishlist struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ListingID uuid.UUID `gorm:"type:uuid;not null" json:"listing_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Listing *Listing `gorm:"foreignKey:ListingID" json:"listing,omitempty"`
}

type ChatRoom struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ListingID *uuid.UUID `gorm:"type:uuid" json:"listing_id"`
	BuyerID   uuid.UUID  `gorm:"type:uuid;not null" json:"buyer_id"`
	SellerID  uuid.UUID  `gorm:"type:uuid;not null" json:"seller_id"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Listing  *Listing      `gorm:"foreignKey:ListingID" json:"listing,omitempty"`
	Buyer    *User         `gorm:"foreignKey:BuyerID" json:"buyer,omitempty"`
	Seller   *User         `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	Messages []ChatMessage `gorm:"foreignKey:RoomID" json:"messages,omitempty"`
}

type ChatMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoomID    uuid.UUID `gorm:"type:uuid;not null" json:"room_id"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}
