package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/ports"
)

// Handler handles HTTP requests for analytics
type Handler struct {
	analytics ports.AnalyticsService
	logger    *zap.Logger
}

// NewHandler creates a new analytics HTTP handler
func NewHandler(analytics ports.AnalyticsService, logger *zap.Logger) *Handler {
	return &Handler{
		analytics: analytics,
		logger:    logger,
	}
}

// RegisterRoutes registers all analytics routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	analytics := router.Group("/api/v1/analytics")
	{
		analytics.GET("/:linkId/summary", h.GetSummary)
		analytics.GET("/:linkId/clicks", h.GetClicks)
		analytics.GET("/:linkId/countries", h.GetCountryDistribution)
		analytics.GET("/:linkId/devices", h.GetDeviceDistribution)
	}

	// Live count endpoint uses short code instead of link ID
	router.GET("/api/v1/live-count/:code", h.GetLiveCount)
}

// GetSummary retrieves analytics summary for a link
// @Summary Get analytics summary
// @Description Get 30-day analytics summary for a link
// @Tags analytics
// @Param linkId path string true "Link ID (UUID)"
// @Success 200 {object} domain.AnalyticsSummary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/analytics/{linkId}/summary [get]
func (h *Handler) GetSummary(c *gin.Context) {
	linkIDStr := c.Param(linkIDParam)
	if linkIDStr == "" {
		h.logger.Warn(errMissingParam)
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errMissingParam})
		return
	}

	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		h.logger.Warn(errLinkIDFormat, zap.String("linkID", linkIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errInvalidUUID})
		return
	}

	// 30 days ago
	since := time.Now().AddDate(0, 0, -30)

	summary, err := h.analytics.GetAnalytics(c.Request.Context(), linkID, since)
	if err != nil {
		h.logger.Error(errGetAnalytics, zap.String("linkID", linkID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{statusKey: errInternalServer})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetClicks retrieves click events for a link within a time range
// @Summary Get click events
// @Description Retrieve click events for a specific link
// @Tags analytics
// @Param linkId path string true "Link ID (UUID)"
// @Param start query string false "Start time (RFC3339)"
// @Param end query string false "End time (RFC3339)"
// @Success 200 {array} domain.ClickEvent
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/analytics/{linkId}/clicks [get]
func (h *Handler) GetClicks(c *gin.Context) {
	linkIDStr := c.Param(linkIDParam)
	if linkIDStr == "" {
		h.logger.Warn(errMissingParam)
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errMissingParam})
		return
	}

	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		h.logger.Warn(errLinkIDFormat, zap.String("linkID", linkIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errInvalidUUID})
		return
	}

	// Parse time range from query parameters
	startStr := c.DefaultQuery(startParam, time.Now().AddDate(0, 0, -7).Format(time.RFC3339))
	endStr := c.DefaultQuery(endParam, time.Now().Format(time.RFC3339))

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		h.logger.Warn(errTimeFormat, zap.String("start", startStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errInvalidTime})
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		h.logger.Warn(errTimeFormat, zap.String("end", endStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errInvalidTime})
		return
	}

	clicks, err := h.analytics.GetClicksByTimeRange(c.Request.Context(), linkID, start, end)
	if err != nil {
		h.logger.Error(errGetClicks, zap.String("linkID", linkID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{statusKey: errInternalServer})
		return
	}

	if clicks == nil {
		clicks = make([]*domain.ClickEvent, 0)
	}

	c.JSON(http.StatusOK, clicks)
}

// GetCountryDistribution retrieves country distribution for a link
// @Summary Get country distribution
// @Description Get click distribution by country for a link
// @Tags analytics
// @Param linkId path string true "Link ID (UUID)"
// @Success 200 {object} map[string]int64
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/analytics/{linkId}/countries [get]
func (h *Handler) GetCountryDistribution(c *gin.Context) {
	linkIDStr := c.Param(linkIDParam)
	if linkIDStr == "" {
		h.logger.Warn(errMissingParam)
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errMissingParam})
		return
	}

	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		h.logger.Warn(errLinkIDFormat, zap.String("linkID", linkIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errInvalidUUID})
		return
	}

	distribution, err := h.analytics.GetCountryDistribution(c.Request.Context(), linkID)
	if err != nil {
		h.logger.Error(errGetCountry, zap.String("linkID", linkID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{statusKey: errInternalServer})
		return
	}

	if distribution == nil {
		distribution = make(map[string]int64)
	}

	c.JSON(http.StatusOK, distribution)
}

// GetDeviceDistribution retrieves device distribution for a link
// @Summary Get device distribution
// @Description Get click distribution by device type for a link
// @Tags analytics
// @Param linkId path string true "Link ID (UUID)"
// @Success 200 {object} map[string]int64
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/analytics/{linkId}/devices [get]
func (h *Handler) GetDeviceDistribution(c *gin.Context) {
	linkIDStr := c.Param(linkIDParam)
	if linkIDStr == "" {
		h.logger.Warn(errMissingParam)
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errMissingParam})
		return
	}

	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		h.logger.Warn(errLinkIDFormat, zap.String("linkID", linkIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errInvalidUUID})
		return
	}

	distribution, err := h.analytics.GetDeviceDistribution(c.Request.Context(), linkID)
	if err != nil {
		h.logger.Error(errGetDevice, zap.String("linkID", linkID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{statusKey: errInternalServer})
		return
	}

	if distribution == nil {
		distribution = make(map[string]int64)
	}

	c.JSON(http.StatusOK, distribution)
}

// GetLiveCount retrieves the current click count for a short code
// @Summary Get live click count
// @Description Get the current total click count for a short code
// @Tags analytics
// @Param code path string true "Short code"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/live-count/{code} [get]
func (h *Handler) GetLiveCount(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		h.logger.Warn(errMissingParam)
		c.JSON(http.StatusBadRequest, gin.H{statusKey: errMissingParam})
		return
	}

	count, err := h.analytics.GetLiveCount(c.Request.Context(), shortCode)
	if err != nil {
		h.logger.Error(errGetLiveCount, zap.String("shortCode", shortCode), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{statusKey: errInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		countKey: count,
	})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
