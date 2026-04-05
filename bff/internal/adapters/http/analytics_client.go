package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bshongwe/linkpulse/bff/internal/domain"
	"go.uber.org/zap"
)

// AnalyticsHTTPClient implements ports.AnalyticsClient
type AnalyticsHTTPClient struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewAnalyticsHTTPClient creates a new analytics HTTP client
func NewAnalyticsHTTPClient(baseURL string, logger *zap.Logger) *AnalyticsHTTPClient {
	return &AnalyticsHTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetLinkAnalytics retrieves analytics for a specific link
func (c *AnalyticsHTTPClient) GetLinkAnalytics(
	ctx context.Context,
	shortCode string,
	jwtToken string,
) (*domain.AnalyticsResponse, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/links/%s/analytics", c.baseURL, shortCode),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call analytics service",
			zap.String("endpoint", fmt.Sprintf("/api/v1/links/%s/analytics", shortCode)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to call analytics service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	var analyticsResp domain.AnalyticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&analyticsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &analyticsResp, nil
}

// GetWorkspaceAnalytics retrieves analytics for a workspace
func (c *AnalyticsHTTPClient) GetWorkspaceAnalytics(
	ctx context.Context,
	workspaceID string,
	jwtToken string,
) (*domain.AnalyticsResponse, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/workspaces/%s/analytics", c.baseURL, workspaceID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call analytics service",
			zap.String("endpoint", fmt.Sprintf("/api/v1/workspaces/%s/analytics", workspaceID)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to call analytics service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	var analyticsResp domain.AnalyticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&analyticsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &analyticsResp, nil
}
