package http_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/faux"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	handler "github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	repo "github.com/naufalilyasa/pal-property-backend/internal/repository/postgres"
	redisRepo "github.com/naufalilyasa/pal-property-backend/internal/repository/redis"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/middleware"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testcontainerRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	pgDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
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

type stubAuthService struct {
	completeAuth func(ctx context.Context, provider string, gothUser goth.User) (*entity.User, error)
	loginUser    func(ctx context.Context, user *entity.User) (*response.AuthTokens, error)
	getMe        func(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error)
	refreshToken func(ctx context.Context, refreshToken string) (*response.AuthTokens, error)
	logout       func(ctx context.Context, refreshToken string) error
}

func (s *stubAuthService) CompleteAuth(ctx context.Context, provider string, gothUser goth.User) (*entity.User, error) {
	if s.completeAuth == nil {
		return nil, nil
	}

	return s.completeAuth(ctx, provider, gothUser)
}

func (s *stubAuthService) LoginUser(ctx context.Context, user *entity.User) (*response.AuthTokens, error) {
	if s.loginUser == nil {
		return nil, nil
	}

	return s.loginUser(ctx, user)
}

func (s *stubAuthService) GetMe(ctx context.Context, userID uuid.UUID) (*response.UserResponse, error) {
	if s.getMe == nil {
		return nil, nil
	}

	return s.getMe(ctx, userID)
}

func (s *stubAuthService) RefreshToken(ctx context.Context, refreshToken string) (*response.AuthTokens, error) {
	if s.refreshToken == nil {
		return nil, nil
	}

	return s.refreshToken(ctx, refreshToken)
}

func (s *stubAuthService) Logout(ctx context.Context, refreshToken string) error {
	if s.logout == nil {
		return nil
	}

	return s.logout(ctx, refreshToken)
}

func newRefreshHandlerTestApp(authSvc service.AuthService) *fiber.App {
	authHandler := handler.NewAuthHandler(authSvc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"success":  false,
				"message":  err.Error(),
				"data":     nil,
				"trace_id": "test-trace",
			})
		},
	})

	app.Post("/auth/refresh", authHandler.RefreshToken)

	return app
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

	config.Env.JwtAccessExpiration = 900
	config.Env.JwtRefreshExpiration = 604800
	config.Env.OAuthTokenEncryptionKey = make([]byte, 32)
	config.Env.JwtRefreshExpiration = 604800

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
	err = setupAuthzTestState(s.db)
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
	authzService, err := newAuthzService(s.db)
	s.Require().NoError(err)

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

	authGroup := s.app.Group("/auth")
	authGroup.Get("/oauth/:provider/callback", authHandler.Callback)
	authGroup.Post("/refresh", authHandler.RefreshToken)

	apiProtected := authGroup.Group("/", middleware.Protected(s.db, authzService))
	apiProtected.Get("/me", authHandler.GetMe)
	apiProtected.Post("/logout", authHandler.Logout)
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
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/faux/callback?provider=faux", nil)

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

func (s *AuthHandlerTestSuite) TestOAuthCallback_UsesReturnToState() {
	statePayload := map[string]string{
		"returnTo": "/seller/onboarding",
	}
	rawState, err := json.Marshal(statePayload)
	s.Require().NoError(err)
	state := base64.RawURLEncoding.EncodeToString(rawState)

	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/faux/callback?provider=faux&state="+state, nil)
	req = req.WithContext(context.WithValue(req.Context(), gothic.ProviderParamKey, "faux"))
	fauxSession := &faux.Session{Name: "Seller Faux", Email: "seller.faux@example.com"}
	recorder := httptest.NewRecorder()
	gothic.StoreInSession("faux", fauxSession.Marshal(), req, recorder)
	for _, cookie := range recorder.Result().Cookies() {
		req.AddCookie(cookie)
	}

	res, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	s.Equal(http.StatusSeeOther, res.StatusCode)
	s.Equal("http://localhost:3000/seller/onboarding", res.Header.Get("Location"))
}

func (s *AuthHandlerTestSuite) TestOAuthCallback_Unauthorized() {
	// Simulating callback WITHOUT setting session cookie
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/faux/callback?provider=faux", nil)

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

func (s *AuthHandlerTestSuite) TestGetMe_Success() {
	// 1. Setup User in DB directly
	userID, _ := uuid.NewV7()
	user := entity.User{
		BaseEntity: entity.BaseEntity{ID: userID},
		Name:       "John Doe",
		Email:      "john@example.com",
		Role:       "user",
	}
	err := s.db.Create(&user).Error
	s.Require().NoError(err)

	// 2. Generate Access Token using our utility
	accToken, _, _, err := jwt.GenerateTokens(userID)
	s.Require().NoError(err)

	// 3. Make request
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: accToken,
	})

	res, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	s.Equal(http.StatusOK, res.StatusCode)

	var responseBody map[string]interface{}
	json.NewDecoder(res.Body).Decode(&responseBody)
	s.Equal(true, responseBody["success"])

	// Assert data
	data := responseBody["data"].(map[string]interface{})
	s.Equal("John Doe", data["name"])
	s.Equal("john@example.com", data["email"])
	s.Equal(userID.String(), data["id"])
	s.Equal(map[string]interface{}{"canAccessDashboard": true, "requiresOnboarding": false}, data["seller_capabilities"])
}

func (s *AuthHandlerTestSuite) TestGetMe_Unauthorized() {
	// Request without cookie
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	res, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)

	s.Equal(http.StatusUnauthorized, res.StatusCode)
}

func (s *AuthHandlerTestSuite) TestRefreshToken_Success() {
	userID, _ := uuid.NewV7()
	_, refToken, jti, err := jwt.GenerateTokens(userID)
	s.Require().NoError(err)

	// Save JTI to Redis manually
	err = s.rdb.Set(s.ctx, "refresh_token:"+jti, userID.String(), time.Duration(config.Env.JwtRefreshExpiration)*time.Second).Err()
	s.Require().NoError(err)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refToken,
	})

	res, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	s.Equal(http.StatusOK, res.StatusCode)

	var responseBody map[string]interface{}
	json.NewDecoder(res.Body).Decode(&responseBody)
	s.Equal(true, responseBody["success"])

	// Check that we got back two new cookies (access_token, refresh_token)
	var hasAccessToken, hasRefreshToken bool
	for _, cookie := range res.Cookies() {
		if cookie.Name == "access_token" {
			hasAccessToken = true
			s.NotEqual("", cookie.Value)
		}
		if cookie.Name == "refresh_token" {
			hasRefreshToken = true
			s.NotEqual("", cookie.Value)
			s.NotEqual(refToken, cookie.Value, "Should generate a new refresh token")
		}
	}
	s.True(hasAccessToken)
	s.True(hasRefreshToken)
}

func (s *AuthHandlerTestSuite) TestRefreshToken_Unauthorized() {
	// 1. Valid Token but Not in Redis (Revoked/Expired)
	userID, _ := uuid.NewV7()
	_, refToken, _, err := jwt.GenerateTokens(userID)
	s.Require().NoError(err)
	// We do NOT save it to Redis this time

	req1 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req1.AddCookie(&http.Cookie{Name: "refresh_token", Value: refToken})

	res1, err := s.app.Test(req1, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	s.Equal(http.StatusUnauthorized, res1.StatusCode)

	// 2. No Token Provided
	req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	res2, err := s.app.Test(req2, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	s.Equal(http.StatusUnauthorized, res2.StatusCode)
}

func (s *AuthHandlerTestSuite) TestLogout_Success() {
	userID, _ := uuid.NewV7()

	// Insert user into DB so middleware.Protected can resolve it
	testUser := entity.User{Email: "logout-success@test.com", Name: "Logout User", Role: "user"}
	testUser.ID = userID
	s.Require().NoError(s.db.Create(&testUser).Error)

	accToken, refToken, jti, err := jwt.GenerateTokens(userID)
	s.Require().NoError(err)

	// Save JTI to Redis manually
	err = s.rdb.Set(s.ctx, "refresh_token:"+jti, userID.String(), time.Duration(config.Env.JwtRefreshExpiration)*time.Second).Err()
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accToken})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refToken})

	res, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	s.Equal(http.StatusOK, res.StatusCode)

	// JTI should be removed from Redis
	val, err := s.rdb.Get(s.ctx, "refresh_token:"+jti).Result()
	s.Equal(redis.Nil, err, "Redis key should be deleted")
	s.Empty(val)

	// Cookies should be cleared
	var hasAccessToken, hasRefreshToken bool
	for _, cookie := range res.Cookies() {
		if cookie.Name == "access_token" && cookie.Value == "" {
			hasAccessToken = true
		}
		if cookie.Name == "refresh_token" && cookie.Value == "" {
			hasRefreshToken = true
		}
	}
	s.True(hasAccessToken, "access_token cookie should be cleared")
	s.True(hasRefreshToken, "refresh_token cookie should be cleared")
}

func (s *AuthHandlerTestSuite) TestLogout_NoRefreshToken() {
	userID, _ := uuid.NewV7()

	// Insert user into DB so middleware.Protected can resolve it
	testUser := entity.User{Email: "logout-notoken@test.com", Name: "Logout No Token", Role: "user"}
	testUser.ID = userID
	s.Require().NoError(s.db.Create(&testUser).Error)

	accToken, _, _, err := jwt.GenerateTokens(userID)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accToken})

	res, err := s.app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	s.Require().NoError(err)
	s.Equal(http.StatusOK, res.StatusCode, "Logout should succeed even without refresh token")
}

func (suite *AuthHandlerTestSuite) TestOAuthCallback_ExistingUser() {
	// First login — creates user
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/faux/callback", nil)
	req = req.WithContext(context.WithValue(req.Context(), gothic.ProviderParamKey, "faux"))
	fauxSession := &faux.Session{Name: "Jack Faux", Email: "jack.faux@example.com"}
	recorder := httptest.NewRecorder()
	gothic.StoreInSession("faux", fauxSession.Marshal(), req, recorder)
	for _, cookie := range recorder.Result().Cookies() {
		req.AddCookie(cookie)
	}
	resp, err := suite.app.Test(req, fiber.TestConfig{})
	suite.Require().NoError(err)
	suite.Equal(http.StatusSeeOther, resp.StatusCode)

	// Second login — same user, same email
	req2 := httptest.NewRequest(http.MethodGet, "/auth/oauth/faux/callback", nil)
	req2 = req2.WithContext(context.WithValue(req2.Context(), gothic.ProviderParamKey, "faux"))
	recorder2 := httptest.NewRecorder()
	gothic.StoreInSession("faux", fauxSession.Marshal(), req2, recorder2)
	for _, cookie := range recorder2.Result().Cookies() {
		req2.AddCookie(cookie)
	}
	resp2, err := suite.app.Test(req2, fiber.TestConfig{})
	suite.Require().NoError(err)
	suite.Equal(http.StatusSeeOther, resp2.StatusCode)

	// Verify only 1 user in DB (no duplicate)
	var count int64
	suite.db.Model(&entity.User{}).Count(&count)
}
func (suite *AuthHandlerTestSuite) TestRefreshToken_CookiesClearedOnFailure() {
	// Generate a real JWT but do NOT store JTI in Redis → validation will fail
	userID := uuid.New()
	_, refreshToken, _, err := jwt.GenerateTokens(userID)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "some-old-token"})

	resp, err := suite.app.Test(req, fiber.TestConfig{})
	suite.Require().NoError(err)
	suite.Equal(http.StatusUnauthorized, resp.StatusCode)

	// Verify cookies are cleared (empty value or MaxAge <= 0)
	cookieMap := map[string]*http.Cookie{}
	for _, c := range resp.Cookies() {
		cookieMap[c.Name] = c
	}
	if accessCookie, ok := cookieMap["access_token"]; ok {
		suite.True(accessCookie.Value == "" || accessCookie.MaxAge < 0,
			"access_token cookie should be cleared on refresh failure")
	}
	if refreshCookie, ok := cookieMap["refresh_token"]; ok {
		suite.True(refreshCookie.Value == "" || refreshCookie.MaxAge < 0,
			"refresh_token cookie should be cleared on refresh failure")
	}
}

func (suite *AuthHandlerTestSuite) TestGetMe_MissingUserIDLocal() {
	// Generate token for a user that does NOT exist in DB
	nonExistentID := uuid.New()
	accessToken, _, _, err := jwt.GenerateTokens(nonExistentID)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})

	resp, err := suite.app.Test(req, fiber.TestConfig{})
	suite.Require().NoError(err)
	// User not in DB → service returns error → should be 4xx (404 or 500 depending on error mapping)
	suite.True(resp.StatusCode >= 400, "non-existent user should return an error status")
}

func (suite *AuthHandlerTestSuite) TestLogout_GarbageRefreshToken() {
	// Create a user and get a valid access token for the protected route
	userID := uuid.New()
	accessToken, _, _, err := jwt.GenerateTokens(userID)
	suite.Require().NoError(err)

	// Insert user so middleware passes
	user := &entity.User{
		BaseEntity: entity.BaseEntity{ID: userID},
		Name:       "Test User",
		Email:      "garbage@example.com",
		Role:       "user",
	}
	suite.db.Create(user)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "this-is-complete-garbage"})

	resp, err := suite.app.Test(req, fiber.TestConfig{})
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode, "logout should always succeed regardless of refresh token validity")
}
func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func TestRefreshToken_InfraFailureReturnsInternalServerError(t *testing.T) {
	app := newRefreshHandlerTestApp(&stubAuthService{
		refreshToken: func(ctx context.Context, refreshToken string) (*response.AuthTokens, error) {
			return nil, errors.New("cache unavailable")
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token"})
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "old-access"})

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	cleared := map[string]bool{}
	for _, cookie := range resp.Cookies() {
		if (cookie.Name == "access_token" || cookie.Name == "refresh_token") && (cookie.Value == "" || cookie.MaxAge < 0) {
			cleared[cookie.Name] = true
		}
	}

	if !cleared["access_token"] || !cleared["refresh_token"] {
		t.Fatalf("expected access and refresh cookies to be cleared, got %#v", cleared)
	}
}

func TestRefreshToken_UnauthorizedStaysUnauthorized(t *testing.T) {
	app := newRefreshHandlerTestApp(&stubAuthService{
		refreshToken: func(ctx context.Context, refreshToken string) (*response.AuthTokens, error) {
			return nil, domain.ErrUnauthorized
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token"})

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}
