package handlers

import (
	"github.com/josephtesla/digital-wallet-go/internal/service"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/gin-gonic/gin"
)

func PaystackWebhookHandler(depositService service.DepositService, withdrawalService service.WithdrawalService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var webhookData map[string]interface{}
		if err := c.ShouldBindJSON(&webhookData); err != nil {
			utils.BadRequestResponse(c, "Invalid webhook data", err.Error())
			return
		}

		// Determine webhook type based on the data
		event, ok := webhookData["event"].(string)
		if !ok {
			utils.BadRequestResponse(c, "Missing event type in webhook", nil)
			return
		}

		switch event {
		case "charge.success":
			// Handle deposit webhook
			if err := depositService.ProcessWebhook(c.Request.Context(), webhookData); err != nil {
				utils.InternalServerErrorResponse(c, "Failed to process deposit webhook", err.Error())
				return
			}

		case "transfer.success", "transfer.failed":
			// Handle withdrawal webhook
			if err := withdrawalService.ProcessWithdrawalWebhook(c.Request.Context(), webhookData); err != nil {
				utils.InternalServerErrorResponse(c, "Failed to process withdrawal webhook", err.Error())
				return
			}

		default:
			utils.SuccessResponse(c, "Webhook received but not processed", gin.H{"event": event})
			return
		}

		utils.SuccessResponse(c, "Webhook processed successfully", gin.H{"event": event})
	}
}
