package handlers

import (
	"github.com/josephtesla/digital-wallet-go/internal/service"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func CreateTransferHandler(transferService service.TransferService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FromWalletID string `json:"from_wallet_id" binding:"required"`
			ToWalletID   string `json:"to_wallet_id" binding:"required"`
			Amount       int64  `json:"amount" binding:"required,min=1"`
			Description  string `json:"description"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			utils.BadRequestResponse(c, "Invalid request body", err.Error())
			return
		}

		// Convert kobo to decimal
		amount := decimal.NewFromInt(req.Amount).Div(decimal.NewFromInt(100))

		transaction, err := transferService.TransferFunds(
			c.Request.Context(),
			req.FromWalletID,
			req.ToWalletID,
			amount,
			req.Description,
		)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Transfer failed", err.Error())
			return
		}

		utils.SuccessResponse(c, "Transfer completed successfully", transaction)
	}
}

func GetTransferHistoryHandler(transferService service.TransferService) gin.HandlerFunc {
	return func(c *gin.Context) {
		walletID := c.Param("walletId")
		if walletID == "" {
			utils.BadRequestResponse(c, "Wallet ID is required", nil)
			return
		}

		history, err := transferService.GetTransferHistory(c.Request.Context(), walletID)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to get transfer history", err.Error())
			return
		}

		utils.SuccessResponse(c, "Transfer history retrieved successfully", history)
	}
}
