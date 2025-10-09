-- Drop RLS policies
DROP POLICY IF EXISTS trackings_tenant_isolation ON trackings;

-- Disable RLS
ALTER TABLE trackings DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_trackings_deleted_at;
DROP INDEX IF EXISTS idx_trackings_created_at;
DROP INDEX IF EXISTS idx_trackings_click_id;
DROP INDEX IF EXISTS idx_trackings_ad_id;
DROP INDEX IF EXISTS idx_trackings_campaign;
DROP INDEX IF EXISTS idx_trackings_platform;
DROP INDEX IF EXISTS idx_trackings_source;
DROP INDEX IF EXISTS idx_trackings_project_id;
DROP INDEX IF EXISTS idx_trackings_tenant_id;
DROP INDEX IF EXISTS idx_trackings_session_id;
DROP INDEX IF EXISTS idx_trackings_contact_id;

-- Drop table
DROP TABLE IF EXISTS trackings;
