package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/mocks"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/gemini"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

type fakeChatDocumentEmbedder struct {
	inputs []string
}

func (f *fakeChatDocumentEmbedder) EmbedDocument(_ context.Context, inputs ...string) ([]gemini.EmbeddingResult, error) {
	f.inputs = append(f.inputs, inputs...)
	return []gemini.EmbeddingResult{{Values: make([]float64, 768)}}, nil
}

func TestRebuildChatRetrievalIndex_UsesRicherPublicLexicalAndEmbeddingText(t *testing.T) {
	var mappingBody map[string]any
	var documentBody map[string]any
	var putPaths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			defer r.Body.Close()
			putPaths = append(putPaths, r.URL.Path)
			decoded := map[string]any{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&decoded))
			if strings.Contains(r.URL.Path, "/_doc/") {
				documentBody = decoded
			} else {
				mappingBody = decoded
			}
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client, err := searchindex.NewClient(server.URL, "", "", server.Client())
	require.NoError(t, err)

	repo := mocks.NewListingRepository(t)
	categoryID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	bedrooms := 4
	bathrooms := 3
	landArea := 180
	buildingArea := 220
	description := "Rumah keluarga dekat sekolah, tol, dan pusat belanja."
	province := "DKI Jakarta"
	city := "Jakarta Selatan"
	district := "Kebayoran Baru"
	village := "Gandaria Utara"

	repo.On("List", mock.Anything, domain.ListingFilter{Page: 1, Limit: 2}).Return([]*entity.Listing{
		{
			BaseEntity:       entity.BaseEntity{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111")},
			Title:            "Rumah Gandaria Mewah",
			Slug:             "rumah-gandaria-mewah",
			Description:      &description,
			TransactionType:  "sale",
			Price:            3250000000,
			Currency:         "IDR",
			LocationProvince: &province,
			LocationCity:     &city,
			LocationDistrict: &district,
			LocationVillage:  &village,
			BedroomCount:     &bedrooms,
			BathroomCount:    &bathrooms,
			LandAreaSqm:      &landArea,
			BuildingAreaSqm:  &buildingArea,
			Status:           "active",
			CategoryID:       &categoryID,
			Category:         &entity.Category{ID: categoryID, Name: "Rumah", Slug: "rumah"},
			Facilities:       datatypes.JSON(`["AC","CCTV"]`),
			Specifications:   datatypes.JSON(`{"bedrooms":4,"bathrooms":3,"land_area_sqm":180,"building_area_sqm":220}`),
		},
		{
			BaseEntity: entity.BaseEntity{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222")},
			Title:      "Rumah Draft",
			Status:     "draft",
		},
	}, int64(2), nil)

	embedder := &fakeChatDocumentEmbedder{}
	err = service.RebuildChatRetrievalIndex(context.Background(), repo, client, embedder, "chat-retrieval", 2)
	require.NoError(t, err)

	require.Len(t, putPaths, 2)
	require.NotNil(t, mappingBody)
	require.NotNil(t, documentBody)

	properties := mappingBody["mappings"].(map[string]any)["properties"].(map[string]any)
	assert.Equal(t, "text", properties["lexical_text"].(map[string]any)["type"])
	assert.Equal(t, float64(768), properties["embedding"].(map[string]any)["dims"])

	lexicalText, ok := documentBody["lexical_text"].(string)
	require.True(t, ok)
	assert.Contains(t, lexicalText, "Rumah Gandaria Mewah")
	assert.Contains(t, lexicalText, "sale")
	assert.Contains(t, lexicalText, "Rumah rumah")
	assert.Contains(t, lexicalText, "Gandaria Utara, Kebayoran Baru, Jakarta Selatan, DKI Jakarta")
	assert.Contains(t, lexicalText, "IDR 3250000000")
	assert.Contains(t, lexicalText, "4")
	assert.Contains(t, lexicalText, "3")
	assert.Contains(t, lexicalText, "180")
	assert.Contains(t, lexicalText, "220")
	assert.Contains(t, lexicalText, description)
	lexicalSearchText, ok := documentBody["lexical_search_text"].(string)
	require.True(t, ok)
	assert.Contains(t, lexicalSearchText, "Facilities: AC, CCTV")
	assert.Contains(t, lexicalSearchText, "Specifications: Bedrooms: 4")
	assert.Contains(t, lexicalText, "Facilities: AC, CCTV")
	assert.Contains(t, lexicalText, "Specifications: Bedrooms: 4")

	require.Len(t, embedder.inputs, 1)
	embeddingInput := embedder.inputs[0]
	assert.Contains(t, embeddingInput, "Title: Rumah Gandaria Mewah")
	assert.Contains(t, embeddingInput, "Transaction type: sale")
	assert.Contains(t, embeddingInput, "Category: Rumah rumah")
	assert.Contains(t, embeddingInput, "Location: Gandaria Utara, Kebayoran Baru, Jakarta Selatan, DKI Jakarta")
	assert.Contains(t, embeddingInput, "Price: IDR 3250000000")
	assert.Contains(t, embeddingInput, "Bedrooms: 4")
	assert.Contains(t, embeddingInput, "Bathrooms: 3")
	assert.Contains(t, embeddingInput, "Land area sqm: 180")
	assert.Contains(t, embeddingInput, "Building area sqm: 220")
	assert.Contains(t, embeddingInput, "Description excerpt: "+description)
	assert.Contains(t, embeddingInput, "Facilities: AC, CCTV")
	assert.Contains(t, embeddingInput, "Specifications: Bedrooms: 4")
	assert.NotContains(t, embeddingInput, "Rumah Draft")
}
