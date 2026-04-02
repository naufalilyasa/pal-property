package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	handler "github.com/naufalilyasa/pal-property-backend/internal/handler/http"
	"github.com/naufalilyasa/pal-property-backend/internal/repository/postgres"
	"github.com/naufalilyasa/pal-property-backend/internal/repository/redis"
	searchrepo "github.com/naufalilyasa/pal-property-backend/internal/repository/search"
	"github.com/naufalilyasa/pal-property-backend/internal/router"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/naufalilyasa/pal-property-backend/pkg/cloudinary"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/gemini"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
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
	if err := postgres.EnsureIndonesiaRegionsSeeded(context.Background(), db, config.Env.WilayahDataPath); err != nil {
		logger.Log.Fatal("Failed to seed wilayah data", zap.Error(err))
	}

	goth.UseProviders(
		google.New(
			config.Env.OAuthClientID,
			config.Env.OAuthClientSecret,
			config.Env.OAuthCallbackURL,
			"email", "profile",
		),
	)

	sessionStore := sessions.NewCookieStore([]byte(config.Env.SessionSecret))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Env.AppEnv == "production",
		SameSite: http.SameSiteLaxMode,
		Domain:   strings.TrimSpace(config.Env.AuthCookieDomain),
	}
	gothic.Store = sessionStore

	rdb := goRedis.NewClient(&goRedis.Options{
		Addr:     config.Env.RedisAddr,
		Password: config.Env.RedisPassword,
		DB:       config.Env.RedisDB,
	})
	cacheRepo := redis.NewCacheRepository(rdb)

	var listingImageStorage domain.ListingImageStorage
	var listingVideoStorage domain.ListingVideoStorage
	if config.Env.CloudinaryEnabled {
		listingImageStorage, err = cloudinary.New(cloudinary.Config{
			CloudName: config.Env.CloudinaryCloudName,
			APIKey:    config.Env.CloudinaryAPIKey,
			APISecret: config.Env.CloudinaryAPISecret,
		})
		if err != nil {
			logger.Log.Fatal("Failed to initialize Cloudinary storage", zap.Error(err))
		}
		if vs, ok := listingImageStorage.(domain.ListingVideoStorage); ok {
			listingVideoStorage = vs
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
	indexJobRepo := postgres.NewSearchIndexJobRepository(db)
	indexTxManager := postgres.NewSearchIndexTransactionManager(db)
	listingAuthzService := service.NewAuthzService(authzService)
	mediaStorages := []domain.ListingImageStorage{}
	if listingImageStorage != nil {
		mediaStorages = append(mediaStorages, listingImageStorage)
	} else if listingVideoStorage != nil {
		if storageAsImage, ok := listingVideoStorage.(domain.ListingImageStorage); ok {
			mediaStorages = append(mediaStorages, storageAsImage)
		}
	}
	listingService := service.NewListingServiceWithAuthzJobsAndTransactions(listingRepo, listingAuthzService, indexJobRepo, indexTxManager, mediaStorages...)
	listingHandler := handler.NewListingHandler(listingService)
	savedListingRepo := postgres.NewSavedListingRepository(db)
	savedListingService := service.NewSavedListingService(savedListingRepo, listingRepo)
	savedListingHandler := handler.NewSavedListingHandler(savedListingService)
	searchClient, err := searchindex.NewClient(config.Env.ElasticAddress, config.Env.ElasticUsername, config.Env.ElasticPassword, nil)
	if err != nil {
		logger.Log.Fatal("Failed to initialize search client", zap.Error(err))
	}
	searchService, err := service.NewSearchReadService(config.Env.ElasticListingsIndex, searchClient)
	if err != nil {
		logger.Log.Fatal("Failed to initialize search service", zap.Error(err))
	}
	searchHandler := handler.NewSearchHandler(searchService)
	chatRetrievalRepo, err := searchrepo.NewChatRetrievalRepository(config.Env.ElasticChatRetrievalIndex, searchClient)
	if err != nil {
		logger.Log.Fatal("Failed to initialize chat retrieval repository", zap.Error(err))
	}
	geminiClient, err := gemini.NewClientFromConfig(context.Background())
	if err != nil {
		logger.Log.Fatal("Failed to initialize Gemini client", zap.Error(err))
	}
	chatRetrievalService, err := service.NewChatRetrievalService(chatRetrievalRepo, geminiClient, config.Env.ChatMaxRetrievalDocs)
	if err != nil {
		logger.Log.Fatal("Failed to initialize chat retrieval service", zap.Error(err))
	}
	chatMemoryRepo := redis.NewChatMemoryRepository(rdb, time.Duration(config.Env.ChatSessionTTLSeconds)*time.Second, config.Env.ChatMaxHistoryTurns)
	chatService, err := service.NewChatService(chatRetrievalService, chatMemoryRepo, geminiClient, config.Env.ChatMaxHistoryTurns)
	if err != nil {
		logger.Log.Fatal("Failed to initialize chat service", zap.Error(err))
	}
	chatHandler := handler.NewChatHandler(chatService)
	regionRepo := postgres.NewRegionRepository(db)
	regionService := service.NewRegionService(regionRepo)
	listingService = service.WithRegionLookupService(listingService, regionService)
	regionHandler := handler.NewRegionHandler(regionService)

	categoryRepo := postgres.NewCategoryRepository(db)
	categoryService := service.NewCategoryServiceWithJobsAndTransactions(categoryRepo, indexJobRepo, indexTxManager)
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
			} else if errors.Is(err, domain.ErrInvalidVideoFile) || errors.Is(err, domain.ErrVideoTooLarge) || errors.Is(err, domain.ErrVideoTooLong) {
				code = fiber.StatusBadRequest
			} else if errors.Is(err, domain.ErrVideoAlreadyExists) {
				code = fiber.StatusConflict
			} else if errors.Is(err, domain.ErrVideoStorageUnset) {
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

	router.Register(app, db, authzService, authHandler, listingHandler, savedListingHandler, searchHandler, chatHandler, regionHandler, categoryHandler)

	logger.Log.Info("Server starting", zap.String("port", config.Env.Port))
	if err := app.Listen(fmt.Sprintf(":%s", config.Env.Port)); err != nil {
		logger.Log.Fatal("Server failed to start", zap.Error(err))
	}
}
