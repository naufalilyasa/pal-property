package main

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	handler "github.com/username/pal-property-backend/internal/handler/http"
	"github.com/username/pal-property-backend/internal/repository/postgres"
	"github.com/username/pal-property-backend/internal/router"
	"github.com/username/pal-property-backend/internal/service"
	"github.com/username/pal-property-backend/pkg/config"
	"github.com/username/pal-property-backend/pkg/logger"
	"go.uber.org/zap"
	pgDriver "gorm.io/driver/postgres"
	gormPkg "gorm.io/gorm"
)

func main() {
	config.LoadConfig()
	logger.InitLogger()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.Env.DBHost, config.Env.DBUser, config.Env.DBPassword,
		config.Env.DBName, config.Env.DBPort, config.Env.DBSSLMode,
	)

	db, err := gormPkg.Open(pgDriver.Open(dsn), &gormPkg.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}

	goth.UseProviders(
		google.New(
			config.Env.ClientID,
			config.Env.ClientSecret,
			config.Env.CallbackURL,
			"email", "profile",
		),
	)

	authRepo := postgres.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService)

	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			logger.Log.Error("Fiber trapped error", zap.Error(err), zap.String("path", c.Path()))

			msg := "An unexpected error occurred"
			if code < 500 {
				msg = err.Error()
			}

			return c.Status(code).JSON(fiber.Map{
				"success":  false,
				"message":  msg,
				"data":     nil,
				"trace_id": c.Locals("requestid"),
			})
		},
	})

	router.Register(app, authHandler)

	logger.Log.Info("Server starting", zap.Int("port", config.Env.Port))
	if err := app.Listen(fmt.Sprintf(":%d", config.Env.Port)); err != nil {
		logger.Log.Fatal("Server failed to start", zap.Error(err))
	}
}
