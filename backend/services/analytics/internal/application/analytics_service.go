package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/ports"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
)

const (
	errInvalidClickEvent = "invalid click event: missing required fields"
	errRecordClick       = "failed to record click"
	errInvalidLinkID     = "invalid link ID"
	errGetSummary        = "failed to get analytics summary"
	errGetAnalytics      = "failed to get analytics"
	errGetLiveCount      = "failed to get live count"
	errInvalidShortCode  = "invalid short code"
	errGetCountry        = "failed to get country distribution"
	errGetDevice         = "failed to get device distribution"
	errGetTimeRange      = "failed to get clicks by time range"
	errTimeRangeOrder    = "start time must be before end time"
	errWrap              = "%s: %w"
)

// AnalyticsService handles analytics business logic
type AnalyticsService struct {
	clickRepo    ports.ClickRepository
	notifier     ports.ClickNotifier
	publisher    ports.EventPublisher
	locService   ports.LocationService
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(
	clickRepo ports.ClickRepository,
	notifier ports.ClickNotifier,
	publisher ports.EventPublisher,
	locService ports.LocationService,
) *AnalyticsService {
	return &AnalyticsService{
		clickRepo:  clickRepo,
		notifier:   notifier,
		publisher:  publisher,
		locService: locService,
	}
}

// RecordClick processes and persists a click event
func (s *AnalyticsService) RecordClick(ctx context.Context, event *domain.ClickEvent) error {
	if event == nil || !event.IsValid() {
		return fmt.Errorf(errInvalidClickEvent)
	}

	// Enrich click event with geolocation if IP hash available
	if event.IPAddressHash != "" && (event.CountryCode == nil || *event.CountryCode == "") && s.locService != nil {
		// Note: IP hash cannot be reversed, so geolocation would need original IP
		// This is a limitation - we'd need to do geolocation at the redirect endpoint
		// and pass country code directly in the click event
	}

	// Persist to database
	if err := s.clickRepo.RecordClick(ctx, event); err != nil {
		logger.Log.Error(errRecordClick, zap.Error(err))
		return fmt.Errorf(errWrap, errRecordClick, err)
	}

	// Publish event for real-time subscribers
	if s.notifier != nil {
		if err := s.notifier.NotifyClick(ctx, event.LinkID, event); err != nil {
			// Log but don't fail - notification is secondary
			logger.Log.Warn("failed to notify click", zap.Error(err))
		}
	}

	// Publish to event stream only if event originated from direct redirect, not from broker
	// This prevents republishing loop when consumer processes its own published events
	if s.publisher != nil && event.Origin != "kafka" {
		if err := s.publisher.PublishClickEvent(ctx, event); err != nil {
			// Log but don't fail - event streaming is secondary
			logger.Log.Warn("failed to publish click event", zap.Error(err))
		}
	}

	logger.Log.Debug("click recorded", 
		zap.String("link_id", event.LinkID.String()),
		zap.String("short_code", event.ShortCode))

	return nil
}

// GetAnalytics retrieves comprehensive analytics for a link
func (s *AnalyticsService) GetAnalytics(ctx context.Context, linkID uuid.UUID, since time.Time) (*domain.AnalyticsSummary, error) {
	if linkID == uuid.Nil {
		return nil, fmt.Errorf(errInvalidLinkID)
	}

	// Get summary from the specified time period
	summary, err := s.clickRepo.GetSummary(ctx, linkID, since)
	if err != nil {
		logger.Log.Error(errGetSummary, zap.Error(err))
		return nil, fmt.Errorf(errWrap, errGetAnalytics, err)
	}

	return summary, nil
}

// GetLiveCount retrieves the current click count for a short code
func (s *AnalyticsService) GetLiveCount(ctx context.Context, shortCode string) (int64, error) {
	if shortCode == "" {
		return 0, fmt.Errorf(errInvalidShortCode)
	}

	count, err := s.clickRepo.GetLiveCount(ctx, shortCode)
	if err != nil {
		logger.Log.Error(errGetLiveCount, zap.Error(err))
		return 0, fmt.Errorf(errWrap, errGetLiveCount, err)
	}

	return count, nil
}

// GetCountryDistribution retrieves click distribution by country
func (s *AnalyticsService) GetCountryDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error) {
	if linkID == uuid.Nil {
		return nil, fmt.Errorf(errInvalidLinkID)
	}

	distribution, err := s.clickRepo.GetCountryDistribution(ctx, linkID)
	if err != nil {
		logger.Log.Error(errGetCountry, zap.Error(err))
		return nil, fmt.Errorf(errWrap, errGetCountry, err)
	}

	return distribution, nil
}

// GetDeviceDistribution retrieves click distribution by device type
func (s *AnalyticsService) GetDeviceDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error) {
	if linkID == uuid.Nil {
		return nil, fmt.Errorf(errInvalidLinkID)
	}

	distribution, err := s.clickRepo.GetDeviceDistribution(ctx, linkID)
	if err != nil {
		logger.Log.Error(errGetDevice, zap.Error(err))
		return nil, fmt.Errorf(errWrap, errGetDevice, err)
	}

	return distribution, nil
}

// GetClicksByTimeRange retrieves detailed click events within a time range
func (s *AnalyticsService) GetClicksByTimeRange(ctx context.Context, linkID uuid.UUID, start, end time.Time) ([]*domain.ClickEvent, error) {
	if linkID == uuid.Nil {
		return nil, fmt.Errorf(errInvalidLinkID)
	}

	if start.After(end) {
		return nil, fmt.Errorf(errTimeRangeOrder)
	}

	clicks, err := s.clickRepo.GetClicksByTimeRange(ctx, linkID, start, end)
	if err != nil {
		logger.Log.Error(errGetTimeRange, zap.Error(err))
		return nil, fmt.Errorf(errWrap, errGetTimeRange, err)
	}

	return clicks, nil
}
