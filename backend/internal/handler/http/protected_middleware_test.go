package http_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/middleware"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/jwt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	pgDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestProtected_UsesCurrentDatabaseRole(t *testing.T) {
	ctx := context.Background()
	logger.Log = zap.NewNop()
	config.Env.AppEnv = "testing"
	setTestJWTConfig(t)

	pgContainer, err := postgres.Run(ctx,
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
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(pgDriver.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.User{}))
	require.NoError(t, setupAuthzTestState(db))

	authzService, err := newAuthzService(db)
	require.NoError(t, err)

	user := entity.User{Name: "Role Flipper", Email: "role@test.com", Role: pkgauthz.RoleUser}
	require.NoError(t, db.Create(&user).Error)

	accessToken, _, _, err := jwt.GenerateTokens(user.ID)
	require.NoError(t, err)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			if err == nil {
				return nil
			}

			if fe, ok := err.(*fiber.Error); ok {
				return c.Status(fe.Code).SendString(fe.Message)
			}

			if err == domain.ErrForbidden {
				return c.Status(fiber.StatusForbidden).SendString(err.Error())
			}

			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		},
	})
	app.Post(
		"/categories",
		middleware.Protected(db, authzService),
		middleware.RequirePermission(authzService, pkgauthz.ResourceCategory, pkgauthz.ActionCreate),
		func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusCreated) },
	)

	response := performProtectedRequest(t, app, accessToken)
	require.Equal(t, fiber.StatusForbidden, response.StatusCode)

	require.NoError(t, db.Model(&entity.User{}).Where("id = ?", user.ID).Update("role", pkgauthz.RoleAdmin).Error)

	response = performProtectedRequest(t, app, accessToken)
	require.Equal(t, fiber.StatusCreated, response.StatusCode)
}

func TestRequirePermission_DeniesWhenAuthzServiceUnavailable(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			if fe, ok := err.(*fiber.Error); ok {
				return c.Status(fe.Code).SendString(fe.Message)
			}

			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		},
	})

	app.Post(
		"/categories",
		middleware.RequirePermission(nil, pkgauthz.ResourceCategory, pkgauthz.ActionCreate),
		func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusCreated) },
	)

	req := httptest.NewRequest(http.MethodPost, "/categories", nil)
	response, err := app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	require.NoError(t, err)
	require.Equal(t, fiber.StatusServiceUnavailable, response.StatusCode)
}

func setTestJWTConfig(t *testing.T) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	config.Env.JwtPrivateKeyBase64 = base64.StdEncoding.EncodeToString(privPEM)

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	config.Env.JwtPublicKeyBase64 = base64.StdEncoding.EncodeToString(pubPEM)

	config.Env.JwtAccessExpiration = 900
	config.Env.JwtRefreshExpiration = 604800
	config.Env.OAuthTokenEncryptionKey = make([]byte, 32)
}

func performProtectedRequest(t *testing.T, app *fiber.App, accessToken string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/categories", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})

	response, err := app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	require.NoError(t, err)

	return response
}
