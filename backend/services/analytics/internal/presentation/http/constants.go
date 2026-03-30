package http

const (
	// URL parameters
	linkIDParam = "linkId"
	startParam  = "start"
	endParam    = "end"

	// Response keys
	statusKey = "status"
	countKey  = "count"

	// Log messages
	errMissingParam   = "missing link ID parameter"
	errInvalidUUID    = "invalid link ID format"
	errInvalidTime    = "invalid time format"
	errInternalServer = "internal server error"

	// Error messages for logging
	errGetAnalytics     = "failed to get analytics summary"
	errGetClicks        = "failed to get clicks"
	errGetCountry       = "failed to get country distribution"
	errGetDevice        = "failed to get device distribution"
	errGetLiveCount     = "failed to get live count"
	errLinkIDFormat     = "invalid link ID format"
	errShortCodeFormat  = "invalid short code format"
	errTimeFormat       = "invalid time format"
)
