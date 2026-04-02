package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/domain"
	"go.uber.org/zap"
)

// ShortenerHTTPClient is an HTTP client to the shortener service
type ShortenerHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewShortenerHTTPClient creates a new shortener HTTP client
func NewShortenerHTTPClient(baseURL string, logger *zap.Logger) *ShortenerHTTPClient {
	return &ShortenerHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// CreateLink calls the shortener service to create a new link
func (c *ShortenerHTTPClient) CreateLink(
	ctx context.Context,
	req domain.CreateLinkRequest,
	workspaceID, userID string,
) (*domain.LinkResponse, error) {
	// TODO: Implement HTTP call to shortener service
	// For now, return a placeholder
	return nil, fmt.Errorf("not implemented")
}

// GetLink retrieves a link from the shortener service
func (c *ShortenerHTTPClient) GetLink(ctx context.Context, shortCode string) (*domain.LinkResponse, error) {
	// TODO: Implement HTTP call to shortener service
	return nil, fmt.Errorf("not implemented")
}

// ListLinksInWorkspace lists all links in a workspace
func (c *ShortenerHTTPClient) ListLinksInWorkspace(
	ctx context.Context,
	workspaceID string,
	page, pageSize int,
) ([]domain.LinkResponse, int64, error) {
	// TODO: Implement HTTP call to shortener service
	return nil, 0, fmt.Errorf("not implemented")
}

// UpdateLink updates a link
func (c *ShortenerHTTPClient) UpdateLink(
	ctx context.Context,
	linkID string,
	req domain.CreateLinkRequest,
	userID string,
) (*domain.LinkResponse, error) {
	// TODO: Implement HTTP call to shortener service
	return nil, fmt.Errorf("not implemented")
}

// DeleteLink deletes a link
func (c *ShortenerHTTPClient) DeleteLink(ctx context.Context, linkID string, userID string) error {
	// TODO: Implement HTTP call to shortener service
	return fmt.Errorf("not implemented")
}

// close closes the HTTP client
func (c *ShortenerHTTPClient) close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// Helper to read response body and close it
func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Helper to check HTTP status
func checkHTTPStatus(resp *http.Response, statusCode int) error {
	if resp.StatusCode != statusCode {
		body, _ := readResponseBody(resp)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
