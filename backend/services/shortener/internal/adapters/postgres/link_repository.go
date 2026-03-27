package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/ports"
	sharedErrors "github.com/bshongwe/linkpulse/backend/shared/errors"
)

// LinkRepository implements ports.LinkRepository using PostgreSQL
type LinkRepository struct {
	db *sqlx.DB
}

const (
	errLinkNotFound          = "link not found"
	errFailedGetRowsAffected = "failed to get rows affected: %w"
)

// NewLinkRepository creates a new PostgreSQL link repository
func NewLinkRepository(db *sqlx.DB) ports.LinkRepository {
	return &LinkRepository{db: db}
}

// Create inserts a new short link into the database
func (r *LinkRepository) Create(ctx context.Context, link *domain.ShortLink) error {
	query := `
		INSERT INTO links (
			id, workspace_id, short_code, original_url, created_by,
			title, description, expires_at, is_active,
			click_count, last_accessed_at, redirect_type,
			qr_code, qr_code_url, tags, campaign_id,
			created_at, updated_at
		) VALUES (
			:id, :workspace_id, :short_code, :original_url, :created_by,
			:title, :description, :expires_at, :is_active,
			:click_count, :last_accessed_at, :redirect_type,
			:qr_code, :qr_code_url, :tags, :campaign_id,
			:created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":              link.ID,
		"workspace_id":    link.WorkspaceID,
		"short_code":      link.ShortCode,
		"original_url":    link.OriginalURL,
		"created_by":      link.CreatedBy,
		"title":           link.Title,
		"description":     link.Description,
		"expires_at":      link.ExpiresAt,
		"is_active":       link.IsActive,
		"click_count":     link.ClickCount,
		"last_accessed_at": link.LastAccessedAt,
		"redirect_type":   link.RedirectType,
		"qr_code":         link.QRCode,
		"qr_code_url":     link.QRCodeURL,
		"tags":            link.Tags,
		"campaign_id":     link.CampaignID,
		"created_at":      link.CreatedAt,
		"updated_at":      link.UpdatedAt,
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return sharedErrors.New(sharedErrors.ErrAlreadyExists, "short code already exists")
		}
		return fmt.Errorf("failed to create link: %w", err)
	}

	return nil
}

// FindByShortCode retrieves a link by short code (public access, no workspace check)
func (r *LinkRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.ShortLink, error) {
	query := `
		SELECT
			id, workspace_id, short_code, original_url, created_by,
			title, description, expires_at, is_active,
			click_count, last_accessed_at, redirect_type,
			qr_code, qr_code_url, tags, campaign_id,
			created_at, updated_at
		FROM links
		WHERE short_code = $1 AND is_active = true
	`

	link := &domain.ShortLink{}
	err := r.db.GetContext(ctx, link, query, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
		}
		return nil, fmt.Errorf("failed to find link: %w", err)
	}

	return link, nil
}

// FindByID retrieves a link by ID with workspace scope check
func (r *LinkRepository) FindByID(ctx context.Context, workspaceID, linkID uuid.UUID) (*domain.ShortLink, error) {
	query := `
		SELECT
			id, workspace_id, short_code, original_url, created_by,
			title, description, expires_at, is_active,
			click_count, last_accessed_at, redirect_type,
			qr_code, qr_code_url, tags, campaign_id,
			created_at, updated_at
		FROM links
		WHERE id = $1 AND workspace_id = $2
	`

	link := &domain.ShortLink{}
	err := r.db.GetContext(ctx, link, query, linkID, workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
		}
		return nil, fmt.Errorf("failed to find link: %w", err)
	}

	return link, nil
}

// FindByCustomAlias finds a link by custom alias (globally, not workspace-scoped)
func (r *LinkRepository) FindByCustomAlias(ctx context.Context, alias string) (*domain.ShortLink, error) {
	query := `
		SELECT
			id, workspace_id, short_code, original_url, created_by,
			title, description, expires_at, is_active,
			click_count, last_accessed_at, redirect_type,
			qr_code, qr_code_url, tags, campaign_id,
			created_at, updated_at
		FROM links
		WHERE short_code = $1
	`

	link := &domain.ShortLink{}
	err := r.db.GetContext(ctx, link, query, alias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
		}
		return nil, fmt.Errorf("failed to find link by alias: %w", err)
	}

	return link, nil
}

// IsCodeAvailable checks if a short code is available globally
func (r *LinkRepository) IsCodeAvailable(ctx context.Context, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM links WHERE short_code = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, code)
	if err != nil {
		return false, fmt.Errorf("failed to check code availability: %w", err)
	}

	return !exists, nil
}

// Update updates an existing link
func (r *LinkRepository) Update(ctx context.Context, link *domain.ShortLink) error {
	query := `
		UPDATE links
		SET
			title = :title,
			description = :description,
			expires_at = :expires_at,
			is_active = :is_active,
			redirect_type = :redirect_type,
			tags = :tags,
			campaign_id = :campaign_id,
			updated_at = :updated_at
		WHERE id = :id AND workspace_id = :workspace_id
	`

	result, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":           link.ID,
		"workspace_id": link.WorkspaceID,
		"title":        link.Title,
		"description":  link.Description,
		"expires_at":   link.ExpiresAt,
		"is_active":    link.IsActive,
		"redirect_type": link.RedirectType,
		"tags":         link.Tags,
		"campaign_id":  link.CampaignID,
		"updated_at":   link.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errFailedGetRowsAffected, err)
	}

	if rows == 0 {
		return sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}

	return nil
}

// Deactivate soft-deletes a link by marking it inactive
func (r *LinkRepository) Deactivate(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	query := `
		UPDATE links
		SET is_active = false, updated_at = $1
		WHERE id = $2 AND workspace_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), linkID, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to deactivate link: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errFailedGetRowsAffected, err)
	}

	if rows == 0 {
		return sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}

	return nil
}

// Delete permanently removes a link
func (r *LinkRepository) Delete(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	query := `DELETE FROM links WHERE id = $1 AND workspace_id = $2`

	result, err := r.db.ExecContext(ctx, query, linkID, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errFailedGetRowsAffected, err)
	}

	if rows == 0 {
		return sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}

	return nil
}

// IncrementClickCount increments the click count for a link
func (r *LinkRepository) IncrementClickCount(ctx context.Context, linkID uuid.UUID) error {
	query := `UPDATE links SET click_count = click_count + 1 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, linkID)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errFailedGetRowsAffected, err)
	}

	if rows == 0 {
		return sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}

	return nil
}

// UpdateLastAccess updates the last accessed timestamp for a link
func (r *LinkRepository) UpdateLastAccess(ctx context.Context, linkID uuid.UUID) error {
	query := `UPDATE links SET last_accessed_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), linkID)
	if err != nil {
		return fmt.Errorf("failed to update last access: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errFailedGetRowsAffected, err)
	}

	if rows == 0 {
		return sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}

	return nil
}

// GetStats retrieves analytics for a single link
func (r *LinkRepository) GetStats(ctx context.Context, workspaceID, linkID uuid.UUID) (*ports.LinkStats, error) {
	query := `
		SELECT
			id, short_code, click_count, created_at, updated_at, last_accessed_at
		FROM links
		WHERE id = $1 AND workspace_id = $2
	`

	var link domain.ShortLink
	err := r.db.GetContext(ctx, &link, query, linkID, workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
		}
		return nil, fmt.Errorf("failed to get link stats: %w", err)
	}

	stats := &ports.LinkStats{
		LinkID:         link.ID,
		ShortCode:      link.ShortCode,
		ClickCount:     link.ClickCount,
		CreatedAt:      link.CreatedAt,
		UpdatedAt:      link.UpdatedAt,
		LastAccessedAt: link.LastAccessedAt,
	}

	return stats, nil
}

// GetWorkspaceStats retrieves aggregate analytics for a workspace
func (r *LinkRepository) GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID) (*ports.WorkspaceStats, error) {
	query := `
		SELECT
			COUNT(*) as total_links,
			COALESCE(SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END), 0) as active_links,
			COALESCE(SUM(click_count), 0) as total_clicks,
			MAX(updated_at) as last_updated
		FROM links
		WHERE workspace_id = $1
	`

	var stats struct {
		TotalLinks   int64      `db:"total_links"`
		ActiveLinks  int64      `db:"active_links"`
		TotalClicks  int64      `db:"total_clicks"`
		LastUpdated  *time.Time `db:"last_updated"`
	}

	err := r.db.GetContext(ctx, &stats, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace stats: %w", err)
	}

	inactiveLinks := stats.TotalLinks - stats.ActiveLinks
	avgClicks := 0.0
	if stats.TotalLinks > 0 {
		avgClicks = float64(stats.TotalClicks) / float64(stats.TotalLinks)
	}

	lastUpdated := time.Now()
	if stats.LastUpdated != nil {
		lastUpdated = *stats.LastUpdated
	}

	return &ports.WorkspaceStats{
		WorkspaceID:   workspaceID,
		TotalLinks:    stats.TotalLinks,
		ActiveLinks:   stats.ActiveLinks,
		InactiveLinks: inactiveLinks,
		TotalClicks:   stats.TotalClicks,
		AverageClicks: avgClicks,
		LastUpdated:   lastUpdated,
	}, nil
}

// ListByWorkspace lists all links in a workspace with pagination
func (r *LinkRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts ports.ListOptions) ([]*domain.ShortLink, int64, error) {
	// Set defaults
	if opts.Limit == 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM links WHERE workspace_id = $1`
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, workspaceID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count links: %w", err)
	}

	// Build query with sorting
	sortField := "created_at"
	sortOrder := "DESC"
	if opts.Sort != "" {
		// Validate sort field to prevent SQL injection
		validFields := map[string]bool{
			"created_at":      true,
			"title":           true,
			"click_count":     true,
			"last_accessed_at": true,
		}
		if validFields[opts.Sort] {
			sortField = opts.Sort
		}
	}
	if opts.Order != "" {
		switch strings.ToUpper(strings.TrimSpace(opts.Order)) {
		case "ASC":
			sortOrder = "ASC"
		case "DESC":
			sortOrder = "DESC"
		}
	}

	query := fmt.Sprintf(`
		SELECT
			id, workspace_id, short_code, original_url, created_by,
			title, description, expires_at, is_active,
			click_count, last_accessed_at, redirect_type,
			qr_code, qr_code_url, tags, campaign_id,
			created_at, updated_at
		FROM links
		WHERE workspace_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortField, sortOrder)

	links := []*domain.ShortLink{}
	err = r.db.SelectContext(ctx, &links, query, workspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list links: %w", err)
	}

	return links, total, nil
}

// ListByCampaign lists all links in a campaign
func (r *LinkRepository) ListByCampaign(ctx context.Context, workspaceID, campaignID uuid.UUID, opts ports.ListOptions) ([]*domain.ShortLink, int64, error) {
	// Set defaults
	if opts.Limit == 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM links WHERE workspace_id = $1 AND campaign_id = $2`
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, workspaceID, campaignID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count campaign links: %w", err)
	}

	// Build query
	sortField := "created_at"
	sortOrder := "DESC"
	if opts.Sort != "" {
		validFields := map[string]bool{
			"created_at":      true,
			"title":           true,
			"click_count":     true,
			"last_accessed_at": true,
		}
		if validFields[opts.Sort] {
			sortField = opts.Sort
		}
	}
	if opts.Order != "" {
		switch strings.ToUpper(strings.TrimSpace(opts.Order)) {
		case "ASC":
			sortOrder = "ASC"
		case "DESC":
			sortOrder = "DESC"
		}
	}

	query := fmt.Sprintf(`
		SELECT
			id, workspace_id, short_code, original_url, created_by,
			title, description, expires_at, is_active,
			click_count, last_accessed_at, redirect_type,
			qr_code, qr_code_url, tags, campaign_id,
			created_at, updated_at
		FROM links
		WHERE workspace_id = $1 AND campaign_id = $2
		ORDER BY %s %s
		LIMIT $3 OFFSET $4
	`, sortField, sortOrder)

	links := []*domain.ShortLink{}
	err = r.db.SelectContext(ctx, &links, query, workspaceID, campaignID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list campaign links: %w", err)
	}

	return links, total, nil
}

// SearchByTag finds links with a specific tag
func (r *LinkRepository) SearchByTag(ctx context.Context, workspaceID uuid.UUID, tag string, opts ports.ListOptions) ([]*domain.ShortLink, int64, error) {
	// Set defaults
	if opts.Limit == 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	// Get total count using PostgreSQL array containment
	countQuery := `
		SELECT COUNT(*) FROM links
		WHERE workspace_id = $1 AND $2 = ANY(tags)
	`
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, workspaceID, tag)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tagged links: %w", err)
	}

	// Build query
	sortField := "created_at"
	sortOrder := "DESC"
	if opts.Sort != "" {
		validFields := map[string]bool{
			"created_at":      true,
			"title":           true,
			"click_count":     true,
			"last_accessed_at": true,
		}
		if validFields[opts.Sort] {
			sortField = opts.Sort
		}
	}
	if opts.Order != "" {
		switch strings.ToUpper(strings.TrimSpace(opts.Order)) {
		case "ASC":
			sortOrder = "ASC"
		case "DESC":
			sortOrder = "DESC"
		}
	}

	query := fmt.Sprintf(`
		SELECT
			id, workspace_id, short_code, original_url, created_by,
			title, description, expires_at, is_active,
			click_count, last_accessed_at, redirect_type,
			qr_code, qr_code_url, tags, campaign_id,
			created_at, updated_at
		FROM links
		WHERE workspace_id = $1 AND $2 = ANY(tags)
		ORDER BY %s %s
		LIMIT $3 OFFSET $4
	`, sortField, sortOrder)

	links := []*domain.ShortLink{}
	err = r.db.SelectContext(ctx, &links, query, workspaceID, tag, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search by tag: %w", err)
	}

	return links, total, nil
}

// ExpiringLinks retrieves links that expire within a certain number of hours
func (r *LinkRepository) ExpiringLinks(ctx context.Context, workspaceID uuid.UUID, withinHours int) ([]*domain.ShortLink, error) {
	query := fmt.Sprintf(`
		SELECT
			id, workspace_id, short_code, original_url, created_by,
			title, description, expires_at, is_active,
			click_count, last_accessed_at, redirect_type,
			qr_code, qr_code_url, tags, campaign_id,
			created_at, updated_at
		FROM links
		WHERE
			workspace_id = $1
			AND expires_at IS NOT NULL
			AND expires_at > NOW()
			AND expires_at <= NOW() + '%d hours'::INTERVAL
			AND is_active = true
		ORDER BY expires_at ASC
	`, withinHours)

	links := []*domain.ShortLink{}
	err := r.db.SelectContext(ctx, &links, query, workspaceID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get expiring links: %w", err)
	}

	return links, nil
}

// CountActiveLinks returns the count of active links in a workspace
func (r *LinkRepository) CountActiveLinks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*) FROM links
		WHERE workspace_id = $1 AND is_active = true
	`

	var count int64
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	if err != nil {
		return 0, fmt.Errorf("failed to count active links: %w", err)
	}

	return count, nil
}
