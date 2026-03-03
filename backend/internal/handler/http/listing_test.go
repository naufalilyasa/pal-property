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
	"net/http"
	"net/http/httptest"
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

type ListingHandlerTestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *gorm.DB
	app         *fiber.App
	ctx         context.Context
}

func (s *ListingHandlerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	logger.Log = zap.NewNop()

	// 1. Setup JWT Keys
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	config.Env.JwtPrivateKeyBase64 = base64.StdEncoding.EncodeToString(privPEM)

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	config.Env.JwtPublicKeyBase64 = base64.StdEncoding.EncodeToString(pubPEM)

	config.Env.JwtAccessExpiration = 900
	config.Env.JwtRefreshExpiration = 604800
	config.Env.AppEnv = "testing"

	// 2. Setup Testcontainers Postgres
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

	// Migrate models
	err = s.db.AutoMigrate(&entity.User{}, &entity.Category{}, &entity.Listing{}, &entity.ListingImage{})
	s.Require().NoError(err)

	// 3. Initialize layers
	listingRepo := repo.NewListingRepository(s.db)
	listingService := service.NewListingService(listingRepo)
	listingHandler := handler.NewListingHandler(listingService)

	// 4. Initialize Fiber
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
			} else if errors.Is(err, domain.ErrInvalidCredential) {
				code = fiber.StatusBadRequest
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

	// Register routes manually
	api := s.app.Group("/api")
	api.Get("/listings", listingHandler.List)
	api.Get("/listings/slug/:slug", listingHandler.GetBySlug)
	api.Get("/listings/:id", listingHandler.GetByID)

	listingProtected := api.Group("/listings", middleware.Protected(s.db))
	listingProtected.Post("/", listingHandler.Create)
	listingProtected.Put("/:id", listingHandler.Update)
	listingProtected.Delete("/:id", listingHandler.Delete)

	authProtected := s.app.Group("/auth", middleware.Protected(s.db))
	authProtected.Get("/me/listings", listingHandler.ListByUserID)
	authProtected.Get("/me/listings", listingHandler.ListByUserID)
	authProtected.Get("/me/listings", listingHandler.ListByUserID)
}

func (s *ListingHandlerTestSuite) TearDownSuite() {
	if s.pgContainer != nil {
		s.pgContainer.Terminate(s.ctx)
	}
}

func (s *ListingHandlerTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE listings CASCADE")
	s.db.Exec("TRUNCATE TABLE categories CASCADE")
	s.db.Exec("TRUNCATE TABLE users CASCADE")
}

// Helpers

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
	cat := entity.Category{
		ID:   uuid.New(),
		Name: "Apartment",
		Slug: "apartment",
	}
	err := s.db.Create(&cat).Error
	s.Require().NoError(err)
	return cat
}

func (s *ListingHandlerTestSuite) makeRequest(method, path string, body interface{}, token string) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: token,
		})
	}

	resp, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	return resp
}

func TestListingHandlerSuite(t *testing.T) {
	suite.Run(t, new(ListingHandlerTestSuite))
}

// Test Cases

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

	var result struct {
		Success bool                     `json:"success"`
		Data    response.ListingResponse `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)
	s.True(result.Success)
	s.Equal(req.Title, result.Data.Title)
	s.NotEmpty(result.Data.Slug)
	s.Equal(user.ID, result.Data.UserID)
}

func (s *ListingHandlerTestSuite) TestCreateListing_InvalidPayload() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)

	// Missing title
	req := request.CreateListingRequest{
		Price:  1000,
		Status: "active",
	}

	resp := s.makeRequest(http.MethodPost, "/api/listings/", req, token)
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	// Price = 0
	req2 := request.CreateListingRequest{
		Title:  "Test Title",
		Price:  0,
		Status: "active",
	}
	resp2 := s.makeRequest(http.MethodPost, "/api/listings/", req2, token)
	s.Equal(http.StatusBadRequest, resp2.StatusCode)
}

func (s *ListingHandlerTestSuite) TestCreateListing_Unauthorized() {
	req := request.CreateListingRequest{
		Title:  "Test Title",
		Price:  1000,
		Status: "active",
	}
	resp := s.makeRequest(http.MethodPost, "/api/listings/", req, "")
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *ListingHandlerTestSuite) TestGetListing_Success() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)

	createReq := request.CreateListingRequest{
		Title:  "Test Listing for Get",
		Price:  1000000,
		Status: "active",
	}
	createResp := s.makeRequest(http.MethodPost, "/api/listings/", createReq, token)
	var createResult struct {
		Data response.ListingResponse `json:"data"`
	}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	listingID := createResult.Data.ID
	slug := createResult.Data.Slug

	// GetByID
	resp := s.makeRequest(http.MethodGet, "/api/listings/"+listingID.String(), nil, "")
	s.Equal(http.StatusOK, resp.StatusCode)
	var getResult struct {
		Data response.ListingResponse `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&getResult)
	s.Equal(listingID, getResult.Data.ID)
	s.Equal(1, getResult.Data.ViewCount) // Incremented

	// GetBySlug
	resp2 := s.makeRequest(http.MethodGet, "/api/listings/slug/"+slug, nil, "")
	s.Equal(http.StatusOK, resp2.StatusCode)
	json.NewDecoder(resp2.Body).Decode(&getResult)
	s.Equal(slug, getResult.Data.Slug)
	s.Equal(2, getResult.Data.ViewCount) // Incremented again
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

	// Create 3 listings
	for i := 1; i <= 3; i++ {
		req := request.CreateListingRequest{
			Title:        fmt.Sprintf("Listing %d", i),
			Price:        int64(i * 1000000),
			Status:       "active",
			LocationCity: ptr("Jakarta"),
		}
		s.makeRequest(http.MethodPost, "/api/listings/", req, token)
	}

	// List with limit=2
	resp := s.makeRequest(http.MethodGet, "/api/listings?limit=2", nil, "")
	s.Equal(http.StatusOK, resp.StatusCode)
	var listResult struct {
		Data response.PaginatedListings `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&listResult)
	s.Len(listResult.Data.Data, 2)
	s.Equal(int64(3), listResult.Data.Total)

	// Filter by city
	resp2 := s.makeRequest(http.MethodGet, "/api/listings?city=Jakarta", nil, "")
	s.Equal(http.StatusOK, resp2.StatusCode)
	json.NewDecoder(resp2.Body).Decode(&listResult)
	s.Len(listResult.Data.Data, 3)

	// Price range edge case: min > max
	resp3 := s.makeRequest(http.MethodGet, "/api/listings?price_min=100&price_max=50", nil, "")
	s.Equal(http.StatusOK, resp3.StatusCode)
	json.NewDecoder(resp3.Body).Decode(&listResult)
	s.Empty(listResult.Data.Data)
}

func (s *ListingHandlerTestSuite) TestUpdateListing() {
	owner := s.createTestUser("user")
	ownerToken := s.mintJWT(owner.ID)

	other := s.createTestUser("user")
	otherToken := s.mintJWT(other.ID)

	admin := s.createTestUser("admin")
	adminToken := s.mintJWT(admin.ID)

	// Create listing
	createReq := request.CreateListingRequest{
		Title:  "Original Title",
		Price:  1000000,
		Status: "active",
	}
	createResp := s.makeRequest(http.MethodPost, "/api/listings/", createReq, ownerToken)
	var createResult struct {
		Data response.ListingResponse `json:"data"`
	}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	listingID := createResult.Data.ID

	// Update by owner
	updateReq := request.UpdateListingRequest{
		Title: ptr("Updated Title"),
	}
	resp := s.makeRequest(http.MethodPut, "/api/listings/"+listingID.String(), updateReq, ownerToken)
	s.Equal(http.StatusOK, resp.StatusCode)

	// Update by non-owner (forbidden)
	resp2 := s.makeRequest(http.MethodPut, "/api/listings/"+listingID.String(), updateReq, otherToken)
	s.Equal(http.StatusForbidden, resp2.StatusCode)

	// Update by admin (success)
	adminUpdateReq := request.UpdateListingRequest{
		Title: ptr("Admin Updated Title"),
	}
	resp3 := s.makeRequest(http.MethodPut, "/api/listings/"+listingID.String(), adminUpdateReq, adminToken)
	s.Equal(http.StatusOK, resp3.StatusCode)
	s.Equal(http.StatusOK, resp3.StatusCode)
}

func (s *ListingHandlerTestSuite) TestDeleteListing() {
	owner := s.createTestUser("user")
	ownerToken := s.mintJWT(owner.ID)

	other := s.createTestUser("user")
	otherToken := s.mintJWT(other.ID)

	// Create listing
	createReq := request.CreateListingRequest{
		Title:  "Delete Me",
		Price:  1000000,
		Status: "active",
	}
	createResp := s.makeRequest(http.MethodPost, "/api/listings/", createReq, ownerToken)
	var createResult struct {
		Data response.ListingResponse `json:"data"`
	}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	listingID := createResult.Data.ID

	// Delete by non-owner (forbidden)
	resp := s.makeRequest(http.MethodDelete, "/api/listings/"+listingID.String(), nil, otherToken)
	s.Equal(http.StatusForbidden, resp.StatusCode)

	// Delete by owner (success)
	resp2 := s.makeRequest(http.MethodDelete, "/api/listings/"+listingID.String(), nil, ownerToken)
	s.Equal(http.StatusOK, resp2.StatusCode)

	// Verify not found after delete
	resp3 := s.makeRequest(http.MethodGet, "/api/listings/"+listingID.String(), nil, "")
	s.Equal(http.StatusNotFound, resp3.StatusCode)
}

func (s *ListingHandlerTestSuite) TestListByUserID() {
	user1 := s.createTestUser("user")
	token1 := s.mintJWT(user1.ID)

	user2 := s.createTestUser("user")
	token2 := s.mintJWT(user2.ID)

	// User 1 creates 2 listings
	s.makeRequest(http.MethodPost, "/api/listings/", request.CreateListingRequest{Title: "U1 L1", Price: 100, Status: "active"}, token1)
	s.makeRequest(http.MethodPost, "/api/listings/", request.CreateListingRequest{Title: "U1 L2", Price: 200, Status: "active"}, token1)

	// List for user 1
	resp := s.makeRequest(http.MethodGet, "/auth/me/listings", nil, token1)
	s.Equal(http.StatusOK, resp.StatusCode)
	var listResult struct {
		Data response.PaginatedListings `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&listResult)
	s.Len(listResult.Data.Data, 2)

	// List for user 2 (empty)
	resp2 := s.makeRequest(http.MethodGet, "/auth/me/listings", nil, token2)
	s.Equal(http.StatusOK, resp2.StatusCode)
	json.NewDecoder(resp2.Body).Decode(&listResult)
	s.Empty(listResult.Data.Data)
}

func ptr[T any](v T) *T {
	return &v
}
