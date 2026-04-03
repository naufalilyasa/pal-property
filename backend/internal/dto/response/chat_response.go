package response

import domain "github.com/naufalilyasa/pal-property-backend/internal/domain"

type ChatResponse struct {
	SessionID       string                          `json:"session_id"`
	Answer          string                          `json:"answer"`
	Grounding       ChatGrounding                   `json:"grounding"`
	Recommendations []ChatRetrievalDocumentResponse `json:"recommendations,omitempty"`
}

type ChatGrounding = domain.ChatGrounding

type ChatGroundingMetadata = domain.ChatGroundingMetadata

type ChatRetrievalDocumentResponse = domain.ChatRetrievalDocument

type ChatDocumentCategoryResponse = domain.ChatDocumentCategory
