-- Remove índices
DROP INDEX IF EXISTS idx_outbox_events_category_status_tenant;
DROP INDEX IF EXISTS idx_outbox_events_category;

-- Remove coluna event_category
ALTER TABLE outbox_events
DROP COLUMN IF EXISTS event_category;
