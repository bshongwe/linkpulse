package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/application"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/domain"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
)

type Handler struct {
	authService *application.AuthService
}

func NewHandler(authService *application.AuthService) *Handler {
	return &Handler{authService: authService}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		logger.Log.Error("Register failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	c.JSON(http.StatusCreated, UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Check if it is an invalid credentials error
		if err == domain.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		// Other errors
		logger.Log.Error("Login failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	// TODO: Generate JWT tokens (I'll update this later)
	c.JSON(http.StatusOK, UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
	})
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
