package ports

import (
	"context"

	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/domain"
)

// ShortenerClient defines the interface for the shortener service
type ShortenerClient interface {
	CreateLink(ctx context.Context, req domain.CreateLinkRequest, workspaceID, userID, jwtToken string) (*domain.LinkResponse, error)
	GetLink(ctx context.Context, shortCode, jwtToken string) (*domain.LinkResponse, error)
	ListLinksInWorkspace(ctx context.Context, workspaceID string, page, pageSize int, jwtToken string) ([]domain.LinkResponse, int64, error)
	UpdateLink(ctx context.Context, linkID string, req domain.CreateLinkRequest, userID, jwtToken string) (*domain.LinkResponse, error)
	DeleteLink(ctx context.Context, linkID string, userID, jwtToken string) error
}

// AnalyticsClient defines the interface for the analytics service
type AnalyticsClient interface {
	GetDashboardStats(ctx context.Context, workspaceID string, jwtToken string) (*domain.DashboardResponse, error)
	GetLinkAnalytics(ctx context.Context, linkID string, jwtToken string) (*domain.AnalyticsResponse, error)
	GetLiveCount(ctx context.Context, shortCode string, jwtToken string) (int64, error)
}

// AuthClient defines the interface for auth service interactions
type AuthClient interface {
	ValidateToken(ctx context.Context, token string) (userID, workspaceID string, err error)
}
