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

type fakeChatRetrievalRepository struct {
	documents []domain.ChatRetrievalDocument
	err       error
	filters   domain.ChatRetrievalFilters
	vector    []float64
	limit     int
}

func (f *fakeChatRetrievalRepository) FetchDocuments(_ context.Context, filters domain.ChatRetrievalFilters, queryVector []float64, limit int) ([]domain.ChatRetrievalDocument, error) {
	f.filters = filters
	f.vector = queryVector
	f.limit = limit
	return f.documents, f.err
}

func (f *fakeChatRetrievalRepository) FetchDocumentByID(_ context.Context, _ uuid.UUID) (*domain.ChatRetrievalDocument, error) {
	return nil, domain.ErrNotFound
}

type fakeChatEmbedder struct {
	results []gemini.EmbeddingResult
	err     error
}

func (f fakeChatEmbedder) EmbedQuery(_ context.Context, _ ...string) ([]gemini.EmbeddingResult, error) {
	return f.results, f.err
}

func TestChatRetrievalServiceRetrieveBuildsGrounding(t *testing.T) {
	repo := &fakeChatRetrievalRepository{documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Title: "Rumah Aktif", Slug: "rumah-aktif", Status: "active"}}}
	svc, err := NewChatRetrievalService(repo, fakeChatEmbedder{results: []gemini.EmbeddingResult{{Values: []float64{0.1, 0.2}}}}, 5)
	require.NoError(t, err)

	result, err := svc.Retrieve(context.Background(), requestdto.ChatRequest{SessionID: "ses-1", Message: "Ada rumah aktif?"})
	require.NoError(t, err)
	require.Len(t, result.Documents, 1)
	require.Equal(t, []string{"rumah-aktif"}, result.Grounding.ListingSlugs)
	require.False(t, result.Grounding.IsDegraded)
	require.Len(t, repo.vector, 2)
}

func TestChatRetrievalServiceRetrieveReturnsDegradedWhenNoActiveResults(t *testing.T) {
	repo := &fakeChatRetrievalRepository{documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Title: "Rumah Draft", Slug: "rumah-draft", Status: "draft"}}}
	svc, err := NewChatRetrievalService(repo, fakeChatEmbedder{results: []gemini.EmbeddingResult{{Values: []float64{0.3}}}}, 5)
	require.NoError(t, err)

	result, err := svc.Retrieve(context.Background(), requestdto.ChatRequest{SessionID: "ses-2", Message: "Ada properti?"})
	require.NoError(t, err)
	require.True(t, result.Grounding.IsDegraded)
	require.Equal(t, "no_active_grounded_results", result.Grounding.DegradedReason)
	require.Empty(t, result.Grounding.ListingIDs)
}

func TestChatRetrievalServiceRetrievePropagatesRepositoryErrors(t *testing.T) {
	repo := &fakeChatRetrievalRepository{err: errors.New("es failure")}
	svc, err := NewChatRetrievalService(repo, fakeChatEmbedder{results: []gemini.EmbeddingResult{{Values: []float64{0.4}}}}, 5)
	require.NoError(t, err)

	_, err = svc.Retrieve(context.Background(), requestdto.ChatRequest{SessionID: "ses-3", Message: "Cari rumah"})
	require.Error(t, err)
	require.ErrorContains(t, err, "fetch chat retrieval documents")
}
