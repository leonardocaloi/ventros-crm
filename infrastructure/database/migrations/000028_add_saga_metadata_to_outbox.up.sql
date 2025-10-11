-- Add Saga metadata to outbox_events for correlation tracking
-- This enables Saga Pattern without additional tables (Event Sourcing approach)

-- Add metadata column (JSONB for flexibility)
ALTER TABLE outbox_events
ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}'::jsonb;

-- Create GIN index on metadata for fast queries by correlation_id
CREATE INDEX IF NOT EXISTS idx_outbox_metadata_correlation_id
ON outbox_events USING gin(metadata jsonb_path_ops);

-- Create index specifically for correlation_id queries (faster than GIN for equality)
CREATE INDEX IF NOT EXISTS idx_outbox_correlation_id
ON outbox_events ((metadata->>'correlation_id'));

-- Create index for saga_type queries
CREATE INDEX IF NOT EXISTS idx_outbox_saga_type
ON outbox_events ((metadata->>'saga_type'));

-- Create composite index for common Saga queries (correlation_id + status)
CREATE INDEX IF NOT EXISTS idx_outbox_saga_tracking
ON outbox_events ((metadata->>'correlation_id'), status, created_at);

-- Add comment
COMMENT ON COLUMN outbox_events.metadata IS 'Saga metadata for correlation tracking: correlation_id, saga_type, saga_step, step_number';
