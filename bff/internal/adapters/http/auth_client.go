package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// AuthHTTPClient implements ports.AuthClient
type AuthHTTPClient struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewAuthHTTPClient creates a new auth HTTP client
func NewAuthHTTPClient(baseURL string, logger *zap.Logger) *AuthHTTPClient {
	return &AuthHTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * context.Background().Done(),
		},
		logger: logger,
	}
}

// ValidateToken validates a JWT token
func (c *AuthHTTPClient) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/auth/validate", c.baseURL),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("failed to call auth service",
			zap.String("endpoint", "/api/v1/auth/validate"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetUserInfo retrieves user information
func (c *AuthHTTPClient) GetUserInfo(ctx context.Context, userID, jwtToken string) (map[string]interface{}, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
