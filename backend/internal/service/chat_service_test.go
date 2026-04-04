package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	requestdto "github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	responsedto "github.com/naufalilyasa/pal-property-backend/internal/dto/response"
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
	require.Equal(t, responsedto.ChatAnswerFormatText, response.AnswerFormat)
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
	require.Equal(t, responsedto.ChatAnswerFormatText, response.AnswerFormat)
	require.Contains(t, gen.request.Question, "Riwayat percakapan singkat")
	require.Len(t, memory.appendCalls, 2)
	require.Len(t, response.Recommendations, 1)
	require.Equal(t, "rumah-aktif", response.Recommendations[0].Slug)
}

func TestChatServiceRespondPassesGroundingMetadata(t *testing.T) {
	listingID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	province := "DKI Jakarta"
	city := "Jakarta Selatan"
	district := "Kebayoran Baru"
	village := "Gandaria"
	category := &domain.ChatDocumentCategory{Name: "Apartemen", Slug: "apartemen"}
	retrieval := fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Documents: []domain.ChatRetrievalDocument{{
		ListingID:          listingID,
		Title:              "Rumah Utama",
		Slug:               "rumah-utama",
		DescriptionExcerpt: "Unit rame, ready to move in",
		Category:           category,
		TransactionType:    "sale",
		Price:              1750000000,
		Currency:           "IDR",
		LocationProvince:   stringPtr(province),
		LocationCity:       stringPtr(city),
		LocationDistrict:   stringPtr(district),
		LocationVillage:    stringPtr(village),
		BedroomCount:       intPtr(3),
		BathroomCount:      intPtr(2),
		LandAreaSqm:        intPtr(120),
		BuildingAreaSqm:    intPtr(90),
		Status:             "active",
	}}, Grounding: domain.ChatGrounding{ListingIDs: []uuid.UUID{listingID}, ListingSlugs: []string{"rumah-utama"}}}}
	gen := &fakeAnswerGenerator{response: &gemini.GroundedAnswerResponse{Answer: "Detail siap"}}
	svc, err := NewChatService(retrieval, &fakeChatMemory{}, gen, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-meta", Message: "Ayo"})
	require.NoError(t, err)
	require.Len(t, gen.request.Documents, 1)
	doc := gen.request.Documents[0]
	require.Equal(t, listingID.String(), doc.ID)
	require.Equal(t, "Rumah Utama", doc.Title)
	require.Equal(t, "rumah-utama", doc.Source)
	require.Equal(t, "Unit rame, ready to move in", doc.Excerpt)
	require.Equal(t, "Apartemen (apartemen)", doc.Category)
	require.Equal(t, "sale", doc.TransactionType)
	require.Equal(t, int64(1750000000), doc.Price)
	require.Equal(t, "IDR", doc.Currency)
	require.NotNil(t, doc.LocationProvince)
	require.Equal(t, province, *doc.LocationProvince)
	require.NotNil(t, doc.LocationCity)
	require.Equal(t, city, *doc.LocationCity)
	require.NotNil(t, doc.LocationDistrict)
	require.Equal(t, district, *doc.LocationDistrict)
	require.NotNil(t, doc.LocationVillage)
	require.Equal(t, village, *doc.LocationVillage)
	require.NotNil(t, doc.BedroomCount)
	require.Equal(t, 3, *doc.BedroomCount)
	require.NotNil(t, doc.BathroomCount)
	require.Equal(t, 2, *doc.BathroomCount)
	require.NotNil(t, doc.LandAreaSqm)
	require.Equal(t, 120, *doc.LandAreaSqm)
	require.NotNil(t, doc.BuildingAreaSqm)
	require.Equal(t, 90, *doc.BuildingAreaSqm)
	require.Equal(t, responsedto.ChatAnswerFormatText, response.AnswerFormat)
}

func TestChatServiceRespondDegradesOnRetrievalError(t *testing.T) {
	svc, err := NewChatService(fakeChatServiceRetrieval{err: errors.New("es timeout")}, &fakeChatMemory{}, &fakeAnswerGenerator{}, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-3", Message: "Cari apartemen"})
	require.NoError(t, err)
	require.True(t, response.Grounding.IsDegraded)
	require.Equal(t, "retrieval_unavailable", response.Grounding.DegradedReason)
	require.Equal(t, responsedto.ChatAnswerFormatText, response.AnswerFormat)
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
	require.Equal(t, responsedto.ChatAnswerFormatText, response.AnswerFormat)
}

func TestChatServiceRespondEmitsMarkdownWhenLinkAllowed(t *testing.T) {
	slug := "rumah-aktif"
	retrieval := fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Title: "Rumah Aktif", Slug: slug, Status: "active"}}, Grounding: domain.ChatGrounding{ListingIDs: []uuid.UUID{uuid.MustParse("22222222-2222-2222-2222-222222222222")}, ListingSlugs: []string{slug}}}}
	gen := &fakeAnswerGenerator{response: &gemini.GroundedAnswerResponse{Answer: "Silakan lihat [Rumah Aktif](/listings/rumah-aktif) untuk detail."}}
	svc, err := NewChatService(retrieval, &fakeChatMemory{}, gen, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-md", Message: "Ada rumah unik?"})
	require.NoError(t, err)
	require.Equal(t, responsedto.ChatAnswerFormatMarkdown, response.AnswerFormat)
	require.Equal(t, "Silakan lihat [Rumah Aktif](/listings/rumah-aktif) untuk detail.", response.Answer)
	require.Len(t, response.Recommendations, 1)
	require.Equal(t, slug, response.Recommendations[0].Slug)
}

func TestChatServiceRespondStripsDisallowedLinks(t *testing.T) {
	retrieval := fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), Title: "Rumah Aktif", Slug: "rumah-aktif", Status: "active"}}, Grounding: domain.ChatGrounding{ListingIDs: []uuid.UUID{uuid.MustParse("33333333-3333-3333-3333-333333333333")}, ListingSlugs: []string{"rumah-aktif"}}}}
	gen := &fakeAnswerGenerator{response: &gemini.GroundedAnswerResponse{Answer: "Lihat [Rumah Luar](/listings/rumah-luar) atau [Ilegal](https://example.com)."}}
	svc, err := NewChatService(retrieval, &fakeChatMemory{}, gen, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-bad", Message: "Pilihan lain?"})
	require.NoError(t, err)
	require.Equal(t, responsedto.ChatAnswerFormatText, response.AnswerFormat)
	require.Equal(t, "Lihat Rumah Luar atau Ilegal.", response.Answer)
	require.Len(t, response.Recommendations, 1)
	require.Equal(t, "rumah-aktif", response.Recommendations[0].Slug)
}

func TestChatServiceRespondSanitizesMixedLinks(t *testing.T) {
	slug := "rumah-aktif"
	retrieval := fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("44444444-4444-4444-4444-444444444444"), Title: "Rumah Aktif", Slug: slug, Status: "active"}}, Grounding: domain.ChatGrounding{ListingIDs: []uuid.UUID{uuid.MustParse("44444444-4444-4444-4444-444444444444")}, ListingSlugs: []string{slug}}}}
	gen := &fakeAnswerGenerator{response: &gemini.GroundedAnswerResponse{Answer: "Cek [Rumah Aktif](/listings/rumah-aktif) dan [Eksternal](https://evil.example)."}}
	svc, err := NewChatService(retrieval, &fakeChatMemory{}, gen, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-mixed", Message: "Tampilkan"})
	require.NoError(t, err)
	require.Equal(t, responsedto.ChatAnswerFormatMarkdown, response.AnswerFormat)
	require.Equal(t, "Cek [Rumah Aktif](/listings/rumah-aktif) dan Eksternal.", response.Answer)
	require.Len(t, response.Recommendations, 1)
	require.Equal(t, slug, response.Recommendations[0].Slug)
}

func TestChatServiceRespondDropsUnsupportedHeadingLevels(t *testing.T) {
	slug := "rumah-utama"
	retrieval := fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("66666666-6666-6666-6666-666666666666"), Title: "Rumah Utama", Slug: slug, Status: "active"}}, Grounding: domain.ChatGrounding{ListingIDs: []uuid.UUID{uuid.MustParse("66666666-6666-6666-6666-666666666666")}, ListingSlugs: []string{slug}}}}
	answer := "## Fitur Unggulan\n- Lokasi strategis dekat sekolah\n- Harga kompetitif\nTerima kasih."
	gen := &fakeAnswerGenerator{response: &gemini.GroundedAnswerResponse{Answer: answer}}
	svc, err := NewChatService(retrieval, &fakeChatMemory{}, gen, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-md-structure", Message: "Fitur"})
	require.NoError(t, err)
	require.Equal(t, responsedto.ChatAnswerFormatMarkdown, response.AnswerFormat)
	require.Equal(t, "Fitur Unggulan\n- Lokasi strategis dekat sekolah\n- Harga kompetitif\nTerima kasih.", response.Answer)
	require.Len(t, response.Recommendations, 1)
	require.Equal(t, slug, response.Recommendations[0].Slug)
}

func TestChatServiceRespondSanitizesUnsupportedHeadings(t *testing.T) {
	slug := "rumah-aktif"
	retrieval := fakeChatServiceRetrieval{result: &domain.ChatRetrievalResult{Documents: []domain.ChatRetrievalDocument{{ListingID: uuid.MustParse("77777777-7777-7777-7777-777777777777"), Title: "Rumah Aktif", Slug: slug, Status: "active"}}, Grounding: domain.ChatGrounding{ListingIDs: []uuid.UUID{uuid.MustParse("77777777-7777-7777-7777-777777777777")}, ListingSlugs: []string{slug}}}}
	answer := "# Top Level\n## Secondary\n### Allowed Title\n#### Allowed Subtitle\n##### Disallowed Highland\n###### Hidden\nBody text."
	gen := &fakeAnswerGenerator{response: &gemini.GroundedAnswerResponse{Answer: answer}}
	svc, err := NewChatService(retrieval, &fakeChatMemory{}, gen, 5)
	require.NoError(t, err)

	response, err := svc.Respond(context.Background(), requestdto.ChatRequest{SessionID: "ses-heading", Message: "Heading"})
	require.NoError(t, err)
	require.Equal(t, responsedto.ChatAnswerFormatMarkdown, response.AnswerFormat)
	require.Equal(t, "Top Level\nSecondary\n### Allowed Title\n#### Allowed Subtitle\nDisallowed Highland\nHidden\nBody text.", response.Answer)
	require.Len(t, response.Recommendations, 1)
	require.Equal(t, slug, response.Recommendations[0].Slug)
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}
