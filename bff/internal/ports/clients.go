package ports

import (
	"context"

	"github.com/bshongwe/linkpulse/bff/internal/domain"
)

// ShortenerClient defines interface for shortener service
type ShortenerClient interface {
	CreateLink(ctx context.Context, req domain.CreateLinkRequest, workspaceID, userID, jwtToken string) (*domain.LinkResponse, error)
	GetLink(ctx context.Context, shortCode, jwtToken string) (*domain.LinkResponse, error)
	ListLinksInWorkspace(ctx context.Context, workspaceID string, page, pageSize int, jwtToken string) ([]domain.LinkResponse, int64, error)
	DeleteLink(ctx context.Context, shortCode, jwtToken string) error
}

// AnalyticsClient defines interface for analytics service
type AnalyticsClient interface {
	GetLinkAnalytics(ctx context.Context, shortCode string, jwtToken string) (*domain.AnalyticsResponse, error)
	GetWorkspaceAnalytics(ctx context.Context, workspaceID string, jwtToken string) (*domain.AnalyticsResponse, error)
}

// AuthClient defines interface for auth service
type AuthClient interface {
	ValidateToken(ctx context.Context, token string) (map[string]interface{}, error)
	GetUserInfo(ctx context.Context, userID, jwtToken string) (map[string]interface{}, error)
}
