package http_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	handler "github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	repo "github.com/naufalilyasa/pal-property-backend/internal/repository/postgres"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/mediaasset"
	"github.com/naufalilyasa/pal-property-backend/pkg/middleware"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	pgDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type fakeListingImageStorage struct {
	mu         sync.Mutex
	uploads    []mediaasset.UploadInput
	destroys   []mediaasset.DestroyInput
	uploadSeq  int
	uploadGate *uploadGate
}

type uploadGate struct {
	mu       sync.Mutex
	expected int
	arrived  int
	ready    chan struct{}
}

func (f *fakeListingImageStorage) UploadListingImage(_ context.Context, input mediaasset.UploadInput) (*mediaasset.UploadResult, error) {
	f.mu.Lock()
	f.uploadSeq++
	f.uploads = append(f.uploads, input)
	seq := f.uploadSeq
	gate := f.uploadGate
	f.mu.Unlock()

	if gate != nil {
		gate.Wait()
	}

	return &mediaasset.UploadResult{
		AssetID:          fmt.Sprintf("asset-%d", seq),
		PublicID:         fmt.Sprintf("listing-image-%d", seq),
		Version:          int64(seq),
		SecureURL:        fmt.Sprintf("https://fake-cloudinary.example/listings/%s/%d", input.PublicID, seq),
		ResourceType:     mediaasset.DefaultResourceType,
		DeliveryType:     mediaasset.DefaultDeliveryType,
		Format:           "jpg",
		Bytes:            int64(1024 * seq),
		Width:            1200,
		Height:           800,
		OriginalFilename: input.File.Filename,
	}, nil
}

func (f *fakeListingImageStorage) DestroyListingImage(_ context.Context, input mediaasset.DestroyInput) (*mediaasset.DestroyResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.destroys = append(f.destroys, input)
	return &mediaasset.DestroyResult{Result: "ok"}, nil
}

func (f *fakeListingImageStorage) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.uploads = nil
	f.destroys = nil
	f.uploadSeq = 0
	f.uploadGate = nil
}

func (f *fakeListingImageStorage) ArmUploadGate(expected int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.uploadGate = &uploadGate{
		expected: expected,
		ready:    make(chan struct{}),
	}
}

func (g *uploadGate) Wait() {
	g.mu.Lock()
	g.arrived++
	if g.arrived == g.expected {
		close(g.ready)
	}
	ready := g.ready
	g.mu.Unlock()

	<-ready
}

func (f *fakeListingImageStorage) UploadCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.uploads)
}

func (f *fakeListingImageStorage) DestroyCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.destroys)
}

func (f *fakeListingImageStorage) LastDestroy() mediaasset.DestroyInput {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.destroys[len(f.destroys)-1]
}

type ListingHandlerTestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *gorm.DB
	app         *fiber.App
	ctx         context.Context
	storage     *fakeListingImageStorage
}

func (s *ListingHandlerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	logger.Log = zap.NewNop()

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	config.Env.JwtPrivateKeyBase64 = base64.StdEncoding.EncodeToString(privPEM)

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	config.Env.JwtPublicKeyBase64 = base64.StdEncoding.EncodeToString(pubPEM)

	config.Env.JwtAccessExpiration = 900
	config.Env.JwtRefreshExpiration = 604800
	config.Env.AppEnv = "testing"

	pgContainer, err := postgres.Run(s.ctx,
		"postgres:17.8-alpine",
		postgres.WithDatabase("pal_db_test"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second),
		),
	)
	s.Require().NoError(err)
	s.pgContainer = pgContainer

	dsn, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	s.db, err = gorm.Open(pgDriver.Open(dsn), &gorm.Config{})
	s.Require().NoError(err)

	err = s.db.AutoMigrate(&entity.User{}, &entity.Category{}, &entity.Listing{}, &entity.ListingImage{})
	s.Require().NoError(err)

	s.storage = &fakeListingImageStorage{}
	listingRepo := repo.NewListingRepository(s.db)
	listingService := service.NewListingService(listingRepo, s.storage)
	listingHandler := handler.NewListingHandler(listingService)

	s.app = fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if errors.Is(err, domain.ErrNotFound) {
				code = fiber.StatusNotFound
			} else if errors.Is(err, domain.ErrForbidden) {
				code = fiber.StatusForbidden
			} else if errors.Is(err, domain.ErrUnauthorized) {
				code = fiber.StatusUnauthorized
			} else if errors.Is(err, domain.ErrConflict) {
				code = fiber.StatusConflict
			} else if errors.Is(err, domain.ErrInvalidCredential) || errors.Is(err, domain.ErrInvalidImageFile) || errors.Is(err, domain.ErrImageOrderInvalid) {
				code = fiber.StatusBadRequest
			} else if errors.Is(err, domain.ErrImageLimitReached) {
				code = fiber.StatusConflict
			} else if errors.Is(err, domain.ErrImageStorageUnset) {
				code = fiber.StatusServiceUnavailable
			}
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			msg := err.Error()
			if code >= 500 {
				msg = "An unexpected error occurred"
			}
			return c.Status(code).JSON(fiber.Map{
				"success":  false,
				"message":  msg,
				"data":     nil,
				"trace_id": "test-trace",
			})
		},
	})

	api := s.app.Group("/api")
	api.Get("/listings", listingHandler.List)
	api.Get("/listings/slug/:slug", listingHandler.GetBySlug)
	api.Get("/listings/:id", listingHandler.GetByID)

	listingProtected := api.Group("/listings", middleware.Protected(s.db))
	listingProtected.Post("/", listingHandler.Create)
	listingProtected.Put("/:id", listingHandler.Update)
	listingProtected.Delete("/:id", listingHandler.Delete)
	listingProtected.Post("/:id/images", listingHandler.UploadImage)
	listingProtected.Delete("/:id/images/:imageId", listingHandler.DeleteImage)
	listingProtected.Patch("/:id/images/:imageId/primary", listingHandler.SetPrimaryImage)
	listingProtected.Patch("/:id/images/reorder", listingHandler.ReorderImages)

	authProtected := s.app.Group("/auth", middleware.Protected(s.db))
	authProtected.Get("/me/listings", listingHandler.ListByUserID)
}

func (s *ListingHandlerTestSuite) TearDownSuite() {
	if s.pgContainer != nil {
		s.pgContainer.Terminate(s.ctx)
	}
}

func (s *ListingHandlerTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE listing_images, listings, categories, users CASCADE")
	s.storage.Reset()
}

func (s *ListingHandlerTestSuite) mintJWT(userID uuid.UUID) string {
	accToken, _, _, err := jwt.GenerateTokens(userID)
	s.Require().NoError(err)
	return accToken
}

func (s *ListingHandlerTestSuite) createTestUser(role string) entity.User {
	user := entity.User{
		Name:  "Test User",
		Email: fmt.Sprintf("test-%s@example.com", uuid.NewString()[:8]),
		Role:  role,
	}
	err := s.db.Create(&user).Error
	s.Require().NoError(err)
	return user
}

func (s *ListingHandlerTestSuite) createTestCategory() entity.Category {
	cat := entity.Category{ID: uuid.New(), Name: "Apartment", Slug: "apartment"}
	err := s.db.Create(&cat).Error
	s.Require().NoError(err)
	return cat
}

func (s *ListingHandlerTestSuite) makeRequest(method, path string, body interface{}, token string) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		s.Require().NoError(err)
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	}

	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	return resp
}

func (s *ListingHandlerTestSuite) makeMultipartRequest(method, path, fieldName, fileName string, fileContent []byte, token string) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	s.Require().NoError(err)
	_, err = part.Write(fileContent)
	s.Require().NoError(err)
	err = writer.Close()
	s.Require().NoError(err)

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	}

	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	return resp
}

func (s *ListingHandlerTestSuite) createListing(token string, req request.CreateListingRequest) response.ListingResponse {
	resp := s.makeRequest(http.MethodPost, "/api/listings/", req, token)
	s.Equal(http.StatusCreated, resp.StatusCode)
	return decodeListingEnvelope(s.T(), resp).Data
}

func TestListingHandlerSuite(t *testing.T) {
	suite.Run(t, new(ListingHandlerTestSuite))
}

func (s *ListingHandlerTestSuite) TestCreateListing_Success() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)
	cat := s.createTestCategory()

	req := request.CreateListingRequest{
		CategoryID: &cat.ID,
		Title:      "Luxury Apartment in Jakarta",
		Price:      2500000000,
		Status:     "active",
		Specifications: request.Specifications{
			Bedrooms:    2,
			Bathrooms:   1,
			LandAreaSqm: 50,
		},
	}

	resp := s.makeRequest(http.MethodPost, "/api/listings/", req, token)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result := decodeListingEnvelope(s.T(), resp)
	s.True(result.Success)
	s.Equal(req.Title, result.Data.Title)
	s.NotEmpty(result.Data.Slug)
	s.Equal(user.ID, result.Data.UserID)
}

func (s *ListingHandlerTestSuite) TestCreateListing_InvalidPayload() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)

	req := request.CreateListingRequest{Price: 1000, Status: "active"}
	resp := s.makeRequest(http.MethodPost, "/api/listings/", req, token)
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	req2 := request.CreateListingRequest{Title: "Test Title", Price: 0, Status: "active"}
	resp2 := s.makeRequest(http.MethodPost, "/api/listings/", req2, token)
	s.Equal(http.StatusBadRequest, resp2.StatusCode)
}

func (s *ListingHandlerTestSuite) TestCreateListing_Unauthorized() {
	req := request.CreateListingRequest{Title: "Test Title", Price: 1000, Status: "active"}
	resp := s.makeRequest(http.MethodPost, "/api/listings/", req, "")
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *ListingHandlerTestSuite) TestGetListing_Success() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)

	listing := s.createListing(token, request.CreateListingRequest{Title: "Test Listing for Get", Price: 1000000, Status: "active"})

	resp := s.makeRequest(http.MethodGet, "/api/listings/"+listing.ID.String(), nil, "")
	s.Equal(http.StatusOK, resp.StatusCode)
	getResult := decodeListingEnvelope(s.T(), resp)
	s.Equal(listing.ID, getResult.Data.ID)
	s.Equal(1, getResult.Data.ViewCount)

	resp2 := s.makeRequest(http.MethodGet, "/api/listings/slug/"+listing.Slug, nil, "")
	s.Equal(http.StatusOK, resp2.StatusCode)
	getResult2 := decodeListingEnvelope(s.T(), resp2)
	s.Equal(listing.Slug, getResult2.Data.Slug)
	s.Equal(2, getResult2.Data.ViewCount)
}

func (s *ListingHandlerTestSuite) TestGetListing_NotFound() {
	resp := s.makeRequest(http.MethodGet, "/api/listings/"+uuid.New().String(), nil, "")
	s.Equal(http.StatusNotFound, resp.StatusCode)

	resp2 := s.makeRequest(http.MethodGet, "/api/listings/slug/non-existent-slug", nil, "")
	s.Equal(http.StatusNotFound, resp2.StatusCode)
}

func (s *ListingHandlerTestSuite) TestGetListing_BadRequest() {
	resp := s.makeRequest(http.MethodGet, "/api/listings/not-a-uuid", nil, "")
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *ListingHandlerTestSuite) TestListListings() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)

	for i := 1; i <= 3; i++ {
		s.createListing(token, request.CreateListingRequest{
			Title:        fmt.Sprintf("Listing %d", i),
			Price:        int64(i * 1000000),
			Status:       "active",
			LocationCity: ptr("Jakarta"),
		})
	}

	resp := s.makeRequest(http.MethodGet, "/api/listings?limit=2", nil, "")
	s.Equal(http.StatusOK, resp.StatusCode)
	listResult := decodePaginatedEnvelope(s.T(), resp)
	s.Len(listResult.Data.Data, 2)
	s.Equal(int64(3), listResult.Data.Total)

	resp2 := s.makeRequest(http.MethodGet, "/api/listings?city=Jakarta", nil, "")
	s.Equal(http.StatusOK, resp2.StatusCode)
	listResult2 := decodePaginatedEnvelope(s.T(), resp2)
	s.Len(listResult2.Data.Data, 3)

	resp3 := s.makeRequest(http.MethodGet, "/api/listings?price_min=100&price_max=50", nil, "")
	s.Equal(http.StatusOK, resp3.StatusCode)
	listResult3 := decodePaginatedEnvelope(s.T(), resp3)
	s.Empty(listResult3.Data.Data)
}

func (s *ListingHandlerTestSuite) TestUpdateListing() {
	owner := s.createTestUser("user")
	ownerToken := s.mintJWT(owner.ID)
	other := s.createTestUser("user")
	otherToken := s.mintJWT(other.ID)
	admin := s.createTestUser("admin")
	adminToken := s.mintJWT(admin.ID)

	listing := s.createListing(ownerToken, request.CreateListingRequest{Title: "Original Title", Price: 1000000, Status: "active"})

	resp := s.makeRequest(http.MethodPut, "/api/listings/"+listing.ID.String(), request.UpdateListingRequest{Title: ptr("Updated Title")}, ownerToken)
	s.Equal(http.StatusOK, resp.StatusCode)

	resp2 := s.makeRequest(http.MethodPut, "/api/listings/"+listing.ID.String(), request.UpdateListingRequest{Title: ptr("Blocked Title")}, otherToken)
	s.Equal(http.StatusForbidden, resp2.StatusCode)

	resp3 := s.makeRequest(http.MethodPut, "/api/listings/"+listing.ID.String(), request.UpdateListingRequest{Title: ptr("Admin Updated Title")}, adminToken)
	s.Equal(http.StatusOK, resp3.StatusCode)
}

func (s *ListingHandlerTestSuite) TestDeleteListing() {
	owner := s.createTestUser("user")
	ownerToken := s.mintJWT(owner.ID)
	other := s.createTestUser("user")
	otherToken := s.mintJWT(other.ID)

	listing := s.createListing(ownerToken, request.CreateListingRequest{Title: "Delete Me", Price: 1000000, Status: "active"})

	resp := s.makeRequest(http.MethodDelete, "/api/listings/"+listing.ID.String(), nil, otherToken)
	s.Equal(http.StatusForbidden, resp.StatusCode)

	resp2 := s.makeRequest(http.MethodDelete, "/api/listings/"+listing.ID.String(), nil, ownerToken)
	s.Equal(http.StatusOK, resp2.StatusCode)

	resp3 := s.makeRequest(http.MethodGet, "/api/listings/"+listing.ID.String(), nil, "")
	s.Equal(http.StatusNotFound, resp3.StatusCode)
}

func (s *ListingHandlerTestSuite) TestListingImageRoutes_WithFakeStorage() {
	owner := s.createTestUser("user")
	ownerToken := s.mintJWT(owner.ID)
	listing := s.createListing(ownerToken, request.CreateListingRequest{Title: "Listing With Images", Price: 1500000, Status: "active"})

	upload1 := s.makeMultipartRequest(http.MethodPost, "/api/listings/"+listing.ID.String()+"/images", "file", "front.jpg", []byte("front-image"), ownerToken)
	s.Equal(http.StatusOK, upload1.StatusCode)
	firstImageResp := decodeListingEnvelope(s.T(), upload1)
	s.Len(firstImageResp.Data.Images, 1)
	firstImage := firstImageResp.Data.Images[0]
	s.True(firstImage.IsPrimary)
	s.Equal(0, firstImage.SortOrder)
	s.NotNil(firstImage.OriginalFilename)
	s.Equal("front.jpg", *firstImage.OriginalFilename)
	s.Equal(1, s.storage.UploadCount())

	upload2 := s.makeMultipartRequest(http.MethodPost, "/api/listings/"+listing.ID.String()+"/images", "file", "living-room.jpg", []byte("living-room-image"), ownerToken)
	s.Equal(http.StatusOK, upload2.StatusCode)
	secondImageResp := decodeListingEnvelope(s.T(), upload2)
	s.Len(secondImageResp.Data.Images, 2)
	secondImage := secondImageResp.Data.Images[1]
	s.False(secondImage.IsPrimary)
	s.Equal(1, secondImage.SortOrder)

	upload3 := s.makeMultipartRequest(http.MethodPost, "/api/listings/"+listing.ID.String()+"/images", "file", "kitchen.jpg", []byte("kitchen-image"), ownerToken)
	s.Equal(http.StatusOK, upload3.StatusCode)
	thirdImageResp := decodeListingEnvelope(s.T(), upload3)
	s.Len(thirdImageResp.Data.Images, 3)
	thirdImage := thirdImageResp.Data.Images[2]
	s.False(thirdImage.IsPrimary)
	s.Equal(2, thirdImage.SortOrder)
	assertImageOrder(s, thirdImageResp.Data.Images, []uuid.UUID{firstImage.ID, secondImage.ID, thirdImage.ID}, true)

	setPrimaryResp := s.makeRequest(http.MethodPatch, "/api/listings/"+listing.ID.String()+"/images/"+thirdImage.ID.String()+"/primary", nil, ownerToken)
	s.Equal(http.StatusOK, setPrimaryResp.StatusCode)
	setPrimaryResult := decodeListingEnvelope(s.T(), setPrimaryResp)
	assertImageOrder(s, setPrimaryResult.Data.Images, []uuid.UUID{firstImage.ID, secondImage.ID, thirdImage.ID}, true)
	s.False(setPrimaryResult.Data.Images[0].IsPrimary)
	s.False(setPrimaryResult.Data.Images[1].IsPrimary)
	s.True(setPrimaryResult.Data.Images[2].IsPrimary)

	reorderReq := request.ReorderListingImagesRequest{OrderedImageIDs: []uuid.UUID{thirdImage.ID, firstImage.ID, secondImage.ID}}
	reorderResp := s.makeRequest(http.MethodPatch, "/api/listings/"+listing.ID.String()+"/images/reorder", reorderReq, ownerToken)
	s.Equal(http.StatusOK, reorderResp.StatusCode)
	reorderResult := decodeListingEnvelope(s.T(), reorderResp)
	assertImageOrder(s, reorderResult.Data.Images, []uuid.UUID{thirdImage.ID, firstImage.ID, secondImage.ID}, true)
	s.True(reorderResult.Data.Images[0].IsPrimary)

	getByIDResp := s.makeRequest(http.MethodGet, "/api/listings/"+listing.ID.String(), nil, "")
	s.Equal(http.StatusOK, getByIDResp.StatusCode)
	getByIDResult := decodeListingEnvelope(s.T(), getByIDResp)
	assertImageOrder(s, getByIDResult.Data.Images, []uuid.UUID{thirdImage.ID, firstImage.ID, secondImage.ID}, true)

	getBySlugResp := s.makeRequest(http.MethodGet, "/api/listings/slug/"+listing.Slug, nil, "")
	s.Equal(http.StatusOK, getBySlugResp.StatusCode)
	getBySlugResult := decodeListingEnvelope(s.T(), getBySlugResp)
	assertImageOrder(s, getBySlugResult.Data.Images, []uuid.UUID{thirdImage.ID, firstImage.ID, secondImage.ID}, true)

	listResp := s.makeRequest(http.MethodGet, "/api/listings?limit=10", nil, "")
	s.Equal(http.StatusOK, listResp.StatusCode)
	listResult := decodePaginatedEnvelope(s.T(), listResp)
	s.Len(listResult.Data.Data, 1)
	assertImageOrder(s, listResult.Data.Data[0].Images, []uuid.UUID{thirdImage.ID, firstImage.ID, secondImage.ID}, true)

	myListingsResp := s.makeRequest(http.MethodGet, "/auth/me/listings", nil, ownerToken)
	s.Equal(http.StatusOK, myListingsResp.StatusCode)
	myListingsResult := decodePaginatedEnvelope(s.T(), myListingsResp)
	s.Len(myListingsResult.Data.Data, 1)
	assertImageOrder(s, myListingsResult.Data.Data[0].Images, []uuid.UUID{thirdImage.ID, firstImage.ID, secondImage.ID}, true)

	deleteResp := s.makeRequest(http.MethodDelete, "/api/listings/"+listing.ID.String()+"/images/"+thirdImage.ID.String(), nil, ownerToken)
	s.Equal(http.StatusOK, deleteResp.StatusCode)
	deleteResult := decodeListingEnvelope(s.T(), deleteResp)
	s.Len(deleteResult.Data.Images, 2)
	assertImageOrder(s, deleteResult.Data.Images, []uuid.UUID{firstImage.ID, secondImage.ID}, true)
	s.True(deleteResult.Data.Images[0].IsPrimary)
	s.False(deleteResult.Data.Images[1].IsPrimary)
	s.Equal(1, s.storage.DestroyCount())
	lastDestroy := s.storage.LastDestroy()
	s.Equal("listing-image-3", lastDestroy.PublicID)
	s.True(lastDestroy.Invalidate)

	finalReadResp := s.makeRequest(http.MethodGet, "/api/listings/"+listing.ID.String(), nil, "")
	s.Equal(http.StatusOK, finalReadResp.StatusCode)
	finalReadResult := decodeListingEnvelope(s.T(), finalReadResp)
	assertImageOrder(s, finalReadResult.Data.Images, []uuid.UUID{firstImage.ID, secondImage.ID}, true)
	s.True(finalReadResult.Data.Images[0].IsPrimary)

	uploadAfterDeleteResp := s.makeMultipartRequest(http.MethodPost, "/api/listings/"+listing.ID.String()+"/images", "file", "balcony.jpg", []byte("balcony-image"), ownerToken)
	s.Equal(http.StatusOK, uploadAfterDeleteResp.StatusCode)
	uploadAfterDeleteResult := decodeListingEnvelope(s.T(), uploadAfterDeleteResp)
	s.Len(uploadAfterDeleteResult.Data.Images, 3)
	assertImageOrder(s, uploadAfterDeleteResult.Data.Images, []uuid.UUID{firstImage.ID, secondImage.ID, uploadAfterDeleteResult.Data.Images[2].ID}, true)
	s.True(uploadAfterDeleteResult.Data.Images[0].IsPrimary)
	s.False(uploadAfterDeleteResult.Data.Images[1].IsPrimary)
	s.False(uploadAfterDeleteResult.Data.Images[2].IsPrimary)
	s.Equal(2, uploadAfterDeleteResult.Data.Images[2].SortOrder)
	s.Equal(4, s.storage.UploadCount())
}

func (s *ListingHandlerTestSuite) TestUploadListingImage_NegativeRoutes() {
	owner := s.createTestUser("user")
	ownerToken := s.mintJWT(owner.ID)
	listing := s.createListing(ownerToken, request.CreateListingRequest{Title: "Listing Upload Guards", Price: 1750000, Status: "active"})

	unauthorizedResp := s.makeMultipartRequest(http.MethodPost, "/api/listings/"+listing.ID.String()+"/images", "file", "front.jpg", []byte("front-image"), "")
	s.Equal(http.StatusUnauthorized, unauthorizedResp.StatusCode)
	unauthorizedResult := decodeErrorEnvelope(s.T(), unauthorizedResp)
	s.False(unauthorizedResult.Success)
	s.Equal("missing access token", unauthorizedResult.Message)
	s.Nil(unauthorizedResult.Data)
	s.Equal(0, s.storage.UploadCount())

	invalidFileResp := s.makeMultipartRequest(http.MethodPost, "/api/listings/"+listing.ID.String()+"/images", "file", "notes.txt", []byte("not-an-image"), ownerToken)
	s.Equal(http.StatusBadRequest, invalidFileResp.StatusCode)
	invalidFileResult := decodeErrorEnvelope(s.T(), invalidFileResp)
	s.False(invalidFileResult.Success)
	s.Equal(domain.ErrInvalidImageFile.Error(), invalidFileResult.Message)
	s.Nil(invalidFileResult.Data)
	s.Equal(0, s.storage.UploadCount())
}

func (s *ListingHandlerTestSuite) TestUploadListingImage_ConcurrentFirstUploadsSerializeEmptyListing() {
	owner := s.createTestUser("user")
	ownerToken := s.mintJWT(owner.ID)
	listing := s.createListing(ownerToken, request.CreateListingRequest{Title: "Concurrent First Uploads", Price: 1850000, Status: "active"})

	s.storage.ArmUploadGate(2)

	type uploadAttempt struct {
		resp *http.Response
		err  error
	}

	upload := func(fileName string, content []byte) uploadAttempt {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			return uploadAttempt{err: err}
		}
		if _, err := part.Write(content); err != nil {
			return uploadAttempt{err: err}
		}
		if err := writer.Close(); err != nil {
			return uploadAttempt{err: err}
		}

		req := httptest.NewRequest(http.MethodPost, "/api/listings/"+listing.ID.String()+"/images", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: "access_token", Value: ownerToken})

		resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
		return uploadAttempt{resp: resp, err: err}
	}

	results := make(chan uploadAttempt, 2)
	var wg sync.WaitGroup
	for i, fileName := range []string{"front.jpg", "side.jpg"} {
		wg.Add(1)
		go func(name string, idx int) {
			defer wg.Done()
			results <- upload(name, []byte(fmt.Sprintf("image-%d", idx)))
		}(fileName, i)
	}
	wg.Wait()
	close(results)

	for result := range results {
		s.Require().NoError(result.err)
		if result.resp.StatusCode != http.StatusOK {
			body, readErr := io.ReadAll(result.resp.Body)
			if readErr == nil {
				_ = result.resp.Body.Close()
				s.FailNow("expected concurrent upload to succeed", string(body))
			}
			s.FailNow("expected concurrent upload to succeed", readErr.Error())
		}
		_ = result.resp.Body.Close()
	}

	finalResp := s.makeRequest(http.MethodGet, "/api/listings/"+listing.ID.String(), nil, "")
	s.Equal(http.StatusOK, finalResp.StatusCode)
	finalResult := decodeListingEnvelope(s.T(), finalResp)
	s.Len(finalResult.Data.Images, 2)
	assertImageOrder(s, finalResult.Data.Images, []uuid.UUID{finalResult.Data.Images[0].ID, finalResult.Data.Images[1].ID}, true)
	s.True(finalResult.Data.Images[0].IsPrimary)
	s.False(finalResult.Data.Images[1].IsPrimary)
	s.Equal(0, finalResult.Data.Images[0].SortOrder)
	s.Equal(1, finalResult.Data.Images[1].SortOrder)
	s.Equal(2, s.storage.UploadCount())
	s.Equal(0, s.storage.DestroyCount())
}

func (s *ListingHandlerTestSuite) TestListByUserID() {
	user1 := s.createTestUser("user")
	token1 := s.mintJWT(user1.ID)
	user2 := s.createTestUser("user")
	token2 := s.mintJWT(user2.ID)

	s.createListing(token1, request.CreateListingRequest{Title: "U1 L1 Listing", Price: 100, Status: "active"})
	s.createListing(token1, request.CreateListingRequest{Title: "U1 L2 Listing", Price: 200, Status: "active"})

	resp := s.makeRequest(http.MethodGet, "/auth/me/listings", nil, token1)
	s.Equal(http.StatusOK, resp.StatusCode)
	listResult := decodePaginatedEnvelope(s.T(), resp)
	s.Len(listResult.Data.Data, 2)

	resp2 := s.makeRequest(http.MethodGet, "/auth/me/listings", nil, token2)
	s.Equal(http.StatusOK, resp2.StatusCode)
	listResult2 := decodePaginatedEnvelope(s.T(), resp2)
	s.Empty(listResult2.Data.Data)
}

func decodeListingEnvelope(t *testing.T, resp *http.Response) struct {
	Success bool                     `json:"success"`
	Data    response.ListingResponse `json:"data"`
} {
	t.Helper()
	defer resp.Body.Close()

	var result struct {
		Success bool                     `json:"success"`
		Data    response.ListingResponse `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("decode listing envelope: %v", err)
	}

	return result
}

func decodePaginatedEnvelope(t *testing.T, resp *http.Response) struct {
	Success bool                       `json:"success"`
	Data    response.PaginatedListings `json:"data"`
} {
	t.Helper()
	defer resp.Body.Close()

	var result struct {
		Success bool                       `json:"success"`
		Data    response.PaginatedListings `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("decode paginated envelope: %v", err)
	}

	return result
}

func decodeErrorEnvelope(t *testing.T, resp *http.Response) struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
} {
	t.Helper()
	defer resp.Body.Close()

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}

	return result
}

func assertImageOrder(s *ListingHandlerTestSuite, images []*response.ListingImageResponse, expected []uuid.UUID, expectNormalizedSort bool) {
	s.Require().Len(images, len(expected))
	for i, image := range images {
		s.Equal(expected[i], image.ID)
		if expectNormalizedSort {
			s.Equal(i, image.SortOrder)
		}
	}
}

func ptr[T any](v T) *T {
	return &v
}
