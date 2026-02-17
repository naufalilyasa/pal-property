package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
	handler "github.com/username/pal-property-backend/internal/handler/http"
	"github.com/username/pal-property-backend/internal/repository/postgres"
	"github.com/username/pal-property-backend/internal/service"
	pgDriver "gorm.io/driver/postgres" // renamed to avoid conflict
	gormPkg "gorm.io/gorm"
)

func main() {
	// 1. Setup Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5433/pal_db?sslmode=disable"
	}

	db, err := gormPkg.Open(pgDriver.Open(dbURL), &gormPkg.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto Migrate (Optional, but good for dev to ensure consistency if not strictly migration-based)
	// For production, rely on migrate tool.
	// We can skip this as we did migration manually.

	// 2. Setup Goth Providers
	// In production, use os.Getenv for these values
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_KEY"),
			os.Getenv("GOOGLE_SECRET"),
			"http://localhost:8080/auth/google/callback",
		),
		facebook.New(
			os.Getenv("FACEBOOK_KEY"),
			os.Getenv("FACEBOOK_SECRET"),
			"http://localhost:8080/auth/facebook/callback",
		),
	)

	// 3. Initialize layers
	authRepo := postgres.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService)

	// 4. Setup Router
	r := gin.Default()

	// Auth Routes
	r.GET("/auth/:provider", authHandler.BeginAuth)
	r.GET("/auth/:provider/callback", authHandler.Callback)

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 5. Run Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
