package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bshongwe/linkpulse/bff/internal/domain"
	"github.com/bshongwe/linkpulse/bff/internal/ports"
	"go.uber.org/zap"
)

// AnalyticsHTTPClient implements ports.AnalyticsClient
type AnalyticsHTTPClient struct {
	baseURL           string
	client            *http.Client
	logger            *zap.Logger
	shortenerClient   ports.ShortenerClient
}

// NewAnalyticsHTTPClient creates a new analytics HTTP client
func NewAnalyticsHTTPClient(baseURL string, logger *zap.Logger, shortenerClient ports.ShortenerClient) *AnalyticsHTTPClient {
	return &AnalyticsHTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:          logger,
		shortenerClient: shortenerClient,
	}
}

// GetLinkAnalytics retrieves analytics for a specific link
func (c *AnalyticsHTTPClient) GetLinkAnalytics(
	ctx context.Context,
	shortCode string,
	jwtToken string,
) (*domain.AnalyticsResponse, error) {
	// Step 1: Resolve shortCode to linkId via shortener service
	link, err := c.shortenerClient.GetLink(ctx, shortCode, jwtToken)
	if err != nil {
		c.logger.Error("failed to resolve short code to link ID",
			zap.String("shortCode", shortCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to resolve link: %w", err)
	}

	// Step 2: Call analytics service with linkId
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/analytics/%s/summary", c.baseURL, link.ID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call analytics service",
			zap.String("endpoint", fmt.Sprintf("/api/v1/analytics/%s/summary", link.ID)),
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
// NOTE: Currently aggregates click counts from all links in the workspace
// since analytics service doesn't have a dedicated workspace analytics endpoint
func (c *AnalyticsHTTPClient) GetWorkspaceAnalytics(
	ctx context.Context,
	workspaceID string,
	jwtToken string,
) (*domain.AnalyticsResponse, error) {
	// Step 1: Get all links for the workspace from shortener service
	links, _, err := c.shortenerClient.ListLinksInWorkspace(ctx, workspaceID, 1, 1000, jwtToken)
	if err != nil {
		c.logger.Error("failed to list links for workspace",
			zap.String("workspaceID", workspaceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list workspace links: %w", err)
	}

	if len(links) == 0 {
		// Return empty analytics response for workspace with no links
		return &domain.AnalyticsResponse{
			TotalClicks: 0,
		}, nil
	}

	// Step 2: Aggregate analytics from all links in the workspace
	totalClicks := int64(0)
	for _, link := range links {
		httpReq, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("%s/api/v1/analytics/%s/summary", c.baseURL, link.ID),
			nil,
		)
		if err != nil {
			c.logger.Warn("failed to create analytics request",
				zap.String("linkID", link.ID),
				zap.Error(err),
			)
			continue
		}

		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

		resp, err := c.client.Do(httpReq)
		if err != nil {
			c.logger.Warn("failed to call analytics service",
				zap.String("linkID", link.ID),
				zap.Error(err),
			)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.logger.Warn("analytics service returned error",
				zap.String("linkID", link.ID),
				zap.Int("status", resp.StatusCode),
			)
			continue
		}

		var analytics domain.AnalyticsResponse
		if err := json.NewDecoder(resp.Body).Decode(&analytics); err != nil {
			c.logger.Warn("failed to decode analytics response",
				zap.String("linkID", link.ID),
				zap.Error(err),
			)
			continue
		}

		totalClicks += analytics.TotalClicks
	}

	return &domain.AnalyticsResponse{
		TotalClicks: totalClicks,
	}, nil
}
