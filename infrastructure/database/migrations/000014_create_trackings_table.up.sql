-- Create trackings table for conversion tracking and ad attribution
CREATE TABLE IF NOT EXISTS trackings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id UUID NOT NULL,
    session_id UUID,
    tenant_id TEXT NOT NULL,
    project_id UUID NOT NULL,

    -- Ad Tracking
    source TEXT NOT NULL,
    platform TEXT NOT NULL,
    campaign TEXT,
    ad_id TEXT,
    ad_url TEXT,

    -- Click & Conversion Tracking
    click_id TEXT UNIQUE, -- CTWA click ID (único)
    conversion_data TEXT,

    -- UTM Parameters
    utm_source TEXT,
    utm_medium TEXT,
    utm_campaign TEXT,
    utm_term TEXT,
    utm_content TEXT,

    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- Foreign Keys
    CONSTRAINT fk_trackings_contact FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    CONSTRAINT fk_trackings_session FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL,
    CONSTRAINT fk_trackings_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_trackings_contact_id ON trackings(contact_id);
CREATE INDEX IF NOT EXISTS idx_trackings_session_id ON trackings(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_trackings_tenant_id ON trackings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_trackings_project_id ON trackings(project_id);
CREATE INDEX IF NOT EXISTS idx_trackings_source ON trackings(source);
CREATE INDEX IF NOT EXISTS idx_trackings_platform ON trackings(platform);
CREATE INDEX IF NOT EXISTS idx_trackings_campaign ON trackings(campaign) WHERE campaign IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_trackings_ad_id ON trackings(ad_id) WHERE ad_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_trackings_click_id ON trackings(click_id) WHERE click_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_trackings_created_at ON trackings(created_at);
CREATE INDEX IF NOT EXISTS idx_trackings_deleted_at ON trackings(deleted_at) WHERE deleted_at IS NULL;

-- Add RLS (Row Level Security) policies
ALTER TABLE trackings ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see trackings from their tenant
CREATE POLICY trackings_tenant_isolation ON trackings
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id', true));

-- Comments
COMMENT ON TABLE trackings IS 'Stores conversion tracking and ad attribution data for contacts';
COMMENT ON COLUMN trackings.source IS 'Tracking source: meta_ads, google_ads, tiktok_ads, linkedin, organic, direct, referral, other';
COMMENT ON COLUMN trackings.platform IS 'Platform: instagram, facebook, google, tiktok, linkedin, whatsapp, other';
COMMENT ON COLUMN trackings.click_id IS 'Click-to-WhatsApp (CTWA) Click ID - unique identifier';
COMMENT ON COLUMN trackings.conversion_data IS 'Encrypted conversion data from platform';
COMMENT ON COLUMN trackings.metadata IS 'Additional platform-specific tracking data';
