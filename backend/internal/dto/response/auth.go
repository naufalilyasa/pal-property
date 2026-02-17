package response

import (
	"time"

	"github.com/google/uuid"
)

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token,omitempty"` // JWT or session token
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
