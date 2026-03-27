package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	ctxkey "github.com/bshongwe/linkpulse/backend/shared/context"
)

// TokenValidator is the subset of token validation the middleware needs.
// Defined here to avoid importing an internal service package from shared/.
// Any concrete type whose ValidateAccessToken returns (userID string, err error) satisfies this.
type TokenValidator interface {
	ValidateAccessToken(token string) (userID string, email string, err error)
}

// AuthMiddleware validates the Bearer token on every request.
// Routes that don't require auth should be registered outside the protected group.
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authorization, "Bearer ")
		userID, _, err := validator.ValidateAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(string(ctxkey.UserID), userID)
		c.Next()
	}
}
