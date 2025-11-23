package api

import (
	"github.com/josephtesla/digital-wallet-go/internal/api/handlers"
	"github.com/josephtesla/digital-wallet-go/internal/api/middleware"
	"github.com/josephtesla/digital-wallet-go/internal/service"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRoutes(router *gin.Engine, services *service.Services, logger *zap.Logger) {

	// Add middleware
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.IdempotencyMiddleware(services.Idempotency, logger))

	// API routes
	api := router.Group("/api/v1")
	{
		// Wallet routes
		wallets := api.Group("/wallets")
		{
			wallets.GET("/:id", handlers.GetWalletHandler(services.Wallet))
			wallets.GET("/:id/balance", handlers.GetWalletBalanceHandler(services.Wallet))
			wallets.POST("/", handlers.CreateWalletHandler(services.Wallet))
		}

		// Transfer routes
		transfers := api.Group("/transfers")
		{
			transfers.POST("/", handlers.CreateTransferHandler(services.Transfer))
			transfers.GET("/:walletId/history", handlers.GetTransferHistoryHandler(services.Transfer))
		}

		// Deposit routes
		deposits := api.Group("/deposits")
		{
			deposits.POST("/init", handlers.InitializeDepositHandler(services.Deposit))
			deposits.GET("/verify/:reference", handlers.VerifyDepositHandler(services.Deposit))
		}

		// Withdrawal routes
		withdrawals := api.Group("/withdrawals")
		{
			withdrawals.POST("/init", handlers.InitializeWithdrawalHandler(services.Withdrawal))
			withdrawals.GET("/:walletId/history", handlers.GetWithdrawalHistoryHandler(services.Withdrawal))
		}

		// Webhook routes
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/paystack", handlers.PaystackWebhookHandler(services.Deposit, services.Withdrawal))
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, "Service is healthy", gin.H{"status": "ok"})
	})
}
