-- ============================================================================
-- Migration: 000046_add_optimistic_locking
-- Purpose: Add version column to all aggregate roots for optimistic locking
--
-- This prevents lost updates in concurrent scenarios by ensuring that updates
-- fail if the aggregate was modified by another transaction since it was loaded.
--
-- Based on: DDD Optimistic Locking Pattern (Vaughn Vernon, IDDD 2013)
-- Priority: P0 - CRITICAL (prevents data loss from race conditions)
-- ============================================================================

-- Add version column to all aggregate root tables
-- Version starts at 1 for new aggregates and increments on every update

-- Core Aggregates
ALTER TABLE contacts ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE sessions ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE channels ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE agents ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE pipelines ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE chats ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE projects ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE billing_accounts ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;

-- Automation Aggregates
ALTER TABLE campaigns ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE sequences ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE broadcasts ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;

-- Other Aggregates
ALTER TABLE credentials ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE contact_lists ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE pipeline_statuses ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE webhook_subscriptions ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;

-- Create indexes for version checking (improves WHERE id = ? AND version = ? performance)
CREATE INDEX CONCURRENTLY idx_contacts_version ON contacts(id, version);
CREATE INDEX CONCURRENTLY idx_sessions_version ON sessions(id, version);
CREATE INDEX CONCURRENTLY idx_channels_version ON channels(id, version);
CREATE INDEX CONCURRENTLY idx_agents_version ON agents(id, version);
CREATE INDEX CONCURRENTLY idx_pipelines_version ON pipelines(id, version);
CREATE INDEX CONCURRENTLY idx_chats_version ON chats(id, version);
CREATE INDEX CONCURRENTLY idx_projects_version ON projects(id, version);
CREATE INDEX CONCURRENTLY idx_billing_accounts_version ON billing_accounts(id, version);
CREATE INDEX CONCURRENTLY idx_campaigns_version ON campaigns(id, version);
CREATE INDEX CONCURRENTLY idx_sequences_version ON sequences(id, version);

-- Comments
COMMENT ON COLUMN contacts.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN sessions.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN channels.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN agents.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN pipelines.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN chats.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN projects.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN billing_accounts.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN campaigns.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN sequences.version IS 'Optimistic locking version - incremented on every update';
