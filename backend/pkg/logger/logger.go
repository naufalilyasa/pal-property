package logger

import (
	"log"
	"os"

	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"go.uber.org/zap"
)

var Log *zap.Logger

func InitLogger() {
	var err error

	if config.Env.AppEnv == "testing" {
		Log = zap.NewNop()
		return
	}

	if config.Env.AppEnv == "development" {
		// Log to stdout AND tmp/logs/app.log
		if err := os.MkdirAll("tmp/logs", 0755); err != nil {
			log.Fatalf("failed to create log directory: %v", err)
		}

		configLog := zap.NewDevelopmentConfig()
		configLog.OutputPaths = []string{"stdout", "tmp/logs/app.log"}
		Log, err = configLog.Build()
	} else {
		// Production logic -> stdout only
		configLog := zap.NewProductionConfig()
		Log, err = configLog.Build()
	}

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	zap.ReplaceGlobals(Log)
}
