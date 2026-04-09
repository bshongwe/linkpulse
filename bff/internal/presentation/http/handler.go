package http

import (
	"net/http"
	"strconv"

	"github.com/bshongwe/linkpulse/bff/internal/application"
	"github.com/bshongwe/linkpulse/bff/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for the BFF
type Handler struct {
	bffService *application.BFFService
	logger     *zap.Logger
}

// NewHandler creates a new handler
func NewHandler(bffService *application.BFFService, logger *zap.Logger) *Handler {
	return &Handler{
		bffService: bffService,
		logger:     logger,
	}
}

// RegisterRoutes registers all BFF routes
func (h *Handler) RegisterRoutes(router *gin.Engine, jwtSecret string) {
	// Protected routes with JWT middleware
	api := router.Group("/api/v1")
	api.Use(JWTMiddleware(jwtSecret))
	{
		// Link endpoints
		api.POST("/links", h.CreateLink)
		api.GET("/links", h.ListLinksForUser)          // Convenience route using workspace_id from JWT
		api.GET("/links/:shortCode", h.GetLink)
		api.GET("/workspaces/:workspaceId/links", h.ListLinks)
		api.DELETE("/links/:shortCode", h.DeleteLink)

		// Analytics endpoints
		api.GET("/links/:shortCode/analytics", h.GetLinkAnalytics)
		api.GET("/workspaces/:workspaceId/analytics", h.GetWorkspaceAnalytics)
	}
}

// CreateLink handles POST /api/v1/links
func (h *Handler) CreateLink(c *gin.Context) {
	var req domain.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	workspaceID := c.GetString("workspace_id")
	userID := c.GetString("user_id")
	jwtToken := c.GetString("jwt_token")

	resp, err := h.bffService.CreateShortLink(c.Request.Context(), req, workspaceID, userID, jwtToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "FAILED_TO_CREATE_LINK",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetLink handles GET /api/v1/links/:shortCode
func (h *Handler) GetLink(c *gin.Context) {
	shortCode := c.Param("shortCode")
	jwtToken := c.GetString("jwt_token")

	resp, err := h.bffService.GetShortLink(c.Request.Context(), shortCode, jwtToken)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Code:    "LINK_NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListLinks handles GET /api/v1/workspaces/:workspaceId/links
func (h *Handler) ListLinks(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	authWorkspaceID := c.GetString("workspace_id")
	
	// SECURITY: Verify requested workspace matches authenticated user's workspace
	if workspaceID != authWorkspaceID {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "You do not have access to this workspace",
		})
		return
	}
	
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	jwtToken := c.GetString("jwt_token")

	links, total, err := h.bffService.ListLinks(c.Request.Context(), workspaceID, page, pageSize, jwtToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "FAILED_TO_LIST_LINKS",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"links":     links,
		"total":     total,
		"page":      page,
		"pageSize":  pageSize,
	})
}

// DeleteLink handles DELETE /api/v1/links/:shortCode
func (h *Handler) DeleteLink(c *gin.Context) {
	shortCode := c.Param("shortCode")
	jwtToken := c.GetString("jwt_token")

	if err := h.bffService.DeleteShortLink(c.Request.Context(), shortCode, jwtToken); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "FAILED_TO_DELETE_LINK",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetLinkAnalytics handles GET /api/v1/links/:shortCode/analytics
func (h *Handler) GetLinkAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	jwtToken := c.GetString("jwt_token")

	resp, err := h.bffService.GetLinkAnalytics(c.Request.Context(), shortCode, jwtToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "FAILED_TO_GET_ANALYTICS",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetWorkspaceAnalytics handles GET /api/v1/workspaces/:workspaceId/analytics
func (h *Handler) GetWorkspaceAnalytics(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	authWorkspaceID := c.GetString("workspace_id")
	
	// SECURITY: Verify requested workspace matches authenticated user's workspace
	if workspaceID != authWorkspaceID {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "You do not have access to this workspace",
		})
		return
	}
	
	jwtToken := c.GetString("jwt_token")

	resp, err := h.bffService.GetWorkspaceAnalytics(c.Request.Context(), workspaceID, jwtToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "FAILED_TO_GET_ANALYTICS",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListLinksForUser handles GET /api/v1/links (convenience route using workspace_id from JWT)
func (h *Handler) ListLinksForUser(c *gin.Context) {
	workspaceID := c.GetString("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "workspace_id not found in token",
		})
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	jwtToken := c.GetString("jwt_token")

	links, total, err := h.bffService.ListLinks(c.Request.Context(), workspaceID, page, pageSize, jwtToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "FAILED_TO_LIST_LINKS",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"links":     links,
		"total":     total,
		"page":      page,
		"pageSize":  pageSize,
	})
}
