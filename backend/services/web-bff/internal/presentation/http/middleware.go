package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthMiddleware extracts JWT claims from the request header
func AuthMiddleware(jwtSecret string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow CORS preflight requests to pass through without authentication
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Extract raw token string
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "missing authorization header",
				"status": http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "invalid authorization header format",
				"status": http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		token, err := validateToken(tokenString, jwtSecret, logger)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "invalid or expired token",
				"status": http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "invalid token claims",
				"status": http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		if !setContextFromClaims(c, claims, tokenString) {
			c.Abort()
			return
		}

		c.Next()
	}
}

func validateToken(tokenString string, jwtSecret string, logger *zap.Logger) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		logger.Warn("invalid token", zap.Error(err))
		return token, err
	}

	return token, nil
}

func setContextFromClaims(c *gin.Context, claims jwt.MapClaims, tokenString string) bool {
	// Check token expiration (exp claim)
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "token expired",
				"status": http.StatusUnauthorized,
			})
			return false
		}
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  "missing user_id in token",
			"status": http.StatusUnauthorized,
		})
		return false
	}

	workspaceID, ok := claims["workspace_id"].(string)
	if !ok || workspaceID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  "missing workspace_id in token",
			"status": http.StatusUnauthorized,
		})
		return false
	}

	c.Set("user_id", userID)
	c.Set("workspace_id", workspaceID)
	c.Set("jwt_token", tokenString)
	return true
}

