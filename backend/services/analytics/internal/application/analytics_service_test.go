package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
)

const (
	errNoError            = "Expected no error, got %v"
	errExpectedOneClick   = "Expected 1 recorded click, got %d"
	errExpectedOneEvent   = "Expected 1 published event, got %d"
	errExpectedNilEvent   = "Expected non-nil summary"
	errExpectedZeroClicks = "Expected 0 total clicks, got %d"
	errExpectedZeroCount  = "Expected 0 live count, got %d"
	errExpectedZeroDist   = "Expected empty distribution, got %d %s"
	errExpectedZeroEvents = "Expected 0 events, got %d"
)

// MockClickRepository mocks the ClickRepository interface
type MockClickRepository struct {
	recordedClicks map[string]*domain.ClickEvent
	summary        *domain.AnalyticsSummary
}

func NewMockClickRepository() *MockClickRepository {
	return &MockClickRepository{
		recordedClicks: make(map[string]*domain.ClickEvent),
		summary: &domain.AnalyticsSummary{
			TotalClicks:      0,
			ClicksLast24h:    0,
			ClicksLast7d:     0,
			ClicksLast30d:    0,
			TopCountries:     make(map[string]int),
			TopDevices:       make(map[string]int),
			TopReferrers:     make(map[string]int),
			TopUTMSources:    make(map[string]int),
		},
	}
}

func (m *MockClickRepository) RecordClick(ctx context.Context, event *domain.ClickEvent) error {
	if event == nil {
		return errRecordFailed
	}
	m.recordedClicks[event.ID.String()] = event
	m.summary.TotalClicks++
	return nil
}

func (m *MockClickRepository) GetSummary(ctx context.Context, linkID uuid.UUID, since time.Time) (*domain.AnalyticsSummary, error) {
	if linkID == uuid.Nil {
		return nil, errGetSummaryFailed
	}
	return m.summary, nil
}

func (m *MockClickRepository) GetLiveCount(ctx context.Context, shortCode string) (int, error) {
	if shortCode == "" {
		return 0, errGetCountFailed
	}
	return len(m.recordedClicks), nil
}

func (m *MockClickRepository) GetClicksByTimeRange(ctx context.Context, linkID uuid.UUID, start, end time.Time) ([]*domain.ClickEvent, error) {
	if linkID == uuid.Nil {
		return nil, errGetDistribution
	}
	var events []*domain.ClickEvent
	for _, event := range m.recordedClicks {
		if event.LinkID == linkID && event.Timestamp.After(start) && event.Timestamp.Before(end) {
			events = append(events, event)
		}
	}
	return events, nil
}

func (m *MockClickRepository) GetCountryDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int, error) {
	if linkID == uuid.Nil {
		return nil, errGetDistribution
	}
	dist := make(map[string]int)
	for _, event := range m.recordedClicks {
		if event.LinkID == linkID {
			dist[event.CountryCode]++
		}
	}
	return dist, nil
}

func (m *MockClickRepository) GetDeviceDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int, error) {
	if linkID == uuid.Nil {
		return nil, errGetDistribution
	}
	dist := make(map[string]int)
	for _, event := range m.recordedClicks {
		if event.LinkID == linkID {
			dist[event.DeviceType]++
		}
	}
	return dist, nil
}

// MockClickNotifier mocks the ClickNotifier interface
type MockClickNotifier struct {
	notifiedClicks []*domain.ClickEvent
}

func NewMockClickNotifier() *MockClickNotifier {
	return &MockClickNotifier{
		notifiedClicks: make([]*domain.ClickEvent, 0),
	}
}

func (m *MockClickNotifier) NotifyClick(ctx context.Context, event *domain.ClickEvent) {
	m.notifiedClicks = append(m.notifiedClicks, event)
}

func (m *MockClickNotifier) Subscribe(linkID uuid.UUID, handler func(*domain.ClickEvent)) (func(), error) {
	if handler == nil {
		return nil, errInvalidNotifier
	}
	// Return no-op unsubscribe function for mock implementation
	return func() {}, nil
}

func (m *MockClickNotifier) Unsubscribe(linkID uuid.UUID) {
	// No-op for mock implementation - test cleanup not needed
}

func (m *MockClickNotifier) UnsubscribeAll(linkID uuid.UUID) {
	// No-op for mock implementation - test cleanup not needed
}

func (m *MockClickNotifier) GetSubscriberCount(linkID uuid.UUID) int {
	return 0
}

// MockEventPublisher mocks the EventPublisher interface
type MockEventPublisher struct {
	publishedEvents []*domain.ClickEvent
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		publishedEvents: make([]*domain.ClickEvent, 0),
	}
}

func (m *MockEventPublisher) PublishClickEvent(ctx context.Context, event *domain.ClickEvent) error {
	if event == nil {
		return errPublishFailed
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *MockEventPublisher) Close() error {
	return nil
}

// MockLocationService mocks the LocationService interface
type MockLocationService struct {
	countryCode string
}

func NewMockLocationService(countryCode string) *MockLocationService {
	return &MockLocationService{
		countryCode: countryCode,
	}
}

func (m *MockLocationService) GetCountryCode(ctx context.Context, ipAddress string) (string, error) {
	if ipAddress == "" {
		return "", errLocationServiceFailed
	}
	return m.countryCode, nil
}

func (m *MockLocationService) Close() error {
	return nil
}

// Test cases

func TestRecordClickSuccess(t *testing.T) {
	ctx := context.Background()
	repo := NewMockClickRepository()
	notifier := NewMockClickNotifier()
	publisher := NewMockEventPublisher()
	location := NewMockLocationService("US")

	service := NewAnalyticsService(repo, notifier, publisher, location)

	linkID := uuid.New()
	event := domain.NewClickEvent(linkID, "short123", "192.168.1.1", "user-agent")

	err := service.RecordClick(ctx, event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(repo.recordedClicks) != 1 {
		t.Errorf(errExpectedOneClick, len(repo.recordedClicks))
	}

	if len(publisher.publishedEvents) != 1 {
		t.Errorf(errExpectedOneEvent, len(publisher.publishedEvents))
	}
}

func TestRecordClickWithNilEvent(t *testing.T) {
	ctx := context.Background()
	repo := NewMockClickRepository()
	notifier := NewMockClickNotifier()
	publisher := NewMockEventPublisher()
	location := NewMockLocationService("US")

	service := NewAnalyticsService(repo, notifier, publisher, location)

	err := service.RecordClick(ctx, nil)
	if err != errInvalidEvent {
		t.Errorf(errNoError, err)
	}
}

func TestGetAnalyticsSummary(t *testing.T) {
	ctx := context.Background()
	repo := NewMockClickRepository()
	notifier := NewMockClickNotifier()
	publisher := NewMockEventPublisher()
	location := NewMockLocationService("US")

	service := NewAnalyticsService(repo, notifier, publisher, location)

	linkID := uuid.New()
	since := time.Now().AddDate(0, 0, -30)

	summary, err := service.GetAnalytics(ctx, linkID, since)
	if err != nil {
		t.Errorf(errNoError, err)
	}

	if summary == nil {
		t.Errorf(errExpectedNilEvent)
	}

	if summary.TotalClicks != 0 {
		t.Errorf(errExpectedZeroClicks, summary.TotalClicks)
	}
}

func TestGetLiveCount(t *testing.T) {
	ctx := context.Background()
	repo := NewMockClickRepository()
	notifier := NewMockClickNotifier()
	publisher := NewMockEventPublisher()
	location := NewMockLocationService("US")

	service := NewAnalyticsService(repo, notifier, publisher, location)

	count, err := service.GetLiveCount(ctx, "short123")
	if err != nil {
		t.Errorf(errNoError, err)
	}

	if count != 0 {
		t.Errorf(errExpectedZeroCount, count)
	}
}

func TestGetCountryDistribution(t *testing.T) {
	ctx := context.Background()
	repo := NewMockClickRepository()
	notifier := NewMockClickNotifier()
	publisher := NewMockEventPublisher()
	location := NewMockLocationService("US")

	service := NewAnalyticsService(repo, notifier, publisher, location)

	linkID := uuid.New()
	dist, err := service.GetCountryDistribution(ctx, linkID)
	if err != nil {
		t.Errorf(errNoError, err)
	}

	if len(dist) != 0 {
		t.Errorf(errExpectedZeroDist, len(dist), "countries")
	}
}

func TestGetDeviceDistribution(t *testing.T) {
	ctx := context.Background()
	repo := NewMockClickRepository()
	notifier := NewMockClickNotifier()
	publisher := NewMockEventPublisher()
	location := NewMockLocationService("US")

	service := NewAnalyticsService(repo, notifier, publisher, location)

	linkID := uuid.New()
	dist, err := service.GetDeviceDistribution(ctx, linkID)
	if err != nil {
		t.Errorf(errNoError, err)
	}

	if len(dist) != 0 {
		t.Errorf(errExpectedZeroDist, len(dist), "devices")
	}
}

func TestGetClicksByTimeRange(t *testing.T) {
	ctx := context.Background()
	repo := NewMockClickRepository()
	notifier := NewMockClickNotifier()
	publisher := NewMockEventPublisher()
	location := NewMockLocationService("US")

	service := NewAnalyticsService(repo, notifier, publisher, location)

	linkID := uuid.New()
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	events, err := service.GetClicksByTimeRange(ctx, linkID, start, end)
	if err != nil {
		t.Errorf(errNoError, err)
	}

	if len(events) != 0 {
		t.Errorf(errExpectedZeroEvents, len(events))
	}
}
