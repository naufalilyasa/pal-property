package response

import domain "github.com/naufalilyasa/pal-property-backend/internal/domain"

// ChatAnswerFormat defines the rendering cue for chat answers.
// When markdown is signaled, only the approved subset (p, br, h3, h4, strong, em, ul, ol, li, a) may appear.
type ChatAnswerFormat string

const (
	ChatAnswerFormatText     ChatAnswerFormat = "text"
	ChatAnswerFormatMarkdown ChatAnswerFormat = "markdown"
)

type ChatResponse struct {
	SessionID       string                          `json:"session_id"`
	Answer          string                          `json:"answer"`
	AnswerFormat    ChatAnswerFormat                `json:"answer_format"`
	Grounding       ChatGrounding                   `json:"grounding"`
	Recommendations []ChatRetrievalDocumentResponse `json:"recommendations,omitempty"`
}

type ChatGrounding = domain.ChatGrounding

type ChatGroundingMetadata = domain.ChatGroundingMetadata

type ChatRetrievalDocumentResponse = domain.ChatRetrievalDocument

type ChatDocumentCategoryResponse = domain.ChatDocumentCategory
