package application

import (
	"context"
	"fmt"

	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/ports"
	"go.uber.org/zap"
)

// BFFService orchestrates calls to core services
type BFFService struct {
	shortenerClient ports.ShortenerClient
	analyticsClient ports.AnalyticsClient
	logger          *zap.Logger
}

// NewBFFService creates a new BFF service
func NewBFFService(
	shortenerClient ports.ShortenerClient,
	analyticsClient ports.AnalyticsClient,
	logger *zap.Logger,
) *BFFService {
	return &BFFService{
		shortenerClient: shortenerClient,
		analyticsClient: analyticsClient,
		logger:          logger,
	}
}

// CreateShortLink creates a new short link and returns the response
func (s *BFFService) CreateShortLink(
	ctx context.Context,
	req domain.CreateLinkRequest,
	workspaceID, userID, jwtToken string,
) (*domain.LinkResponse, error) {
	if workspaceID == "" || userID == "" {
		return nil, fmt.Errorf("workspace_id and user_id are required")
	}

	resp, err := s.shortenerClient.CreateLink(ctx, req, workspaceID, userID, jwtToken)
	if err != nil {
		s.logger.Error("failed to create short link",
			zap.String("workspace_id", workspaceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	return resp, nil
}

// GetShortLink retrieves a short link
func (s *BFFService) GetShortLink(ctx context.Context, shortCode, jwtToken string) (*domain.LinkResponse, error) {
	if shortCode == "" {
		return nil, fmt.Errorf("short_code is required")
	}

	resp, err := s.shortenerClient.GetLink(ctx, shortCode, jwtToken)
	if err != nil {
		s.logger.Error("failed to get short link",
			zap.String("short_code", shortCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get link: %w", err)
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
		return nil, 0, fmt.Errorf("workspace_id is required")
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
		return nil, 0, fmt.Errorf("failed to list links: %w", err)
	}

	return links, total, nil
}

// GetDashboard retrieves complete dashboard data by aggregating from multiple services
func (s *BFFService) GetDashboard(ctx context.Context, workspaceID string) (*domain.DashboardResponse, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace_id is required")
	}

	// Call analytics service for aggregated stats
	dashboard, err := s.analyticsClient.GetDashboardStats(ctx, workspaceID)
	if err != nil {
		s.logger.Error("failed to get dashboard stats",
			zap.String("workspace_id", workspaceID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	return dashboard, nil
}

// GetLinkAnalytics retrieves detailed analytics for a specific link
func (s *BFFService) GetLinkAnalytics(ctx context.Context, linkID string) (*domain.AnalyticsResponse, error) {
	if linkID == "" {
		return nil, fmt.Errorf("link_id is required")
	}

	analytics, err := s.analyticsClient.GetLinkAnalytics(ctx, linkID)
	if err != nil {
		s.logger.Error("failed to get link analytics",
			zap.String("link_id", linkID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return analytics, nil
}

// UpdateLink updates an existing link
func (s *BFFService) UpdateLink(
	ctx context.Context,
	linkID string,
	req domain.CreateLinkRequest,
	userID, jwtToken string,
) (*domain.LinkResponse, error) {
	if linkID == "" || userID == "" {
		return nil, fmt.Errorf("link_id and user_id are required")
	}

	resp, err := s.shortenerClient.UpdateLink(ctx, linkID, req, userID, jwtToken)
	if err != nil {
		s.logger.Error("failed to update link",
			zap.String("link_id", linkID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update link: %w", err)
	}

	return resp, nil
}

// DeleteLink deletes a link
func (s *BFFService) DeleteLink(ctx context.Context, linkID string, userID, jwtToken string) error {
	if linkID == "" || userID == "" {
		return fmt.Errorf("link_id and user_id are required")
	}

	err := s.shortenerClient.DeleteLink(ctx, linkID, userID, jwtToken)
	if err != nil {
		s.logger.Error("failed to delete link",
			zap.String("link_id", linkID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil
}
