CREATE TABLE IF NOT EXISTS links (
    id              UUID PRIMARY KEY,
    workspace_id    UUID NOT NULL,
    short_code      VARCHAR(64) UNIQUE NOT NULL,
    original_url    TEXT NOT NULL,
    created_by      UUID NOT NULL,
    title           VARCHAR(200),
    description     VARCHAR(500),
    expires_at      TIMESTAMPTZ,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    click_count     BIGINT NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMPTZ,
    redirect_type   VARCHAR(3) NOT NULL DEFAULT '302',
    qr_code         TEXT,
    qr_code_url     TEXT,
    tags            TEXT[] NOT NULL DEFAULT '{}',
    campaign_id     UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_links_workspace_id  ON links(workspace_id);
CREATE INDEX IF NOT EXISTS idx_links_short_code    ON links(short_code);
CREATE INDEX IF NOT EXISTS idx_links_campaign_id   ON links(campaign_id) WHERE campaign_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_links_expires_at    ON links(expires_at)  WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_links_tags          ON links USING GIN(tags);
