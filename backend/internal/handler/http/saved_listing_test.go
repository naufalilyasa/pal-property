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
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	handler "github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	repo "github.com/naufalilyasa/pal-property-backend/internal/repository/postgres"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/middleware"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	pd "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SavedListingHandlerTestSuite struct {
	suite.Suite
	ctx          context.Context
	pgContainer  *postgres.PostgresContainer
	db           *gorm.DB
	app          *fiber.App
	authzService *authz.Service
	handler      *handler.SavedListingHandler
}

func (s *SavedListingHandlerTestSuite) SetupSuite() {
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

	s.db, err = gorm.Open(pd.Open(dsn), &gorm.Config{})
	s.Require().NoError(err)

	err = s.db.AutoMigrate(&entity.User{}, &entity.Category{}, &entity.Listing{}, &entity.ListingImage{}, &entity.ListingVideo{}, &entity.SavedListing{})
	s.Require().NoError(err)
	s.Require().NoError(s.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_saved_listings_user_listing ON saved_listings (user_id, listing_id)").Error)

	s.authzService, err = newAuthzService(s.db)
	s.Require().NoError(err)

	listingRepo := repo.NewListingRepository(s.db)
	savedListingRepo := repo.NewSavedListingRepository(s.db)
	s.handler = handler.NewSavedListingHandler(service.NewSavedListingService(savedListingRepo, listingRepo))

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
			} else if errors.Is(err, domain.ErrInvalidImageFile) || errors.Is(err, domain.ErrImageOrderInvalid) {
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
			return c.Status(code).JSON(fiber.Map{"success": false, "message": msg, "data": nil, "trace_id": "test-trace"})
		},
	})

	api := s.app.Group("/api")
	savedProtected := api.Group("/me", middleware.Protected(s.db, s.authzService))
	savedProtected.Get("/saved-listings", s.handler.List)
	savedProtected.Get("/saved-listings/contains", s.handler.Contains)
	savedProtected.Post("/saved-listings", s.handler.Save)
	savedProtected.Delete("/saved-listings/:listingId", s.handler.Remove)
}

func (s *SavedListingHandlerTestSuite) TearDownSuite() {
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(s.ctx)
	}
}

func (s *SavedListingHandlerTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE saved_listings, listing_images, listings, categories, users CASCADE")
}

func (s *SavedListingHandlerTestSuite) TestListSavedListings_Unauthorized() {
	resp := s.makeRequest(http.MethodGet, "/api/me/saved-listings", nil, "")
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *SavedListingHandlerTestSuite) TestSaveInvalidPayload() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)
	resp := s.makeRequest(http.MethodPost, "/api/me/saved-listings", map[string]string{}, token)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	resp2 := s.makeRequest(http.MethodPost, "/api/me/saved-listings", map[string]string{"listing_id": "not-uuid"}, token)
	s.Equal(http.StatusBadRequest, resp2.StatusCode)
}

func (s *SavedListingHandlerTestSuite) TestContainsInvalidUUID() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)
	resp := s.makeRequest(http.MethodGet, "/api/me/saved-listings/contains?listing_ids=not-uuid", nil, token)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *SavedListingHandlerTestSuite) TestContainsLimitExceeded() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)
	ids := make([]string, service.SavedListingContainsLimit+1)
	for i := range ids {
		ids[i] = uuid.NewString()
	}
	query := fmt.Sprintf("/api/me/saved-listings/contains?listing_ids=%s", strings.Join(ids, ","))
	resp := s.makeRequest(http.MethodGet, query, nil, token)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *SavedListingHandlerTestSuite) TestDeleteInvalidUUID() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)
	resp := s.makeRequest(http.MethodDelete, "/api/me/saved-listings/not-uuid", nil, token)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *SavedListingHandlerTestSuite) TestSaveContainsListAndDeleteFlow() {
	user := s.createTestUser("user")
	token := s.mintJWT(user.ID)
	listingA := s.createListing(user.ID)
	listingB := s.createListing(user.ID)
	listingC := s.createListing(user.ID)

	resp := s.makeRequest(http.MethodPost, "/api/me/saved-listings", map[string]string{"listing_id": listingA.ID.String()}, token)
	s.Equal(http.StatusCreated, resp.StatusCode)
	saved := decodeSavedListingEnvelope(s.T(), resp)
	s.Equal(listingA.ID, saved.Data.ListingID)
	s.True(saved.Data.Saved)

	dup := s.makeRequest(http.MethodPost, "/api/me/saved-listings", map[string]string{"listing_id": listingA.ID.String()}, token)
	s.Equal(http.StatusCreated, dup.StatusCode)
	dupResult := decodeSavedListingEnvelope(s.T(), dup)
	s.Equal(listingA.ID, dupResult.Data.ListingID)
	s.True(dupResult.Data.Saved)

	respB := s.makeRequest(http.MethodPost, "/api/me/saved-listings", map[string]string{"listing_id": listingB.ID.String()}, token)
	s.Equal(http.StatusCreated, respB.StatusCode)

	listResp := s.makeRequest(http.MethodGet, "/api/me/saved-listings?page=1&limit=5", nil, token)
	s.Equal(http.StatusOK, listResp.StatusCode)
	listResult := decodePaginatedEnvelope(s.T(), listResp)
	s.Len(listResult.Data.Data, 2)

	containsResp := s.makeRequest(http.MethodGet, fmt.Sprintf("/api/me/saved-listings/contains?listing_ids=%s,%s,%s", listingA.ID, listingB.ID, listingC.ID), nil, token)
	s.Equal(http.StatusOK, containsResp.StatusCode)
	containsResult := decodeSavedListingContainsEnvelope(s.T(), containsResp)
	s.Equal([]uuid.UUID{listingA.ID, listingB.ID}, containsResult.Data.ListingIDs)

	deleteResp := s.makeRequest(http.MethodDelete, fmt.Sprintf("/api/me/saved-listings/%s", listingA.ID), nil, token)
	s.Equal(http.StatusOK, deleteResp.StatusCode)
	deleteResult := decodeSavedListingEnvelope(s.T(), deleteResp)
	s.False(deleteResult.Data.Saved)

	repeat := s.makeRequest(http.MethodDelete, fmt.Sprintf("/api/me/saved-listings/%s", listingA.ID), nil, token)
	s.Equal(http.StatusOK, repeat.StatusCode)
	repeatResult := decodeSavedListingEnvelope(s.T(), repeat)
	s.False(repeatResult.Data.Saved)

	afterDelete := s.makeRequest(http.MethodGet, "/api/me/saved-listings", nil, token)
	s.Equal(http.StatusOK, afterDelete.StatusCode)
	afterResult := decodePaginatedEnvelope(s.T(), afterDelete)
	s.Len(afterResult.Data.Data, 1)
	s.Equal(listingB.ID, afterResult.Data.Data[0].ID)
}

func (s *SavedListingHandlerTestSuite) makeRequest(method, path string, body interface{}, token string) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		s.Require().NoError(err)
		bodyReader = bytes.NewReader(jsonBytes)
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

func (s *SavedListingHandlerTestSuite) mintJWT(userID uuid.UUID) string {
	accToken, _, _, err := jwt.GenerateTokens(userID)
	s.Require().NoError(err)
	return accToken
}

func (s *SavedListingHandlerTestSuite) createTestUser(role string) entity.User {
	user := entity.User{Name: "Saved User", Email: fmt.Sprintf("saved-%s@example.com", uuid.NewString()[:8]), Role: role}
	s.Require().NoError(s.db.Create(&user).Error)
	return user
}

func (s *SavedListingHandlerTestSuite) createListing(ownerID uuid.UUID) entity.Listing {
	listing := entity.Listing{
		BaseEntity: entity.BaseEntity{ID: uuid.Must(uuid.NewV7())},
		UserID:     ownerID,
		Title:      fmt.Sprintf("Saved Listing %s", uuid.NewString()[:6]),
		Slug:       fmt.Sprintf("saved-%s", uuid.NewString()),
		Price:      1000000,
		Status:     "active",
	}
	s.Require().NoError(s.db.Create(&listing).Error)
	return listing
}

func decodeSavedListingEnvelope(t *testing.T, resp *http.Response) struct {
	Success bool `json:"success"`
	Data    struct {
		ListingID uuid.UUID `json:"listing_id"`
		Saved     bool      `json:"saved"`
	} `json:"data"`
} {
	t.Helper()
	defer resp.Body.Close()
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ListingID uuid.UUID `json:"listing_id"`
			Saved     bool      `json:"saved"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode saved listing envelope: %v", err)
	}
	return result
}

func decodeSavedListingContainsEnvelope(t *testing.T, resp *http.Response) struct {
	Success bool `json:"success"`
	Data    struct {
		ListingIDs []uuid.UUID `json:"listing_ids"`
	} `json:"data"`
} {
	t.Helper()
	defer resp.Body.Close()
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ListingIDs []uuid.UUID `json:"listing_ids"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode saved listing contains envelope: %v", err)
	}
	return result
}

func TestSavedListingHandlerSuite(t *testing.T) {
	suite.Run(t, new(SavedListingHandlerTestSuite))
}
