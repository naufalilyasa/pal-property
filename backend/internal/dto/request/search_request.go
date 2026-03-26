package request

import "github.com/google/uuid"

type SearchListingsRequest struct {
	Query            string     `json:"q" validate:"omitempty,max=255"`
	TransactionType  string     `json:"transaction_type" validate:"omitempty,oneof=sale rent"`
	CategoryID       *uuid.UUID `json:"category_id"`
	LocationProvince string     `json:"location_province" validate:"omitempty,max=100"`
	LocationCity     string     `json:"location_city" validate:"omitempty,max=100"`
	PriceMin         *int64     `json:"price_min" validate:"omitempty,gte=0"`
	PriceMax         *int64     `json:"price_max" validate:"omitempty,gte=0"`
	Page             int        `json:"page" validate:"omitempty,gte=1"`
	Limit            int        `json:"limit" validate:"omitempty,gte=1,lte=100"`
	Sort             string     `json:"sort" validate:"omitempty,oneof=relevance newest price_asc price_desc"`
}
