package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/application"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/ports"
	httphandler "github.com/bshongwe/linkpulse/backend/services/shortener/internal/presentation/http"
	sharedErrors "github.com/bshongwe/linkpulse/backend/shared/errors"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.Init("test")
}

// ---- minimal mock repo & cache (same pattern as service tests) ----

type mockRepo struct {
	mu    sync.Mutex
	links map[string]*domain.ShortLink
	byID  map[uuid.UUID]*domain.ShortLink
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		links: make(map[string]*domain.ShortLink),
		byID:  make(map[uuid.UUID]*domain.ShortLink),
	}
}

func (r *mockRepo) Create(ctx context.Context, link *domain.ShortLink) error {
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
		return nil, sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	return l, nil
}
func (r *mockRepo) FindByID(ctx context.Context, workspaceID, linkID uuid.UUID) (*domain.ShortLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[linkID]
	if !ok {
		return nil, sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	// Enforce workspace isolation: link must belong to requested workspace
	if l.WorkspaceID != workspaceID {
		return nil, sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	return l, nil
}
func (r *mockRepo) FindByCustomAlias(ctx context.Context, alias string) (*domain.ShortLink, error) {
	return r.FindByShortCode(ctx, alias)
}
func (r *mockRepo) IsCodeAvailable(ctx context.Context, code string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, taken := r.links[code]
	return !taken, nil
}
func (r *mockRepo) Update(ctx context.Context, link *domain.ShortLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[link.ID]; !ok {
		return sharedErrors.New(sharedErrors.ErrNotFound, "not found")
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
		return sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	// Enforce workspace isolation
	if l.WorkspaceID != workspaceID {
		return sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	l.IsActive = false
	return nil
}
func (r *mockRepo) Delete(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[linkID]
	if !ok {
		return sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	// Enforce workspace isolation
	if l.WorkspaceID != workspaceID {
		return sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	delete(r.links, l.ShortCode)
	delete(r.byID, linkID)
	return nil
}
func (r *mockRepo) IncrementClickCount(ctx context.Context, linkID uuid.UUID) error { return nil }
func (r *mockRepo) UpdateLastAccess(ctx context.Context, linkID uuid.UUID) error    { return nil }
func (r *mockRepo) GetStats(ctx context.Context, workspaceID, linkID uuid.UUID) (*ports.LinkStats, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[linkID]
	if !ok {
		return nil, sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	// Enforce workspace isolation
	if l.WorkspaceID != workspaceID {
		return nil, sharedErrors.New(sharedErrors.ErrNotFound, "not found")
	}
	return &ports.LinkStats{LinkID: l.ID, ShortCode: l.ShortCode, CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt}, nil
}
func (r *mockRepo) GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID) (*ports.WorkspaceStats, error) {
	return &ports.WorkspaceStats{}, nil
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
	return nil, 0, nil
}
func (r *mockRepo) SearchByTag(ctx context.Context, workspaceID uuid.UUID, tag string, opts ports.ListOptions) ([]*domain.ShortLink, int64, error) {
	return nil, 0, nil
}
func (r *mockRepo) ExpiringLinks(ctx context.Context, workspaceID uuid.UUID, withinHours int) ([]*domain.ShortLink, error) {
	return nil, nil
}
func (r *mockRepo) CountActiveLinks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	return 0, nil
}

type mockCache struct {
	mu    sync.Mutex
	store map[string]interface{}
}

func newMockCache() *mockCache { return &mockCache{store: make(map[string]interface{})} }
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
		return nil, errors.New("miss")
	}
	return v, nil
}
func (c *mockCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
	return nil
}
func (c *mockCache) Exists(ctx context.Context, key string) (bool, error)              { return false, nil }
func (c *mockCache) IncrementInt(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, nil
}
func (c *mockCache) SetWithoutTTL(ctx context.Context, key string, value interface{}) error {
	return c.Set(ctx, key, value, 0)
}

// ---- test router setup ----

func newRouter() (*gin.Engine, *mockRepo) {
	repo := newMockRepo()
	cache := newMockCache()
	svc := application.NewShortenerService(repo, cache)
	handler := httphandler.NewShortenerHandler(svc)
	r := gin.New()
	httphandler.RegisterRoutes(r, handler)
	return r, repo
}

func jsonBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	return bytes.NewBuffer(b)
}

func do(r *gin.Engine, method, path string, body *bytes.Buffer) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	r.ServeHTTP(w, req)
	return w
}

// ---- tests ----

func TestCreateShortLink_201(t *testing.T) {
	r, _ := newRouter()
	body := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": uuid.New().String(),
		"created_by":   uuid.New().String(),
	})
	w := do(r, http.MethodPost, "/api/v1/shorten", body)
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d — body: %s", w.Code, http.StatusCreated, w.Body)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["short_code"] == "" {
		t.Error("expected non-empty short_code in response")
	}
}

func TestCreateShortLink_400_MissingURL(t *testing.T) {
	r, _ := newRouter()
	body := jsonBody(t, map[string]interface{}{
		"workspace_id": uuid.New().String(),
		"created_by":   uuid.New().String(),
	})
	w := do(r, http.MethodPost, "/api/v1/shorten", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateShortLink_400_InvalidWorkspaceID(t *testing.T) {
	r, _ := newRouter()
	body := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": "not-a-uuid",
		"created_by":   uuid.New().String(),
	})
	w := do(r, http.MethodPost, "/api/v1/shorten", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateShortLink_409_DuplicateAlias(t *testing.T) {
	r, _ := newRouter()
	wsID := uuid.New().String()
	body := func() *bytes.Buffer {
		return jsonBody(t, map[string]interface{}{
			"original_url": "https://example.com",
			"workspace_id": wsID,
			"created_by":   uuid.New().String(),
			"custom_alias": "duplicate",
		})
	}
	do(r, http.MethodPost, "/api/v1/shorten", body())
	w := do(r, http.MethodPost, "/api/v1/shorten", body())
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestGetShortLink_200(t *testing.T) {
	r, _ := newRouter()
	wsID := uuid.New().String()
	// Create first
	createBody := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": wsID,
		"created_by":   uuid.New().String(),
	})
	cw := do(r, http.MethodPost, "/api/v1/shorten", createBody)
	var createResp map[string]interface{}
	json.Unmarshal(cw.Body.Bytes(), &createResp)
	shortCode := createResp["data"].(map[string]interface{})["short_code"].(string)

	w := do(r, http.MethodGet, "/api/v1/shorten?short_code="+shortCode, nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — body: %s", w.Code, http.StatusOK, w.Body)
	}
}

func TestGetShortLink_404(t *testing.T) {
	r, _ := newRouter()
	w := do(r, http.MethodGet, "/api/v1/shorten?short_code=doesnotexist", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetShortLink_400_MissingParam(t *testing.T) {
	r, _ := newRouter()
	w := do(r, http.MethodGet, "/api/v1/shorten", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetShortLink_410_Expired(t *testing.T) {
	r, repo := newRouter()
	wsID := uuid.New()
	createBody := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": wsID.String(),
		"created_by":   uuid.New().String(),
	})
	cw := do(r, http.MethodPost, "/api/v1/shorten", createBody)
	var createResp map[string]interface{}
	json.Unmarshal(cw.Body.Bytes(), &createResp)
	shortCode := createResp["data"].(map[string]interface{})["short_code"].(string)

	// Manually expire the link in the mock repo
	past := time.Now().Add(-time.Second)
	repo.links[shortCode].ExpiresAt = &past

	w := do(r, http.MethodGet, "/api/v1/shorten?short_code="+shortCode, nil)
	if w.Code != http.StatusGone {
		t.Errorf("status = %d, want %d", w.Code, http.StatusGone)
	}
}

func TestUpdateShortLink_200(t *testing.T) {
	r, _ := newRouter()
	wsID := uuid.New().String()
	createBody := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": wsID,
		"created_by":   uuid.New().String(),
	})
	cw := do(r, http.MethodPost, "/api/v1/shorten", createBody)
	var createResp map[string]interface{}
	json.Unmarshal(cw.Body.Bytes(), &createResp)
	linkID := createResp["data"].(map[string]interface{})["id"].(string)

	updateBody := jsonBody(t, map[string]interface{}{
		"workspace_id": wsID,
		"title":        "New Title",
	})
	w := do(r, http.MethodPut, "/api/v1/shorten/"+linkID, updateBody)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — body: %s", w.Code, http.StatusOK, w.Body)
	}
}

func TestUpdateShortLink_404(t *testing.T) {
	r, _ := newRouter()
	body := jsonBody(t, map[string]interface{}{
		"workspace_id": uuid.New().String(),
		"title":        "x",
	})
	w := do(r, http.MethodPut, "/api/v1/shorten/"+uuid.New().String(), body)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestDeactivateLink_204(t *testing.T) {
	r, _ := newRouter()
	wsID := uuid.New().String()
	createBody := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": wsID,
		"created_by":   uuid.New().String(),
	})
	cw := do(r, http.MethodPost, "/api/v1/shorten", createBody)
	var createResp map[string]interface{}
	json.Unmarshal(cw.Body.Bytes(), &createResp)
	linkID := createResp["data"].(map[string]interface{})["id"].(string)

	deactivateBody := jsonBody(t, map[string]interface{}{"workspace_id": wsID})
	w := do(r, http.MethodPost, "/api/v1/shorten/"+linkID+"/deactivate", deactivateBody)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d — body: %s", w.Code, http.StatusNoContent, w.Body)
	}
}

func TestDeleteLink_204(t *testing.T) {
	r, _ := newRouter()
	wsID := uuid.New().String()
	createBody := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": wsID,
		"created_by":   uuid.New().String(),
	})
	cw := do(r, http.MethodPost, "/api/v1/shorten", createBody)
	var createResp map[string]interface{}
	json.Unmarshal(cw.Body.Bytes(), &createResp)
	linkID := createResp["data"].(map[string]interface{})["id"].(string)

	deleteBody := jsonBody(t, map[string]interface{}{"workspace_id": wsID})
	w := do(r, http.MethodDelete, "/api/v1/shorten/"+linkID, deleteBody)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d — body: %s", w.Code, http.StatusNoContent, w.Body)
	}
}

func TestDeleteLink_404(t *testing.T) {
	r, _ := newRouter()
	body := jsonBody(t, map[string]interface{}{"workspace_id": uuid.New().String()})
	w := do(r, http.MethodDelete, "/api/v1/shorten/"+uuid.New().String(), body)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetLinkStats_200(t *testing.T) {
	r, _ := newRouter()
	wsID := uuid.New().String()
	createBody := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": wsID,
		"created_by":   uuid.New().String(),
	})
	cw := do(r, http.MethodPost, "/api/v1/shorten", createBody)
	var createResp map[string]interface{}
	json.Unmarshal(cw.Body.Bytes(), &createResp)
	linkID := createResp["data"].(map[string]interface{})["id"].(string)

	w := do(r, http.MethodGet, fmt.Sprintf("/api/v1/shorten/%s/stats?workspace_id=%s", linkID, wsID), nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — body: %s", w.Code, http.StatusOK, w.Body)
	}
}

func TestListLinksInWorkspace_200(t *testing.T) {
	r, _ := newRouter()
	wsID := uuid.New().String()
	for i := 0; i < 3; i++ {
		body := jsonBody(t, map[string]interface{}{
			"original_url": fmt.Sprintf("https://example%d.com", i),
			"workspace_id": wsID,
			"created_by":   uuid.New().String(),
		})
		do(r, http.MethodPost, "/api/v1/shorten", body)
	}

	// Add required pagination query parameters
	w := do(r, http.MethodGet, "/api/v1/shorten/workspace/"+wsID+"?page=1&page_size=10", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — body: %s", w.Code, http.StatusOK, w.Body)
		return // Early return on error to prevent panic on nil decode
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if int(data["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", data["total"])
	}
}

func TestSearchByTag_200(t *testing.T) {
	r, _ := newRouter()
	wsID := uuid.New().String()
	body := jsonBody(t, map[string]interface{}{
		"original_url": "https://example.com",
		"workspace_id": wsID,
		"created_by":   uuid.New().String(),
		"tags":         []string{"sale"},
	})
	do(r, http.MethodPost, "/api/v1/shorten", body)

	w := do(r, http.MethodGet, fmt.Sprintf("/api/v1/shorten/search/tag?tag=sale&workspace_id=%s", wsID), nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — body: %s", w.Code, http.StatusOK, w.Body)
	}
}
