package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/domain"
	"go.uber.org/zap"
)

// AnalyticsHTTPClient is an HTTP client to the analytics service
type AnalyticsHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewAnalyticsHTTPClient creates a new analytics HTTP client
func NewAnalyticsHTTPClient(baseURL string, logger *zap.Logger) *AnalyticsHTTPClient {
	return &AnalyticsHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetDashboardStats retrieves dashboard statistics for a workspace
func (c *AnalyticsHTTPClient) GetDashboardStats(
	ctx context.Context,
	workspaceID string,
) (*domain.DashboardResponse, error) {
	// For now, return mock data until analytics service exposes this endpoint
	// In a real scenario, you would call the analytics service here
	dashboard := &domain.DashboardResponse{
		TotalLinks:  0,
		TotalClicks: 0,
		RecentLinks: []domain.LinkResponse{},
		TopLinks:    []domain.LinkResponse{},
	}

	return dashboard, nil
}

// GetLinkAnalytics retrieves detailed analytics for a specific link
func (c *AnalyticsHTTPClient) GetLinkAnalytics(
	ctx context.Context,
	linkID string,
) (*domain.AnalyticsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/%s/summary", c.baseURL, linkID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("failed to create request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call analytics service", zap.Error(err))
		return nil, fmt.Errorf("failed to call analytics service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("analytics service error", zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		c.logger.Error("invalid response format")
		return nil, fmt.Errorf("invalid response format from analytics service")
	}

	analytics := &domain.AnalyticsResponse{
		LinkID:         linkID,
		ClicksByCountry: make(map[string]int64),
		ClicksByDevice:  make(map[string]int64),
	}

	if v, ok := data["total_clicks"].(float64); ok {
		analytics.TotalClicks = int64(v)
	}

	if v, ok := data["unique_clicks"].(float64); ok {
		analytics.UniqueClicks = int64(v)
	}

	return analytics, nil
}

// GetLiveCount retrieves the live click count for a short code
func (c *AnalyticsHTTPClient) GetLiveCount(ctx context.Context, shortCode string) (int64, error) {
	url := fmt.Sprintf("%s/api/v1/live-count/%s", c.baseURL, shortCode)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("failed to create request", zap.Error(err))
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call analytics service", zap.Error(err))
		return 0, fmt.Errorf("failed to call analytics service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("analytics service error", zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return 0, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode response", zap.Error(err))
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if count, ok := result["count"].(float64); ok {
		return int64(count), nil
	}

	return 0, nil
}
