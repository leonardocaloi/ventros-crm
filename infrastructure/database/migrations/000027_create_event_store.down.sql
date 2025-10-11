-- Rollback Event Store tables

DROP INDEX IF EXISTS idx_contact_snapshots_tenant;
DROP INDEX IF EXISTS idx_contact_snapshots_aggregate;
DROP TABLE IF EXISTS contact_snapshots;

DROP INDEX IF EXISTS idx_contact_events_data;
DROP INDEX IF EXISTS idx_contact_events_occurred;
DROP INDEX IF EXISTS idx_contact_events_correlation;
DROP INDEX IF EXISTS idx_contact_events_tenant;
DROP INDEX IF EXISTS idx_contact_events_type;
DROP INDEX IF EXISTS idx_contact_events_aggregate;
DROP TABLE IF EXISTS contact_event_store;
