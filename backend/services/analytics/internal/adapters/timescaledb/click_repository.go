package timescaledb

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/ports"
)

// ClickRepository implements the ports.ClickRepository interface using TimescaleDB
type ClickRepository struct {
	pool *pgxpool.Pool
}

// NewClickRepository creates a new click repository for TimescaleDB
func NewClickRepository(pool *pgxpool.Pool) ports.ClickRepository {
	return &ClickRepository{pool: pool}
}

const (
	errRecordFailed      = "failed to record click"
	errGetSummaryFailed  = "failed to get summary"
	errGetCountFailed    = "failed to get count"
	errGetDistribution   = "failed to get distribution"
	errScanFailed        = "failed to scan row"
	errWrap              = "%s: %w"
	noRowsResult         = "no rows in result set"
)

// RecordClick persists a single click event to the hypertable
func (r *ClickRepository) RecordClick(ctx context.Context, event *domain.ClickEvent) error {
	id := uuid.New()
	query := `INSERT INTO click_events (id, time, link_id, short_code, ip_hash, country_code, device_type, referrer, utm_source, utm_medium, utm_campaign)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		id,
		event.Timestamp,
		event.LinkID,
		event.ShortCode,
		nullableString(event.IPAddressHash),
		nullableString(event.CountryCode),
		nullableString(event.DeviceType),
		nullableString(event.Referrer),
		nullableString(event.UTMSource),
		nullableString(event.UTMMedium),
		nullableString(event.UTMCampaign),
	)

	if err != nil {
		return fmt.Errorf(errWrap, errRecordFailed, err)
	}

	return nil
}

// getSummaryCounts retrieves all the count metrics for a link
func (r *ClickRepository) getSummaryCounts(ctx context.Context, linkID uuid.UUID, since time.Time) (total, last24h, last7d, last30d int64, err error) {
	// Get total clicks
	totalErr := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM click_events WHERE link_id = $1 AND time >= $2`,
		linkID, since).Scan(&total)
	if totalErr != nil && totalErr.Error() != noRowsResult {
		return 0, 0, 0, 0, fmt.Errorf(errWrap, errGetSummaryFailed, totalErr)
	}

	// Get clicks in last 24h
	last24hErr := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM click_events WHERE link_id = $1 AND time >= NOW() - INTERVAL '24 hours'`,
		linkID).Scan(&last24h)
	if last24hErr != nil && last24hErr.Error() != noRowsResult {
		return 0, 0, 0, 0, fmt.Errorf(errWrap, errGetSummaryFailed, last24hErr)
	}

	// Get clicks in last 7 days
	last7dErr := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM click_events WHERE link_id = $1 AND time >= NOW() - INTERVAL '7 days'`,
		linkID).Scan(&last7d)
	if last7dErr != nil && last7dErr.Error() != noRowsResult {
		return 0, 0, 0, 0, fmt.Errorf(errWrap, errGetSummaryFailed, last7dErr)
	}

	// Get clicks in last 30 days
	last30dErr := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM click_events WHERE link_id = $1 AND time >= NOW() - INTERVAL '30 days'`,
		linkID).Scan(&last30d)
	if last30dErr != nil && last30dErr.Error() != noRowsResult {
		return 0, 0, 0, 0, fmt.Errorf(errWrap, errGetSummaryFailed, last30dErr)
	}

	return total, last24h, last7d, last30d, nil
}

// getSummaryTopMetrics retrieves top countries, devices, referrers, and UTM sources
func (r *ClickRepository) getSummaryTopMetrics(ctx context.Context, linkID uuid.UUID, since time.Time, summary *domain.AnalyticsSummary) error {
	// Get top countries
	if err := r.populateTopValues(ctx, linkID, since, "country_code", summary.TopCountries); err != nil {
		return err
	}

	// Get top devices
	if err := r.populateTopValues(ctx, linkID, since, "device_type", summary.TopDevices); err != nil {
		return err
	}

	// Get top referrers
	if err := r.populateTopValues(ctx, linkID, since, "referrer", summary.TopReferrers); err != nil {
		return err
	}

	// Get top UTM sources
	if err := r.populateTopValues(ctx, linkID, since, "utm_source", summary.TopUTMSources); err != nil {
		return err
	}

	return nil
}

// populateTopValues is a helper that queries and populates a map with top values for a given column
func (r *ClickRepository) populateTopValues(ctx context.Context, linkID uuid.UUID, since time.Time, column string, target map[string]int64) error {
	query := fmt.Sprintf(
		`SELECT %s, COUNT(*) as count FROM click_events 
		 WHERE link_id = $1 AND %s IS NOT NULL AND time >= $2
		 GROUP BY %s ORDER BY count DESC LIMIT 10`,
		column, column, column)

	rows, err := r.pool.Query(ctx, query, linkID, since)
	if err != nil {
		return fmt.Errorf(errWrap, "failed to query top "+column, err)
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var count int64
		if err := rows.Scan(&key, &count); err != nil {
			return fmt.Errorf(errWrap, errScanFailed, err)
		}
		target[key] = count
	}

	return nil
}

// GetSummary retrieves aggregated analytics for a link
func (r *ClickRepository) GetSummary(ctx context.Context, linkID uuid.UUID, since time.Time) (*domain.AnalyticsSummary, error) {
	summary := &domain.AnalyticsSummary{
		LinkID:        linkID,
		TopCountries:  make(map[string]int64),
		TopDevices:    make(map[string]int64),
		TopReferrers:  make(map[string]int64),
		TopUTMSources: make(map[string]int64),
	}

	// Get counts for all time windows
	total, last24h, last7d, last30d, err := r.getSummaryCounts(ctx, linkID, since)
	if err != nil {
		return nil, err
	}

	summary.TotalClicks = total
	summary.ClicksLast24h = last24h
	summary.ClicksLast7d = last7d
	summary.ClicksLast30d = last30d

	// Get top metrics across all dimensions
	if err := r.getSummaryTopMetrics(ctx, linkID, since, summary); err != nil {
		return nil, err
	}

	// Get last click time
	lastErr := r.pool.QueryRow(ctx,
		`SELECT MAX(time) FROM click_events WHERE link_id = $1`,
		linkID).Scan(&summary.LastClickTime)
	if lastErr != nil && lastErr.Error() != noRowsResult {
		return nil, fmt.Errorf(errWrap, errGetSummaryFailed, lastErr)
	}

	return summary, nil
}

func (r *ClickRepository) GetLiveCount(ctx context.Context, shortCode string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM click_events WHERE short_code = $1`,
		shortCode).Scan(&count)

	if err != nil && err.Error() != noRowsResult {
		return 0, fmt.Errorf(errWrap, errGetCountFailed, err)
	}

	return count, nil
}

// GetClicksByTimeRange retrieves click events within a time range
func (r *ClickRepository) GetClicksByTimeRange(ctx context.Context, linkID uuid.UUID, start, end time.Time) ([]*domain.ClickEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, link_id, short_code, time, ip_hash, country_code, device_type, referrer, utm_source, utm_medium, utm_campaign
		 FROM click_events WHERE link_id = $1 AND time BETWEEN $2 AND $3
		 ORDER BY time DESC LIMIT 1000`,
		linkID, start, end)

	if err != nil {
		return nil, fmt.Errorf(errWrap, errGetDistribution, err)
	}
	defer rows.Close()

	var clicks []*domain.ClickEvent
	for rows.Next() {
		event := &domain.ClickEvent{}
		if err := rows.Scan(
			&event.ID, &event.LinkID, &event.ShortCode, &event.Timestamp,
			&event.IPAddressHash, &event.CountryCode, &event.DeviceType,
			&event.Referrer, &event.UTMSource, &event.UTMMedium, &event.UTMCampaign,
		); err != nil {
			return nil, fmt.Errorf(errWrap, errScanFailed, err)
		}
		clicks = append(clicks, event)
	}

	return clicks, nil
}

// GetCountryDistribution retrieves click distribution by country
func (r *ClickRepository) GetCountryDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error) {
	distribution := make(map[string]int64)

	rows, err := r.pool.Query(ctx,
		`SELECT country_code, COUNT(*) as count FROM click_events 
		 WHERE link_id = $1 AND country_code IS NOT NULL
		 GROUP BY country_code ORDER BY count DESC`,
		linkID)

	if err != nil {
		return nil, fmt.Errorf(errWrap, errGetDistribution, err)
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		var count int64
		if err := rows.Scan(&code, &count); err != nil {
			return nil, fmt.Errorf(errWrap, errScanFailed, err)
		}
		distribution[code] = count
	}

	return distribution, nil
}

// GetDeviceDistribution retrieves click distribution by device type
func (r *ClickRepository) GetDeviceDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error) {
	distribution := make(map[string]int64)

	rows, err := r.pool.Query(ctx,
		`SELECT device_type, COUNT(*) as count FROM click_events 
		 WHERE link_id = $1 AND device_type IS NOT NULL
		 GROUP BY device_type ORDER BY count DESC`,
		linkID)

	if err != nil {
		return nil, fmt.Errorf(errWrap, errGetDistribution, err)
	}
	defer rows.Close()

	for rows.Next() {
		var dtype string
		var count int64
		if err := rows.Scan(&dtype, &count); err != nil {
			return nil, fmt.Errorf(errWrap, errScanFailed, err)
		}
		distribution[dtype] = count
	}

	return distribution, nil
}

// nullableString converts empty strings to nil for optional fields
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
