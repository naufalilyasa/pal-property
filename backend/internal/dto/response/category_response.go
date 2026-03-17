package response

import (
	"github.com/google/uuid"
	"time"
)

// CategoryShortResponse is embedded inside ListingResponse — lean, no children.
type CategoryShortResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	IconURL *string   `json:"icon_url"`
}

// CategoryResponse is returned for GET /api/categories and GET /api/categories/:slug.
type CategoryResponse struct {
	ID        uuid.UUID               `json:"id"`
	Name      string                  `json:"name"`
	Slug      string                  `json:"slug"`
	ParentID  *uuid.UUID              `json:"parent_id"`
	IconURL   *string                 `json:"icon_url"`
	CreatedAt time.Time               `json:"created_at"`
	Parent    *CategoryShortResponse  `json:"parent,omitempty"`
	Children  []CategoryShortResponse `json:"children,omitempty"`
}
