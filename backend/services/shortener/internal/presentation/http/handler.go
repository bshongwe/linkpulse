package http

import (
	"net/http"
	"time"

	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/application"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/ports"
	sharedErrors "github.com/bshongwe/linkpulse/backend/shared/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	errInvalidRequestPayload   = "invalid request payload"
	errLinkNotFound            = "short link not found"
	errFailedCreateShortLink   = "failed to create short link"
	errFailedRetrieveShortLink = "failed to retrieve short link"
	errLinkExpired             = "short link has expired"
	errFailedUpdateShortLink   = "failed to update short link"
	errFailedDeactivateLink    = "failed to deactivate short link"
	errFailedDeleteLink        = "failed to delete short link"
	errFailedRetrieveStats     = "failed to retrieve link stats"
	errFailedListLinks         = "failed to list links"
	errFailedSearchLinks       = "failed to search links"
	errInvalidQueryParameters  = "invalid query parameters"
	errMissingQueryParameters  = "missing required query parameters"
)

type ShortenerHandler struct {
	service *application.ShortenerService
}

// NewShortenerHandler creates a new shortener HTTP handler
func NewShortenerHandler(service *application.ShortenerService) *ShortenerHandler {
	return &ShortenerHandler{
		service: service,
	}
}

// CreateShortLinkRequest represents the request payload for creating a short link
type CreateShortLinkRequest struct {
	OriginalURL  string   `json:"original_url" binding:"required,url"`
	WorkspaceID  string   `json:"workspace_id" binding:"required"`
	CreatedBy    string   `json:"created_by" binding:"required"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	CustomAlias  string   `json:"custom_alias"`
	ExpiresAt    *int64   `json:"expires_at"`
	RedirectType string   `json:"redirect_type"`
	Tags         []string `json:"tags"`
	CampaignID   *string  `json:"campaign_id"`
}

// CreateShortLinkResponse represents the response payload for a created short link
type CreateShortLinkResponse struct {
	ID          string  `json:"id"`
	ShortCode   string  `json:"short_code"`
	OriginalURL string  `json:"original_url"`
	WorkspaceID string  `json:"workspace_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ExpiresAt   *int64  `json:"expires_at"`
	IsActive    bool    `json:"is_active"`
	QRCode      string  `json:"qr_code"`
	CreatedAt   int64   `json:"created_at"`
}

// CreateShortLink handles POST /shorten requests
func (h *ShortenerHandler) CreateShortLink(c *gin.Context) {
	var req CreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errInvalidRequestPayload,
			"details": err.Error(),
		})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}
	userID, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid created_by"})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t := time.Unix(*req.ExpiresAt, 0)
		expiresAt = &t
	}

	var campaignID *uuid.UUID
	if req.CampaignID != nil {
		id, err := uuid.Parse(*req.CampaignID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign_id"})
			return
		}
		campaignID = &id
	}

	domainReq := &domain.CreateShortLinkRequest{
		OriginalURL:  req.OriginalURL,
		CustomAlias:  req.CustomAlias,
		Title:        req.Title,
		Description:  req.Description,
		ExpiresAt:    expiresAt,
		RedirectType: domain.RedirectType(req.RedirectType),
		Tags:         req.Tags,
		CampaignID:   campaignID,
	}

	link, err := h.service.CreateShortLink(c.Request.Context(), domainReq, userID, workspaceID)
	if err != nil {
		if sharedErrors.IsAlreadyExists(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "short code already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedCreateShortLink,
			"details": err.Error(),
		})
		return
	}

	var respExpiresAt *int64
	if link.ExpiresAt != nil {
		t := link.ExpiresAt.Unix()
		respExpiresAt = &t
	}

	response := CreateShortLinkResponse{
		ID:          link.ID.String(),
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
		WorkspaceID: link.WorkspaceID.String(),
		Title:       link.Title,
		Description: link.Description,
		ExpiresAt:   respExpiresAt,
		IsActive:    link.IsActive,
		QRCode:      link.QRCode,
		CreatedAt:   link.CreatedAt.Unix(),
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// GetShortLinkRequest represents query parameters for retrieving a short link
type GetShortLinkRequest struct {
	ShortCode string `form:"short_code" binding:"required"`
}

// GetShortLinkResponse represents the response for retrieving a short link
type GetShortLinkResponse struct {
	ID             string   `json:"id"`
	ShortCode      string   `json:"short_code"`
	OriginalURL    string   `json:"original_url"`
	WorkspaceID    string   `json:"workspace_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ExpiresAt      *int64   `json:"expires_at"`
	IsActive       bool     `json:"is_active"`
	ClickCount     int64    `json:"click_count"`
	LastAccessedAt *int64   `json:"last_accessed_at"`
	RedirectType   string   `json:"redirect_type"`
	QRCode         string   `json:"qr_code"`
	Tags           []string `json:"tags"`
	CampaignID     *string  `json:"campaign_id"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
}

// GetShortLink handles GET /shorten requests
func (h *ShortenerHandler) GetShortLink(c *gin.Context) {
	var req GetShortLinkRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errMissingQueryParameters,
			"details": err.Error(),
		})
		return
	}

	link, err := h.service.GetShortLink(c.Request.Context(), req.ShortCode)
	if err != nil {
		if sharedErrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": errLinkNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedRetrieveShortLink,
			"details": err.Error(),
		})
		return
	}

	if link.IsExpired() {
		c.JSON(http.StatusGone, gin.H{"error": errLinkExpired})
		return
	}

	var lastAccessedAt *int64
	if link.LastAccessedAt != nil {
		t := link.LastAccessedAt.Unix()
		lastAccessedAt = &t
	}
	var expiresAt *int64
	if link.ExpiresAt != nil {
		t := link.ExpiresAt.Unix()
		expiresAt = &t
	}
	var campaignID *string
	if link.CampaignID != nil {
		s := link.CampaignID.String()
		campaignID = &s
	}

	response := GetShortLinkResponse{
		ID:             link.ID.String(),
		ShortCode:      link.ShortCode,
		OriginalURL:    link.OriginalURL,
		WorkspaceID:    link.WorkspaceID.String(),
		Title:          link.Title,
		Description:    link.Description,
		ExpiresAt:      expiresAt,
		IsActive:       link.IsActive,
		ClickCount:     link.ClickCount,
		LastAccessedAt: lastAccessedAt,
		RedirectType:   string(link.RedirectType),
		QRCode:         link.QRCode,
		Tags:           link.Tags,
		CampaignID:     campaignID,
		CreatedAt:      link.CreatedAt.Unix(),
		UpdatedAt:      link.UpdatedAt.Unix(),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateShortLinkRequest represents the request payload for updating a short link
type UpdateShortLinkRequest struct {
	WorkspaceID  string   `json:"workspace_id" binding:"required"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ExpiresAt    *int64   `json:"expires_at"`
	IsActive     *bool    `json:"is_active"` // pointer so omitted != false
	RedirectType string   `json:"redirect_type"`
	Tags         []string `json:"tags"`
	CampaignID   *string  `json:"campaign_id"`
}

// UpdateShortLink handles PUT /shorten/:id requests
func (h *ShortenerHandler) UpdateShortLink(c *gin.Context) {
	var req UpdateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errInvalidRequestPayload,
			"details": err.Error(),
		})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}
	// Read link ID from path param, not body
	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t := time.Unix(*req.ExpiresAt, 0)
		expiresAt = &t
	}

	var campaignID *uuid.UUID
	if req.CampaignID != nil {
		id, err := uuid.Parse(*req.CampaignID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign_id"})
			return
		}
		campaignID = &id
	}

	domainReq := &domain.UpdateShortLinkRequest{
		Title:        req.Title,
		Description:  req.Description,
		ExpiresAt:    expiresAt,
		IsActive:     req.IsActive,
		RedirectType: domain.RedirectType(req.RedirectType),
		Tags:         req.Tags,
		CampaignID:   campaignID,
	}

	link, err := h.service.UpdateShortLink(c.Request.Context(), workspaceID, linkID, domainReq)
	if err != nil {
		if sharedErrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": errLinkNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedUpdateShortLink,
			"details": err.Error(),
		})
		return
	}

	var respExpiresAt *int64
	if link.ExpiresAt != nil {
		t := link.ExpiresAt.Unix()
		respExpiresAt = &t
	}

	response := CreateShortLinkResponse{
		ID:          link.ID.String(),
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
		WorkspaceID: link.WorkspaceID.String(),
		Title:       link.Title,
		Description: link.Description,
		ExpiresAt:   respExpiresAt,
		IsActive:    link.IsActive,
		QRCode:      link.QRCode,
		CreatedAt:   link.CreatedAt.Unix(),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// DeactivateLinkRequest represents the request payload for deactivating a short link
type DeactivateLinkRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
}

// DeactivateLink handles POST /shorten/:id/deactivate requests
func (h *ShortenerHandler) DeactivateLink(c *gin.Context) {
	var req DeactivateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errInvalidRequestPayload,
			"details": err.Error(),
		})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}
	// Read link ID from path param
	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeactivateLink(c.Request.Context(), workspaceID, linkID); err != nil {
		if sharedErrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": errLinkNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedDeactivateLink,
			"details": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteLinkRequest represents the request payload for deleting a short link
type DeleteLinkRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
}

// DeleteLink handles DELETE /shorten/:id requests
func (h *ShortenerHandler) DeleteLink(c *gin.Context) {
	var req DeleteLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errInvalidRequestPayload,
			"details": err.Error(),
		})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}
	// Read link ID from path param
	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteLink(c.Request.Context(), workspaceID, linkID); err != nil {
		if sharedErrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": errLinkNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedDeleteLink,
			"details": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// LinkStatsResponse represents the response for link statistics
type LinkStatsResponse struct {
	ID             string `json:"id"`
	ShortCode      string `json:"short_code"`
	ClickCount     int64  `json:"click_count"`
	LastAccessedAt *int64 `json:"last_accessed_at"`
	CreatedAt      int64  `json:"created_at"`
}

// GetLinkStats handles GET /shorten/:id/stats requests
func (h *ShortenerHandler) GetLinkStats(c *gin.Context) {
	var req struct {
		WorkspaceID string `form:"workspace_id" binding:"required"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errMissingQueryParameters,
			"details": err.Error(),
		})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}
	// Read link ID from path param
	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	stats, err := h.service.GetLinkStats(c.Request.Context(), workspaceID, linkID)
	if err != nil {
		if sharedErrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": errLinkNotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedRetrieveStats,
			"details": err.Error(),
		})
		return
	}

	var lastAccessedAt *int64
	if stats.LastAccessedAt != nil {
		t := stats.LastAccessedAt.Unix()
		lastAccessedAt = &t
	}

	response := LinkStatsResponse{
		ID:             stats.LinkID.String(),
		ShortCode:      stats.ShortCode,
		ClickCount:     stats.ClickCount,
		LastAccessedAt: lastAccessedAt,
		CreatedAt:      stats.CreatedAt.Unix(),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// ListLinksResponse represents paginated list of links
type ListLinksResponse struct {
	Links      []LinkSummary `json:"links"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int64         `json:"total_pages"`
}

// LinkSummary is a lightweight link representation used in list responses
type LinkSummary struct {
	ID             string `json:"id"`
	ShortCode      string `json:"short_code"`
	OriginalURL    string `json:"original_url"`
	ClickCount     int64  `json:"click_count"`
	LastAccessedAt *int64 `json:"last_accessed_at"`
	CreatedAt      int64  `json:"created_at"`
}

func toLinkSummary(link *domain.ShortLink) LinkSummary {
	var lastAccessedAt *int64
	if link.LastAccessedAt != nil {
		t := link.LastAccessedAt.Unix()
		lastAccessedAt = &t
	}
	return LinkSummary{
		ID:             link.ID.String(),
		ShortCode:      link.ShortCode,
		OriginalURL:    link.OriginalURL,
		ClickCount:     link.ClickCount,
		LastAccessedAt: lastAccessedAt,
		CreatedAt:      link.CreatedAt.Unix(),
	}
}

// ListLinksInWorkspace handles GET /shorten/workspace/:workspace_id requests
func (h *ShortenerHandler) ListLinksInWorkspace(c *gin.Context) {
	var req struct {
		Page     int    `form:"page" binding:"min=1"`
		PageSize int    `form:"page_size" binding:"min=1,max=100"`
		Sort     string `form:"sort"`
		Order    string `form:"order"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errInvalidQueryParameters,
			"details": err.Error(),
		})
		return
	}

	// Read workspace ID from path param
	workspaceID, err := uuid.Parse(c.Param("workspace_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	opts := ports.ListOptions{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
		Sort:   req.Sort,
		Order:  req.Order,
	}

	links, total, err := h.service.ListLinksInWorkspace(c.Request.Context(), workspaceID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedListLinks,
			"details": err.Error(),
		})
		return
	}

	responseLinks := make([]LinkSummary, len(links))
	for i, link := range links {
		responseLinks[i] = toLinkSummary(link)
	}

	c.JSON(http.StatusOK, gin.H{"data": ListLinksResponse{
		Links:      responseLinks,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}})
}

// ListLinksByCampaign handles GET /shorten/campaign/:campaign_id requests
func (h *ShortenerHandler) ListLinksByCampaign(c *gin.Context) {
	var req struct {
		WorkspaceID string `form:"workspace_id" binding:"required"`
		Page        int    `form:"page" binding:"min=1"`
		PageSize    int    `form:"page_size" binding:"min=1,max=100"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errInvalidQueryParameters,
			"details": err.Error(),
		})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}
	// Read campaign ID from path param
	campaignID, err := uuid.Parse(c.Param("campaign_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign_id"})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	opts := ports.ListOptions{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
	}

	links, total, err := h.service.ListLinksByCampaign(c.Request.Context(), workspaceID, campaignID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedListLinks,
			"details": err.Error(),
		})
		return
	}

	responseLinks := make([]LinkSummary, len(links))
	for i, link := range links {
		responseLinks[i] = toLinkSummary(link)
	}

	c.JSON(http.StatusOK, gin.H{"data": ListLinksResponse{
		Links:      responseLinks,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}})
}

// SearchByTagRequest represents query parameters for tag search
type SearchByTagRequest struct {
	Tag         string `form:"tag" binding:"required"`
	WorkspaceID string `form:"workspace_id" binding:"required"`
	Page        int    `form:"page" binding:"min=1"`
	PageSize    int    `form:"page_size" binding:"min=1,max=100"`
}

// SearchByTag handles GET /shorten/search/tag requests
func (h *ShortenerHandler) SearchByTag(c *gin.Context) {
	var req SearchByTagRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errInvalidQueryParameters,
			"details": err.Error(),
		})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	opts := ports.ListOptions{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
	}

	links, total, err := h.service.SearchByTag(c.Request.Context(), workspaceID, req.Tag, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errFailedSearchLinks,
			"details": err.Error(),
		})
		return
	}

	responseLinks := make([]LinkSummary, len(links))
	for i, link := range links {
		responseLinks[i] = toLinkSummary(link)
	}

	c.JSON(http.StatusOK, gin.H{"data": ListLinksResponse{
		Links:      responseLinks,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}})
}

// RegisterRoutes registers all shortener routes on the given router group
func RegisterRoutes(router *gin.Engine, handler *ShortenerHandler) {
	group := router.Group("/api/v1/shorten")

	group.POST("", handler.CreateShortLink)
	group.GET("", handler.GetShortLink)
	group.PUT("/:id", handler.UpdateShortLink)
	group.POST("/:id/deactivate", handler.DeactivateLink)
	group.DELETE("/:id", handler.DeleteLink)
	group.GET("/:id/stats", handler.GetLinkStats)
	group.GET("/workspace/:workspace_id", handler.ListLinksInWorkspace)
	group.GET("/campaign/:campaign_id", handler.ListLinksByCampaign)
	group.GET("/search/tag", handler.SearchByTag)
}
