package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	requestdto "github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	responsedto "github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/gemini"
	validator "github.com/naufalilyasa/pal-property-backend/pkg/validator"
)

type chatAnswerGenerator interface {
	GenerateGroundedAnswer(ctx context.Context, req gemini.GroundedAnswerRequest) (*gemini.GroundedAnswerResponse, error)
}

type ChatService interface {
	Respond(ctx context.Context, req requestdto.ChatRequest) (*responsedto.ChatResponse, error)
}

type chatService struct {
	retrieval    ChatRetrievalService
	memory       domain.ChatMemoryRepository
	generator    chatAnswerGenerator
	historyTurns int
}

func NewChatService(retrieval ChatRetrievalService, memory domain.ChatMemoryRepository, generator chatAnswerGenerator, historyTurns int) (ChatService, error) {
	if retrieval == nil {
		return nil, fmt.Errorf("chat service: retrieval service is required")
	}
	if memory == nil {
		return nil, fmt.Errorf("chat service: memory repository is required")
	}
	if generator == nil {
		return nil, fmt.Errorf("chat service: answer generator is required")
	}
	if historyTurns <= 0 {
		historyTurns = config.Env.ChatMaxHistoryTurns
		if historyTurns <= 0 {
			historyTurns = 10
		}
	}
	return &chatService{retrieval: retrieval, memory: memory, generator: generator, historyTurns: historyTurns}, nil
}

func (s *chatService) Respond(ctx context.Context, req requestdto.ChatRequest) (*responsedto.ChatResponse, error) {
	if err := validator.Validate.Struct(req); err != nil {
		return nil, err
	}

	turns, memoryIssue := s.loadRecentTurns(ctx, req.SessionID)
	retrievalResult, err := s.retrieval.Retrieve(ctx, req)
	if err != nil {
		response := degradedChatResponse(req.SessionID, "Maaf, saya belum bisa mencari properti sekarang. Silakan coba lagi sebentar lagi.", "retrieval_unavailable")
		appendConversation(ctx, s.memory, req.SessionID, req.Message, response.Answer)
		return response, nil
	}

	if len(retrievalResult.Grounding.ListingIDs) == 0 {
		response := degradedChatResponse(req.SessionID, "Maaf, saya belum menemukan properti aktif yang cocok dengan permintaan Anda saat ini.", retrievalResult.Grounding.DegradedReason)
		appendConversation(ctx, s.memory, req.SessionID, req.Message, response.Answer)
		return response, nil
	}

	groundedDocs := make([]gemini.GroundingDocument, 0, len(retrievalResult.Documents))
	for _, document := range retrievalResult.Documents {
		if !isPublicListingStatus(document.Status) {
			continue
		}
		groundedDocs = append(groundedDocs, gemini.GroundingDocument{
			ID:               document.ListingID.String(),
			Title:            document.Title,
			Source:           document.Slug,
			Excerpt:          document.DescriptionExcerpt,
			Category:         buildChatGroundingCategory(document.Category),
			TransactionType:  document.TransactionType,
			Price:            document.Price,
			Currency:         document.Currency,
			LocationProvince: document.LocationProvince,
			LocationCity:     document.LocationCity,
			LocationDistrict: document.LocationDistrict,
			LocationVillage:  document.LocationVillage,
			BedroomCount:     document.BedroomCount,
			BathroomCount:    document.BathroomCount,
			LandAreaSqm:      document.LandAreaSqm,
			BuildingAreaSqm:  document.BuildingAreaSqm,
		})
	}

	modelResponse, err := s.generator.GenerateGroundedAnswer(ctx, gemini.GroundedAnswerRequest{
		Question:       buildGroundedQuestion(turns, req.Message),
		Documents:      groundedDocs,
		CandidateCount: 1,
	})
	if err != nil {
		response := degradedChatResponse(req.SessionID, "Maaf, saya belum bisa merangkai jawaban saat ini. Silakan coba lagi sebentar lagi.", "generation_unavailable")
		response.Grounding.ListingIDs = retrievalResult.Grounding.ListingIDs
		response.Grounding.ListingSlugs = retrievalResult.Grounding.ListingSlugs
		response.Grounding.Citations = retrievalResult.Grounding.Citations
		appendConversation(ctx, s.memory, req.SessionID, req.Message, response.Answer)
		return response, nil
	}

	recommendations := toChatRecommendations(retrievalResult.Documents)
	allowedSlugs := buildAllowedListingSlugs(recommendations)
	sanitizedAnswer, hasValidMarkdownLinks := sanitizeChatAnswer(modelResponse.Answer, allowedSlugs)
	answerFormat := responsedto.ChatAnswerFormatText
	if hasValidMarkdownLinks {
		answerFormat = responsedto.ChatAnswerFormatMarkdown
	}
	response := &responsedto.ChatResponse{
		SessionID:       req.SessionID,
		Answer:          sanitizedAnswer,
		AnswerFormat:    answerFormat,
		Grounding:       retrievalResult.Grounding,
		Recommendations: recommendations,
	}
	if memoryIssue != "" {
		response.Grounding.IsDegraded = true
		if response.Grounding.DegradedReason == "" {
			response.Grounding.DegradedReason = memoryIssue
		}
	}
	appendConversation(ctx, s.memory, req.SessionID, req.Message, response.Answer)
	return response, nil
}

func toChatRecommendations(documents []domain.ChatRetrievalDocument) []responsedto.ChatRetrievalDocumentResponse {
	recommendations := make([]responsedto.ChatRetrievalDocumentResponse, 0, len(documents))
	for _, document := range documents {
		if !isPublicListingStatus(document.Status) {
			continue
		}
		recommendations = append(recommendations, responsedto.ChatRetrievalDocumentResponse(document))
	}
	return recommendations
}

func (s *chatService) loadRecentTurns(ctx context.Context, sessionID string) ([]domain.ChatTurn, string) {
	turns, err := s.memory.GetTurns(ctx, sessionID)
	if err == nil {
		if len(turns) > s.historyTurns {
			return turns[len(turns)-s.historyTurns:], ""
		}
		return turns, ""
	}
	if errors.Is(err, domain.ErrChatMemoryNotFound) {
		return nil, ""
	}
	return nil, "memory_unavailable"
}

func degradedChatResponse(sessionID, answer, reason string) *responsedto.ChatResponse {
	return &responsedto.ChatResponse{
		SessionID:    sessionID,
		Answer:       answer,
		AnswerFormat: responsedto.ChatAnswerFormatText,
		Grounding: responsedto.ChatGrounding{
			IsDegraded:     true,
			DegradedReason: reason,
		},
	}
}

func buildGroundedQuestion(turns []domain.ChatTurn, message string) string {
	if len(turns) == 0 {
		return message
	}
	parts := make([]string, 0, len(turns)+2)
	parts = append(parts, "Riwayat percakapan singkat:")
	for _, turn := range turns {
		parts = append(parts, fmt.Sprintf("- %s: %s", turn.Role, turn.Message))
	}
	parts = append(parts, fmt.Sprintf("Pertanyaan terbaru: %s", message))
	return strings.Join(parts, "\n")
}

func appendConversation(ctx context.Context, memory domain.ChatMemoryRepository, sessionID, question, answer string) {
	_ = memory.AppendTurn(ctx, sessionID, domain.ChatTurn{Role: "user", Message: question})
	_ = memory.AppendTurn(ctx, sessionID, domain.ChatTurn{Role: "assistant", Message: answer})
}

func buildChatGroundingCategory(category *domain.ChatDocumentCategory) string {
	if category == nil {
		return ""
	}
	name := strings.TrimSpace(category.Name)
	slug := strings.TrimSpace(category.Slug)
	switch {
	case name != "" && slug != "":
		return fmt.Sprintf("%s (%s)", name, slug)
	case name != "":
		return name
	default:
		return slug
	}
}

var markdownLinkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
var markdownHeadingRegex = regexp.MustCompile(`(?m)^\s*#{3,4}\s+\S`)
var markdownListRegex = regexp.MustCompile(`(?m)^\s*(?:[-+*]|\d+\.)\s+\S`)
var markdownEmphasisRegex = regexp.MustCompile(`(?m)(?:\*\*[^\n*]+\*\*|\*[^\n*]+\*|__[^\n_]+__|_[^\n_]+_)`)

func buildAllowedListingSlugs(recommendations []responsedto.ChatRetrievalDocumentResponse) map[string]struct{} {
	if len(recommendations) == 0 {
		return nil
	}
	slugs := make(map[string]struct{}, len(recommendations))
	for _, rec := range recommendations {
		if slug := strings.TrimSpace(rec.Slug); slug != "" {
			slugs[slug] = struct{}{}
		}
	}
	return slugs
}

func sanitizeChatAnswer(answer string, allowedSlugs map[string]struct{}) (string, bool) {
	if answer == "" {
		return answer, false
	}
	answer = sanitizeUnsupportedMarkdownHeadings(answer)
	matches := markdownLinkRegex.FindAllStringSubmatchIndex(answer, -1)
	if len(matches) == 0 {
		return answer, containsMarkdownStructure(answer)
	}
	var builder strings.Builder
	lastIndex := 0
	validLinkFound := false
	for _, match := range matches {
		builder.WriteString(answer[lastIndex:match[0]])
		text := answer[match[2]:match[3]]
		url := answer[match[4]:match[5]]
		if isAllowedListingLink(url, allowedSlugs) {
			builder.WriteString(answer[match[0]:match[1]])
			validLinkFound = true
		} else {
			builder.WriteString(text)
		}
		lastIndex = match[1]
	}
	builder.WriteString(answer[lastIndex:])
	sanitized := builder.String()
	return sanitized, validLinkFound || containsMarkdownStructure(sanitized)
}

func containsMarkdownStructure(answer string) bool {
	if strings.TrimSpace(answer) == "" {
		return false
	}
	if markdownHeadingRegex.MatchString(answer) {
		return true
	}
	if markdownListRegex.MatchString(answer) {
		return true
	}
	if markdownEmphasisRegex.MatchString(answer) {
		return true
	}
	return false
}

func sanitizeUnsupportedMarkdownHeadings(answer string) string {
	if answer == "" {
		return answer
	}
	parts := strings.Split(answer, "\n")
	var builder strings.Builder
	for i, part := range parts {
		if i > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(sanitizeUnsupportedHeadingLine(part))
	}
	return builder.String()
}

func sanitizeUnsupportedHeadingLine(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" || trimmed[0] != '#' {
		return line
	}
	leading := line[:len(line)-len(trimmed)]
	count := 0
	for count < len(trimmed) && trimmed[count] == '#' {
		count++
	}
	if count != 1 && count != 2 && count != 5 && count != 6 {
		return line
	}
	rest := strings.TrimLeft(trimmed[count:], " \t")
	return leading + rest
}

func isAllowedListingLink(url string, allowedSlugs map[string]struct{}) bool {
	if !strings.HasPrefix(url, "/listings/") {
		return false
	}
	slug := strings.TrimPrefix(url, "/listings/")
	slug = strings.TrimSuffix(slug, "/")
	if slug == "" || strings.Contains(slug, "/") || strings.ContainsAny(slug, " ?#") {
		return false
	}
	if allowedSlugs == nil {
		return false
	}
	_, ok := allowedSlugs[slug]
	return ok
}
