package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	ctxkey "github.com/bshongwe/linkpulse/backend/shared/context"
)

// TokenValidator is the subset of token validation the middleware needs.
// Defined here to avoid importing an internal service package from shared/.
// Any concrete type whose ValidateAccessToken returns (userID string, email string, err error) satisfies this.
type TokenValidator interface {
	ValidateAccessToken(token string) (userID string, email string, err error)
}

// AuthMiddleware validates the Bearer token on every request.
// Routes that don't require auth should be registered outside the protected group.
// Panics at startup if validator is nil to catch misconfiguration early.
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	if validator == nil {
		panic("AuthMiddleware: validator must not be nil")
	}
	return func(c *gin.Context) {
		authorization := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenStr := parts[1]
		userID, _, err := validator.ValidateAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(string(ctxkey.UserID), userID)
		c.Next()
	}
}
