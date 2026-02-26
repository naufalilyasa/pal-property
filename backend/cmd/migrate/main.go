package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	config.LoadConfig()
	logger.InitLogger()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.Env.DBUser, config.Env.DBPassword, config.Env.DBHost,
		config.Env.DBPort, config.Env.DBName, config.Env.DBSSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Log.Fatal("Could not connect to database", zap.Error(err))
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		logger.Log.Fatal("Could not ping database", zap.Error(err))
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Log.Fatal("Could not create migrate driver", zap.Error(err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		logger.Log.Fatal("Could not create migrate instance", zap.Error(err))
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	if cmd == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			logger.Log.Fatal("Migration down failed", zap.Error(err))
		}
		logger.Log.Info("Migration down successful")
	} else {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Log.Fatal("Migration up failed", zap.Error(err))
		}
		logger.Log.Info("Migration up successful")
	}
}
