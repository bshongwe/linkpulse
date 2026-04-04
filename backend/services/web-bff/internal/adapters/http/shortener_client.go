package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/domain"
	"go.uber.org/zap"
)

const (
	authorizationHeader         = "Authorization"
	bearerPrefix                = "Bearer "
	contentTypeHeader           = "Content-Type"
	contentTypeJSON             = "application/json"
	errFailedMarshal            = "failed to marshal request"
	errFailedCreateRequest      = "failed to create request"
	errFailedCallService        = "failed to call shortener service"
	errServiceError             = "shortener service error"
	errFailedDecode             = "failed to decode response"
	errInvalidResponseFormat    = "invalid response format from shortener service"
	errFormat                   = "%s: %w"
	errServiceReturnedStatus    = "shortener service returned status %d"
	errInvalidResponseFormatMsg = "invalid response format"
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
	workspaceID, userID, jwtToken string,
) (*domain.LinkResponse, error) {
	payload := map[string]interface{}{
		"original_url": req.OriginalURL,
		"workspace_id": workspaceID,
		"created_by":   userID,
	}

	if req.CustomAlias != nil {
		payload["custom_alias"] = *req.CustomAlias
	}
	if req.Title != nil {
		payload["title"] = *req.Title
	}
	if req.Description != nil {
		payload["description"] = *req.Description
	}
	if req.RedirectType != nil {
		payload["redirect_type"] = *req.RedirectType
	}
	if len(req.Tags) > 0 {
		payload["tags"] = req.Tags
	}
	if req.CampaignID != nil {
		payload["campaign_id"] = *req.CampaignID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error(errFailedMarshal, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedMarshal, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/shorten", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error(errFailedCreateRequest, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedCreateRequest, err)
	}

	httpReq.Header.Set(contentTypeHeader, contentTypeJSON)
	if jwtToken != "" {
		authHeader := bearerPrefix + jwtToken
		httpReq.Header.Set(authorizationHeader, authHeader)
		c.logger.Info("CreateLink: forwarding JWT token to shortener", 
			zap.Int("token_length", len(jwtToken)))
	} else {
		c.logger.Warn("CreateLink: no JWT token provided")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error(errFailedCallService, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error(errServiceError, zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return nil, fmt.Errorf(errServiceReturnedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(errFailedDecode, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedDecode, err)
	}

	// Extract the link data
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		c.logger.Error(errInvalidResponseFormatMsg)
		return nil, fmt.Errorf(errInvalidResponseFormat)
	}

	// Map response to LinkResponse
	link := &domain.LinkResponse{
		ID:          getString(data, "id"),
		ShortCode:   getString(data, "short_code"),
		OriginalURL: getString(data, "original_url"),
		Title:       getString(data, "title"),
		Description: getString(data, "description"),
		IsActive:    getBool(data, "is_active"),
	}

	if v, ok := data["click_count"].(float64); ok {
		link.Clicks = int64(v)
	}

	return link, nil
}

// GetLink retrieves a link from the shortener service
func (c *ShortenerHTTPClient) GetLink(ctx context.Context, shortCode, jwtToken string) (*domain.LinkResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/shorten?short_code="+shortCode, nil)
	if err != nil {
		c.logger.Error(errFailedCreateRequest, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedCreateRequest, err)
	}

	if jwtToken != "" {
		httpReq.Header.Set(authorizationHeader, bearerPrefix+jwtToken)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error(errFailedCallService, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error(errServiceError, zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return nil, fmt.Errorf(errServiceReturnedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(errFailedDecode, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedDecode, err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		c.logger.Error(errInvalidResponseFormatMsg)
		return nil, fmt.Errorf(errInvalidResponseFormat)
	}

	link := &domain.LinkResponse{
		ID:          getString(data, "id"),
		ShortCode:   getString(data, "short_code"),
		OriginalURL: getString(data, "original_url"),
		Title:       getString(data, "title"),
		Description: getString(data, "description"),
		IsActive:    getBool(data, "is_active"),
	}

	if v, ok := data["click_count"].(float64); ok {
		link.Clicks = int64(v)
	}

	return link, nil
}

// ListLinksInWorkspace lists all links in a workspace
func (c *ShortenerHTTPClient) ListLinksInWorkspace(
	ctx context.Context,
	workspaceID string,
	page, pageSize int,
	jwtToken string,
) ([]domain.LinkResponse, int64, error) {
	url := fmt.Sprintf("%s/api/v1/shorten/workspace/%s?page=%d&page_size=%d", c.baseURL, workspaceID, page, pageSize)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error(errFailedCreateRequest, zap.Error(err))
		return nil, 0, fmt.Errorf(errFormat, errFailedCreateRequest, err)
	}

	if jwtToken != "" {
		httpReq.Header.Set(authorizationHeader, bearerPrefix+jwtToken)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error(errFailedCallService, zap.Error(err))
		return nil, 0, fmt.Errorf(errFormat, errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error(errServiceError, zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return nil, 0, fmt.Errorf(errServiceReturnedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(errFailedDecode, zap.Error(err))
		return nil, 0, fmt.Errorf(errFormat, errFailedDecode, err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		c.logger.Error(errInvalidResponseFormatMsg)
		return nil, 0, fmt.Errorf(errInvalidResponseFormat)
	}

	links := c.parseLinksFromData(data)

	var total int64
	if v, ok := data["total"].(float64); ok {
		total = int64(v)
	}

	return links, total, nil
}

// parseLinksFromData extracts links from response data
func (c *ShortenerHTTPClient) parseLinksFromData(data map[string]interface{}) []domain.LinkResponse {
	var links []domain.LinkResponse
	if linksData, ok := data["links"].([]interface{}); ok {
		for _, item := range linksData {
			if linkMap, ok := item.(map[string]interface{}); ok {
				link := domain.LinkResponse{
					ID:          getString(linkMap, "id"),
					ShortCode:   getString(linkMap, "short_code"),
					OriginalURL: getString(linkMap, "original_url"),
					Title:       getString(linkMap, "title"),
					Description: getString(linkMap, "description"),
					IsActive:    getBool(linkMap, "is_active"),
				}
				if v, ok := linkMap["click_count"].(float64); ok {
					link.Clicks = int64(v)
				}
				links = append(links, link)
			}
		}
	}
	return links
}

// UpdateLink updates a link
func (c *ShortenerHTTPClient) UpdateLink(
	ctx context.Context,
	linkID string,
	req domain.CreateLinkRequest,
	userID, jwtToken string,
) (*domain.LinkResponse, error) {
	payload := make(map[string]interface{})

	if req.Title != nil {
		payload["title"] = *req.Title
	}
	if req.Description != nil {
		payload["description"] = *req.Description
	}
	if req.RedirectType != nil {
		payload["redirect_type"] = *req.RedirectType
	}
	if len(req.Tags) > 0 {
		payload["tags"] = req.Tags
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error(errFailedMarshal, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedMarshal, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/api/v1/shorten/"+linkID, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error(errFailedCreateRequest, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedCreateRequest, err)
	}

	httpReq.Header.Set(contentTypeHeader, contentTypeJSON)
	if jwtToken != "" {
		httpReq.Header.Set(authorizationHeader, bearerPrefix+jwtToken)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error(errFailedCallService, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error(errServiceError, zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return nil, fmt.Errorf(errServiceReturnedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(errFailedDecode, zap.Error(err))
		return nil, fmt.Errorf(errFormat, errFailedDecode, err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		c.logger.Error(errInvalidResponseFormatMsg)
		return nil, fmt.Errorf(errInvalidResponseFormat)
	}

	link := &domain.LinkResponse{
		ID:          getString(data, "id"),
		ShortCode:   getString(data, "short_code"),
		OriginalURL: getString(data, "original_url"),
		Title:       getString(data, "title"),
		Description: getString(data, "description"),
		IsActive:    getBool(data, "is_active"),
	}

	if v, ok := data["click_count"].(float64); ok {
		link.Clicks = int64(v)
	}

	return link, nil
}

// DeleteLink deletes a link
func (c *ShortenerHTTPClient) DeleteLink(ctx context.Context, linkID string, userID, jwtToken string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/shorten/"+linkID, nil)
	if err != nil {
		c.logger.Error(errFailedCreateRequest, zap.Error(err))
		return fmt.Errorf(errFormat, errFailedCreateRequest, err)
	}

	if jwtToken != "" {
		httpReq.Header.Set(authorizationHeader, bearerPrefix+jwtToken)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error(errFailedCallService, zap.Error(err))
		return fmt.Errorf(errFormat, errFailedCallService, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error(errServiceError, zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
		return fmt.Errorf(errServiceReturnedStatus, resp.StatusCode)
	}

	return nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
