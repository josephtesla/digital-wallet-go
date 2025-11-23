package main

import (
	"context"
	"log"

	"github.com/josephtesla/digital-wallet-go/db"
	"github.com/josephtesla/digital-wallet-go/internal/api"
	"github.com/josephtesla/digital-wallet-go/internal/infra"
	"github.com/josephtesla/digital-wallet-go/internal/repository"
	"github.com/josephtesla/digital-wallet-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	config := infra.LoadConfig()

	// Initialize logger
	logger := infra.InitLogger(config.LogLevel)
	defer logger.Sync()
	logger.Info("Starting Digital Wallet API server")

	appContext = context.Background()

	// Initialize database
	database, err := infra.InitDB(config.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Run migrations
	if err := db.Migrate(database, logger); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Seed database
	if err := db.Seed(database, logger); err != nil {
		logger.Fatal("Failed to seed database", zap.Error(err))
	}

	// Initialize Redis
	redisClient, err := infra.InitRedis(config.RedisURL)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}

	// Initialize Paystack client
	paystackClient := infra.InitPaystackClient(config.PaystackSecretKey)

	// Initialize repositories
	repos := repository.NewRepositories(database)

	// Initialize services
	services := service.NewServices(repos, redisClient, paystackClient, logger)

	// Initialize router
	router := gin.New()
	api.SetupRoutes(router, services, logger)

	// Start server
	port := config.Port
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting", zap.String("port", port))
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
