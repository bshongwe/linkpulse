package kafka

import "time"

// Constants for Kafka configuration and error handling
const (
	// Error messages
	errInvalidEvent    = "invalid click event"
	errMarshalFailed   = "failed to marshal click event"
	errPublishFailed   = "failed to publish click event"
	errNoHandlers      = "no event handlers registered"
	errWrap            = "%s: %w"

	// Kafka configuration
	readTimeout = 5 * time.Second
)
