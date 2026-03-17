package http_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
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

type CategoryHandlerTestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *gorm.DB
	app         *fiber.App
	ctx         context.Context
}

func (s *CategoryHandlerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	logger.Log = zap.NewNop()
	config.Env.AppEnv = "testing"

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

	// 2. Start postgres testcontainer
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

	// 3. AutoMigrate
	err = s.db.AutoMigrate(&entity.User{}, &entity.Category{}, &entity.Listing{}, &entity.ListingImage{})
	s.Require().NoError(err)

	// 4. Create real layers
	categoryRepo := repo.NewCategoryRepository(s.db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	// 5. Create fiber app
	s.app = fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := err.Error()
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
				msg = fe.Message
			} else {
				switch {
				case errors.Is(err, domain.ErrNotFound):
					code = fiber.StatusNotFound
				case errors.Is(err, domain.ErrConflict):
					code = fiber.StatusConflict
				case errors.Is(err, domain.ErrForbidden):
					code = fiber.StatusForbidden
				case errors.Is(err, domain.ErrUnauthorized):
					code = fiber.StatusUnauthorized
				}
			}
			traceID := uuid.New().String()
			return c.Status(code).JSON(fiber.Map{
				"success":  false,
				"message":  msg,
				"data":     nil,
				"trace_id": traceID,
			})
		},
	})

	// Register routes
	s.app.Get("/api/categories", categoryHandler.List)
	s.app.Get("/api/categories/:slug", categoryHandler.GetBySlug)
	s.app.Post("/api/categories/", middleware.Protected(s.db), middleware.RequireRole("admin"), categoryHandler.Create)
	s.app.Put("/api/categories/:id", middleware.Protected(s.db), middleware.RequireRole("admin"), categoryHandler.Update)
	s.app.Delete("/api/categories/:id", middleware.Protected(s.db), middleware.RequireRole("admin"), categoryHandler.Delete)
}

func (s *CategoryHandlerTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE listings CASCADE")
	s.db.Exec("TRUNCATE categories CASCADE")
	s.db.Exec("TRUNCATE users CASCADE")
}

func (s *CategoryHandlerTestSuite) TearDownSuite() {
	if s.pgContainer != nil {
		s.pgContainer.Terminate(s.ctx)
	}
}

func TestCategoryHandlerSuite(t *testing.T) {
	suite.Run(t, new(CategoryHandlerTestSuite))
}

// Helpers

func (s *CategoryHandlerTestSuite) mintJWT(userID uuid.UUID) string {
	accessToken, _, _, err := jwt.GenerateTokens(userID)
	s.Require().NoError(err)
	return accessToken
}

func (s *CategoryHandlerTestSuite) createAdmin() entity.User {
	u := entity.User{Name: "Admin", Email: "admin@test.com", Role: "admin"}
	s.Require().NoError(s.db.Create(&u).Error)
	return u
}

func (s *CategoryHandlerTestSuite) createUser() entity.User {
	u := entity.User{Name: "User", Email: "user@test.com", Role: "user"}
	s.Require().NoError(s.db.Create(&u).Error)
	return u
}

func (s *CategoryHandlerTestSuite) createCategory(name, slug string, parentID *uuid.UUID) entity.Category {
	cat := entity.Category{Name: name, Slug: slug, ParentID: parentID}
	s.Require().NoError(s.db.Create(&cat).Error)
	return cat
}

func (s *CategoryHandlerTestSuite) doRequest(method, path string, body interface{}, token string) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		b, _ := sonic.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	}
	res, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	return res
}

// Test Cases

func (s *CategoryHandlerTestSuite) TestList_Empty() {
	res := s.doRequest(http.MethodGet, "/api/categories", nil, "")
	s.Equal(http.StatusOK, res.StatusCode)

	var result map[string]interface{}
	sonic.ConfigDefault.NewDecoder(res.Body).Decode(&result)
	s.Equal(true, result["success"])
}

func (s *CategoryHandlerTestSuite) TestList_WithRootsAndChildren() {
	parent := s.createCategory("Property", "property", nil)
	s.createCategory("House", "house", &parent.ID)
	s.createCategory("Apartment", "apartment", &parent.ID)

	res := s.doRequest(http.MethodGet, "/api/categories", nil, "")
	s.Equal(http.StatusOK, res.StatusCode)

	var result struct {
		Success bool `json:"success"`
		Data    []struct {
			ID       uuid.UUID `json:"id"`
			Children []struct {
				ID uuid.UUID `json:"id"`
			} `json:"children"`
		} `json:"data"`
	}
	sonic.ConfigDefault.NewDecoder(res.Body).Decode(&result)
	s.True(result.Success)
	s.Len(result.Data, 1)
	s.Len(result.Data[0].Children, 2)
}

func (s *CategoryHandlerTestSuite) TestGetBySlug_Found() {
	cat := s.createCategory("House", "house", nil)
	res := s.doRequest(http.MethodGet, "/api/categories/house", nil, "")
	s.Equal(http.StatusOK, res.StatusCode)

	var result map[string]interface{}
	sonic.ConfigDefault.NewDecoder(res.Body).Decode(&result)
	data := result["data"].(map[string]interface{})
	s.Equal(cat.ID.String(), data["id"])
	s.Equal("house", data["slug"])
}

func (s *CategoryHandlerTestSuite) TestGetBySlug_NotFound() {
	res := s.doRequest(http.MethodGet, "/api/categories/not-found", nil, "")
	s.Equal(http.StatusNotFound, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestCreate_AsAdmin_201() {
	admin := s.createAdmin()
	token := s.mintJWT(admin.ID)
	req := request.CreateCategoryRequest{Name: "Kondominium"}

	res := s.doRequest(http.MethodPost, "/api/categories/", req, token)
	s.Equal(http.StatusCreated, res.StatusCode)

	var result map[string]interface{}
	sonic.ConfigDefault.NewDecoder(res.Body).Decode(&result)
	data := result["data"].(map[string]interface{})
	s.Equal("kondominium", data["slug"])
}

func (s *CategoryHandlerTestSuite) TestCreate_AsUser_403() {
	user := s.createUser()
	token := s.mintJWT(user.ID)
	req := request.CreateCategoryRequest{Name: "Kondominium"}

	res := s.doRequest(http.MethodPost, "/api/categories/", req, token)
	s.Equal(http.StatusForbidden, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestCreate_NoAuth_401() {
	req := request.CreateCategoryRequest{Name: "Kondominium"}
	res := s.doRequest(http.MethodPost, "/api/categories/", req, "")
	s.Equal(http.StatusUnauthorized, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestCreate_MissingName_400() {
	admin := s.createAdmin()
	token := s.mintJWT(admin.ID)
	req := map[string]interface{}{}

	res := s.doRequest(http.MethodPost, "/api/categories/", req, token)
	s.Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestCreate_DuplicateName_201() {
	admin := s.createAdmin()
	token := s.mintJWT(admin.ID)
	s.createCategory("Rumah", "rumah", nil)
	req := request.CreateCategoryRequest{Name: "Rumah"}

	res := s.doRequest(http.MethodPost, "/api/categories/", req, token)
	s.Equal(http.StatusCreated, res.StatusCode)

	var result map[string]interface{}
	sonic.ConfigDefault.NewDecoder(res.Body).Decode(&result)
	data := result["data"].(map[string]interface{})
	s.Equal("rumah-2", data["slug"])
}


func (s *CategoryHandlerTestSuite) TestUpdate_AsAdmin_200() {
	admin := s.createAdmin()
	token := s.mintJWT(admin.ID)
	cat := s.createCategory("Old Name", "old-name", nil)
	newName := "New Name"
	req := request.UpdateCategoryRequest{Name: &newName}

	res := s.doRequest(http.MethodPut, "/api/categories/"+cat.ID.String(), req, token)
	s.Equal(http.StatusOK, res.StatusCode)

	var result map[string]interface{}
	sonic.ConfigDefault.NewDecoder(res.Body).Decode(&result)
	data := result["data"].(map[string]interface{})
	s.Equal("New Name", data["name"])
	s.Equal("old-name", data["slug"]) // Slug unchanged
}

func (s *CategoryHandlerTestSuite) TestUpdate_AsUser_403() {
	user := s.createUser()
	token := s.mintJWT(user.ID)
	cat := s.createCategory("Old Name", "old-name", nil)
	newName := "New Name"
	req := request.UpdateCategoryRequest{Name: &newName}

	res := s.doRequest(http.MethodPut, "/api/categories/"+cat.ID.String(), req, token)
	s.Equal(http.StatusForbidden, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestUpdate_NotFound_404() {
	admin := s.createAdmin()
	token := s.mintJWT(admin.ID)
	newName := "New Name"
	req := request.UpdateCategoryRequest{Name: &newName}

	res := s.doRequest(http.MethodPut, "/api/categories/"+uuid.New().String(), req, token)
	s.Equal(http.StatusNotFound, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestDelete_AsAdmin_200() {
	admin := s.createAdmin()
	token := s.mintJWT(admin.ID)
	cat := s.createCategory("Leaf", "leaf", nil)

	res := s.doRequest(http.MethodDelete, "/api/categories/"+cat.ID.String(), nil, token)
	s.Equal(http.StatusOK, res.StatusCode)

	// Verify deleted
	var count int64
	s.db.Model(&entity.Category{}).Where("id = ?", cat.ID).Count(&count)
	s.Equal(int64(0), count)
}

func (s *CategoryHandlerTestSuite) TestDelete_HasChildren_409() {
	admin := s.createAdmin()
	token := s.mintJWT(admin.ID)
	parent := s.createCategory("Parent", "parent", nil)
	s.createCategory("Child", "child", &parent.ID)

	res := s.doRequest(http.MethodDelete, "/api/categories/"+parent.ID.String(), nil, token)
	s.Equal(http.StatusConflict, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestDelete_HasListings_409() {
	cat := s.createCategory("Rumah", "rumah", nil)
	admin := s.createAdmin()
	// Create a listing that references this category
	listing := entity.Listing{
		UserID:     admin.ID,
		CategoryID: &cat.ID,
		Title:      "Test Listing",
		Slug:       "test-listing",
		Price:      500000000,
		Status:     "active",
	}
	s.Require().NoError(s.db.Create(&listing).Error)

	// Now try to delete the category
	token := s.mintJWT(admin.ID)
	res := s.doRequest(http.MethodDelete, "/api/categories/"+cat.ID.String(), nil, token)
	s.Equal(http.StatusConflict, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestDelete_AsUser_403() {
	user := s.createUser()
	token := s.mintJWT(user.ID)
	cat := s.createCategory("Leaf", "leaf", nil)

	res := s.doRequest(http.MethodDelete, "/api/categories/"+cat.ID.String(), nil, token)
	s.Equal(http.StatusForbidden, res.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestDelete_NotFound_404() {
	admin := s.createAdmin()
	token := s.mintJWT(admin.ID)

	res := s.doRequest(http.MethodDelete, "/api/categories/"+uuid.New().String(), nil, token)
	s.Equal(http.StatusNotFound, res.StatusCode)
}
