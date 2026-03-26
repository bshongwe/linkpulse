package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/ports"
	"github.com/bshongwe/linkpulse/backend/shared/errors"
)

// ShortenerService handles URL shortening business logic
type ShortenerService struct {
	linkRepo ports.LinkRepository
	cache    ports.CachePort // for fast redirects and analytics caching
}

const (
	cacheKeyPrefix = "short:"
)

// NewShortenerService creates a new shortener service
func NewShortenerService(linkRepo ports.LinkRepository, cache ports.CachePort) *ShortenerService {
	return &ShortenerService{
		linkRepo: linkRepo,
		cache:    cache,
	}
}

// CreateShortLink creates a new shortened URL
func (s *ShortenerService) CreateShortLink(
	ctx context.Context,
	req *domain.CreateShortLinkRequest,
	userID, workspaceID uuid.UUID,
) (*domain.ShortLink, error) {
	// Determine short code
	shortCode := req.CustomAlias
	if shortCode == "" {
		shortCode = generateShortCode()
	}

	// Check availability
	available, err := s.linkRepo.IsCodeAvailable(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to check code availability: %w", err)
	}
	if !available {
		return nil, errors.New(errors.ErrAlreadyExists, "short code already taken")
	}

	// Validate redirect type
	redirectType := req.RedirectType
	if redirectType == "" {
		redirectType = domain.RedirectTemporary
	}

	// Create link entity
	link := &domain.ShortLink{
		ID:           uuid.New(),
		ShortCode:    shortCode,
		OriginalURL:  req.OriginalURL,
		WorkspaceID:  workspaceID,
		CreatedBy:    userID,
		Title:        req.Title,
		Description:  req.Description,
		ExpiresAt:    req.ExpiresAt,
		RedirectType: redirectType,
		IsActive:     true,
		ClickCount:   0,
		Tags:         req.Tags,
		CampaignID:   req.CampaignID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Persist to database
	if err := s.linkRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create short link: %w", err)
	}

	// Cache the mapping for fast redirects (24 hour TTL)
	cacheKey := fmt.Sprintf("%s%s", cacheKeyPrefix, shortCode)
	s.cache.Set(ctx, cacheKey, req.OriginalURL, 24*time.Hour)

	return link, nil
}

// GetShortLink retrieves a link by short code and records analytics
func (s *ShortenerService) GetShortLink(ctx context.Context, shortCode string) (*domain.ShortLink, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("%s%s", cacheKeyPrefix, shortCode)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		// Cache hit - still fetch from DB for complete data and analytics
		// but at least we can serve the redirect quickly
		return s.getLinkFromDB(ctx, shortCode)
	}

	return s.getLinkFromDB(ctx, shortCode)
}

// getLinkFromDB retrieves a link from database
func (s *ShortenerService) getLinkFromDB(ctx context.Context, shortCode string) (*domain.ShortLink, error) {
	link, err := s.linkRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to find short link: %w", err)
	}

	// Check if link can be accessed
	if !link.CanAccess() {
		return nil, errors.New(errors.ErrNotFound, "link is inactive or expired")
	}

	return link, nil
}

// RecordAccess records a click/access to a link
func (s *ShortenerService) RecordAccess(ctx context.Context, linkID uuid.UUID) error {
	// Increment click count in database
	if err := s.linkRepo.IncrementClickCount(ctx, linkID); err != nil {
		// Log but don't fail - analytics is secondary
		fmt.Printf("warning: failed to increment click count for link %s: %v\n", linkID, err)
	}

	// Update last accessed time
	if err := s.linkRepo.UpdateLastAccess(ctx, linkID); err != nil {
		// Log but don't fail
		fmt.Printf("warning: failed to update last access for link %s: %v\n", linkID, err)
	}

	return nil
}

// UpdateShortLink updates an existing link
func (s *ShortenerService) UpdateShortLink(
	ctx context.Context,
	workspaceID, linkID uuid.UUID,
	req *domain.UpdateShortLinkRequest,
) (*domain.ShortLink, error) {
	// Fetch existing link
	link, err := s.linkRepo.FindByID(ctx, workspaceID, linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to find link: %w", err)
	}

	// Update fields
	if req.Title != "" {
		link.Title = req.Title
	}
	if req.Description != "" {
		link.Description = req.Description
	}
	if req.ExpiresAt != nil {
		link.ExpiresAt = req.ExpiresAt
	}
	if req.RedirectType != "" {
		link.RedirectType = req.RedirectType
	}
	link.IsActive = req.IsActive
	link.Tags = req.Tags
	link.CampaignID = req.CampaignID
	link.UpdatedAt = time.Now()

	// Persist changes
	if err := s.linkRepo.Update(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to update link: %w", err)
	}

	return link, nil
}

// DeactivateLink soft-deletes a link
func (s *ShortenerService) DeactivateLink(
	ctx context.Context,
	workspaceID, linkID uuid.UUID,
) error {
	if err := s.linkRepo.Deactivate(ctx, workspaceID, linkID); err != nil {
		return fmt.Errorf("failed to deactivate link: %w", err)
	}

	// Invalidate cache
	link, _ := s.linkRepo.FindByID(ctx, workspaceID, linkID)
	if link != nil {
		cacheKey := fmt.Sprintf("%s%s", cacheKeyPrefix, link.ShortCode)
		s.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// DeleteLink permanently deletes a link
func (s *ShortenerService) DeleteLink(
	ctx context.Context,
	workspaceID, linkID uuid.UUID,
) error {
	// Get link before deletion for cache cleanup
	link, _ := s.linkRepo.FindByID(ctx, workspaceID, linkID)

	if err := s.linkRepo.Delete(ctx, workspaceID, linkID); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	// Invalidate cache
	if link != nil {
		cacheKey := fmt.Sprintf("%s%s", cacheKeyPrefix, link.ShortCode)
		s.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// GetLinkStats retrieves analytics for a link
func (s *ShortenerService) GetLinkStats(
	ctx context.Context,
	workspaceID, linkID uuid.UUID,
) (*ports.LinkStats, error) {
	stats, err := s.linkRepo.GetStats(ctx, workspaceID, linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch link stats: %w", err)
	}

	return stats, nil
}

// ListLinksInWorkspace lists all links in a workspace
func (s *ShortenerService) ListLinksInWorkspace(
	ctx context.Context,
	workspaceID uuid.UUID,
	opts ports.ListOptions,
) ([]*domain.ShortLink, int64, error) {
	links, total, err := s.linkRepo.ListByWorkspace(ctx, workspaceID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list links: %w", err)
	}

	return links, total, nil
}

// ListLinksByCampaign lists all links in a campaign
func (s *ShortenerService) ListLinksByCampaign(
	ctx context.Context,
	workspaceID, campaignID uuid.UUID,
	opts ports.ListOptions,
) ([]*domain.ShortLink, int64, error) {
	links, total, err := s.linkRepo.ListByCampaign(ctx, workspaceID, campaignID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list campaign links: %w", err)
	}

	return links, total, nil
}

// SearchByTag finds links with a specific tag
func (s *ShortenerService) SearchByTag(
	ctx context.Context,
	workspaceID uuid.UUID,
	tag string,
	opts ports.ListOptions,
) ([]*domain.ShortLink, int64, error) {
	links, total, err := s.linkRepo.SearchByTag(ctx, workspaceID, tag, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search by tag: %w", err)
	}

	return links, total, nil
}

// GenerateShortCode generates a random 8-character short code
// Uses URL-safe base64 encoding to ensure compatibility
func generateShortCode() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based code if random fails
		return fmt.Sprintf("link%d", time.Now().UnixNano())
	}
	// Encode to base64 URL-safe and take first 8 chars
	return base64.RawURLEncoding.EncodeToString(b)[:8]
}
