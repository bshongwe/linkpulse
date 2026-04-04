package domain

import "time"

// CreateLinkRequest represents a request to create a short link
type CreateLinkRequest struct {
	OriginalURL  string  `json:"original_url" binding:"required,url"`
	CustomAlias  *string `json:"custom_alias,omitempty" binding:"omitempty,alphanum"`
	Title        *string `json:"title,omitempty" binding:"omitempty,max=200"`
	Description  *string `json:"description,omitempty" binding:"omitempty,max=500"`
	CampaignID   *string `json:"campaign_id,omitempty"`
	Tags         []string `json:"tags,omitempty" binding:"omitempty,max=10"`
	RedirectType *string `json:"redirect_type,omitempty" binding:"omitempty,oneof=301 302"`
}

// LinkResponse represents a short link in the BFF response
type LinkResponse struct {
	ID          string    `json:"id"`
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Clicks      int64     `json:"click_count"`
	IsActive    bool      `json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Tags        []string  `json:"tags,omitempty"`
	CampaignID  *string   `json:"campaign_id,omitempty"`
}

// DashboardResponse represents the complete dashboard data
type DashboardResponse struct {
	TotalLinks   int64          `json:"total_links"`
	TotalClicks  int64          `json:"total_clicks"`
	RecentLinks  []LinkResponse `json:"recent_links"`
	TopLinks     []LinkResponse `json:"top_links"`
}

// AnalyticsResponse represents analytics for a specific link
type AnalyticsResponse struct {
	LinkID           string            `json:"link_id"`
	ShortCode        string            `json:"short_code"`
	TotalClicks      int64             `json:"total_clicks"`
	UniqueClicks     int64             `json:"unique_clicks"`
	ClicksByCountry  map[string]int64  `json:"clicks_by_country,omitempty"`
	ClicksByDevice   map[string]int64  `json:"clicks_by_device,omitempty"`
	TopReferrers     []string          `json:"top_referrers,omitempty"`
	LastAccessedAt   *time.Time        `json:"last_accessed_at,omitempty"`
}

// ErrorResponse is the standard error format
type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
	Path   string `json:"path,omitempty"`
}
