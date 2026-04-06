package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// extractBearerToken pulls the token string from the Authorization header.
// Returns empty string if the header is missing or malformed.
func extractBearerToken(c *gin.Context) string {
	parts := strings.Fields(strings.TrimSpace(c.GetHeader("Authorization")))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// parseToken validates the token signature and returns the parsed claims.
func parseToken(tokenString, secretKey string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// setContextClaims stores identity values from the token claims into the Gin context.
func setContextClaims(c *gin.Context, tokenString string, claims jwt.MapClaims) {
	c.Set("jwt_token", tokenString)

	userID, _ := claims["user_id"].(string)
	c.Set("user_id", userID)

	// Fall back to user_id as workspace scope until workspaces are fully implemented
	wsID, _ := claims["workspace_id"].(string)
	if wsID == "" {
		wsID = userID
	}
	c.Set("workspace_id", wsID)
	c.Set("claims", claims)
}

// JWTMiddleware validates Bearer tokens and populates the request context with identity claims.
func JWTMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		claims, err := parseToken(tokenString, secretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		setContextClaims(c, tokenString, claims)
		c.Next()
	}
}
