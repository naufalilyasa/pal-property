package http_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/faux"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testcontainerRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	handler "github.com/username/pal-property-backend/internal/handler/http"
	repo "github.com/username/pal-property-backend/internal/repository/postgres"
	redisRepo "github.com/username/pal-property-backend/internal/repository/redis"
	"github.com/username/pal-property-backend/internal/service"
	"github.com/username/pal-property-backend/pkg/config"
	"github.com/username/pal-property-backend/pkg/logger"
	"go.uber.org/zap"
	pgDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/username/pal-property-backend/internal/domain/entity"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	app            *fiber.App
	db             *gorm.DB
	pgContainer    *postgres.PostgresContainer
	redisContainer *testcontainerRedis.RedisContainer
	rdb            *redis.Client
	ctx            context.Context
	fauxProvider   *faux.Provider
}

// SetupSuite runs once before the tests in the suite
func (s *AuthHandlerTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// 1. Disable Zap logging during tests
	logger.Log = zap.NewNop()

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

	config.Env.JwtAccessExpiration = time.Minute * 15
	config.Env.JwtRefreshExpiration = time.Hour * 168

	// 2. Setup Testcontainers Postgres (pal_db_test)
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

	// Migrate essential models
	err = s.db.AutoMigrate(&entity.User{}, &entity.OAuthAccount{})
	s.Require().NoError(err)

	redisContainer, err := testcontainerRedis.Run(s.ctx, "redis:8.2-alpine")
	s.Require().NoError(err)
	s.redisContainer = redisContainer

	redisAddr, err := redisContainer.ConnectionString(s.ctx)
	s.Require().NoError(err)

	opt, err := redis.ParseURL(redisAddr)
	s.Require().NoError(err)
	s.rdb = redis.NewClient(opt)

	// Setup Faux Provider
	s.fauxProvider = &faux.Provider{}
	goth.UseProviders(s.fauxProvider)

	// Setup Gothic Cookie Store using Gorilla Sessions
	store := sessions.NewCookieStore([]byte("test-session-secret"))
	store.MaxAge(86400)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = false
	gothic.Store = store

	// 3. Initialize layers
	authRepo := repo.NewAuthRepository(s.db)
	cacheRepo := redisRepo.NewCacheRepository(s.rdb)
	authService := service.NewAuthService(authRepo, cacheRepo)
	authHandler := handler.NewAuthHandler(authService)

	// 4. Initialize Fiber
	s.app = fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			msg := err.Error()
			return c.Status(code).JSON(fiber.Map{
				"success":  false,
				"message":  msg,
				"data":     nil,
				"trace_id": "test-trace",
			})
		},
	})

	auth := s.app.Group("/auth")
	auth.Get("/:provider/callback", authHandler.Callback)
}

func (s *AuthHandlerTestSuite) TearDownSuite() {
	if s.pgContainer != nil {
		s.pgContainer.Terminate(s.ctx)
	}
	if s.redisContainer != nil {
		s.redisContainer.Terminate(s.ctx)
	}
}

// SetupTest runs before each test -> Clean Tables
func (s *AuthHandlerTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE users CASCADE")
	s.db.Exec("TRUNCATE TABLE oauth_accounts CASCADE")
	s.rdb.FlushDB(s.ctx)
}

// The Feature Test implementation
func (s *AuthHandlerTestSuite) TestOAuthCallback_Success() {
	// 1. Setup Request to callback handler
	req := httptest.NewRequest(http.MethodGet, "/auth/faux/callback?provider=faux", nil)

	// Gothic natively extracts the "provider" context key from the request Context.
	// Since Fiber isn't natively populating this raw *http.Request context for Gothic in the test environment, we must inject it.
	req = req.WithContext(context.WithValue(req.Context(), gothic.ProviderParamKey, "faux"))
	fauxSession := &faux.Session{
		Name:  "Jack Faux",
		Email: "jack.faux@example.com",
	}
	recorder := httptest.NewRecorder()

	// Use Gothic's native StoreInSession, which applies gzip to the mock session string before saving it
	gothic.StoreInSession("faux", fauxSession.Marshal(), req, recorder)

	// Inject the cookie generated by gothic.Store back into the mock Fiber request
	for _, cookie := range recorder.Result().Cookies() {
		req.AddCookie(cookie)
	}

	// 3. Execute Fiber Request
	res, err := s.app.Test(req, fiber.TestConfig{
		Timeout: 30 * time.Second,
	})
	s.Require().NoError(err)

	// 4. Verify HTTP and generic properties
	bodyBytes, _ := io.ReadAll(res.Body)
	s.Equal(http.StatusSeeOther, res.StatusCode, "Expected 303 Redirect from Auth Callback, got: "+string(bodyBytes))
	s.Equal("http://localhost:3000/dashboard", res.Header.Get("Location"), "Should redirect to frontend dashboard")

	// 5. Assert Cookies are set correctly
	var hasAccessToken, hasRefreshToken bool
	for _, cookie := range res.Cookies() {
		if cookie.Name == "access_token" {
			hasAccessToken = true
			s.True(cookie.HttpOnly)
			s.Equal(http.SameSiteLaxMode, cookie.SameSite)
		}
		if cookie.Name == "refresh_token" {
			hasRefreshToken = true
			s.True(cookie.HttpOnly)
		}
	}
	s.True(hasAccessToken, "Should set access_token cookie")
	s.True(hasRefreshToken, "Should set refresh_token cookie")

	// 6. Assert Database record persists the generated user correctly
	var user entity.User
	err = s.db.Where("email = ?", "jack.faux@example.com").First(&user).Error
	s.Require().NoError(err, "User record must exist in DB")
	s.Equal("jack.faux@example.com", user.Email)

	// 7. Verify Redis Key exists
	keys, err := s.rdb.Keys(s.ctx, "refresh_token:*").Result()
	s.Require().NoError(err)
	s.Len(keys, 1, "Should have exactly 1 refresh token in redis")
}

func (s *AuthHandlerTestSuite) TestOAuthCallback_Unauthorized() {
	// Simulating callback WITHOUT setting session cookie
	req := httptest.NewRequest(http.MethodGet, "/auth/faux/callback?provider=faux", nil)

	res, err := s.app.Test(req, fiber.TestConfig{
		Timeout: 30 * time.Second,
	})
	s.Require().NoError(err)

	// 401 occurs because Gothic evaluates `CompleteUserAuth` and fails due to missing provider session
	s.Equal(http.StatusUnauthorized, res.StatusCode)

	var responseBody map[string]interface{}
	json.NewDecoder(res.Body).Decode(&responseBody)
	s.Equal(false, responseBody["success"])
	s.NotNil(responseBody["message"])
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
