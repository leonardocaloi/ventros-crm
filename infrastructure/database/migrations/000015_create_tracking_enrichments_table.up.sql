-- Create tracking_enrichments table for storing enriched data from Meta Ads API, etc
CREATE TABLE IF NOT EXISTS tracking_enrichments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_id UUID NOT NULL UNIQUE,
    tenant_id TEXT NOT NULL,

    -- Source of enrichment
    source TEXT NOT NULL, -- meta_ads, google_ads, tiktok_ads, etc

    -- Meta Ads API enriched data
    ad_account_id TEXT,
    ad_account_name TEXT,
    campaign_id TEXT,
    campaign_name TEXT,
    adset_id TEXT,
    adset_name TEXT,
    ad_id TEXT,
    ad_name TEXT,
    ad_creative_id TEXT,

    -- Creative information
    creative_type TEXT,   -- image, video, carousel, collection
    creative_format TEXT, -- stories, feed, reels
    creative_body TEXT,
    creative_title TEXT,
    creative_url TEXT,

    -- Targeting & Audience
    targeting_data JSONB DEFAULT '{}'::jsonb,
    audience_name TEXT,

    -- Metrics snapshot at enrichment time
    impressions BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    spend DECIMAL(10,2) DEFAULT 0,
    ctr DECIMAL(5,2) DEFAULT 0,    -- Click-through rate
    cpc DECIMAL(10,2) DEFAULT 0,   -- Cost per click

    -- Raw API data
    raw_api_data JSONB DEFAULT '{}'::jsonb,

    -- Enrichment metadata
    enriched_at TIMESTAMP NOT NULL DEFAULT NOW(),
    enrichment_type TEXT, -- automatic, manual, scheduled
    api_version TEXT,     -- e.g., v18.0

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- Foreign Keys
    CONSTRAINT fk_enrichments_tracking FOREIGN KEY (tracking_id) REFERENCES trackings(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE UNIQUE INDEX IF NOT EXISTS idx_enrichments_tracking_id ON tracking_enrichments(tracking_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_enrichments_tenant_id ON tracking_enrichments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_enrichments_source ON tracking_enrichments(source);
CREATE INDEX IF NOT EXISTS idx_enrichments_enriched_at ON tracking_enrichments(enriched_at);
CREATE INDEX IF NOT EXISTS idx_enrichments_campaign_id ON tracking_enrichments(campaign_id) WHERE campaign_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_enrichments_ad_id ON tracking_enrichments(ad_id) WHERE ad_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_enrichments_deleted_at ON tracking_enrichments(deleted_at) WHERE deleted_at IS NULL;

-- Add RLS (Row Level Security) policies
ALTER TABLE tracking_enrichments ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see enrichments from their tenant
CREATE POLICY enrichments_tenant_isolation ON tracking_enrichments
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id', true));

-- Comments
COMMENT ON TABLE tracking_enrichments IS 'Stores enriched tracking data from advertising platforms APIs (Meta Ads, Google Ads, etc)';
COMMENT ON COLUMN tracking_enrichments.source IS 'Source of enrichment: meta_ads, google_ads, tiktok_ads, etc';
COMMENT ON COLUMN tracking_enrichments.raw_api_data IS 'Complete raw response from platform API';
COMMENT ON COLUMN tracking_enrichments.enrichment_type IS 'How enrichment was triggered: automatic (on tracking creation), manual (user request), scheduled (batch job)';
