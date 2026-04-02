package http

import (
	"context"
	"fmt"
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

// GetDashboardStats retrieves dashboard statistics
func (c *AnalyticsHTTPClient) GetDashboardStats(
	ctx context.Context,
	workspaceID string,
) (*domain.DashboardResponse, error) {
	// TODO: Implement HTTP call to analytics service
	return nil, fmt.Errorf("not implemented")
}

// GetLinkAnalytics retrieves detailed analytics for a link
func (c *AnalyticsHTTPClient) GetLinkAnalytics(
	ctx context.Context,
	linkID string,
) (*domain.AnalyticsResponse, error) {
	// TODO: Implement HTTP call to analytics service
	return nil, fmt.Errorf("not implemented")
}

// GetLiveCount retrieves the live click count for a short code
func (c *AnalyticsHTTPClient) GetLiveCount(ctx context.Context, shortCode string) (int64, error) {
	// TODO: Implement HTTP call to analytics service
	return 0, fmt.Errorf("not implemented")
}

// close closes the HTTP client
func (c *AnalyticsHTTPClient) close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
