package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/bshongwe/linkpulse/backend/shared/middleware"
)

const testSecret = "test-secret-key"

func makeToken(t *testing.T, secret string, userID, email string, expiry time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(expiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestJWTValidator_ValidateAccessToken(t *testing.T) {
	validator := middleware.NewJWTValidator(testSecret)
	userID := uuid.New().String()

	t.Run("valid token", func(t *testing.T) {
		tok := makeToken(t, testSecret, userID, "user@example.com", 15*time.Minute)
		gotUserID, gotEmail, err := validator.ValidateAccessToken(tok)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotUserID != userID {
			t.Errorf("userID = %q, want %q", gotUserID, userID)
		}
		if gotEmail != "user@example.com" {
			t.Errorf("email = %q, want %q", gotEmail, "user@example.com")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		tok := makeToken(t, testSecret, userID, "user@example.com", -1*time.Minute)
		_, _, err := validator.ValidateAccessToken(tok)
		if err == nil {
			t.Fatal("expected error for expired token, got nil")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		tok := makeToken(t, "wrong-secret", userID, "user@example.com", 15*time.Minute)
		_, _, err := validator.ValidateAccessToken(tok)
		if err == nil {
			t.Fatal("expected error for wrong secret, got nil")
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		_, _, err := validator.ValidateAccessToken("not.a.token")
		if err == nil {
			t.Fatal("expected error for malformed token, got nil")
		}
	})

	t.Run("wrong signing method rejected", func(t *testing.T) {
		// Test that tokens with non-HMAC signing methods are rejected.
		// We construct a minimal RS256 token header without needing RSA keys.
		// The validator should reject it before even trying to verify the signature.
		// This token is deliberately malformed (invalid signature) for testing only.
		malformedToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCIsImV4cCI6OTk5OTk5OTk5OX0.invalidsignature"
		_, _, err := validator.ValidateAccessToken(malformedToken)
		if err == nil {
			t.Fatal("expected error for wrong signing method, got nil")
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := middleware.NewJWTValidator(testSecret)
	userID := uuid.New().String()

	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(middleware.AuthMiddleware(validator))
		r.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		return r
	}

	t.Run("valid bearer token passes", func(t *testing.T) {
		tok := makeToken(t, testSecret, userID, "user@example.com", 15*time.Minute)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		newRouter().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		newRouter().ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("malformed header scheme returns 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Token abc123")
		newRouter().ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		tok := makeToken(t, testSecret, userID, "user@example.com", -1*time.Minute)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		newRouter().ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("nil validator panics at startup", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for nil validator, got none")
			}
		}()
		middleware.AuthMiddleware(nil)
	})
}
