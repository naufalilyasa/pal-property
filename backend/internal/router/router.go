package router

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/middleware"
	"go.uber.org/zap"
)

func ZapLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		traceID := requestid.FromContext(c)

		logger.Log.Info("HTTP Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", duration),
			zap.String("ip", c.IP()),
			zap.String("trace_id", traceID),
		)
		return err
	}
}

func Register(app *fiber.App, authHandler *http.AuthHandler, listingHandler *http.ListingHandler) {
	// Global Middlewares
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.Env.CorsAllowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.Use(limiter.New(limiter.Config{
		Next: func(c fiber.Ctx) bool {
			env := config.Env.AppEnv
			return env == "testing" || env == "development"
		},
		Max:        config.Env.RateLimitMax,
		Expiration: time.Duration(config.Env.RateLimitExp) * time.Second,
	}))

	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.New().String()
		},
	}))

	app.Use(ZapLogger())

	// Health Check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"status": "ok"})
	})

	// Buat SATU base group untuk auth
	authGroup := app.Group("/auth")

	// ==========================================
	// 1. Auth Routes (Public)
	// ==========================================
	// Tambahkan prefix spesifik (misal /oauth) untuk provider
	// agar terhindar dari tabrakan dengan route "/me" atau route statis lain ke depannya.
	authGroup.Get("/oauth/:provider", authHandler.BeginAuth)
	authGroup.Get("/oauth/:provider/callback", authHandler.Callback)
	authGroup.Post("/refresh", authHandler.RefreshToken)

	// ==========================================
	// 2. Auth Routes (Protected)
	// ==========================================
	// Buat sub-group dari authGroup, dan pasang middleware Protected() di sini.
	// Semua route di bawah apiProtected otomatis membutuhkan autentikasi.
	apiProtected := authGroup.Group("/", middleware.Protected())
	apiProtected.Get("/me", authHandler.GetMe)
	apiProtected.Post("/logout", authHandler.Logout)
	// ==========================================
	// 3. Listing Routes (Public)
	// ==========================================
	api := app.Group("/api")
	api.Get("/listings", listingHandler.List)
	api.Get("/listings/slug/:slug", listingHandler.GetBySlug)
	api.Get("/listings/:id", listingHandler.GetByID)

	// ==========================================
	// 4. Listing Routes (Protected)
	// ==========================================
	listingProtected := api.Group("/listings", middleware.Protected())
	listingProtected.Post("/", listingHandler.Create)
	listingProtected.Put("/:id", listingHandler.Update)
	listingProtected.Delete("/:id", listingHandler.Delete)

	apiProtected.Get("/me/listings", listingHandler.ListByUserID)
}
