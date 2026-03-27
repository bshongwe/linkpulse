package domain

import (
	"time"

	"github.com/google/uuid"
)

// RedirectType specifies how the short link redirects
type RedirectType string

const (
	RedirectPermanent RedirectType = "301" // 301 Moved Permanently
	RedirectTemporary RedirectType = "302" // 302 Found (temporary redirect)
)

// ShortLink represents a shortened URL in the system
type ShortLink struct {
	ID             uuid.UUID     `json:"id"`
	ShortCode      string        `json:"short_code"`           // 6-10 character unique code
	OriginalURL    string        `json:"original_url"`
	WorkspaceID    uuid.UUID     `json:"workspace_id"`         // Multi-tenant support
	CreatedBy      uuid.UUID     `json:"created_by"`           // User who created the link
	Title          string        `json:"title,omitempty"`
	Description    string        `json:"description,omitempty"`
	ExpiresAt      *time.Time    `json:"expires_at,omitempty"`
	IsActive       bool          `json:"is_active"`
	ClickCount     int64         `json:"click_count"`          // Analytics
	LastAccessedAt *time.Time    `json:"last_accessed_at,omitempty"` // Last click timestamp
	RedirectType   RedirectType  `json:"redirect_type"`        // 301 or 302
	QRCode         string        `json:"qr_code,omitempty"`    // Base64 encoded QR code image
	QRCodeURL      string        `json:"qr_code_url,omitempty"` // URL to QR code image
	Tags           []string      `json:"tags,omitempty"`       // For filtering/organization
	CampaignID     *uuid.UUID    `json:"campaign_id,omitempty"` // Link to marketing campaigns
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`           // For tracking modifications
}

// CreateShortLinkRequest is the payload for creating a new short link
type CreateShortLinkRequest struct {
	OriginalURL string     `json:"original_url" validate:"required,url"`
	CustomAlias string     `json:"custom_alias,omitempty" validate:"omitempty,alphanum,max=50"` // Vanity URL
	Title       string     `json:"title,omitempty" validate:"omitempty,max=200"`
	Description string     `json:"description,omitempty" validate:"omitempty,max=500"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" validate:"omitempty,gtfield=Now"` // Must be in future
	RedirectType RedirectType `json:"redirect_type,omitempty"` // Defaults to 302
	Tags        []string   `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"` // Max 10 tags, each max 50 chars
	CampaignID  *uuid.UUID `json:"campaign_id,omitempty"`     // Optional campaign reference
}

// UpdateShortLinkRequest is the payload for updating a short link
type UpdateShortLinkRequest struct {
	Title        string       `json:"title,omitempty" validate:"omitempty,max=200"`
	Description  string       `json:"description,omitempty" validate:"omitempty,max=500"`
	ExpiresAt    *time.Time   `json:"expires_at,omitempty" validate:"omitempty,gtfield=Now"`
	IsActive     *bool        `json:"is_active,omitempty"` // pointer so omitted != false
	RedirectType RedirectType `json:"redirect_type,omitempty"`
	Tags         []string     `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	CampaignID   *uuid.UUID   `json:"campaign_id,omitempty"`
}

// ShortLinkResponse is the response model for short link endpoints
type ShortLinkResponse struct {
	ID             uuid.UUID    `json:"id"`
	ShortCode      string       `json:"short_code"`
	ShortURL       string       `json:"short_url"`           // Full shortened URL (e.g., https://link.pulse/abc123)
	OriginalURL    string       `json:"original_url"`
	Title          string       `json:"title,omitempty"`
	Description    string       `json:"description,omitempty"`
	ExpiresAt      *time.Time   `json:"expires_at,omitempty"`
	IsActive       bool         `json:"is_active"`
	ClickCount     int64        `json:"click_count"`
	LastAccessedAt *time.Time   `json:"last_accessed_at,omitempty"`
	RedirectType   RedirectType `json:"redirect_type"`
	QRCodeURL      string       `json:"qr_code_url,omitempty"`
	Tags           []string     `json:"tags,omitempty"`
	CampaignID     *uuid.UUID   `json:"campaign_id,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// IsExpired checks if the short link has expired
func (sl *ShortLink) IsExpired() bool {
	if sl.ExpiresAt == nil {
		return false // No expiry set
	}
	return time.Now().After(*sl.ExpiresAt)
}

// CanAccess checks if the link can be accessed (active and not expired)
func (sl *ShortLink) CanAccess() bool {
	return sl.IsActive && !sl.IsExpired()
}
