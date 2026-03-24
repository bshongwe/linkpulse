package middleware

import (
	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a placeholder for auth middleware
// This will be implemented when JWT token handling is added
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement JWT token validation
		c.Next()
	}
}
