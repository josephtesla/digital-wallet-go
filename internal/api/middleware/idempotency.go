package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/josephtesla/digital-wallet-go/internal/service"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func IdempotencyMiddleware(idempotencyService service.IdempotencyService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip idempotency check for GET requests
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// Get idempotency key from header
		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			c.Next()
			return
		}

		// Read request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("Failed to read request body", zap.Error(err))
			utils.InternalServerErrorResponse(c, "Failed to read request body", err.Error())
			c.Abort()
			return
		}

		// Restore request body
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		// Check if key exists
		existingKey, err := idempotencyService.CheckKey(c.Request.Context(), idempotencyKey)
		if err != nil {
			logger.Error("Failed to check idempotency key", zap.Error(err))
			utils.InternalServerErrorResponse(c, "Failed to check idempotency key", err.Error())
			c.Abort()
			return
		}

		if existingKey != nil {
			// Key exists, return cached response
			if existingKey.Response != nil {
				c.Header("Content-Type", "application/json")
				c.String(http.StatusOK, string(existingKey.Response))
				c.Abort()
				return
			}
		}

		// Generate request hash
		requestHash := idempotencyService.GenerateRequestHash(
			c.Request.Method,
			c.Request.URL.Path,
			string(body),
		)

		// Store the key for later use
		c.Set("idempotency_key", idempotencyKey)
		c.Set("request_hash", requestHash)

		c.Next()
	}
}
