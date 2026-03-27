package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/naufalilyasa/pal-property-backend/internal/repository/postgres"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
	"go.uber.org/zap"
	pgDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	logger.InitLogger()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Env.DBHost,
		config.Env.DBUser,
		config.Env.DBPassword,
		config.Env.DBName,
		config.Env.DBPort,
		config.Env.DBSSLMode,
	)
	db, err := gorm.Open(pgDriver.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect database", zap.Error(err))
	}
	searchClient, err := searchindex.NewClient(config.Env.ElasticAddress, config.Env.ElasticUsername, config.Env.ElasticPassword, nil)
	if err != nil {
		logger.Log.Fatal("Failed to initialize search client", zap.Error(err))
	}
	projector, err := service.NewElasticsearchSearchProjector(config.Env.ElasticListingsIndex, searchClient, postgres.NewListingRepository(db))
	if err != nil {
		logger.Log.Fatal("Failed to initialize search projector", zap.Error(err))
	}
	jobs := postgres.NewSearchIndexJobRepository(db)
	processor, err := service.NewIndexingJobProcessor(jobs, projector)
	if err != nil {
		logger.Log.Fatal("Failed to initialize indexing job processor", zap.Error(err))
	}
	if err := searchClient.EnsureIndex(context.Background(), config.Env.ElasticListingsIndex, service.ListingIndexMapping()); err != nil {
		logger.Log.Fatal("Failed to ensure search index", zap.Error(err))
	}
	if len(os.Args) > 1 && os.Args[1] == "rebuild" {
		logger.Log.Info("Listing index rebuild starting")
		if err := service.RebuildListingIndex(context.Background(), postgres.NewListingRepository(db), searchClient, config.Env.ElasticListingsIndex, 200); err != nil {
			logger.Log.Fatal("Listing index rebuild failed", zap.Error(err))
		}
		logger.Log.Info("Listing index rebuild complete")
		return
	}

	logger.Log.Info("Listing indexer worker starting")
	if err := processor.Run(context.Background(), 3*time.Second, 100); err != nil {
		logger.Log.Fatal("Listing indexer worker stopped", zap.Error(err))
	}
}
