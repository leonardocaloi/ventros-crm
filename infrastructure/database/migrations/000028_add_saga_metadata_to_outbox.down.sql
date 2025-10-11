-- Rollback: Remove Saga metadata from outbox_events

-- Drop indexes
DROP INDEX IF EXISTS idx_outbox_saga_tracking;
DROP INDEX IF EXISTS idx_outbox_saga_type;
DROP INDEX IF EXISTS idx_outbox_correlation_id;
DROP INDEX IF EXISTS idx_outbox_metadata_correlation_id;

-- Drop column
ALTER TABLE outbox_events
DROP COLUMN IF EXISTS metadata;
