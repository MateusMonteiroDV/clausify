package main

import (
	"log"

	"github.com/clausify/backend/internal/config"
	"github.com/clausify/backend/internal/database"
	"github.com/clausify/backend/internal/routes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	// Setup Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := routes.Setup(db, cfg, logger)

	logger.Info("Clausify API starting", zap.String("port", cfg.Port))
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
