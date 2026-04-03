package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthMiddleware extracts JWT claims from the request header
func AuthMiddleware(jwtSecret string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractAndValidateToken(c, jwtSecret, logger)
		if err != nil || !token.Valid {
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

		if !setContextFromClaims(c, claims) {
			c.Abort()
			return
		}

		c.Next()
	}
}

func extractAndValidateToken(c *gin.Context, jwtSecret string, logger *zap.Logger) (*jwt.Token, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  "missing authorization header",
			"status": http.StatusUnauthorized,
		})
		return nil, fmt.Errorf("missing auth header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  "invalid authorization header format",
			"status": http.StatusUnauthorized,
		})
		return nil, fmt.Errorf("invalid auth format")
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		logger.Warn("invalid token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  "invalid or expired token",
			"status": http.StatusUnauthorized,
		})
		return token, err
	}

	return token, nil
}

func setContextFromClaims(c *gin.Context, claims jwt.MapClaims) bool {
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
	return true
}

