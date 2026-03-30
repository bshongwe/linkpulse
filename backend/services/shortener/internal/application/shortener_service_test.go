package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/application"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/ports"
	sharedErrors "github.com/bshongwe/linkpulse/backend/shared/errors"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
)

const (
	errLinkNotFound       = "link not found"
	exampleURL            = "https://example.com"
	errUnexpectedErrorFmt = "unexpected error: %v"
	cacheKeyPrefix        = "short:"
)

func init() {
	// ShortenerService uses logger.Log — initialise it once for all tests.
	logger.Init("test")
}

// --- mock LinkRepository ---

type mockRepo struct {
	mu     sync.Mutex
	links  map[string]*domain.ShortLink // keyed by short_code
	byID   map[uuid.UUID]*domain.ShortLink
	failOn string // method name to force an error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		links: make(map[string]*domain.ShortLink),
		byID:  make(map[uuid.UUID]*domain.ShortLink),
	}
}

func (r *mockRepo) Create(ctx context.Context, link *domain.ShortLink) error {
	if r.failOn == "Create" {
		return errors.New("mock create error")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ShortCode] = link
	r.byID[link.ID] = link
	return nil
}

func (r *mockRepo) FindByShortCode(ctx context.Context, code string) (*domain.ShortLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.links[code]
	if !ok {
		return nil, sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}
	return l, nil
}

func (r *mockRepo) FindByID(ctx context.Context, workspaceID, linkID uuid.UUID) (*domain.ShortLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[linkID]
	if !ok {
		return nil, sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}
	return l, nil
}

func (r *mockRepo) FindByCustomAlias(ctx context.Context, alias string) (*domain.ShortLink, error) {
	return r.FindByShortCode(ctx, alias)
}

func (r *mockRepo) IsCodeAvailable(ctx context.Context, code string) (bool, error) {
	if r.failOn == "IsCodeAvailable" {
		return false, errors.New("mock availability error")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_, taken := r.links[code]
	return !taken, nil
}

func (r *mockRepo) Update(ctx context.Context, link *domain.ShortLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[link.ID]; !ok {
		return sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}
	r.links[link.ShortCode] = link
	r.byID[link.ID] = link
	return nil
}

func (r *mockRepo) Deactivate(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[linkID]
	if !ok {
		return sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}
	l.IsActive = false
	return nil
}

func (r *mockRepo) Delete(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[linkID]
	if !ok {
		return sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}
	delete(r.links, l.ShortCode)
	delete(r.byID, linkID)
	return nil
}

func (r *mockRepo) IncrementClickCount(ctx context.Context, linkID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if l, ok := r.byID[linkID]; ok {
		l.ClickCount++
	}
	return nil
}

func (r *mockRepo) UpdateLastAccess(ctx context.Context, linkID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if l, ok := r.byID[linkID]; ok {
		now := time.Now()
		l.LastAccessedAt = &now
	}
	return nil
}

func (r *mockRepo) GetStats(ctx context.Context, workspaceID, linkID uuid.UUID) (*ports.LinkStats, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[linkID]
	if !ok {
		return nil, sharedErrors.New(sharedErrors.ErrNotFound, errLinkNotFound)
	}
	return &ports.LinkStats{
		LinkID:     l.ID,
		ShortCode:  l.ShortCode,
		ClickCount: l.ClickCount,
		CreatedAt:  l.CreatedAt,
		UpdatedAt:  l.UpdatedAt,
	}, nil
}

func (r *mockRepo) GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID) (*ports.WorkspaceStats, error) {
	return &ports.WorkspaceStats{WorkspaceID: workspaceID}, nil
}

func (r *mockRepo) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts ports.ListOptions) ([]*domain.ShortLink, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.ShortLink
	for _, l := range r.byID {
		if l.WorkspaceID == workspaceID {
			out = append(out, l)
		}
	}
	return out, int64(len(out)), nil
}

func (r *mockRepo) ListByCampaign(ctx context.Context, workspaceID, campaignID uuid.UUID, opts ports.ListOptions) ([]*domain.ShortLink, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.ShortLink
	for _, l := range r.byID {
		if l.CampaignID != nil && *l.CampaignID == campaignID {
			out = append(out, l)
		}
	}
	return out, int64(len(out)), nil
}

func (r *mockRepo) SearchByTag(ctx context.Context, workspaceID uuid.UUID, tag string, opts ports.ListOptions) ([]*domain.ShortLink, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.ShortLink
	for _, l := range r.byID {
		for _, t := range l.Tags {
			if t == tag {
				out = append(out, l)
				break
			}
		}
	}
	return out, int64(len(out)), nil
}

func (r *mockRepo) ExpiringLinks(ctx context.Context, workspaceID uuid.UUID, withinHours int) ([]*domain.ShortLink, error) {
	return nil, nil
}

func (r *mockRepo) CountActiveLinks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	return 0, nil
}

// --- mock CachePort ---

type mockCache struct {
	mu    sync.Mutex
	store map[string]interface{}
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string]interface{})}
}

func (c *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
	return nil
}

func (c *mockCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.store[key]
	if !ok {
		return nil, errors.New("cache miss")
	}
	return v, nil
}

func (c *mockCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
	return nil
}

func (c *mockCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.store[key]
	return ok, nil
}

func (c *mockCache) IncrementInt(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, nil
}

func (c *mockCache) SetWithoutTTL(ctx context.Context, key string, value interface{}) error {
	return c.Set(ctx, key, value, 0)
}

// --- helpers ---

func newService() (*application.ShortenerService, *mockRepo, *mockCache) {
	repo := newMockRepo()
	cache := newMockCache()
	svc := application.NewShortenerService(repo, cache)
	return svc, repo, cache
}

func createReq(url string) *domain.CreateShortLinkRequest {
	return &domain.CreateShortLinkRequest{
		OriginalURL: url,
	}
}

// --- tests ---

func TestCreateShortLinkSuccess(t *testing.T) {
	svc, _, cache := newService()
	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()

	link, err := svc.CreateShortLink(ctx, createReq(exampleURL), userID, workspaceID)
	if err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if link.ShortCode == "" {
		t.Error("expected non-empty ShortCode")
	}
	if link.OriginalURL != exampleURL {
		t.Errorf("OriginalURL = %q, want %q", link.OriginalURL, exampleURL)
	}
	if link.WorkspaceID != workspaceID {
		t.Error("WorkspaceID mismatch")
	}
	if !link.IsActive {
		t.Error("expected link to be active")
	}
	if link.RedirectType != domain.RedirectTemporary {
		t.Errorf("RedirectType = %q, want %q", link.RedirectType, domain.RedirectTemporary)
	}

	// Cache should be populated
	cacheKey := cacheKeyPrefix + link.ShortCode
	if _, err := cache.Get(ctx, cacheKey); err != nil {
		t.Error("expected cache to be populated after create")
	}
}

func TestCreateShortLinkCustomAlias(t *testing.T) {
	svc, _, _ := newService()
	req := createReq(exampleURL)
	req.CustomAlias = "myalias"

	link, err := svc.CreateShortLink(context.Background(), req, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if link.ShortCode != "myalias" {
		t.Errorf("ShortCode = %q, want %q", link.ShortCode, "myalias")
	}
}

func TestCreateShortLinkDuplicateCode(t *testing.T) {
	svc, _, _ := newService()
	ctx := context.Background()
	req := createReq(exampleURL)
	req.CustomAlias = "taken"

	if _, err := svc.CreateShortLink(ctx, req, uuid.New(), uuid.New()); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	req2 := createReq("https://other.com")
	req2.CustomAlias = "taken"
	_, err := svc.CreateShortLink(ctx, req2, uuid.New(), uuid.New())
	if !sharedErrors.IsAlreadyExists(err) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestCreateShortLinkRepoError(t *testing.T) {
	svc, repo, _ := newService()
	repo.failOn = "Create"

	_, err := svc.CreateShortLink(context.Background(), createReq(exampleURL), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error when repo.Create fails")
	}
}

func TestGetShortLinkFound(t *testing.T) {
	svc, _, _ := newService()
	ctx := context.Background()

	created, _ := svc.CreateShortLink(ctx, createReq(exampleURL), uuid.New(), uuid.New())

	got, err := svc.GetShortLink(ctx, created.ShortCode)
	if err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if got.ShortCode != created.ShortCode {
		t.Errorf("ShortCode = %q, want %q", got.ShortCode, created.ShortCode)
	}
}

func TestGetShortLinkNotFound(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.GetShortLink(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing short code")
	}
}

func TestGetShortLinkInactiveLink(t *testing.T) {
	svc, repo, _ := newService()
	ctx := context.Background()

	created, _ := svc.CreateShortLink(ctx, createReq(exampleURL), uuid.New(), uuid.New())
	repo.links[created.ShortCode].IsActive = false

	_, err := svc.GetShortLink(ctx, created.ShortCode)
	if !sharedErrors.IsNotFound(err) {
		t.Errorf("expected ErrNotFound for inactive link, got %v", err)
	}
}

func TestGetShortLinkExpiredLink(t *testing.T) {
	svc, repo, _ := newService()
	ctx := context.Background()

	created, _ := svc.CreateShortLink(ctx, createReq(exampleURL), uuid.New(), uuid.New())
	past := time.Now().Add(-time.Second)
	repo.links[created.ShortCode].ExpiresAt = &past

	// Service should return the expired link (handler will return 410)
	link, err := svc.GetShortLink(ctx, created.ShortCode)
	if err != nil {
		t.Errorf("expected no error for expired link, got %v", err)
	}
	if link == nil {
		t.Fatal("expected link to be returned, got nil")
	}
	if link.IsExpired() == false {
		t.Error("expected link to be expired")
	}
}

func TestUpdateShortLink(t *testing.T) {
	svc, _, _ := newService()
	ctx := context.Background()
	workspaceID := uuid.New()

	created, _ := svc.CreateShortLink(ctx, createReq(exampleURL), uuid.New(), workspaceID)

	newTitle := "Updated Title"
	updated, err := svc.UpdateShortLink(ctx, workspaceID, created.ID, &domain.UpdateShortLinkRequest{
		Title: newTitle,
	})
	if err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if updated.Title != newTitle {
		t.Errorf("Title = %q, want %q", updated.Title, newTitle)
	}
}

func TestUpdateShortLinkIsActivePointer(t *testing.T) {
	svc, _, _ := newService()
	ctx := context.Background()
	workspaceID := uuid.New()

	created, _ := svc.CreateShortLink(ctx, createReq(exampleURL), uuid.New(), workspaceID)

	f := false
	updated, err := svc.UpdateShortLink(ctx, workspaceID, created.ID, &domain.UpdateShortLinkRequest{
		IsActive: &f,
	})
	if err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if updated.IsActive {
		t.Error("expected IsActive = false after update")
	}
}

func TestDeactivateLink(t *testing.T) {
	svc, repo, cache := newService()
	ctx := context.Background()
	workspaceID := uuid.New()

	created, _ := svc.CreateShortLink(ctx, createReq(exampleURL), uuid.New(), workspaceID)
	cacheKey := cacheKeyPrefix + created.ShortCode

	if err := svc.DeactivateLink(ctx, workspaceID, created.ID); err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if repo.links[created.ShortCode].IsActive {
		t.Error("expected link to be inactive after deactivation")
	}
	// Cache should be evicted
	if _, err := cache.Get(ctx, cacheKey); err == nil {
		t.Error("expected cache entry to be deleted after deactivation")
	}
}

func TestDeleteLink(t *testing.T) {
	svc, repo, cache := newService()
	ctx := context.Background()
	workspaceID := uuid.New()

	created, _ := svc.CreateShortLink(ctx, createReq(exampleURL), uuid.New(), workspaceID)
	cacheKey := cacheKeyPrefix + created.ShortCode

	if err := svc.DeleteLink(ctx, workspaceID, created.ID); err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if _, ok := repo.byID[created.ID]; ok {
		t.Error("expected link to be removed from repo after delete")
	}
	if _, err := cache.Get(ctx, cacheKey); err == nil {
		t.Error("expected cache entry to be deleted after delete")
	}
}

func TestGetLinkStats(t *testing.T) {
	svc, _, _ := newService()
	ctx := context.Background()
	workspaceID := uuid.New()

	created, _ := svc.CreateShortLink(ctx, createReq(exampleURL), uuid.New(), workspaceID)

	stats, err := svc.GetLinkStats(ctx, workspaceID, created.ID)
	if err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if stats.LinkID != created.ID {
		t.Errorf("LinkID = %v, want %v", stats.LinkID, created.ID)
	}
}

func TestListLinksInWorkspace(t *testing.T) {
	svc, _, _ := newService()
	ctx := context.Background()
	workspaceID := uuid.New()

	svc.CreateShortLink(ctx, createReq("https://a.com"), uuid.New(), workspaceID)
	svc.CreateShortLink(ctx, createReq("https://b.com"), uuid.New(), workspaceID)
	// Link in a different workspace — should not appear
	svc.CreateShortLink(ctx, createReq("https://c.com"), uuid.New(), uuid.New())

	links, total, err := svc.ListLinksInWorkspace(ctx, workspaceID, ports.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(links) != 2 {
		t.Errorf("len(links) = %d, want 2", len(links))
	}
}

func TestSearchByTag(t *testing.T) {
	svc, _, _ := newService()
	ctx := context.Background()
	workspaceID := uuid.New()

	req := createReq(exampleURL)
	req.Tags = []string{"promo", "summer"}
	svc.CreateShortLink(ctx, req, uuid.New(), workspaceID)

	req2 := createReq("https://other.com")
	req2.Tags = []string{"winter"}
	svc.CreateShortLink(ctx, req2, uuid.New(), workspaceID)

	links, total, err := svc.SearchByTag(ctx, workspaceID, "promo", ports.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf(errUnexpectedErrorFmt, err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(links) != 1 || links[0].Tags[0] != "promo" {
		t.Error("unexpected search result")
	}
}
