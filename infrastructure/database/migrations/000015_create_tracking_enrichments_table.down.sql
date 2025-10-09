-- Drop RLS policies
DROP POLICY IF EXISTS enrichments_tenant_isolation ON tracking_enrichments;

-- Disable RLS
ALTER TABLE tracking_enrichments DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_enrichments_deleted_at;
DROP INDEX IF EXISTS idx_enrichments_ad_id;
DROP INDEX IF EXISTS idx_enrichments_campaign_id;
DROP INDEX IF EXISTS idx_enrichments_enriched_at;
DROP INDEX IF EXISTS idx_enrichments_source;
DROP INDEX IF EXISTS idx_enrichments_tenant_id;
DROP INDEX IF EXISTS idx_enrichments_tracking_id;

-- Drop table
DROP TABLE IF EXISTS tracking_enrichments;
