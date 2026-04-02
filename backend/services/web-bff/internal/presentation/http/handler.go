package http

import (
	"net/http"
	"strconv"

	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/application"
	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for the BFF
type Handler struct {
	bffService *application.BFFService
	logger     *zap.Logger
}

// NewHandler creates a new BFF HTTP handler
func NewHandler(bffService *application.BFFService, logger *zap.Logger) *Handler {
	return &Handler{
		bffService: bffService,
		logger:     logger,
	}
}

// RegisterRoutes registers all BFF routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/bff")
	{
		// Link management
		api.POST("/links", h.CreateLink)
		api.GET("/links", h.ListLinks)
		api.GET("/links/:linkID", h.GetLink)
		api.PUT("/links/:linkID", h.UpdateLink)
		api.DELETE("/links/:linkID", h.DeleteLink)

		// Analytics
		api.GET("/dashboard", h.GetDashboard)
		api.GET("/links/:linkID/analytics", h.GetLinkAnalytics)
	}
}

// CreateLink creates a new short link
func (h *Handler) CreateLink(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.GetString("workspace_id")

	if userID == "" || workspaceID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Error:  "missing user or workspace context",
			Status: http.StatusUnauthorized,
		})
		return
	}

	var req domain.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:  err.Error(),
			Status: http.StatusBadRequest,
		})
		return
	}

	link, err := h.bffService.CreateShortLink(c.Request.Context(), req, workspaceID, userID)
	if err != nil {
		h.logger.Error("failed to create link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:  "failed to create link",
			Status: http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, link)
}

// ListLinks lists all links in a workspace
func (h *Handler) ListLinks(c *gin.Context) {
	workspaceID := c.GetString("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Error:  "missing workspace context",
			Status: http.StatusUnauthorized,
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

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	links, total, err := h.bffService.ListLinks(c.Request.Context(), workspaceID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:  "failed to list links",
			Status: http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": links,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetLink retrieves a single link
func (h *Handler) GetLink(c *gin.Context) {
	linkID := c.Param("linkID")
	if linkID == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:  "missing link ID",
			Status: http.StatusBadRequest,
		})
		return
	}

	link, err := h.bffService.GetShortLink(c.Request.Context(), linkID)
	if err != nil {
		h.logger.Error("failed to get link", zap.String("linkID", linkID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:  "failed to get link",
			Status: http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, link)
}

// UpdateLink updates an existing link
func (h *Handler) UpdateLink(c *gin.Context) {
	userID := c.GetString("user_id")
	linkID := c.Param("linkID")

	if userID == "" || linkID == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:  "missing user ID or link ID",
			Status: http.StatusBadRequest,
		})
		return
	}

	var req domain.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:  err.Error(),
			Status: http.StatusBadRequest,
		})
		return
	}

	link, err := h.bffService.UpdateLink(c.Request.Context(), linkID, req, userID)
	if err != nil {
		h.logger.Error("failed to update link", zap.String("linkID", linkID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:  "failed to update link",
			Status: http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, link)
}

// DeleteLink deletes a link
func (h *Handler) DeleteLink(c *gin.Context) {
	userID := c.GetString("user_id")
	linkID := c.Param("linkID")

	if userID == "" || linkID == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:  "missing user ID or link ID",
			Status: http.StatusBadRequest,
		})
		return
	}

	err := h.bffService.DeleteLink(c.Request.Context(), linkID, userID)
	if err != nil {
		h.logger.Error("failed to delete link", zap.String("linkID", linkID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:  "failed to delete link",
			Status: http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetDashboard retrieves dashboard data
func (h *Handler) GetDashboard(c *gin.Context) {
	workspaceID := c.GetString("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Error:  "missing workspace context",
			Status: http.StatusUnauthorized,
		})
		return
	}

	dashboard, err := h.bffService.GetDashboard(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:  "failed to load dashboard",
			Status: http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetLinkAnalytics retrieves analytics for a link
func (h *Handler) GetLinkAnalytics(c *gin.Context) {
	linkID := c.Param("linkID")
	if linkID == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:  "missing link ID",
			Status: http.StatusBadRequest,
		})
		return
	}

	analytics, err := h.bffService.GetLinkAnalytics(c.Request.Context(), linkID)
	if err != nil {
		h.logger.Error("failed to get link analytics", zap.String("linkID", linkID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:  "failed to get analytics",
			Status: http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
