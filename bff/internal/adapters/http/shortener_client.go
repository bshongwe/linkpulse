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

// API endpoints
const (
	shortenEndpoint          = "/api/v1/shorten"
	shortenWorkspaceEndpoint = "/api/v1/shorten/workspace"
)

// Error messages
const (
	errFailedCreateRequest  = "failed to create request: %w"
	errFailedCallService    = "failed to call shortener service: %w"
	errFailedDecodeResponse = "failed to decode response: %w"
	errServiceStatusMsg     = "shortener service returned status %d"
)

// Header names
const (
	headerContentType   = "Content-Type"
	headerAuthorization = "Authorization"
	headerWorkspaceID   = "X-Workspace-ID"
	headerUserID        = "X-User-ID"
)

// Header values
const (
	contentTypeJSON = "application/json"
	bearerPrefix    = "Bearer %s"
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
		shortenerReq["expires_at"] = req.ExpiresAt.Unix() // Unix seconds, not milliseconds
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
		fmt.Sprintf("%s%s", c.baseURL, shortenEndpoint),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf(errFailedCreateRequest, err)
	}

	httpReq.Header.Set(headerContentType, contentTypeJSON)
	httpReq.Header.Set(headerAuthorization, fmt.Sprintf(bearerPrefix, jwtToken))
	httpReq.Header.Set(headerWorkspaceID, workspaceID)
	httpReq.Header.Set(headerUserID, userID)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call shortener service",
			zap.String("endpoint", shortenEndpoint),
			zap.Error(err),
		)
		return nil, fmt.Errorf(errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(errServiceStatusMsg+": %s", resp.StatusCode, string(bodyBytes))
	}

	// Shortener returns {"data": {...}}
	var envelope struct {
		Data domain.LinkResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf(errFailedDecodeResponse, err)
	}

	return &envelope.Data, nil
}

// GetLink retrieves a short link
func (c *ShortenerHTTPClient) GetLink(ctx context.Context, shortCode, jwtToken string) (*domain.LinkResponse, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s%s?short_code=%s", c.baseURL, shortenEndpoint, shortCode),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(errFailedCreateRequest, err)
	}

	httpReq.Header.Set(headerAuthorization, fmt.Sprintf(bearerPrefix, jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call shortener service",
			zap.String("endpoint", fmt.Sprintf("%s (short_code=%s)", shortenEndpoint, shortCode)),
			zap.Error(err),
		)
		return nil, fmt.Errorf(errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errServiceStatusMsg, resp.StatusCode)
	}

	// Shortener returns {"data": {...}}
	var envelope struct {
		Data domain.LinkResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf(errFailedDecodeResponse, err)
	}

	return &envelope.Data, nil
}

// ListLinksInWorkspace lists all links in a workspace
func (c *ShortenerHTTPClient) ListLinksInWorkspace(
	ctx context.Context,
	workspaceID string,
	page, pageSize int,
	jwtToken string,
) ([]domain.LinkResponse, int64, error) {
	url := fmt.Sprintf("%s%s/%s?page=%d&page_size=%d", c.baseURL, shortenWorkspaceEndpoint, workspaceID, page, pageSize)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf(errFailedCreateRequest, err)
	}

	httpReq.Header.Set(headerAuthorization, fmt.Sprintf(bearerPrefix, jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf(errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf(errServiceStatusMsg, resp.StatusCode)
	}

	// Shortener returns {"data": {"links": [...], "total": ...}}
	var envelope struct {
		Data struct {
			Links []domain.LinkResponse `json:"links"`
			Total int64                 `json:"total"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, 0, fmt.Errorf(errFailedDecodeResponse, err)
	}

	return envelope.Data.Links, envelope.Data.Total, nil
}

// DeleteLink deletes a short link
func (c *ShortenerHTTPClient) DeleteLink(ctx context.Context, shortCode, jwtToken string) error {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s%s/%s", c.baseURL, shortenEndpoint, shortCode),
		nil,
	)
	if err != nil {
		return fmt.Errorf(errFailedCreateRequest, err)
	}

	httpReq.Header.Set(headerAuthorization, fmt.Sprintf(bearerPrefix, jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf(errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf(errServiceStatusMsg, resp.StatusCode)
	}

	return nil
}
