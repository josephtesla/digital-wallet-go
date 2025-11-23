package handlers

import (
	"github.com/josephtesla/digital-wallet-go/internal/service"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func InitializeDepositHandler(depositService service.DepositService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID   string `json:"user_id" binding:"required"`
			WalletID string `json:"wallet_id" binding:"required"`
			Amount   int64  `json:"amount" binding:"required,min=1"`
			Email    string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			utils.BadRequestResponse(c, "Invalid request body", err.Error())
			return
		}

		// Convert kobo to decimal
		amount := decimal.NewFromInt(req.Amount).Div(decimal.NewFromInt(100))

		response, err := depositService.InitializeDeposit(
			c.Request.Context(),
			req.UserID,
			req.WalletID,
			amount,
			req.Email,
		)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to initialize deposit", err.Error())
			return
		}

		utils.SuccessResponse(c, "Deposit initialized successfully", response)
	}
}

func VerifyDepositHandler(depositService service.DepositService) gin.HandlerFunc {
	return func(c *gin.Context) {
		reference := c.Param("reference")
		if reference == "" {
			utils.BadRequestResponse(c, "Reference is required", nil)
			return
		}

		payment, err := depositService.VerifyDeposit(c.Request.Context(), reference)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to verify deposit", err.Error())
			return
		}

		utils.SuccessResponse(c, "Deposit verification completed", payment)
	}
}
