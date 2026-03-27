package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/domain"
)

// Repository errors
var (
	ErrLinkNotFound    = errors.New("link not found")
	ErrLinkAlreadyExists = errors.New("link already exists")
	ErrCodeUnavailable = errors.New("short code is not available")
	ErrInvalidWorkspace = errors.New("link does not belong to workspace")
)

// Define data access methods
type LinkRepository interface {
	// Create creates a new short link
	Create(ctx context.Context, link *domain.ShortLink) error

	// FindByShortCode finds a link by its short code (public lookup)
	FindByShortCode(ctx context.Context, shortCode string) (*domain.ShortLink, error)

	// FindByID finds a link by ID within a specific workspace
	FindByID(ctx context.Context, workspaceID, linkID uuid.UUID) (*domain.ShortLink, error)

	// FindByCustomAlias finds a link by custom alias (if set)
	FindByCustomAlias(ctx context.Context, alias string) (*domain.ShortLink, error)

	// IsCodeAvailable checks if a short code is available for use
	IsCodeAvailable(ctx context.Context, code string) (bool, error)

	// ListByWorkspace lists all links in a workspace with optional filtering
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts ListOptions) ([]*domain.ShortLink, int64, error)

	// ListByCampaign lists all links in a campaign
	ListByCampaign(ctx context.Context, workspaceID, campaignID uuid.UUID, opts ListOptions) ([]*domain.ShortLink, int64, error)

	// Update updates an existing link
	Update(ctx context.Context, link *domain.ShortLink) error

	// Deactivate soft-deletes a link (marks as inactive)
	Deactivate(ctx context.Context, workspaceID, linkID uuid.UUID) error

	// Delete permanently deletes a link (hard delete)
	Delete(ctx context.Context, workspaceID, linkID uuid.UUID) error

	// IncrementClickCount increments the click count for a link
	IncrementClickCount(ctx context.Context, linkID uuid.UUID) error

	// UpdateLastAccess updates the last accessed timestamp
	UpdateLastAccess(ctx context.Context, linkID uuid.UUID) error

	// GetStats returns analytics stats for a link
	GetStats(ctx context.Context, workspaceID, linkID uuid.UUID) (*LinkStats, error)

	// GetWorkspaceStats returns aggregated stats for a workspace
	GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID) (*WorkspaceStats, error)

	// SearchByTag finds links with a specific tag in a workspace
	SearchByTag(ctx context.Context, workspaceID uuid.UUID, tag string, opts ListOptions) ([]*domain.ShortLink, int64, error)

	// ExpiringLinks returns links expiring within the given duration
	ExpiringLinks(ctx context.Context, workspaceID uuid.UUID, withinHours int) ([]*domain.ShortLink, error)

	// CountActiveLinks returns the count of active links in a workspace
	CountActiveLinks(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

// ListOptions provides filtering and pagination for list operations
type ListOptions struct {
	Limit  int    `json:"limit"`   // Max results (default: 20, max enforced by repository: 100)
	Offset int    `json:"offset"`  // Pagination offset
	Sort   string `json:"sort"`    // Sort field: "created_at", "click_count", "title", "last_accessed_at" (default: "created_at")
	Order  string `json:"order"`   // Sort order: "asc" or "desc", case-insensitive (default: "desc")
	Active *bool  `json:"active"`  // Filter by active status (nil = all)
}

// LinkStats contains analytics for a single link
type LinkStats struct {
	LinkID         uuid.UUID  `json:"link_id"`
	ShortCode      string     `json:"short_code"`
	ClickCount     int64      `json:"click_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// WorkspaceStats contains aggregated analytics for a workspace
type WorkspaceStats struct {
	WorkspaceID    uuid.UUID `json:"workspace_id"`
	TotalLinks     int64     `json:"total_links"`
	ActiveLinks    int64     `json:"active_links"`
	InactiveLinks  int64     `json:"inactive_links"`
	TotalClicks    int64     `json:"total_clicks"`
	AverageClicks  float64   `json:"average_clicks"`
	TopLink        *domain.ShortLink `json:"top_link,omitempty"`
	LastUpdated    time.Time `json:"last_updated"`
}
