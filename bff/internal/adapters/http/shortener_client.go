package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bshongwe/linkpulse/bff/internal/domain"
	"go.uber.org/zap"
)

// ShortenerHTTPClient implements ports.ShortenerClient
type ShortenerHTTPClient struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewShortenerHTTPClient creates a new shortener HTTP client
func NewShortenerHTTPClient(baseURL string, logger *zap.Logger) *ShortenerHTTPClient {
	return &ShortenerHTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// CreateLink creates a new short link
func (c *ShortenerHTTPClient) CreateLink(
	ctx context.Context,
	req domain.CreateLinkRequest,
	workspaceID, userID, jwtToken string,
) (*domain.LinkResponse, error) {
	// Map BFF request to shortener service request
	shortenerReq := map[string]interface{}{
		"original_url": req.URL,
		"workspace_id": workspaceID,
		"created_by":   userID,
	}
	if req.Custom != nil {
		shortenerReq["custom_alias"] = *req.Custom
	}
	if req.ExpiresAt != nil {
		shortenerReq["expires_at"] = req.ExpiresAt.Unix() * 1000 // Convert to milliseconds
	}
	if len(req.Tags) > 0 {
		shortenerReq["tags"] = req.Tags
	}

	body, err := json.Marshal(shortenerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/shorten", c.baseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	httpReq.Header.Set("X-Workspace-ID", workspaceID)
	httpReq.Header.Set("X-User-ID", userID)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call shortener service",
			zap.String("endpoint", "/api/v1/links"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to call shortener service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shortener service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var linkResp domain.LinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&linkResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &linkResp, nil
}

// GetLink retrieves a short link
func (c *ShortenerHTTPClient) GetLink(ctx context.Context, shortCode, jwtToken string) (*domain.LinkResponse, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/shorten?short_code=%s", c.baseURL, shortCode),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call shortener service",
			zap.String("endpoint", fmt.Sprintf("/api/v1/links/%s", shortCode)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to call shortener service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shortener service returned status %d", resp.StatusCode)
	}

	var linkResp domain.LinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&linkResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &linkResp, nil
}

// ListLinksInWorkspace lists all links in a workspace
func (c *ShortenerHTTPClient) ListLinksInWorkspace(
	ctx context.Context,
	workspaceID string,
	page, pageSize int,
	jwtToken string,
) ([]domain.LinkResponse, int64, error) {
	url := fmt.Sprintf("%s/api/v1/shorten/workspace/%s?page=%d&pageSize=%d", c.baseURL, workspaceID, page, pageSize)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to call shortener service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("shortener service returned status %d", resp.StatusCode)
	}

	var result struct {
		Links []domain.LinkResponse `json:"links"`
		Total int64                 `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Links, result.Total, nil
}

// DeleteLink deletes a short link
func (c *ShortenerHTTPClient) DeleteLink(ctx context.Context, shortCode, jwtToken string) error {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/api/v1/shorten/%s", c.baseURL, shortCode),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call shortener service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("shortener service returned status %d", resp.StatusCode)
	}

	return nil
}
