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

func TestBuildGenerateConfigUsesMinaSingleResultRules(t *testing.T) {
	config := buildGenerateConfig(GroundedAnswerRequest{
		Documents: []GroundingDocument{{Title: "Rumah Keluarga", Source: "rumah-keluarga"}},
	})

	require.NotNil(t, config.SystemInstruction)
	instruction := flattenContent(config.SystemInstruction)
	lowerInstruction := strings.ToLower(instruction)
	require.Contains(t, instruction, "Mina")
	require.Contains(t, instruction, "single-result detail")
	require.Contains(t, lowerInstruction, "jangan mengarang harga")
	require.Contains(t, instruction, "raw ID")
	require.Contains(t, instruction, "[Nama Properti](/listings/<slug>)")
	require.Contains(t, instruction, "p, br, h3, h4, strong, em, ul, ol, li, a")
	require.Contains(t, instruction, "survey/kunjungan")
	require.NotContains(t, instruction, "multi-result comparison")
}

func TestBuildGenerateConfigUsesMinaMultiResultRules(t *testing.T) {
	config := buildGenerateConfig(GroundedAnswerRequest{
		Documents: []GroundingDocument{{Title: "Rumah A", Source: "rumah-a"}, {Title: "Rumah B", Source: "rumah-b"}},
	})

	require.NotNil(t, config.SystemInstruction)
	instruction := flattenContent(config.SystemInstruction)
	lowerInstruction := strings.ToLower(instruction)
	require.Contains(t, instruction, "Mina")
	require.Contains(t, instruction, "multi-result comparison")
	require.Contains(t, lowerInstruction, "perbandingan singkat")
	require.Contains(t, lowerInstruction, "pertanyaan follow-up singkat")
	require.Contains(t, lowerInstruction, "tautan relatif")
	require.NotContains(t, instruction, "single-result detail")
}

func TestBuildDocumentSummaryFormatsPublicSafeGroundingContext(t *testing.T) {
	province := "Jawa Barat"
	city := "Bogor"
	district := "Bogor Barat"
	village := "Menteng"
	bedrooms := 3
	bathrooms := 2
	landArea := 120
	buildingArea := 90

	summary := buildDocumentSummary([]GroundingDocument{{
		ID:               "11111111-1111-1111-1111-111111111111",
		Title:            "Rumah Keluarga Bogor",
		Source:           "rumah-keluarga-bogor",
		Excerpt:          "Dekat sekolah dan akses tol.",
		Category:         "Rumah",
		TransactionType:  "sale",
		Price:            1500000000,
		Currency:         "IDR",
		LocationProvince: &province,
		LocationCity:     &city,
		LocationDistrict: &district,
		LocationVillage:  &village,
		BedroomCount:     &bedrooms,
		BathroomCount:    &bathrooms,
		LandAreaSqm:      &landArea,
		BuildingAreaSqm:  &buildingArea,
	}, {
		ID:      "22222222-2222-2222-2222-222222222222",
		Title:   "Cluster Sentul",
		Source:  "cluster-sentul",
		Excerpt: "Cocok untuk keluarga muda.",
	}})

	require.Contains(t, summary, "Jumlah properti kandidat: 2")
	require.Contains(t, summary, "Mode jawaban Mina: comparison-style singkat untuk beberapa properti")
	require.Contains(t, summary, "Link listing: /listings/rumah-keluarga-bogor")
	require.Contains(t, summary, "Link listing: /listings/cluster-sentul")
	require.Contains(t, summary, "Kategori: Rumah")
	require.Contains(t, summary, "Harga: Rp 1.500.000.000")
	require.Contains(t, summary, "Lokasi: Menteng, Bogor Barat, Bogor, Jawa Barat")
	require.Contains(t, summary, "Kamar tidur: 3")
	require.Contains(t, summary, "Kamar mandi: 2")
	require.Contains(t, summary, "Luas tanah: 120 m²")
	require.Contains(t, summary, "Luas bangunan: 90 m²")
	require.Contains(t, summary, "Deskripsi singkat: Dekat sekolah dan akses tol.")
	require.NotContains(t, summary, "11111111-1111-1111-1111-111111111111")
	require.NotContains(t, summary, "Slug:")
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
