package application

import (
	"context"
	"fmt"

	"github.com/bshongwe/linkpulse/bff/internal/domain"
	"github.com/bshongwe/linkpulse/bff/internal/ports"
	"go.uber.org/zap"
)

// Validation error messages
const (
	errWorkspaceIDRequired = "workspace_id is required"
	errUserIDRequired      = "user_id are required"
	errShortCodeRequired   = "short_code is required"
)

// Error message formats
const (
	errFailedCreateLink    = "failed to create link: %w"
	errFailedGetLink       = "failed to get link: %w"
	errFailedListLinks     = "failed to list links: %w"
	errFailedDeleteLink    = "failed to delete link: %w"
	errFailedGetAnalytics  = "failed to get analytics: %w"
)

// BFFService orchestrates calls to core backend services
type BFFService struct {
	shortenerClient ports.ShortenerClient
	analyticsClient ports.AnalyticsClient
	authClient      ports.AuthClient
	logger          *zap.Logger
}

// NewBFFService creates a new BFF service
func NewBFFService(
	shortenerClient ports.ShortenerClient,
	analyticsClient ports.AnalyticsClient,
	authClient ports.AuthClient,
	logger *zap.Logger,
) *BFFService {
	return &BFFService{
		shortenerClient: shortenerClient,
		analyticsClient: analyticsClient,
		authClient:      authClient,
		logger:          logger,
	}
}

// CreateShortLink creates a new short link
func (s *BFFService) CreateShortLink(
	ctx context.Context,
	req domain.CreateLinkRequest,
	workspaceID, userID, jwtToken string,
) (*domain.LinkResponse, error) {
	if workspaceID == "" || userID == "" {
		return nil, fmt.Errorf(errWorkspaceIDRequired + " and " + errUserIDRequired)
	}

	resp, err := s.shortenerClient.CreateLink(ctx, req, workspaceID, userID, jwtToken)
	if err != nil {
		s.logger.Error("failed to create short link",
			zap.String("workspace_id", workspaceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf(errFailedCreateLink, err)
	}

	return resp, nil
}

// GetShortLink retrieves a short link
func (s *BFFService) GetShortLink(ctx context.Context, shortCode, jwtToken string) (*domain.LinkResponse, error) {
	if shortCode == "" {
		return nil, fmt.Errorf(errShortCodeRequired)
	}

	resp, err := s.shortenerClient.GetLink(ctx, shortCode, jwtToken)
	if err != nil {
		s.logger.Error("failed to get short link",
			zap.String("short_code", shortCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf(errFailedGetLink, err)
	}

	return resp, nil
}

// ListLinks lists all links in a workspace
func (s *BFFService) ListLinks(
	ctx context.Context,
	workspaceID string,
	page, pageSize int,
	jwtToken string,
) ([]domain.LinkResponse, int64, error) {
	if workspaceID == "" {
		return nil, 0, fmt.Errorf(errWorkspaceIDRequired)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	links, total, err := s.shortenerClient.ListLinksInWorkspace(ctx, workspaceID, page, pageSize, jwtToken)
	if err != nil {
		s.logger.Error("failed to list links",
			zap.String("workspace_id", workspaceID),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf(errFailedListLinks, err)
	}

	return links, total, nil
}

// DeleteShortLink deletes a short link
func (s *BFFService) DeleteShortLink(ctx context.Context, shortCode, jwtToken string) error {
	if shortCode == "" {
		return fmt.Errorf(errShortCodeRequired)
	}

	err := s.shortenerClient.DeleteLink(ctx, shortCode, jwtToken)
	if err != nil {
		s.logger.Error("failed to delete short link",
			zap.String("short_code", shortCode),
			zap.Error(err),
		)
		return fmt.Errorf(errFailedDeleteLink, err)
	}

	return nil
}

// GetLinkAnalytics retrieves analytics for a link
func (s *BFFService) GetLinkAnalytics(ctx context.Context, shortCode, jwtToken string) (*domain.AnalyticsResponse, error) {
	if shortCode == "" {
		return nil, fmt.Errorf(errShortCodeRequired)
	}

	resp, err := s.analyticsClient.GetLinkAnalytics(ctx, shortCode, jwtToken)
	if err != nil {
		s.logger.Error("failed to get link analytics",
			zap.String("short_code", shortCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf(errFailedGetAnalytics, err)
	}

	return resp, nil
}

// GetWorkspaceAnalytics retrieves analytics for a workspace
func (s *BFFService) GetWorkspaceAnalytics(ctx context.Context, workspaceID, jwtToken string) (*domain.AnalyticsResponse, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf(errWorkspaceIDRequired)
	}

	resp, err := s.analyticsClient.GetWorkspaceAnalytics(ctx, workspaceID, jwtToken)
	if err != nil {
		s.logger.Error("failed to get workspace analytics",
			zap.String("workspace_id", workspaceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf(errFailedGetAnalytics, err)
	}

	return resp, nil
}
