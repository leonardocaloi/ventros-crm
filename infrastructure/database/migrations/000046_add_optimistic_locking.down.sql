-- ============================================================================
-- Migration Rollback: 000046_add_optimistic_locking
-- Purpose: Remove version column from all aggregate roots
-- ============================================================================

-- Drop indexes first
DROP INDEX CONCURRENTLY IF EXISTS idx_contacts_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_channels_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_agents_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_pipelines_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_chats_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_projects_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_billing_accounts_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_campaigns_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_sequences_version;

-- Remove version columns
ALTER TABLE contacts DROP COLUMN IF EXISTS version;
ALTER TABLE sessions DROP COLUMN IF EXISTS version;
ALTER TABLE channels DROP COLUMN IF EXISTS version;
ALTER TABLE agents DROP COLUMN IF EXISTS version;
ALTER TABLE pipelines DROP COLUMN IF EXISTS version;
ALTER TABLE chats DROP COLUMN IF EXISTS version;
ALTER TABLE projects DROP COLUMN IF EXISTS version;
ALTER TABLE billing_accounts DROP COLUMN IF EXISTS version;
ALTER TABLE campaigns DROP COLUMN IF EXISTS version;
ALTER TABLE sequences DROP COLUMN IF EXISTS version;
ALTER TABLE broadcasts DROP COLUMN IF EXISTS version;
ALTER TABLE credentials DROP COLUMN IF EXISTS version;
ALTER TABLE contact_lists DROP COLUMN IF EXISTS version;
ALTER TABLE pipeline_statuses DROP COLUMN IF EXISTS version;
ALTER TABLE webhook_subscriptions DROP COLUMN IF EXISTS version;
