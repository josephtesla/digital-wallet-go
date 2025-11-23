package middleware

import (
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			utils.UnauthorizedResponse(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := authHeader[7:]
		if token == "" {
			utils.UnauthorizedResponse(c, "Token is required")
			c.Abort()
			return
		}

		// TODO: Implement JWT token validation
		// For now, we'll just log the token and continue
		logger.Info("Auth token received", zap.String("token", token[:10]+"..."))

		c.Next()
	}
}
