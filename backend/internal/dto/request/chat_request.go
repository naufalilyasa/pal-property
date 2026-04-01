package request

import (
	"github.com/google/uuid"
)

type ChatRetrievalFilters struct {
	Query            string     `json:"query" validate:"omitempty,max=255"`
	TransactionType  string     `json:"transaction_type" validate:"omitempty,oneof=sale rent"`
	CategoryID       *uuid.UUID `json:"category_id"`
	LocationProvince string     `json:"location_province" validate:"omitempty,max=100"`
	LocationCity     string     `json:"location_city" validate:"omitempty,max=100"`
	PriceMin         *int64     `json:"price_min" validate:"omitempty,gte=0"`
	PriceMax         *int64     `json:"price_max" validate:"omitempty,gte=0"`
}

type ChatRequest struct {
	SessionID    string               `json:"session_id" validate:"required,max=255"`
	Message      string               `json:"message" validate:"required,max=2048"`
	ListingID    *uuid.UUID           `json:"listing_id"`
	Filters      ChatRetrievalFilters `json:"filters"`
	MaxDocuments int                  `json:"max_documents" validate:"omitempty,gte=1,lte=20"`
}
