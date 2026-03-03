package request

import "github.com/google/uuid"

type CreateCategoryRequest struct {
	Name     string     `json:"name"     validate:"required,min=2,max=100"`
	ParentID *uuid.UUID `json:"parent_id"`
	IconURL  *string    `json:"icon_url"`
}

type UpdateCategoryRequest struct {
	Name    *string `json:"name"     validate:"omitempty,min=2,max=100"`
	IconURL *string `json:"icon_url"`
	// Slug intentionally excluded — locked after creation
}
