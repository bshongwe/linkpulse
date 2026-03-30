package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/application"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
)

func init() {
	// AnalyticsService uses logger.Log — initialise it once for all tests
	logger.Init("test")
}

// --- mock ClickRepository ---

type mockClickRepo struct {
	mu       sync.Mutex
	clicks   map[string]*domain.ClickEvent // keyed by click ID
	summary  *domain.AnalyticsSummary
	failOn   string // method name to force an error
}

func newMockClickRepo() *mockClickRepo {
	return &mockClickRepo{
		clicks: make(map[string]*domain.ClickEvent),
		summary: &domain.AnalyticsSummary{
			TotalClicks:   0,
			ClicksLast24h: 0,
			ClicksLast7d:  0,
			ClicksLast30d: 0,
			TopCountries:  make(map[string]int64),
			TopDevices:    make(map[string]int64),
			TopReferrers:  make(map[string]int64),
			TopUTMSources: make(map[string]int64),
		},
	}
}

func (m *mockClickRepo) RecordClick(ctx context.Context, event *domain.ClickEvent) error {
	if m.failOn == "RecordClick" {
		return errors.New("mock record click error")
	}
	if event == nil {
		return errors.New("event cannot be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clicks[event.ID.String()] = event
	m.summary.TotalClicks++
	return nil
}

func (m *mockClickRepo) GetSummary(ctx context.Context, linkID uuid.UUID, since time.Time) (*domain.AnalyticsSummary, error) {
	if m.failOn == "GetSummary" {
		return nil, errors.New("mock get summary error")
	}
	if linkID == uuid.Nil {
		return nil, errors.New("invalid link ID")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.summary, nil
}

func (m *mockClickRepo) GetLiveCount(ctx context.Context, shortCode string) (int64, error) {
	if m.failOn == "GetLiveCount" {
		return 0, errors.New("mock get live count error")
	}
	if shortCode == "" {
		return 0, errors.New("invalid short code")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.clicks)), nil
}

func (m *mockClickRepo) GetClicksByTimeRange(ctx context.Context, linkID uuid.UUID, start, end time.Time) ([]*domain.ClickEvent, error) {
	if m.failOn == "GetClicksByTimeRange" {
		return nil, errors.New("mock get clicks by time range error")
	}
	if linkID == uuid.Nil {
		return nil, errors.New("invalid link ID")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var events []*domain.ClickEvent
	for _, event := range m.clicks {
		if event.LinkID == linkID && event.Timestamp.After(start) && event.Timestamp.Before(end) {
			events = append(events, event)
		}
	}
	return events, nil
}

func (m *mockClickRepo) GetCountryDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error) {
	if m.failOn == "GetCountryDistribution" {
		return nil, errors.New("mock get country distribution error")
	}
	if linkID == uuid.Nil {
		return nil, errors.New("invalid link ID")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	dist := make(map[string]int64)
	for _, event := range m.clicks {
		if event.LinkID == linkID && event.CountryCode != nil {
			dist[*event.CountryCode]++
		}
	}
	return dist, nil
}

func (m *mockClickRepo) GetDeviceDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error) {
	if m.failOn == "GetDeviceDistribution" {
		return nil, errors.New("mock get device distribution error")
	}
	if linkID == uuid.Nil {
		return nil, errors.New("invalid link ID")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	dist := make(map[string]int64)
	for _, event := range m.clicks {
		if event.LinkID == linkID && event.DeviceType != nil {
			dist[*event.DeviceType]++
		}
	}
	return dist, nil
}

// --- mock ClickNotifier ---

type mockClickNotifier struct {
	mu              sync.Mutex
	notifiedClicks  []*domain.ClickEvent
	failOn          string
}

func newMockClickNotifier() *mockClickNotifier {
	return &mockClickNotifier{
		notifiedClicks: make([]*domain.ClickEvent, 0),
	}
}

func (m *mockClickNotifier) NotifyClick(ctx context.Context, linkID uuid.UUID, event *domain.ClickEvent) error {
	if m.failOn == "NotifyClick" {
		return errors.New("mock notify click error")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifiedClicks = append(m.notifiedClicks, event)
	return nil
}

func (m *mockClickNotifier) Subscribe(linkID uuid.UUID, handler func(*domain.ClickEvent)) (func(), error) {
	if m.failOn == "Subscribe" {
		return nil, errors.New("mock subscribe error")
	}
	if handler == nil {
		return nil, errors.New("handler cannot be nil")
	}
	// Return no-op unsubscribe function for testing
	return func() {}, nil
}

// --- mock EventPublisher ---

type mockEventPublisher struct {
	mu              sync.Mutex
	publishedEvents []*domain.ClickEvent
	failOn          string
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{
		publishedEvents: make([]*domain.ClickEvent, 0),
	}
}

func (m *mockEventPublisher) PublishClickEvent(ctx context.Context, event *domain.ClickEvent) error {
	if m.failOn == "PublishClickEvent" {
		return errors.New("mock publish click event error")
	}
	if event == nil {
		return errors.New("event cannot be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventPublisher) Close() error {
	return nil
}

// --- mock LocationService ---

type mockLocationService struct {
	mu          sync.Mutex
	countryCode string
	failOn      string
}

func newMockLocationService(countryCode string) *mockLocationService {
	return &mockLocationService{
		countryCode: countryCode,
	}
}

func (m *mockLocationService) GetCountryCode(ctx context.Context, ipAddress string) (string, error) {
	if m.failOn == "GetCountryCode" {
		return "", errors.New("mock get country code error")
	}
	if ipAddress == "" {
		return "", errors.New("invalid IP address")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.countryCode, nil
}

func (m *mockLocationService) Close() error {
	return nil
}

// --- Test Cases ---

func TestRecordClickSuccess(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	publisher := newMockEventPublisher()
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()
	linkID := uuid.New()
	event := domain.NewClickEvent(linkID, "short123")

	err := service.RecordClick(ctx, event)
	if err != nil {
		t.Fatalf("RecordClick failed: %v", err)
	}

	// Verify click was persisted
	if len(repo.clicks) != 1 {
		t.Errorf("expected 1 recorded click, got %d", len(repo.clicks))
	}

	// Verify click was notified
	if len(notifier.notifiedClicks) != 1 {
		t.Errorf("expected 1 notified click, got %d", len(notifier.notifiedClicks))
	}

	// Verify click was published
	if len(publisher.publishedEvents) != 1 {
		t.Errorf("expected 1 published event, got %d", len(publisher.publishedEvents))
	}
}

func TestRecordClickWithNilEvent(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	publisher := newMockEventPublisher()
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()
	var event *domain.ClickEvent

	err := service.RecordClick(ctx, event)
	if err == nil {
		t.Error("expected error for nil event, got nil")
	}
}

func TestGetAnalyticsSummary(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	publisher := newMockEventPublisher()
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()
	linkID := uuid.New()
	since := time.Now().AddDate(0, 0, -30)

	summary, err := service.GetAnalytics(ctx, linkID, since)
	if err != nil {
		t.Fatalf("GetAnalytics failed: %v", err)
	}

	if summary == nil {
		t.Error("expected non-nil summary")
	}

	if summary.TotalClicks != 0 {
		t.Errorf("expected 0 total clicks, got %d", summary.TotalClicks)
	}
}

func TestGetLiveCount(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	publisher := newMockEventPublisher()
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()

	count, err := service.GetLiveCount(ctx, "short123")
	if err != nil {
		t.Fatalf("GetLiveCount failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 live count, got %d", count)
	}
}

func TestGetCountryDistribution(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	publisher := newMockEventPublisher()
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()
	linkID := uuid.New()

	dist, err := service.GetCountryDistribution(ctx, linkID)
	if err != nil {
		t.Fatalf("GetCountryDistribution failed: %v", err)
	}

	if len(dist) != 0 {
		t.Errorf("expected empty distribution, got %d countries", len(dist))
	}
}

func TestGetDeviceDistribution(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	publisher := newMockEventPublisher()
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()
	linkID := uuid.New()

	dist, err := service.GetDeviceDistribution(ctx, linkID)
	if err != nil {
		t.Fatalf("GetDeviceDistribution failed: %v", err)
	}

	if len(dist) != 0 {
		t.Errorf("expected empty distribution, got %d devices", len(dist))
	}
}

func TestGetClicksByTimeRange(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	publisher := newMockEventPublisher()
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()
	linkID := uuid.New()
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	events, err := service.GetClicksByTimeRange(ctx, linkID, start, end)
	if err != nil {
		t.Fatalf("GetClicksByTimeRange failed: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestRecordClickNotificationFailureDoesNotFailRequest(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	notifier.failOn = "NotifyClick"
	publisher := newMockEventPublisher()
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()
	linkID := uuid.New()
	event := domain.NewClickEvent(linkID, "short123")

	err := service.RecordClick(ctx, event)
	// Notification failure should not cause RecordClick to fail
	if err != nil {
		t.Fatalf("RecordClick should not fail when NotifyClick fails, got: %v", err)
	}

	// But click should still be recorded
	if len(repo.clicks) != 1 {
		t.Errorf("expected 1 recorded click, got %d", len(repo.clicks))
	}
}

func TestRecordClickPublishFailureDoesNotFailRequest(t *testing.T) {
	repo := newMockClickRepo()
	notifier := newMockClickNotifier()
	publisher := newMockEventPublisher()
	publisher.failOn = "PublishClickEvent"
	location := newMockLocationService("US")
	service := application.NewAnalyticsService(repo, notifier, publisher, location)

	ctx := context.Background()
	linkID := uuid.New()
	event := domain.NewClickEvent(linkID, "short123")

	err := service.RecordClick(ctx, event)
	// Publish failure should not cause RecordClick to fail
	if err != nil {
		t.Fatalf("RecordClick should not fail when PublishClickEvent fails, got: %v", err)
	}

	// But click should still be recorded
	if len(repo.clicks) != 1 {
		t.Errorf("expected 1 recorded click, got %d", len(repo.clicks))
	}
}
