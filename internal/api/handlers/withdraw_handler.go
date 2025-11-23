package handlers

import (
	"github.com/josephtesla/digital-wallet-go/internal/service"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func InitializeWithdrawalHandler(withdrawalService service.WithdrawalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID        string `json:"user_id" binding:"required"`
			WalletID      string `json:"wallet_id" binding:"required"`
			BankAccountID string `json:"bank_account_id" binding:"required"`
			Amount        int64  `json:"amount" binding:"required,min=1"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			utils.BadRequestResponse(c, "Invalid request body", err.Error())
			return
		}

		// Convert kobo to decimal
		amount := decimal.NewFromInt(req.Amount).Div(decimal.NewFromInt(100))

		response, err := withdrawalService.InitializeWithdrawal(
			c.Request.Context(),
			req.UserID,
			req.WalletID,
			req.BankAccountID,
			amount,
		)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to initialize withdrawal", err.Error())
			return
		}

		utils.SuccessResponse(c, "Withdrawal initialized successfully", response)
	}
}

func GetWithdrawalHistoryHandler(withdrawalService service.WithdrawalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		walletID := c.Param("walletId")
		if walletID == "" {
			utils.BadRequestResponse(c, "Wallet ID is required", nil)
			return
		}

		history, err := withdrawalService.GetWithdrawalHistory(c.Request.Context(), walletID)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to get withdrawal history", err.Error())
			return
		}

		utils.SuccessResponse(c, "Withdrawal history retrieved successfully", history)
	}
}
