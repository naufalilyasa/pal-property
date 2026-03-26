package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	handler "github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	"github.com/naufalilyasa/pal-property-backend/internal/repository/postgres"
	"github.com/naufalilyasa/pal-property-backend/internal/repository/redis"
	"github.com/naufalilyasa/pal-property-backend/internal/router"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/naufalilyasa/pal-property-backend/pkg/cloudinary"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	pkgeventing "github.com/naufalilyasa/pal-property-backend/pkg/eventing"
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
	eventPublisher, err := pkgeventing.NewKafkaPublisher(config.Env.KafkaBrokers, config.Env.KafkaClientID, config.Env.KafkaTopicListingEvents, config.Env.KafkaTopicCategoryEvents)
	if err != nil {
		logger.Log.Fatal("Failed to initialize event publisher", zap.Error(err))
	}
	defer func() { _ = eventPublisher.Close() }()

	var listingImageStorage domain.ListingImageStorage
	if config.Env.CloudinaryEnabled {
		listingImageStorage, err = cloudinary.New(cloudinary.Config{
			CloudName: config.Env.CloudinaryCloudName,
			APIKey:    config.Env.CloudinaryAPIKey,
			APISecret: config.Env.CloudinaryAPISecret,
		})
		if err != nil {
			logger.Log.Fatal("Failed to initialize Cloudinary storage", zap.Error(err))
		}
	}

	authRepo := postgres.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo, cacheRepo)
	authHandler := handler.NewAuthHandler(authService)

	authzService, err := authz.NewService(db)
	if err != nil {
		logger.Log.Fatal("Failed to initialize authz service", zap.Error(err))
	}

	listingRepo := postgres.NewListingRepository(db)
	listingAuthzService := service.NewAuthzService(authzService)
	listingService := service.NewListingServiceWithAuthzAndPublisher(listingRepo, listingAuthzService, eventPublisher, listingImageStorage)
	listingHandler := handler.NewListingHandler(listingService)

	categoryRepo := postgres.NewCategoryRepository(db)
	categoryService := service.NewCategoryServiceWithPublisher(categoryRepo, eventPublisher)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
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

	router.Register(app, db, authzService, authHandler, listingHandler, categoryHandler)

	logger.Log.Info("Server starting", zap.String("port", config.Env.Port))
	if err := app.Listen(fmt.Sprintf(":%s", config.Env.Port)); err != nil {
		logger.Log.Fatal("Server failed to start", zap.Error(err))
	}
}
