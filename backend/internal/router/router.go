package router

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/google/uuid"
	"github.com/username/pal-property-backend/internal/handler/http"
	"github.com/username/pal-property-backend/pkg/config"
	"github.com/username/pal-property-backend/pkg/logger"
	"go.uber.org/zap"
)

func ZapLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		requestID := c.Locals("requestid")
		var traceID string
		if requestID != nil {
			if idStr, ok := requestID.(string); ok {
				traceID = idStr
			}
		}

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

func Register(app *fiber.App, authHandler *http.AuthHandler) {
	// Global Middlewares
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{config.Env.CorsAllowedOrigins},
		AllowHeaders: []string{"Origin, Content-Type, Accept, Authorization"},
	}))

	app.Use(limiter.New(limiter.Config{
		Next: func(c fiber.Ctx) bool {
			env := config.Env.AppEnv
			return env == "testing" || env == "development"
		},
		Max:        config.Env.RateLimitMax,
		Expiration: config.Env.RateLimitExp,
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

	// Auth Routes
	auth := app.Group("/auth")
	auth.Get("/:provider", authHandler.BeginAuth)
	auth.Get("/:provider/callback", authHandler.Callback)
}
