package domain

import (
	"time"

	"github.com/google/uuid"
)

// ClickEvent represents a single click/access to a short link
type ClickEvent struct {
	ID            uuid.UUID `json:"id"`
	LinkID        uuid.UUID `json:"link_id"`
	ShortCode     string    `json:"short_code"`
	Timestamp     time.Time `json:"timestamp"`
	IPAddressHash string    `json:"ip_hash"`        // anonymized
	CountryCode   *string   `json:"country_code,omitempty"`
	DeviceType    *string   `json:"device_type,omitempty"`
	Referrer      *string   `json:"referrer,omitempty"`
	UTMSource     *string   `json:"utm_source,omitempty"`
	UTMMedium     *string   `json:"utm_medium,omitempty"`
	UTMCampaign   *string   `json:"utm_campaign,omitempty"`
}

// AnalyticsSummary represents aggregated analytics for a link
type AnalyticsSummary struct {
	LinkID          uuid.UUID            `json:"link_id"`
	TotalClicks     int64                `json:"total_clicks"`
	ClicksLast24h   int64                `json:"clicks_last_24h"`
	ClicksLast7d    int64                `json:"clicks_last_7d"`
	ClicksLast30d   int64                `json:"clicks_last_30d"`
	TopCountries    map[string]int64     `json:"top_countries"`
	TopDevices      map[string]int64     `json:"top_devices"`
	TopReferrers    map[string]int64     `json:"top_referrers"`
	TopUTMSources   map[string]int64     `json:"top_utm_sources"`
	LastClickTime   *time.Time           `json:"last_click_time,omitempty"`
}

// NewClickEvent creates a new click event with defaults
func NewClickEvent(linkID uuid.UUID, shortCode string) *ClickEvent {
	return &ClickEvent{
		ID:        uuid.New(),
		LinkID:    linkID,
		ShortCode: shortCode,
		Timestamp: time.Now().UTC(),
	}
}

// IsValid checks if the click event has required fields
func (ce *ClickEvent) IsValid() bool {
	return ce.LinkID != uuid.Nil && ce.ShortCode != "" && !ce.Timestamp.IsZero()
}
