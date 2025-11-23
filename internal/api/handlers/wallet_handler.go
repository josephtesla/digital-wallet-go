package handlers

import (
	"github.com/josephtesla/digital-wallet-go/internal/service"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/gin-gonic/gin"
)

func CreateWalletHandler(walletService service.WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID   string `json:"user_id" binding:"required"`
			Currency string `json:"currency" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			utils.BadRequestResponse(c, "Invalid request body", err.Error())
			return
		}

		wallet, err := walletService.CreateWallet(c.Request.Context(), req.UserID, req.Currency)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to create wallet", err.Error())
			return
		}

		utils.CreatedResponse(c, "Wallet created successfully", wallet)
	}
}

func GetWalletHandler(walletService service.WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		walletID := c.Param("id")
		if walletID == "" {
			utils.BadRequestResponse(c, "Wallet ID is required", nil)
			return
		}

		wallet, err := walletService.GetWallet(c.Request.Context(), walletID)
		if err != nil {
			utils.NotFoundResponse(c, "Wallet not found")
			return
		}

		utils.SuccessResponse(c, "Wallet retrieved successfully", wallet)
	}
}

func GetWalletBalanceHandler(walletService service.WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		walletID := c.Param("id")
		if walletID == "" {
			utils.BadRequestResponse(c, "Wallet ID is required", nil)
			return
		}

		balance, err := walletService.GetBalance(c.Request.Context(), walletID)
		if err != nil {
			utils.NotFoundResponse(c, "Wallet not found")
			return
		}

		utils.SuccessResponse(c, "Balance retrieved successfully", gin.H{
			"balance":  balance.ToKobo(),
			"currency": balance.Currency,
		})
	}
}
