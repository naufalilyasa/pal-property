package main

import (
	"fmt"
	"log"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	handler "github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	"github.com/naufalilyasa/pal-property-backend/internal/repository/postgres"
	"github.com/naufalilyasa/pal-property-backend/internal/repository/redis"
	"github.com/naufalilyasa/pal-property-backend/internal/router"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	goRedis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	pgDriver "gorm.io/driver/postgres"
	gormPkg "gorm.io/gorm"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	logger.InitLogger()

	loggerDev, _ := zap.NewDevelopment()
	defer loggerDev.Sync()
	sugar := loggerDev.Sugar()

	sugar.Infow("Pemeriksaan Konfigurasi Environment",
		"jwtprivatekeybase64", config.Env.JwtPrivateKeyBase64,
	)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Env.DBHost, config.Env.DBUser, config.Env.DBPassword,
		config.Env.DBName, config.Env.DBPort, config.Env.DBSSLMode,
	)

	db, err := gormPkg.Open(pgDriver.Open(dsn), &gormPkg.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}

	goth.UseProviders(
		google.New(
			config.Env.OAuthClientID,
			config.Env.OAuthClientSecret,
			config.Env.OAuthCallbackURL,
			"email", "profile",
		),
	)

	rdb := goRedis.NewClient(&goRedis.Options{
		Addr:     config.Env.RedisAddr,
		Password: config.Env.RedisPassword,
		DB:       config.Env.RedisDB,
	})
	cacheRepo := redis.NewCacheRepository(rdb)

	authRepo := postgres.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo, cacheRepo)
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
				"trace_id": requestid.FromContext(c),
			})
		},
	})

	router.Register(app, authHandler)

	logger.Log.Info("Server starting", zap.String("port", config.Env.Port))
	if err := app.Listen(fmt.Sprintf(":%s", config.Env.Port)); err != nil {
		logger.Log.Fatal("Server failed to start", zap.Error(err))
	}
}
