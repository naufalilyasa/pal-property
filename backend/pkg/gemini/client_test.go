package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestEmbedQueryAndDocumentTags(t *testing.T) {
	ctx := context.Background()
	fake := newFakeModelsAPI()
	client, err := NewClient(ctx, "key", WithModelsAPI(fake))
	require.NoError(t, err)

	queryResults, err := client.EmbedQuery(ctx, "alpha", "beta")
	require.NoError(t, err)
	require.Len(t, queryResults, 2)
	require.Equal(t, EmbeddingTaskQuery, queryResults[0].Task)
	require.Equal(t, defaultEmbeddingModel, fake.lastEmbedModel)

	docResults, err := client.EmbedDocument(ctx, "document")
	require.NoError(t, err)
	require.Len(t, docResults, 1)
	require.Equal(t, EmbeddingTaskDocument, docResults[0].Task)
}

func TestGenerateGroundedAnswerRetriesAndHook(t *testing.T) {
	ctx := context.Background()
	fake := newFakeModelsAPI()
	fake.generateResp = &genai.GenerateContentResponse{
		ModelVersion: defaultGenerationModel,
		Candidates: []*genai.Candidate{
			{Content: genai.NewContentFromText("Answer text", genai.RoleModel)},
		},
	}
	fake.failGenerate = 1
	hookCalled := false
	client, err := NewClient(ctx, "key", WithModelsAPI(fake), WithSafetyHook(func(ctx context.Context, req GroundedAnswerRequest) error {
		hookCalled = true
		return nil
	}))
	require.NoError(t, err)

	req := GroundedAnswerRequest{
		Question:        "How much is this?",
		Documents:       []GroundingDocument{{ID: "L1", Title: "Listing", Source: "Listing Source", Excerpt: "Prime land"}},
		MaxOutputTokens: 128,
	}

	resp, err := client.GenerateGroundedAnswer(ctx, req)
	require.NoError(t, err)
	require.True(t, hookCalled)
	require.Equal(t, 2, fake.generateCalls)
	require.Equal(t, req.MaxOutputTokens, fake.lastGenerateConfig.MaxOutputTokens)
	require.Equal(t, defaultGenerationModel, fake.lastGenerateModel)
	require.Contains(t, resp.Answer, "Answer text")

	foundQuestion := false
	for _, content := range fake.lastGenerateContents {
		for _, part := range content.Parts {
			if strings.Contains(part.Text, req.Question) {
				foundQuestion = true
			}
		}
	}
	require.True(t, foundQuestion)
}

func TestGenerateGroundedAnswerRejectsUnsafeHook(t *testing.T) {
	ctx := context.Background()
	fake := newFakeModelsAPI()
	blockErr := errors.New("blocked request")
	client, err := NewClient(ctx, "key", WithModelsAPI(fake), WithSafetyHook(func(ctx context.Context, req GroundedAnswerRequest) error {
		return blockErr
	}))
	require.NoError(t, err)

	_, err = client.GenerateGroundedAnswer(ctx, GroundedAnswerRequest{Question: "hi"})
	require.ErrorIs(t, err, blockErr)
	require.Zero(t, fake.generateCalls)
}

type fakeModelsAPI struct {
	embedCall            int
	lastEmbedModel       string
	lastEmbedContents    []*genai.Content
	generateResp         *genai.GenerateContentResponse
	failGenerate         int
	generateCalls        int
	lastGenerateModel    string
	lastGenerateContents []*genai.Content
	lastGenerateConfig   *genai.GenerateContentConfig
}

func newFakeModelsAPI() *fakeModelsAPI {
	return &fakeModelsAPI{}
}

func (f *fakeModelsAPI) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	f.generateCalls++
	f.lastGenerateModel = model
	f.lastGenerateContents = contents
	f.lastGenerateConfig = config
	if f.generateCalls <= f.failGenerate {
		return nil, errors.New("transient error")
	}
	return f.generateResp, nil
}

func (f *fakeModelsAPI) EmbedContent(ctx context.Context, model string, contents []*genai.Content, config *genai.EmbedContentConfig) (*genai.EmbedContentResponse, error) {
	f.embedCall++
	f.lastEmbedModel = model
	f.lastEmbedContents = contents
	resp := &genai.EmbedContentResponse{}
	for i := range contents {
		resp.Embeddings = append(resp.Embeddings, &genai.ContentEmbedding{Values: []float32{float32(i + 1)}})
	}
	return resp, nil
}
