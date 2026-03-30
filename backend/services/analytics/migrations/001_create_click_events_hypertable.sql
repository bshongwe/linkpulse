-- Create hypertable for time-series click event data
-- This migration must be run on a PostgreSQL instance with TimescaleDB extension enabled

-- Ensure TimescaleDB extension is installed
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Create the click_events hypertable
-- This table stores all click events from the shortener service
CREATE TABLE IF NOT EXISTS click_events (
    time               TIMESTAMP NOT NULL DEFAULT NOW(),
    id                 UUID NOT NULL,
    link_id            UUID NOT NULL,
    short_code         VARCHAR(255) NOT NULL,
    ip_hash            VARCHAR(64),
    country_code       VARCHAR(2),
    device_type        VARCHAR(50),
    referrer           TEXT,
    utm_source         VARCHAR(255),
    utm_medium         VARCHAR(255),
    utm_campaign       VARCHAR(255),
    created_at         TIMESTAMP DEFAULT NOW()
);

-- Convert to hypertable with time as the time column
-- Chunk interval: 1 week (604800000 milliseconds)
SELECT create_hypertable('click_events', 'time', 
    if_not_exists => TRUE,
    chunk_time_interval => INTERVAL '1 week');

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_click_events_link_id_time 
    ON click_events (link_id, time DESC);

CREATE INDEX IF NOT EXISTS idx_click_events_short_code_time 
    ON click_events (short_code, time DESC);

CREATE INDEX IF NOT EXISTS idx_click_events_country_code 
    ON click_events (country_code) 
    WHERE country_code IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_click_events_device_type 
    ON click_events (device_type) 
    WHERE device_type IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_click_events_utm_source 
    ON click_events (utm_source) 
    WHERE utm_source IS NOT NULL;

-- Enable compression for old data (older than 7 days)
-- This significantly reduces storage footprint
ALTER TABLE click_events SET (
    timescaledb.compress,
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('click_events', INTERVAL '7 days', if_not_exists => TRUE);

-- Enable continuous aggregates for fast analytics queries
-- This will pre-aggregate data at regular intervals
CREATE MATERIALIZED VIEW IF NOT EXISTS click_events_1h AS
SELECT
    time_bucket(INTERVAL '1 hour', time) AS bucket,
    link_id,
    short_code,
    country_code,
    device_type,
    COUNT(*) AS clicks,
    COUNT(DISTINCT ip_hash) AS unique_visitors
FROM click_events
GROUP BY bucket, link_id, short_code, country_code, device_type;

CREATE MATERIALIZED VIEW IF NOT EXISTS click_events_1d AS
SELECT
    time_bucket(INTERVAL '1 day', time) AS bucket,
    link_id,
    short_code,
    COUNT(*) AS clicks,
    COUNT(DISTINCT ip_hash) AS unique_visitors
FROM click_events
GROUP BY bucket, link_id, short_code;

-- Add refresh policies for continuous aggregates (optional but recommended)
-- Uncomment if you want automatic refresh:
-- SELECT add_continuous_aggregate_policy('click_events_1h',
--     start_offset => INTERVAL '2 hours',
--     end_offset => INTERVAL '1 hour',
--     schedule_interval => INTERVAL '30 minutes');

-- Grant permissions (if using separate database users)
-- GRANT SELECT ON click_events TO analytics_user;
-- GRANT SELECT ON click_events_1h TO analytics_user;
-- GRANT SELECT ON click_events_1d TO analytics_user;
