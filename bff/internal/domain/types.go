package domain

import "time"

// CreateLinkRequest represents request to create a short link
type CreateLinkRequest struct {
	URL          string     `json:"url" binding:"required,url,max=2048"`
	Title        *string    `json:"title" binding:"omitempty,max=200"`
	Description  *string    `json:"description" binding:"omitempty,max=500"`
	Custom       *string    `json:"custom" binding:"omitempty,max=50"`
	ExpiresAt    *int64     `json:"expires_at" binding:"omitempty"`          // Unix seconds from frontend
	RedirectType *string    `json:"redirect_type" binding:"omitempty"`       // "301" or "302"
	Tags         []string   `json:"tags" binding:"omitempty"`
	CampaignID   *string    `json:"campaign_id" binding:"omitempty"`
}

// LinkResponse represents the response for a link
type LinkResponse struct {
	ID          string     `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   int64      `json:"created_at"`
	ExpiresAt   *int64     `json:"expires_at,omitempty"`
	Clicks      int64      `json:"click_count"`
	Tags        []string   `json:"tags,omitempty"`
	IsActive    bool       `json:"is_active"`
	WorkspaceID string     `json:"workspace_id"`
}

// AnalyticsResponse represents click analytics
type AnalyticsResponse struct {
	ShortCode      string                `json:"short_code"`
	TotalClicks    int64                 `json:"total_clicks"`
	UniqueClicks   int64                 `json:"unique_clicks"`
	ClicksByCountry map[string]int64     `json:"clicks_by_country,omitempty"`
	ClicksByDevice  map[string]int64     `json:"clicks_by_device,omitempty"`
	ClicksOverTime  []TimeSeriesDataPoint `json:"clicks_over_time,omitempty"`
}

// TimeSeriesDataPoint represents a point in time series data
type TimeSeriesDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     int64     `json:"value"`
}

// ErrorResponse represents error response format
type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}
