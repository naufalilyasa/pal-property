package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	requestdto "github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/pkg/gemini"
	"github.com/stretchr/testify/require"
)

type fakeChatServiceRetrieval struct {
	result *domain.ChatRetrievalResult
	err    error
}

func (f fakeChatServiceRetrieval) Retrieve(_ context.Context, _ requestdto.ChatRequest) (*domain.ChatRetrievalResult, error) {
	return f.result, f.err
}

type fakeChatMemory struct {
	turns       []domain.ChatTurn
	getErr      error
	appendCalls []domain.ChatTurn
}

func (f *fakeChatMemory) AppendTurn(_ context.Context, _ string, turn domain.ChatTurn) error {
	f.appendCalls = append(f.appendCalls, turn)
	return nil
}

func (f *fakeChatMemory) GetTurns(_ context.Context, _ string) ([]domain.ChatTurn, error) {
	return f.turns, f.getErr
}

func (f *fakeChatMemory) ClearSession(_ context.Context, _ string) error { return nil }

type fakeAnswerGenerator struct {
	response *gemini.GroundedAnswerResponse
	err      error
	request  gemini.GroundedAnswerRequest
}

func (f *fakeAnswerGenerator) GenerateGroundedAnswer(_ context.Context, req gemini.GroundedAnswerRequest) (*gemini.GroundedAnswerResponse, error) {
	f.request = req
	return f.response, f.err
}

func TestChatServiceRespondReturnsFallbackOnNoGrounding(t *testing.T) {
	svc, err := NewChatService(
		fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Grounding: domain.ChatGrounding{IsDegraded: true, DegradedReason: "no_active_grounded_results"}}},
		&fakeChatMemory{},
		&fakeAnswerGenerator{},
		5,
	)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-1", Message: "Ada rumah?"})
	require.NoError(t, err)
	require.True(t, response.Grounding.IsDegraded)
	require.Contains(t, response.Answer, "belum menemukan properti aktif")
}

func TestChatServiceRespondUsesMemoryAndGenerator(t *testing.T) {
	memory := &fakeChatMemory{turns: []domain.ChatTurn{{Role: "user", Message: "Cari rumah"}, {Role: "assistant", Message: "Di kota mana?"}}}
	gen := &fakeAnswerGenerator{response: &gemini.GroundedAnswerResponse{Answer: "Saya rekomendasikan Rumah Aktif."}}
	retrieval := fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Title: "Rumah Aktif", Slug: "rumah-aktif", Status: "active", DescriptionExcerpt: "Dekat sekolah"}}, Grounding: domain.ChatGrounding{ListingIDs: []uuid.UUID{uuid.MustParse("11111111-1111-1111-1111-111111111111")}, ListingSlugs: []string{"rumah-aktif"}}}}
	svc, err := NewChatService(retrieval, memory, gen, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-2", Message: "Ada rumah aktif?"})
	require.NoError(t, err)
	require.Equal(t, "Saya rekomendasikan Rumah Aktif.", response.Answer)
	require.Contains(t, gen.request.Question, "Riwayat percakapan singkat")
	require.Len(t, memory.appendCalls, 2)
}

func TestChatServiceRespondDegradesOnRetrievalError(t *testing.T) {
	svc, err := NewChatService(fakeChatServiceRetrieval{err: errors.New("es timeout")}, &fakeChatMemory{}, &fakeAnswerGenerator{}, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-3", Message: "Cari apartemen"})
	require.NoError(t, err)
	require.True(t, response.Grounding.IsDegraded)
	require.Equal(t, "retrieval_unavailable", response.Grounding.DegradedReason)
}

func TestChatServiceRespondDegradesOnGenerationError(t *testing.T) {
	retrieval := fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Title: "Rumah Aktif", Slug: "rumah-aktif", Status: "active"}}, Grounding: domain.ChatGrounding{ListingIDs: []uuid.UUID{uuid.MustParse("11111111-1111-1111-1111-111111111111")}, ListingSlugs: []string{"rumah-aktif"}}}}
	gen := &fakeAnswerGenerator{err: errors.New("gemini timeout")}
	svc, err := NewChatService(retrieval, &fakeChatMemory{}, gen, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-4", Message: "Cari rumah aktif"})
	require.NoError(t, err)
	require.True(t, response.Grounding.IsDegraded)
	require.Equal(t, "generation_unavailable", response.Grounding.DegradedReason)
	require.Equal(t, []string{"rumah-aktif"}, response.Grounding.ListingSlugs)
}
